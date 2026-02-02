package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudflareTunnelIngressSpec defines the desired state of CloudflareTunnelIngress
type CloudflareTunnelIngressSpec struct {
	// Hostname is the public hostname to expose (e.g., grafana.lucena.cloud)
	// +kubebuilder:validation:Required
	Hostname string `json:"hostname"`

	// Service defines the backend Kubernetes service to expose
	// +kubebuilder:validation:Required
	Service ServiceReference `json:"service"`

	// SyncInterval defines how often to check and sync the tunnel configuration
	// +kubebuilder:default="5m"
	// +optional
	SyncInterval string `json:"syncInterval,omitempty"`

	// Enabled determines if this ingress should be active
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// ServiceReference defines the backend service to expose
type ServiceReference struct {
	// Name is the name of the Kubernetes service
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the namespace of the service (defaults to CR namespace)
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Port is the service port to use (defaults to first port)
	// +optional
	Port *int32 `json:"port,omitempty"`

	// Protocol is the protocol to use (http or https)
	// +kubebuilder:validation:Enum=http;https
	// +kubebuilder:default="http"
	// +optional
	Protocol string `json:"protocol,omitempty"`
}

// CloudflareTunnelIngressStatus defines the observed state of CloudflareTunnelIngress
type CloudflareTunnelIngressStatus struct {
	// Phase represents the current phase of the ingress
	// +kubebuilder:validation:Enum=Pending;Syncing;Ready;Failed
	Phase string `json:"phase,omitempty"`

	// LastSyncTime is the last time the tunnel was synced
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// CurrentEndpoint is the current service endpoint configured in the tunnel
	// +optional
	CurrentEndpoint string `json:"currentEndpoint,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the most recent generation observed
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=cfti;cftunnel
// +kubebuilder:printcolumn:name="Hostname",type="string",JSONPath=".spec.hostname",description="Exposed hostname"
// +kubebuilder:printcolumn:name="Service",type="string",JSONPath=".spec.service.name",description="Backend service"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Current phase"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".status.currentEndpoint",description="Current endpoint"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CloudflareTunnelIngress is the Schema for the cloudflaretunnelingresses API
type CloudflareTunnelIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudflareTunnelIngressSpec   `json:"spec,omitempty"`
	Status CloudflareTunnelIngressStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudflareTunnelIngressList contains a list of CloudflareTunnelIngress
type CloudflareTunnelIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflareTunnelIngress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudflareTunnelIngress{}, &CloudflareTunnelIngressList{})
}

// IsEnabled returns whether the ingress is enabled
func (c *CloudflareTunnelIngress) IsEnabled() bool {
	return c.Spec.Enabled == nil || *c.Spec.Enabled
}

// GetServiceNamespace returns the service namespace, defaulting to CR namespace
func (c *CloudflareTunnelIngress) GetServiceNamespace() string {
	if c.Spec.Service.Namespace != "" {
		return c.Spec.Service.Namespace
	}
	return c.Namespace
}

// GetProtocol returns the protocol, defaulting to http
func (c *CloudflareTunnelIngress) GetProtocol() string {
	if c.Spec.Service.Protocol != "" {
		return c.Spec.Service.Protocol
	}
	return "http"
}
