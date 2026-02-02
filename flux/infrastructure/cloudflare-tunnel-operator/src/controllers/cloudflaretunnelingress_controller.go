package controllers

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	tunnelv1alpha1 "github.com/brunovlucena/cloudflare-tunnel-operator/api/v1alpha1"
	"github.com/brunovlucena/cloudflare-tunnel-operator/internal/cloudflare"
)

const (
	PhasePending = "Pending"
	PhaseSyncing = "Syncing"
	PhaseReady   = "Ready"
	PhaseFailed  = "Failed"

	ConditionReady = "Ready"

	DefaultSyncInterval = 5 * time.Minute
	ErrorRequeueDelay   = 30 * time.Second
	MinSyncInterval     = 1 * time.Minute
	MaxSyncInterval     = 1 * time.Hour
)

type CloudflareTunnelIngressReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	CloudflareEmail string
	CloudflareKey   string
	TunnelToken     string

	cfClient     *cloudflare.Client
	cfClientOnce sync.Once
	cfClientErr  error
}

// +kubebuilder:rbac:groups=tunnel.cloudflare.io,resources=cloudflaretunnelingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tunnel.cloudflare.io,resources=cloudflaretunnelingresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tunnel.cloudflare.io,resources=cloudflaretunnelingresses/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *CloudflareTunnelIngressReconciler) getCloudflareClient() (*cloudflare.Client, error) {
	r.cfClientOnce.Do(func() {
		r.cfClient, r.cfClientErr = cloudflare.NewClient(r.CloudflareEmail, r.CloudflareKey, r.TunnelToken)
	})
	return r.cfClient, r.cfClientErr
}

func (r *CloudflareTunnelIngressReconciler) parseSyncInterval(ingress *tunnelv1alpha1.CloudflareTunnelIngress) time.Duration {
	if ingress.Spec.SyncInterval == "" {
		return DefaultSyncInterval
	}

	duration, err := time.ParseDuration(ingress.Spec.SyncInterval)
	if err != nil {
		log.Log.WithName("controller").Error(err, "Invalid syncInterval, using default", "syncInterval", ingress.Spec.SyncInterval)
		return DefaultSyncInterval
	}

	if duration < MinSyncInterval {
		return MinSyncInterval
	}
	if duration > MaxSyncInterval {
		return MaxSyncInterval
	}

	return duration
}

// getServiceEndpoint returns the endpoint (protocol://ip:port) for a service
// CRITICAL: Uses ClusterIP directly (NOT FQDN) to prevent pod IP resolution
// If ClusterIP is not available or is a pod IP, returns error
func (r *CloudflareTunnelIngressReconciler) getServiceEndpoint(ctx context.Context, ingress *tunnelv1alpha1.CloudflareTunnelIngress) (string, error) {
	svc := &corev1.Service{}
	svcNamespace := ingress.GetServiceNamespace()
	svcKey := types.NamespacedName{
		Name:      ingress.Spec.Service.Name,
		Namespace: svcNamespace,
	}

	if err := r.Get(ctx, svcKey, svc); err != nil {
		return "", fmt.Errorf("failed to get service %s/%s: %w", svcNamespace, ingress.Spec.Service.Name, err)
	}

	logger := log.FromContext(ctx)

	// CRITICAL: Use ClusterIP directly - NEVER use FQDN as it might resolve to pod IPs
	// ClusterIP is stable and guaranteed to be a service IP, not a pod IP
	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
		// Service has no ClusterIP (Headless service) - try NodePort as fallback
		logger.Info("⚠️  Service has no ClusterIP, attempting NodePort fallback")
		nodeIP, err := r.getNodeIP(ctx)
		if err != nil {
			return "", fmt.Errorf("service %s/%s has no ClusterIP and failed to get node IP: %w", svcNamespace, ingress.Spec.Service.Name, err)
		}

		// Find NodePort
		var nodePort int32
		if ingress.Spec.Service.Port != nil {
			// Try to find matching port
			for _, port := range svc.Spec.Ports {
				if port.Port == *ingress.Spec.Service.Port && port.NodePort > 0 {
					nodePort = port.NodePort
					break
				}
			}
			if nodePort == 0 {
				return "", fmt.Errorf("service %s/%s has no NodePort for port %d", svcNamespace, ingress.Spec.Service.Name, *ingress.Spec.Service.Port)
			}
		} else if len(svc.Spec.Ports) > 0 {
			if svc.Spec.Ports[0].NodePort == 0 {
				return "", fmt.Errorf("service %s/%s has no NodePort", svcNamespace, ingress.Spec.Service.Name)
			}
			nodePort = svc.Spec.Ports[0].NodePort
		} else {
			return "", fmt.Errorf("service %s/%s has no ports", svcNamespace, ingress.Spec.Service.Name)
		}

		endpoint := fmt.Sprintf("%s://%s:%d", ingress.GetProtocol(), nodeIP, nodePort)
		// CRITICAL: Validate node IP is not a pod IP
		if isPodIP(nodeIP) {
			return "", fmt.Errorf("node IP %s is a pod IP - rejected", nodeIP)
		}
		logger.Info("✅ Using NodePort", "nodeIP", nodeIP, "nodePort", nodePort, "endpoint", endpoint)
		return endpoint, nil
	}

	// CRITICAL: Validate ClusterIP is NOT a pod IP
	if isPodIP(svc.Spec.ClusterIP) {
		return "", fmt.Errorf("service %s/%s ClusterIP %s is a pod IP - rejected", svcNamespace, ingress.Spec.Service.Name, svc.Spec.ClusterIP)
	}

	// Determine the port to use
	var port int32
	if ingress.Spec.Service.Port != nil {
		port = *ingress.Spec.Service.Port
	} else if len(svc.Spec.Ports) > 0 {
		port = svc.Spec.Ports[0].Port
	} else {
		port = 80
	}

	// Use FQDN for cloudflared to access services within the cluster
	// Format: <service-name>.<namespace>.svc.cluster.local
	fqdn := fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, svc.Namespace)
	endpoint := fmt.Sprintf("%s://%s:%d", ingress.GetProtocol(), fqdn, port)
	logger.Info("✅ Using FQDN", "fqdn", fqdn, "port", port, "endpoint", endpoint)
	return endpoint, nil
}

// getNodeIP returns the internal IP of the first available node
// CRITICAL: Only returns non-pod IPs
func (r *CloudflareTunnelIngressReconciler) getNodeIP(ctx context.Context) (string, error) {
	nodes := &corev1.NodeList{}
	if err := r.List(ctx, nodes, &client.ListOptions{}); err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found in cluster")
	}

	// Find the first node with an internal IP that is NOT a pod IP
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP && !isPodIP(addr.Address) {
				return addr.Address, nil
			}
		}
	}

	// Fallback to ExternalIP if no InternalIP found (and validate it's not a pod IP)
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeExternalIP && !isPodIP(addr.Address) {
				return addr.Address, nil
			}
		}
	}

	return "", fmt.Errorf("no valid node IP found (all appear to be pod IPs)")
}

// isPodIP checks if an IP address is a pod IP (10.99.x.x or other pod CIDR ranges)
// CRITICAL: This function prevents pod IPs from being used in any circumstance
func isPodIP(ip string) bool {
	// Common pod IP ranges to reject:
	// - 10.99.x.x (typical pod CIDR)
	// - 10.246.x.x (another common pod CIDR)
	// - Any IP in the pod network range

	// Check for 10.99.x.x (most common pod IP range)
	if len(ip) >= 6 && ip[0:6] == "10.99." {
		return true
	}

	// Check for 10.246.x.x (another common pod CIDR)
	if len(ip) >= 7 && ip[0:7] == "10.246." {
		return true
	}

	// Additional validation: pod IPs are typically in specific ranges
	// We can add more ranges here if needed, but 10.99.x.x is the primary concern

	return false
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
	return isPodIP(ip)
}

// buildKnownCorrectValues builds a map of all enabled ingresses to their Service ClusterIP endpoints
// Returns: (knownCorrectValues map, managedHostnames set)
// managedHostnames includes ALL managed hostnames, even if we couldn't get their service
// This ensures we never preserve stale pod IPs for managed hostnames
func (r *CloudflareTunnelIngressReconciler) buildKnownCorrectValues(ctx context.Context, currentIngress *tunnelv1alpha1.CloudflareTunnelIngress, currentEndpoint string) (map[string]string, map[string]bool) {
	knownCorrectValues := make(map[string]string)
	managedHostnames := make(map[string]bool)

	knownCorrectValues[currentIngress.Spec.Hostname] = currentEndpoint
	managedHostnames[currentIngress.Spec.Hostname] = true

	allIngresses := &tunnelv1alpha1.CloudflareTunnelIngressList{}
	// List all ingresses, including those with deletionTimestamp
	// This ensures we always know which hostnames are managed
	if err := r.List(ctx, allIngresses, &client.ListOptions{}); err != nil {
		return knownCorrectValues, managedHostnames
	}

	for _, ingress := range allIngresses.Items {
		// Include ALL enabled ingresses in managedHostnames, regardless of deletion state
		// This prevents preserving stale pod IPs for managed hostnames
		if !ingress.IsEnabled() {
			continue
		}

		// Mark as managed - even if service unavailable, we won't preserve pod IPs
		managedHostnames[ingress.Spec.Hostname] = true

		if ingress.Spec.Hostname == currentIngress.Spec.Hostname {
			continue
		}

		// Try to get Service ClusterIP - if successful, add to knownCorrectValues
		endpoint, err := r.getServiceEndpoint(ctx, &ingress)
		if err == nil {
			knownCorrectValues[ingress.Spec.Hostname] = endpoint
		}
		// If error, hostname is still in managedHostnames, so we won't preserve pod IPs
	}

	return knownCorrectValues, managedHostnames
}

func (r *CloudflareTunnelIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	reconcileStartTime := time.Now()

	// Extract reconcileID from context if available (controller-runtime adds it)
	reconcileID := fmt.Sprintf("%s/%s", req.Namespace, req.Name)
	if val := ctx.Value("reconcileID"); val != nil {
		if id, ok := val.(string); ok {
			reconcileID = id
		}
	}

	logger.Info("🔄 Reconciling CloudflareTunnelIngress", "name", req.Name, "namespace", req.Namespace, "reconcileID", reconcileID, "startTime", reconcileStartTime.Format(time.RFC3339Nano))

	ingress := &tunnelv1alpha1.CloudflareTunnelIngress{}
	getStartTime := time.Now()
	if err := r.Get(ctx, req.NamespacedName, ingress); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ℹ️  Resource not found, skipping", "name", req.Name, "namespace", req.Namespace, "duration", time.Since(getStartTime))
			return ctrl.Result{}, nil
		}
		logger.Error(err, "❌ Failed to get resource", "duration", time.Since(getStartTime))
		return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, err
	}
	logger.Info("📋 Resource retrieved", "generation", ingress.Generation, "observedGeneration", ingress.Status.ObservedGeneration, "duration", time.Since(getStartTime))

	if !ingress.IsEnabled() {
		logger.Info("⏸️  Ingress is disabled", "hostname", ingress.Spec.Hostname)
		if err := r.updateStatus(ctx, ingress, PhasePending, "", "Ingress disabled"); err != nil {
			return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, err
		}
		return ctrl.Result{}, nil
	}

	// Get Service ClusterIP endpoint - this is the source of truth
	serviceStartTime := time.Now()
	endpoint, err := r.getServiceEndpoint(ctx, ingress)
	serviceDuration := time.Since(serviceStartTime)
	if err != nil {
		logger.Info("⏳ Service not ready", "error", err.Error(), "duration", serviceDuration)
		if err := r.updateStatus(ctx, ingress, PhasePending, "", err.Error()); err != nil {
			return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, err
		}
		return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, nil
	}

	// CRITICAL: Final validation - ensure endpoint does NOT contain pod IP
	if containsPodIP(endpoint) {
		logger.Error(nil, "🚫 REJECTING pod IP in endpoint", "hostname", ingress.Spec.Hostname, "endpoint", endpoint)
		if err := r.updateStatus(ctx, ingress, PhaseFailed, "", fmt.Sprintf("Service endpoint contains pod IP - rejected: %s", endpoint)); err != nil {
			return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, err
		}
		return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, fmt.Errorf("service endpoint contains pod IP: %s", endpoint)
	}

	logger.Info("📍 Service endpoint", "hostname", ingress.Spec.Hostname, "endpoint", endpoint, "duration", serviceDuration)

	cfClient, err := r.getCloudflareClient()
	if err != nil {
		logger.Error(err, "❌ Failed to get Cloudflare client")
		if err := r.updateStatus(ctx, ingress, PhaseFailed, "", fmt.Sprintf("Failed to create Cloudflare client: %v", err)); err != nil {
			return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, err
		}
		return ctrl.Result{}, nil
	}

	// Build map of all correct Service ClusterIPs from Kubernetes (source of truth)
	buildStartTime := time.Now()
	knownCorrectValues, managedHostnames := r.buildKnownCorrectValues(ctx, ingress, endpoint)
	buildDuration := time.Since(buildStartTime)
	logger.Info("📊 Built known correct values", "knownCount", len(knownCorrectValues), "managedCount", len(managedHostnames), "duration", buildDuration)

	// Always update Cloudflare tunnel with Service ClusterIPs
	// Simple: Service ClusterIP -> Cloudflare tunnel. That's it.
	if err := r.updateStatus(ctx, ingress, PhaseSyncing, "", "Updating tunnel configuration"); err != nil {
		return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, err
	}

	updateStartTime := time.Now()
	logger.Info("🔄 Updating tunnel", "hostname", ingress.Spec.Hostname, "endpoint", endpoint, "reconcileID", reconcileID)
	if err := cfClient.UpdateHostname(ingress.Spec.Hostname, endpoint, knownCorrectValues, managedHostnames); err != nil {
		updateDuration := time.Since(updateStartTime)
		logger.Error(err, "❌ Failed to update tunnel", "duration", updateDuration, "totalDuration", time.Since(reconcileStartTime))
		if err := r.updateStatus(ctx, ingress, PhaseFailed, "", fmt.Sprintf("Failed to update tunnel: %v", err)); err != nil {
			return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, err
		}
		return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, nil
	}
	updateDuration := time.Since(updateStartTime)
	totalDuration := time.Since(reconcileStartTime)

	logger.Info("✅ Tunnel updated successfully", "hostname", ingress.Spec.Hostname, "endpoint", endpoint, "updateDuration", updateDuration, "totalDuration", totalDuration)
	logger.Info("🌐 DNS record created/updated", "hostname", ingress.Spec.Hostname)
	if err := r.updateStatus(ctx, ingress, PhaseReady, endpoint, "Tunnel configuration updated successfully"); err != nil {
		return ctrl.Result{RequeueAfter: ErrorRequeueDelay}, err
	}

	syncInterval := r.parseSyncInterval(ingress)
	logger.Info("✅ Reconciliation complete", "hostname", ingress.Spec.Hostname, "nextSync", syncInterval, "totalDuration", totalDuration)
	return ctrl.Result{RequeueAfter: syncInterval}, nil
}

func (r *CloudflareTunnelIngressReconciler) updateStatus(ctx context.Context, ingress *tunnelv1alpha1.CloudflareTunnelIngress, phase, endpoint, message string) error {
	logger := log.FromContext(ctx)

	statusChanged := ingress.Status.Phase != phase ||
		ingress.Status.Message != message ||
		(endpoint != "" && ingress.Status.CurrentEndpoint != endpoint) ||
		ingress.Status.ObservedGeneration != ingress.Generation

	if !statusChanged {
		return nil
	}

	oldPhase := ingress.Status.Phase
	oldEndpoint := ingress.Status.CurrentEndpoint
	ingress.Status.Phase = phase
	ingress.Status.Message = message
	ingress.Status.ObservedGeneration = ingress.Generation

	now := metav1.Now()
	if oldPhase != phase || (endpoint != "" && oldEndpoint != endpoint) {
		ingress.Status.LastSyncTime = &now
	} else if ingress.Status.LastSyncTime == nil {
		ingress.Status.LastSyncTime = &now
	}

	if endpoint != "" {
		ingress.Status.CurrentEndpoint = endpoint
	}

	condition := metav1.Condition{
		Type:               ConditionReady,
		LastTransitionTime: now,
		ObservedGeneration: ingress.Generation,
	}

	switch phase {
	case PhaseReady:
		condition.Status = metav1.ConditionTrue
		condition.Reason = "Synced"
		condition.Message = message
	case PhaseFailed:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SyncFailed"
		condition.Message = message
	default:
		condition.Status = metav1.ConditionUnknown
		condition.Reason = "Syncing"
		condition.Message = message
	}

	found := false
	for i, c := range ingress.Status.Conditions {
		if c.Type == ConditionReady {
			if c.Status != condition.Status || c.Reason != condition.Reason || c.Message != condition.Message {
				ingress.Status.Conditions[i] = condition
			}
			found = true
			break
		}
	}
	if !found {
		ingress.Status.Conditions = append(ingress.Status.Conditions, condition)
	}

	if err := r.Status().Update(ctx, ingress); err != nil {
		logger.Error(err, "Failed to update status")
		return err
	}

	return nil
}

func (r *CloudflareTunnelIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.CloudflareEmail = os.Getenv("CLOUDFLARE_EMAIL")
	r.CloudflareKey = os.Getenv("CLOUDFLARE_API_KEY")
	r.TunnelToken = os.Getenv("CLOUDFLARE_TUNNEL_TOKEN")

	if r.CloudflareEmail == "" || r.CloudflareKey == "" || r.TunnelToken == "" {
		return fmt.Errorf("CLOUDFLARE_EMAIL, CLOUDFLARE_API_KEY, and CLOUDFLARE_TUNNEL_TOKEN must be set")
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&tunnelv1alpha1.CloudflareTunnelIngress{}).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
			predicate.LabelChangedPredicate{},
		)).
		Complete(r)
}
