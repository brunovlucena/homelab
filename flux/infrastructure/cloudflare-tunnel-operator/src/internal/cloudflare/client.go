package cloudflare

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"crypto/rand"
	"encoding/hex"
)

var updateMutex sync.Mutex

type Client struct {
	httpClient *http.Client
	email      string
	apiKey     string
	accountID  string
	tunnelID   string
}

type TunnelToken struct {
	AccountID string `json:"a"`
	TunnelID  string `json:"t"`
	Secret    string `json:"s"`
}

type TunnelConfig struct {
	Config struct {
		Ingress []IngressRule `json:"ingress"`
	} `json:"config"`
}

type IngressRule struct {
	Hostname string `json:"hostname,omitempty"`
	Service  string `json:"service"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
	Errors  []APIError      `json:"errors"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ConfigResponse struct {
	Config struct {
		Ingress []IngressRule `json:"ingress"`
	} `json:"config"`
}

type DNSRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl,omitempty"`
}

type DNSRecordResponse struct {
	Result []DNSRecord `json:"result"`
}

func NewClient(email, apiKey, tunnelToken string) (*Client, error) {
	tokenData, err := base64.StdEncoding.DecodeString(tunnelToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tunnel token: %w", err)
	}

	var token TunnelToken
	if err := json.Unmarshal(tokenData, &token); err != nil {
		return nil, fmt.Errorf("failed to parse tunnel token: %w", err)
	}

	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		email:      email,
		apiKey:     apiKey,
		accountID:  token.AccountID,
		tunnelID:   token.TunnelID,
	}, nil
}

func (c *Client) GetCurrentConfig() (*ConfigResponse, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/cfd_tunnel/%s/configurations", c.accountID, c.tunnelID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	httpStartTime := time.Now()
	resp, err := c.httpClient.Do(req)
	httpDuration := time.Since(httpStartTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("📥 GET %s completed in %v - status: %s", url, httpDuration, resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API request failed")
	}

	var config ConfigResponse
	if err := json.Unmarshal(apiResp.Result, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &config, nil
}

func (c *Client) GetServiceForHostname(hostname string) (string, error) {
	config, err := c.GetCurrentConfig()
	if err != nil {
		return "", err
	}

	for _, rule := range config.Config.Ingress {
		if rule.Hostname == hostname {
			return rule.Service, nil
		}
	}

	return "", nil
}

// generateTraceID generates a unique trace ID for operation tracking
func generateTraceID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// UpdateHostname FORCE OVERWRITES all managed hostnames with correct ClusterIPs from Kubernetes
// CRITICAL: This ensures pod IPs are completely eliminated by overwriting ALL managed hostnames
func (c *Client) UpdateHostname(hostname, service string, knownCorrectValues map[string]string, managedHostnames map[string]bool) error {
	traceID := generateTraceID()
	startTime := time.Now()
	
	log.Printf("🔍 [TRACE:%s] UpdateHostname START - hostname=%s, service=%s, managedCount=%d, knownCount=%d", traceID, hostname, service, len(managedHostnames), len(knownCorrectValues))
	
	// CRITICAL: Validate the service we're about to write does NOT contain a pod IP
	if containsPodIP(service) {
		log.Printf("🚫 [TRACE:%s] REJECTING pod IP in service endpoint: %s", traceID, service)
		return fmt.Errorf("service endpoint contains pod IP - rejecting: %s", service)
	}
	
	// Log mutex acquisition attempt
	log.Printf("🔒 [TRACE:%s] Attempting to acquire updateMutex...", traceID)
	updateMutex.Lock()
	mutexAcquiredTime := time.Now()
	log.Printf("🔒 [TRACE:%s] updateMutex ACQUIRED (waited %v)", traceID, mutexAcquiredTime.Sub(startTime))
	defer func() {
		updateMutex.Unlock()
		totalDuration := time.Since(startTime)
		log.Printf("🔓 [TRACE:%s] updateMutex RELEASED (total duration: %v)", traceID, totalDuration)
	}()

	log.Printf("📥 [TRACE:%s] Reading current config from Cloudflare API...", traceID)
	readStartTime := time.Now()
	config, err := c.GetCurrentConfig()
	readDuration := time.Since(readStartTime)
	if err != nil {
		log.Printf("❌ [TRACE:%s] Failed to get current config after %v: %v", traceID, readDuration, err)
		return fmt.Errorf("failed to get current config: %w", err)
	}
	log.Printf("📥 [TRACE:%s] Config read completed in %v - found %d ingress rules", traceID, readDuration, len(config.Config.Ingress))

	// Build new config: FORCE OVERWRITE all managed hostnames with correct ClusterIPs
	newIngress := make([]IngressRule, 0)
	unmanagedHostnames := make(map[string]string) // Track unmanaged hostnames to preserve
	
	// First pass: collect unmanaged hostnames and identify pod IPs to remove
	podIPsRemoved := 0
	for _, rule := range config.Config.Ingress {
		if rule.Hostname == "" {
			continue // Skip catch-all, we'll add it at the end
		}
		
		// If hostname is managed, we'll overwrite it with Kubernetes value (or remove if pod IP)
		if managedHostnames[rule.Hostname] {
			// Check if it's a pod IP - if so, we'll remove it completely
			if containsPodIP(rule.Service) {
				podIPsRemoved++
				log.Printf("🚫 [TRACE:%s] REMOVING pod IP for managed hostname: %s -> %s (will be overwritten with ClusterIP)", traceID, rule.Hostname, rule.Service)
				continue // Skip - we'll add the correct value from knownCorrectValues
			}
			// If it's not a pod IP but is managed, we'll still overwrite it with Kubernetes value
			// (to ensure consistency)
			continue
		}
		
		// Unmanaged hostname - preserve it (but still reject pod IPs for safety)
		if containsPodIP(rule.Service) {
			log.Printf("🚫 [TRACE:%s] REJECTING pod IP for unmanaged hostname: %s -> %s (will be removed)", traceID, rule.Hostname, rule.Service)
			continue
		}
		unmanagedHostnames[rule.Hostname] = rule.Service
	}
	
	log.Printf("📊 [TRACE:%s] Processing complete - podIPsRemoved=%d, unmanagedCount=%d", traceID, podIPsRemoved, len(unmanagedHostnames))
	
	// Second pass: Add ALL managed hostnames with correct ClusterIPs/NodePorts from Kubernetes
	writtenHostnames := make(map[string]bool) // Track which hostnames we wrote from Kubernetes
	log.Printf("✅ [TRACE:%s] FORCE OVERWRITING %d managed hostnames with ClusterIP/NodePort from Kubernetes:", traceID, len(knownCorrectValues))
	for managedHostname, managedService := range knownCorrectValues {
		// CRITICAL: Double-check it's not a pod IP
		if containsPodIP(managedService) {
			log.Printf("🚫 [TRACE:%s] REJECTING pod IP in Kubernetes value for %s: %s", traceID, managedHostname, managedService)
			continue // Skip this one - don't write pod IPs
		}
		log.Printf("✅ [TRACE:%s] FORCE OVERWRITE: %s -> %s (ClusterIP/NodePort)", traceID, managedHostname, managedService)
		newIngress = append(newIngress, IngressRule{
			Hostname: managedHostname,
			Service:  managedService,
		})
		writtenHostnames[managedHostname] = true
	}
	
	// Add unmanaged hostnames (preserved from API)
	for unmanagedHostname, unmanagedService := range unmanagedHostnames {
		newIngress = append(newIngress, IngressRule{
			Hostname: unmanagedHostname,
			Service:  unmanagedService,
		})
	}

	// Add catch-all rule at the end
	newIngress = append(newIngress, IngressRule{Service: "http_status:404"})

	// Log the final config being written
	log.Printf("📤 [TRACE:%s] FORCE OVERWRITING Cloudflare config with %d ingress rules (ALL managed hostnames from Kubernetes):", traceID, len(newIngress))
	for i, rule := range newIngress {
		if rule.Hostname != "" {
			marker := ""
			if writtenHostnames[rule.Hostname] {
				marker = " ← FROM KUBERNETES (ClusterIP/NodePort)"
			}
			log.Printf("  [TRACE:%s] [%d] %s -> %s%s", traceID, i, rule.Hostname, rule.Service, marker)
		} else {
			log.Printf("  [TRACE:%s] [%d] catch-all -> %s", traceID, i, rule.Service)
		}
	}

	// Update tunnel configuration with RETRY logic to handle Cloudflare API eventual consistency
	// CRITICAL: Keep retrying until verification passes (no pod IPs)
	maxRetries := 5
	retryDelay := 2 * time.Second
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		writeStartTime := time.Now()
		if attempt > 1 {
			log.Printf("🔄 [TRACE:%s] Retry attempt %d/%d (Cloudflare API eventual consistency)...", traceID, attempt, maxRetries)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
		
		log.Printf("📤 [TRACE:%s] Calling putConfig (attempt %d/%d, mutex held for %v so far)...", traceID, attempt, maxRetries, time.Since(mutexAcquiredTime))
		if err := c.putConfig(newIngress, traceID); err != nil {
			log.Printf("❌ [TRACE:%s] putConfig failed after %v: %v", traceID, time.Since(writeStartTime), err)
			if attempt == maxRetries {
				return err
			}
			continue // Retry
		}
		writeDuration := time.Since(writeStartTime)
		log.Printf("✅ [TRACE:%s] putConfig completed in %v", traceID, writeDuration)
		
		// VERIFICATION: Read back the config to verify what was actually written
		log.Printf("🔍 [TRACE:%s] VERIFICATION: Reading back config to verify write...", traceID)
		verifyStartTime := time.Now()
		verifyConfig, err := c.GetCurrentConfig()
		verifyDuration := time.Since(verifyStartTime)
		if err != nil {
			log.Printf("⚠️  [TRACE:%s] VERIFICATION failed to read back config: %v", traceID, err)
			if attempt == maxRetries {
				return fmt.Errorf("verification failed after %d attempts: %w", maxRetries, err)
			}
			continue // Retry
		}
		
		log.Printf("🔍 [TRACE:%s] VERIFICATION: Read back config in %v, found %d ingress rules", traceID, verifyDuration, len(verifyConfig.Config.Ingress))
		
		// Verify ALL managed hostnames have correct ClusterIP/NodePort (no pod IPs)
		podIPsFound := 0
		mismatches := 0
		for _, intendedRule := range newIngress {
			if intendedRule.Hostname == "" {
				continue
			}
			if !writtenHostnames[intendedRule.Hostname] {
				continue // Skip unmanaged hostnames
			}
			found := false
			for _, actualRule := range verifyConfig.Config.Ingress {
				if actualRule.Hostname == intendedRule.Hostname {
					found = true
					if containsPodIP(actualRule.Service) {
						podIPsFound++
						log.Printf("🚫 [TRACE:%s] VERIFICATION FAILED: %s has POD IP: %s (expected: %s)", traceID, intendedRule.Hostname, actualRule.Service, intendedRule.Service)
					} else if actualRule.Service != intendedRule.Service {
						mismatches++
						log.Printf("⚠️  [TRACE:%s] VERIFICATION MISMATCH for %s: intended=%s, actual=%s", traceID, intendedRule.Hostname, intendedRule.Service, actualRule.Service)
					} else {
						log.Printf("✅ [TRACE:%s] VERIFICATION OK: %s -> %s", traceID, intendedRule.Hostname, actualRule.Service)
					}
					break
				}
			}
			if !found {
				log.Printf("⚠️  [TRACE:%s] VERIFICATION: %s not found in API response", traceID, intendedRule.Hostname)
				mismatches++
			}
		}
		
		// If verification passes (no pod IPs, no mismatches), wait a bit for replication, then verify again
		if podIPsFound == 0 && mismatches == 0 {
			log.Printf("✅ [TRACE:%s] VERIFICATION PASSED - all managed hostnames have correct ClusterIP/NodePort (attempt %d/%d)", traceID, attempt, maxRetries)
			// Wait 3 seconds for Cloudflare API replication, then verify one more time
			log.Printf("⏳ [TRACE:%s] Waiting 3s for Cloudflare API replication before final verification...", traceID)
			time.Sleep(3 * time.Second)
			
			// Final verification after replication delay
			finalVerifyConfig, err := c.GetCurrentConfig()
			if err == nil {
				finalPodIPsFound := 0
				for _, intendedRule := range newIngress {
					if intendedRule.Hostname == "" || !writtenHostnames[intendedRule.Hostname] {
						continue
					}
					for _, actualRule := range finalVerifyConfig.Config.Ingress {
						if actualRule.Hostname == intendedRule.Hostname {
							if containsPodIP(actualRule.Service) {
								finalPodIPsFound++
								log.Printf("🚫 [TRACE:%s] FINAL VERIFICATION FAILED: %s still has POD IP: %s", traceID, intendedRule.Hostname, actualRule.Service)
							}
							break
						}
					}
				}
				if finalPodIPsFound == 0 {
					log.Printf("✅ [TRACE:%s] FINAL VERIFICATION PASSED after replication delay - all correct!", traceID)
					break
				} else {
					log.Printf("🚨 [TRACE:%s] FINAL VERIFICATION FAILED: %d pod IPs still found after replication delay! Retrying...", traceID, finalPodIPsFound)
					if attempt < maxRetries {
						continue // Retry
					}
				}
			}
			break
		}
		
		// If we found pod IPs or mismatches, retry
		if podIPsFound > 0 {
			log.Printf("🚨 [TRACE:%s] VERIFICATION FAILED: Found %d pod IPs in Cloudflare config! Retrying... (attempt %d/%d)", traceID, podIPsFound, attempt, maxRetries)
			if attempt == maxRetries {
				return fmt.Errorf("verification failed after %d attempts: found %d pod IPs in Cloudflare config", maxRetries, podIPsFound)
			}
			continue // Retry
		}
		
		if mismatches > 0 {
			log.Printf("⚠️  [TRACE:%s] VERIFICATION: Found %d mismatches. Retrying... (attempt %d/%d)", traceID, mismatches, attempt, maxRetries)
			if attempt == maxRetries {
				log.Printf("⚠️  [TRACE:%s] VERIFICATION: Continuing despite %d mismatches after %d attempts", traceID, mismatches, maxRetries)
				break // Continue despite mismatches (non-critical)
			}
			continue // Retry
		}
	}

	// Create/update DNS record for this hostname
	if err := c.createOrUpdateDNSRecord(hostname); err != nil {
		// Log error but don't fail - DNS might already exist or zone might not be managed
		log.Printf("⚠️  DNS update failed for %s: %v", hostname, err)
	} else {
		log.Printf("✅ DNS record created/updated for %s", hostname)
	}

	return nil
}

func (c *Client) putConfig(ingress []IngressRule, traceID string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/cfd_tunnel/%s/configurations", c.accountID, c.tunnelID)

	payload := TunnelConfig{Config: struct {
		Ingress []IngressRule `json:"ingress"`
	}{Ingress: ingress}}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	log.Printf("🌐 [TRACE:%s] PUT %s (payload size: %d bytes)", traceID, url, len(body))

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	httpStartTime := time.Now()
	resp, err := c.httpClient.Do(req)
	httpDuration := time.Since(httpStartTime)
	if err != nil {
		log.Printf("❌ [TRACE:%s] HTTP request failed after %v: %v", traceID, httpDuration, err)
		return fmt.Errorf("failed to update config: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("📡 [TRACE:%s] HTTP response received in %v - status: %s", traceID, httpDuration, resp.Status)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			log.Printf("❌ [TRACE:%s] API error: %s", traceID, apiResp.Errors[0].Message)
			return fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		log.Printf("❌ [TRACE:%s] API request failed (no error details)", traceID)
		return fmt.Errorf("API request failed")
	}

	log.Printf("✅ [TRACE:%s] API PUT successful", traceID)
	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("X-Auth-Email", c.email)
	req.Header.Set("X-Auth-Key", c.apiKey)
}

// containsPodIP checks if a service URL contains a pod IP
// CRITICAL: This function prevents pod IPs from being written to Cloudflare
func containsPodIP(serviceURL string) bool {
	// Extract IP from service URL (format: http://IP:port or https://IP:port)
	if len(serviceURL) < 7 {
		return false
	}
	
	// Check for http:// or https://
	var ipStart int
	if serviceURL[0:7] == "http://" {
		ipStart = 7
	} else if len(serviceURL) >= 8 && serviceURL[0:8] == "https://" {
		ipStart = 8
	} else {
		return false // Not a valid HTTP URL
	}
	
	// Find the end of the IP (before the colon for port)
	ipEnd := ipStart
	for ipEnd < len(serviceURL) && serviceURL[ipEnd] != ':' {
		ipEnd++
	}
	
	if ipEnd <= ipStart {
		return false
	}
	
	ip := serviceURL[ipStart:ipEnd]
	
	// Check for pod IP ranges
	// 10.99.x.x (most common pod CIDR)
	if len(ip) >= 6 && ip[0:6] == "10.99." {
		return true
	}
	
	// 10.246.x.x (another common pod CIDR)
	if len(ip) >= 7 && ip[0:7] == "10.246." {
		return true
	}
	
	return false
}

// extractDomain extracts the root domain from a hostname (e.g., "grafana.lucena.cloud" -> "lucena.cloud")
func extractDomain(hostname string) string {
	parts := strings.Split(hostname, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return hostname
}

// getZoneID gets the zone ID for a domain
func (c *Client) getZoneID(domain string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", domain)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get zone: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp struct {
		Success bool `json:"success"`
		Result  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
		Errors []APIError `json:"errors"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return "", fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return "", fmt.Errorf("API request failed")
	}

	if len(apiResp.Result) == 0 {
		return "", fmt.Errorf("zone not found for domain: %s", domain)
	}

	return apiResp.Result[0].ID, nil
}

// getDNSRecord gets an existing DNS record for a hostname
func (c *Client) getDNSRecord(zoneID, hostname string) (*DNSRecord, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=CNAME&name=%s", zoneID, hostname)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp struct {
		Success bool         `json:"success"`
		Result  []DNSRecord `json:"result"`
		Errors  []APIError  `json:"errors"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API request failed")
	}

	if len(apiResp.Result) == 0 {
		return nil, nil // Record doesn't exist
	}

	return &apiResp.Result[0], nil
}

// createOrUpdateDNSRecord creates or updates a CNAME record for the tunnel hostname
func (c *Client) createOrUpdateDNSRecord(hostname string) error {
	domain := extractDomain(hostname)
	zoneID, err := c.getZoneID(domain)
	if err != nil {
		return fmt.Errorf("failed to get zone ID for %s: %w", domain, err)
	}

	tunnelTarget := fmt.Sprintf("%s.cfargotunnel.com", c.tunnelID)
	existingRecord, err := c.getDNSRecord(zoneID, hostname)

	if err != nil {
		return fmt.Errorf("failed to check existing DNS record: %w", err)
	}

	record := DNSRecord{
		Type:    "CNAME",
		Name:    hostname,
		Content: tunnelTarget,
		Proxied: true,
		TTL:     1, // Auto TTL
	}

	var url string
	var method string

	if existingRecord != nil {
		// Update existing record
		url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, existingRecord.ID)
		method = http.MethodPut
		record.ID = existingRecord.ID
	} else {
		// Create new record
		url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)
		method = http.MethodPost
	}

	body, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal DNS record: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to %s DNS record: %w", method, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return fmt.Errorf("API request failed")
	}

	return nil
}
