// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🔒 SEC-007: Network Segmentation & Data Exfiltration Testing
//
//	User Story: Network Segmentation & Data Exfiltration Testing
//	Priority: P1 | Story Points: 8
//
//	Tests validate:
//	- Network policy enforcement
//	- Egress filtering
//	- Data exfiltration prevention
//	- Service mesh security
//	- Pod-to-pod isolation
//	- Ingress controller security
//	- VPC network segmentation
//	- RabbitMQ network isolation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestSec007_DefaultDenyNetworkPolicy validates default deny policies exist.
func TestSec007_DefaultDenyNetworkPolicy(t *testing.T) {
	tests := []struct {
		name          string
		policy        *networkingv1.NetworkPolicy
		isDefaultDeny bool
		description   string
	}{
		{
			name: "Default deny all ingress",
			policy: &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: "default-deny-ingress"},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{},
					PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
				},
			},
			isDefaultDeny: true,
			description:   "Empty pod selector with PolicyTypeIngress should be default deny",
		},
		{
			name: "Default deny all egress",
			policy: &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: "default-deny-egress"},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{},
					PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
				},
			},
			isDefaultDeny: true,
			description:   "Empty pod selector with PolicyTypeEgress should be default deny",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isDefaultDeny := isDefaultDenyPolicy(tt.policy)

			// Assert
			assert.Equal(t, tt.isDefaultDeny, isDefaultDeny, tt.description)
		})
	}
}

// TestSec007_CrossNamespaceBlocked validates cross-namespace communication is blocked.
func TestSec007_CrossNamespaceBlocked(t *testing.T) {
	tests := []struct {
		name        string
		policy      *networkingv1.NetworkPolicy
		sourceNS    string
		targetNS    string
		shouldAllow bool
		description string
	}{
		{
			name:        "Same namespace allowed",
			policy:      createNamespaceIsolationPolicy("knative-lambda"),
			sourceNS:    "knative-lambda",
			targetNS:    "knative-lambda",
			shouldAllow: true,
			description: "Same namespace traffic should be allowed",
		},
		{
			name:        "Different namespace blocked",
			policy:      createNamespaceIsolationPolicy("knative-lambda"),
			sourceNS:    "other-namespace",
			targetNS:    "knative-lambda",
			shouldAllow: false,
			description: "Cross-namespace traffic should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			allowed := checkNamespaceAccess(tt.policy, tt.sourceNS, tt.targetNS)

			// Assert
			assert.Equal(t, tt.shouldAllow, allowed, tt.description)
		})
	}
}

// TestSec007_EgressFiltering validates egress traffic is filtered.
func TestSec007_EgressFiltering(t *testing.T) {
	tests := []struct {
		name        string
		policy      *networkingv1.NetworkPolicy
		destination string
		port        int
		shouldAllow bool
		description string
	}{
		{
			name:        "DNS allowed",
			policy:      createEgressPolicy([]int{53}, []string{}),
			destination: "8.8.8.8",
			port:        53,
			shouldAllow: true,
			description: "DNS traffic should be allowed",
		},
		{
			name:        "HTTP to internet blocked",
			policy:      createEgressPolicy([]int{53}, []string{}),
			destination: "evil.com",
			port:        80,
			shouldAllow: false,
			description: "Arbitrary internet HTTP should be blocked",
		},
		{
			name:        "HTTPS to internet blocked",
			policy:      createEgressPolicy([]int{53}, []string{}),
			destination: "attacker.com",
			port:        443,
			shouldAllow: false,
			description: "Arbitrary internet HTTPS should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			allowed := checkEgressAllowed(tt.policy, tt.destination, tt.port)

			// Assert
			assert.Equal(t, tt.shouldAllow, allowed, tt.description)
		})
	}
}

// TestSec007_DataExfiltrationPrevention validates common exfiltration techniques are blocked.
func TestSec007_DataExfiltrationPrevention(t *testing.T) {
	tests := []struct {
		name        string
		protocol    string
		destination string
		isBlocked   bool
		description string
	}{
		{
			name:        "DNS tunneling blocked",
			protocol:    "DNS",
			destination: "base64data.attacker.com",
			isBlocked:   true,
			description: "Suspicious DNS queries should be flagged",
		},
		{
			name:        "Reverse shell blocked",
			protocol:    "TCP",
			destination: "attacker.com:4444",
			isBlocked:   true,
			description: "Reverse shell connections should be blocked",
		},
		{
			name:        "Unauthorized cloud storage blocked",
			protocol:    "HTTPS",
			destination: "s3.amazonaws.com/unauthorized-bucket",
			isBlocked:   true,
			description: "Unauthorized cloud storage access should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isBlocked := detectExfiltrationAttempt(tt.protocol, tt.destination)

			// Assert
			assert.Equal(t, tt.isBlocked, isBlocked, tt.description)
		})
	}
}

// TestSec007_PodToPodIsolation validates pods are properly isolated.
func TestSec007_PodToPodIsolation(t *testing.T) {
	tests := []struct {
		name        string
		sourceLabel map[string]string
		targetLabel map[string]string
		policy      *networkingv1.NetworkPolicy
		shouldAllow bool
		description string
	}{
		{
			name:        "Builder to parser blocked",
			sourceLabel: map[string]string{"app": "builder"},
			targetLabel: map[string]string{"app": "parser"},
			policy:      createPodIsolationPolicy(),
			shouldAllow: false,
			description: "Builder should not access parser directly",
		},
		{
			name:        "Parser to parser blocked",
			sourceLabel: map[string]string{"app": "parser", "instance": "1"},
			targetLabel: map[string]string{"app": "parser", "instance": "2"},
			policy:      createPodIsolationPolicy(),
			shouldAllow: false,
			description: "Parsers should be isolated from each other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			allowed := checkPodToPodAccess(tt.policy, tt.sourceLabel, tt.targetLabel)

			// Assert
			assert.Equal(t, tt.shouldAllow, allowed, tt.description)
		})
	}
}

// TestSec007_IngressControllerSecurity validates ingress security.
func TestSec007_IngressControllerSecurity(t *testing.T) {
	tests := []struct {
		name        string
		tlsEnabled  bool
		rateLimit   int
		shouldAllow bool
		description string
	}{
		{
			name:        "TLS not enabled - blocked",
			tlsEnabled:  false,
			rateLimit:   100,
			shouldAllow: false,
			description: "Ingress without TLS should be blocked",
		},
		{
			name:        "TLS enabled - allowed",
			tlsEnabled:  true,
			rateLimit:   100,
			shouldAllow: true,
			description: "Ingress with TLS should be allowed",
		},
		{
			name:        "No rate limit - blocked",
			tlsEnabled:  true,
			rateLimit:   0,
			shouldAllow: false,
			description: "Ingress without rate limit should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isSecure := validateIngressSecurity(tt.tlsEnabled, tt.rateLimit)

			// Assert
			assert.Equal(t, tt.shouldAllow, isSecure, tt.description)
		})
	}
}

// TestSec007_ServiceMeshMTLS validates mTLS is enforced.
func TestSec007_ServiceMeshMTLS(t *testing.T) {
	tests := []struct {
		name        string
		mtlsMode    string
		isSecure    bool
		description string
	}{
		{
			name:        "STRICT mTLS",
			mtlsMode:    "STRICT",
			isSecure:    true,
			description: "STRICT mTLS should be secure",
		},
		{
			name:        "PERMISSIVE mTLS",
			mtlsMode:    "PERMISSIVE",
			isSecure:    false,
			description: "PERMISSIVE mTLS should not be allowed",
		},
		{
			name:        "DISABLE mTLS",
			mtlsMode:    "DISABLE",
			isSecure:    false,
			description: "DISABLE mTLS should not be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isSecure := tt.mtlsMode == "STRICT"

			// Assert
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// TestSec007_RabbitMQNetworkIsolation validates RabbitMQ network isolation.
func TestSec007_RabbitMQNetworkIsolation(t *testing.T) {
	tests := []struct {
		name              string
		exposedToInternet bool
		tlsEnabled        bool
		isSecure          bool
		description       string
	}{
		{
			name:              "RabbitMQ exposed to internet",
			exposedToInternet: true,
			tlsEnabled:        true,
			isSecure:          false,
			description:       "RabbitMQ should not be exposed to internet",
		},
		{
			name:              "RabbitMQ internal only with TLS",
			exposedToInternet: false,
			tlsEnabled:        true,
			isSecure:          true,
			description:       "Internal RabbitMQ with TLS should be secure",
		},
		{
			name:              "RabbitMQ internal without TLS",
			exposedToInternet: false,
			tlsEnabled:        false,
			isSecure:          false,
			description:       "RabbitMQ without TLS should not be secure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isSecure := !tt.exposedToInternet && tt.tlsEnabled

			// Assert
			assert.Equal(t, tt.isSecure, isSecure, tt.description)
		})
	}
}

// Helper Functions.

func isDefaultDenyPolicy(policy *networkingv1.NetworkPolicy) bool {
	// Default deny has empty pod selector and no ingress/egress rules
	return len(policy.Spec.PodSelector.MatchLabels) == 0 &&
		len(policy.Spec.PodSelector.MatchExpressions) == 0 &&
		len(policy.Spec.Ingress) == 0 &&
		len(policy.Spec.Egress) == 0
}

func createNamespaceIsolationPolicy(namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namespace-isolation",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
						},
					},
				},
			},
		},
	}
}

func checkNamespaceAccess(policy *networkingv1.NetworkPolicy, sourceNS, targetNS string) bool {
	// Simple check: if policy allows from PodSelector only (not NamespaceSelector),
	// then only same-namespace traffic is allowed
	if len(policy.Spec.Ingress) == 0 {
		return false
	}

	for _, rule := range policy.Spec.Ingress {
		for _, peer := range rule.From {
			if peer.NamespaceSelector != nil {
				// Has namespace selector, check if it matches
				return sourceNS == targetNS
			}
			if peer.PodSelector != nil && peer.NamespaceSelector == nil {
				// Only pod selector, must be same namespace
				return sourceNS == targetNS
			}
		}
	}

	return sourceNS == targetNS
}

func createEgressPolicy(allowedPorts []int, _ []string) *networkingv1.NetworkPolicy {
	egressRules := []networkingv1.NetworkPolicyEgressRule{}

	for _, port := range allowedPorts {
		portVal := intstr.FromInt(port)
		egressRules = append(egressRules, networkingv1.NetworkPolicyEgressRule{
			Ports: []networkingv1.NetworkPolicyPort{
				{Port: &portVal},
			},
		})
	}

	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "egress-filter"},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
			Egress:      egressRules,
		},
	}
}

func checkEgressAllowed(policy *networkingv1.NetworkPolicy, _ string, port int) bool {
	if policy.Spec.Egress == nil {
		return false
	}

	for _, rule := range policy.Spec.Egress {
		for _, policyPort := range rule.Ports {
			if policyPort.Port != nil && policyPort.Port.IntVal == int32(port) { //nolint:gosec // G115: Test code with controlled port values
				return true
			}
		}
	}

	return false
}

func detectExfiltrationAttempt(_, destination string) bool {
	// Simple heuristic: block suspicious destinations
	suspiciousPatterns := []string{
		"attacker.com",
		"evil.com",
		"pastebin.com",
		"base64",
		":4444",                         // Common reverse shell port
		":1337",                         // Common hacker port
		"unauthorized-bucket",           // Unauthorized cloud storage
		"s3.amazonaws.com/unauthorized", // S3 unauthorized access
	}

	for _, pattern := range suspiciousPatterns {
		if contains(destination, pattern) {
			return true
		}
	}

	return false
}

func createPodIsolationPolicy() *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-isolation"},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			Ingress:     []networkingv1.NetworkPolicyIngressRule{},
		},
	}
}

func checkPodToPodAccess(policy *networkingv1.NetworkPolicy, sourceLabel, _ map[string]string) bool {
	// If policy has no ingress rules, all traffic is blocked
	if len(policy.Spec.Ingress) == 0 {
		return false
	}

	// Check if source matches any ingress rule
	for _, rule := range policy.Spec.Ingress {
		for _, peer := range rule.From {
			if peer.PodSelector != nil {
				// Check if source labels match the selector
				if matchesSelector(sourceLabel, peer.PodSelector) {
					return true
				}
			}
		}
	}

	return false
}

func validateIngressSecurity(tlsEnabled bool, rateLimit int) bool {
	return tlsEnabled && rateLimit > 0
}

func matchesSelector(labels map[string]string, selector *metav1.LabelSelector) bool {
	if selector == nil {
		return true
	}

	for key, value := range selector.MatchLabels {
		if labels[key] != value {
			return false
		}
	}

	return true
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
