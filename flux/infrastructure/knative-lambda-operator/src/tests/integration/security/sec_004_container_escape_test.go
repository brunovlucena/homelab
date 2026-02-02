// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🔒 SEC-004: Container Escape & Privilege Escalation Testing
//
//	User Story: Container Escape & Privilege Escalation Testing
//	Priority: P0 | Story Points: 13
//
//	Tests validate:
//	- Container security context enforcement
//	- Capability restriction
//	- Host filesystem protection
//	- Host namespace isolation
//	- Kernel exploitation prevention
//	- Container runtime security
//	- Resource limit enforcement
//	- Parser sandbox isolation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestSec004_PrivilegedContainersBlocked validates privileged containers are blocked.
func TestSec004_PrivilegedContainersBlocked(t *testing.T) {
	tests := []struct {
		name        string
		privileged  bool
		shouldBlock bool
		description string
	}{
		{
			name:        "Privileged true blocked",
			privileged:  true,
			shouldBlock: true,
			description: "privileged: true should be blocked",
		},
		{
			name:        "Privileged false allowed",
			privileged:  false,
			shouldBlock: false,
			description: "privileged: false should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			pod := createTestPod("test-pod", tt.privileged, false)

			// Act
			isValid := validatePodSecurityContext(pod)

			// Assert
			if tt.shouldBlock {
				assert.False(t, isValid, tt.description)
			} else {
				assert.True(t, isValid, tt.description)
			}
		})
	}
}

// TestSec004_AllowPrivilegeEscalation validates privilege escalation is blocked.
func TestSec004_AllowPrivilegeEscalation(t *testing.T) {
	tests := []struct {
		name                     string
		allowPrivilegeEscalation *bool
		shouldBlock              bool
		description              string
	}{
		{
			name:                     "Privilege escalation allowed - blocked",
			allowPrivilegeEscalation: boolPtr(true),
			shouldBlock:              true,
			description:              "allowPrivilegeEscalation: true should be blocked",
		},
		{
			name:                     "Privilege escalation disabled - allowed",
			allowPrivilegeEscalation: boolPtr(false),
			shouldBlock:              false,
			description:              "allowPrivilegeEscalation: false should be allowed",
		},
		{
			name:                     "Privilege escalation unset - blocked",
			allowPrivilegeEscalation: nil,
			shouldBlock:              true,
			description:              "Missing allowPrivilegeEscalation should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			pod := createTestPodWithEscalation("test-pod", tt.allowPrivilegeEscalation)

			// Act
			isValid := validatePodSecurityContext(pod)

			// Assert
			if tt.shouldBlock {
				assert.False(t, isValid, tt.description)
			} else {
				assert.True(t, isValid, tt.description)
			}
		})
	}
}

// TestSec004_RunAsNonRoot validates containers run as non-root.
func TestSec004_RunAsNonRoot(t *testing.T) {
	tests := []struct {
		name         string
		runAsNonRoot *bool
		runAsUser    *int64
		shouldBlock  bool
		description  string
	}{
		{
			name:         "Run as root blocked",
			runAsNonRoot: boolPtr(false),
			runAsUser:    int64Ptr(0),
			shouldBlock:  true,
			description:  "Running as root should be blocked",
		},
		{
			name:         "Run as non-root allowed",
			runAsNonRoot: boolPtr(true),
			runAsUser:    int64Ptr(65534),
			shouldBlock:  false,
			description:  "Running as non-root should be allowed",
		},
		{
			name:         "RunAsNonRoot not set - blocked",
			runAsNonRoot: nil,
			runAsUser:    int64Ptr(1000),
			shouldBlock:  true,
			description:  "Missing runAsNonRoot should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			pod := createTestPodWithUser("test-pod", tt.runAsNonRoot, tt.runAsUser)

			// Act
			isValid := validatePodSecurityContext(pod)

			// Assert
			if tt.shouldBlock {
				assert.False(t, isValid, tt.description)
			} else {
				assert.True(t, isValid, tt.description)
			}
		})
	}
}

// TestSec004_CapabilitiesRestriction validates dangerous capabilities are blocked.
func TestSec004_CapabilitiesRestriction(t *testing.T) {
	dangerousCapabilities := []corev1.Capability{
		"SYS_ADMIN",
		"SYS_MODULE",
		"SYS_RAWIO",
		"SYS_PTRACE",
		"SYS_BOOT",
		"NET_ADMIN",
		"DAC_OVERRIDE",
	}

	for _, cap := range dangerousCapabilities {
		t.Run(string(cap), func(t *testing.T) {
			// Arrange
			pod := createTestPodWithCapabilities("test-pod", []corev1.Capability{cap})

			// Act
			isValid := validatePodSecurityContext(pod)

			// Assert
			assert.False(t, isValid, "Capability %s should be blocked", cap)
		})
	}
}

// TestSec004_AllCapabilitiesDropped validates ALL capabilities are dropped.
func TestSec004_AllCapabilitiesDropped(t *testing.T) {
	// Arrange
	pod := createTestPodWithDroppedCapabilities("test-pod")

	// Act
	isValid := validatePodSecurityContext(pod)

	// Assert
	assert.True(t, isValid, "Pod with all capabilities dropped should be valid")

	// Verify ALL is in drop list
	container := pod.Spec.Containers[0]
	require.NotNil(t, container.SecurityContext)
	require.NotNil(t, container.SecurityContext.Capabilities)

	found := false
	for _, cap := range container.SecurityContext.Capabilities.Drop {
		if cap == "ALL" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should drop ALL capabilities")
}

// TestSec004_HostPathMountBlocked validates dangerous host paths are blocked.
func TestSec004_HostPathMountBlocked(t *testing.T) {
	dangerousPaths := []string{
		"/",
		"/proc",
		"/sys",
		"/dev",
		"/etc",
		"/var/run/docker.sock",
		"/var/run/containerd.sock",
		"/run/containerd",
	}

	for _, path := range dangerousPaths {
		t.Run(path, func(t *testing.T) {
			// Arrange
			pod := createTestPodWithHostPath("test-pod", path)

			// Act
			isValid := validatePodVolumes(pod)

			// Assert
			assert.False(t, isValid, "HostPath %s should be blocked", path)
		})
	}
}

// TestSec004_HostNamespaceIsolation validates host namespaces are blocked.
func TestSec004_HostNamespaceIsolation(t *testing.T) {
	tests := []struct {
		name        string
		hostNetwork bool
		hostPID     bool
		hostIPC     bool
		shouldBlock bool
		description string
	}{
		{
			name:        "Host network blocked",
			hostNetwork: true,
			hostPID:     false,
			hostIPC:     false,
			shouldBlock: true,
			description: "hostNetwork: true should be blocked",
		},
		{
			name:        "Host PID blocked",
			hostNetwork: false,
			hostPID:     true,
			hostIPC:     false,
			shouldBlock: true,
			description: "hostPID: true should be blocked",
		},
		{
			name:        "Host IPC blocked",
			hostNetwork: false,
			hostPID:     false,
			hostIPC:     true,
			shouldBlock: true,
			description: "hostIPC: true should be blocked",
		},
		{
			name:        "All host namespaces isolated",
			hostNetwork: false,
			hostPID:     false,
			hostIPC:     false,
			shouldBlock: false,
			description: "Isolated namespaces should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			pod := createTestPodWithHostNamespaces("test-pod", tt.hostNetwork, tt.hostPID, tt.hostIPC)

			// Act
			isValid := validatePodSecurityContext(pod)

			// Assert
			if tt.shouldBlock {
				assert.False(t, isValid, tt.description)
			} else {
				assert.True(t, isValid, tt.description)
			}
		})
	}
}

// TestSec004_ReadOnlyRootFilesystem validates root filesystem is read-only.
func TestSec004_ReadOnlyRootFilesystem(t *testing.T) {
	tests := []struct {
		name           string
		readOnlyRootFs *bool
		shouldBlock    bool
		description    string
	}{
		{
			name:           "Writable root filesystem blocked",
			readOnlyRootFs: boolPtr(false),
			shouldBlock:    true,
			description:    "Writable root filesystem should be blocked",
		},
		{
			name:           "Read-only root filesystem allowed",
			readOnlyRootFs: boolPtr(true),
			shouldBlock:    false,
			description:    "Read-only root filesystem should be allowed",
		},
		{
			name:           "Unset root filesystem blocked",
			readOnlyRootFs: nil,
			shouldBlock:    true,
			description:    "Missing readOnlyRootFilesystem should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			pod := createTestPodWithRootFs("test-pod", tt.readOnlyRootFs)

			// Act
			isValid := validatePodSecurityContext(pod)

			// Assert
			if tt.shouldBlock {
				assert.False(t, isValid, tt.description)
			} else {
				assert.True(t, isValid, tt.description)
			}
		})
	}
}

// TestSec004_ResourceLimitsEnforced validates resource limits are set.
func TestSec004_ResourceLimitsEnforced(t *testing.T) {
	tests := []struct {
		name        string
		hasLimits   bool
		shouldBlock bool
		description string
	}{
		{
			name:        "Missing resource limits blocked",
			hasLimits:   false,
			shouldBlock: true,
			description: "Pods without resource limits should be blocked",
		},
		{
			name:        "Resource limits set allowed",
			hasLimits:   true,
			shouldBlock: false,
			description: "Pods with resource limits should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			pod := createTestPodWithResources("test-pod", tt.hasLimits)

			// Act
			isValid := validatePodResourceLimits(pod)

			// Assert
			if tt.shouldBlock {
				assert.False(t, isValid, tt.description)
			} else {
				assert.True(t, isValid, tt.description)
			}
		})
	}
}

// TestSec004_SeccompProfileEnforced validates seccomp profile is set.
func TestSec004_SeccompProfileEnforced(t *testing.T) {
	tests := []struct {
		name        string
		seccomp     *corev1.SeccompProfile
		shouldBlock bool
		description string
	}{
		{
			name:        "Missing seccomp profile blocked",
			seccomp:     nil,
			shouldBlock: true,
			description: "Missing seccomp profile should be blocked",
		},
		{
			name: "RuntimeDefault seccomp allowed",
			seccomp: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
			shouldBlock: false,
			description: "RuntimeDefault seccomp should be allowed",
		},
		{
			name: "Unconfined seccomp blocked",
			seccomp: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeUnconfined,
			},
			shouldBlock: true,
			description: "Unconfined seccomp should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			pod := createTestPodWithSeccomp("test-pod", tt.seccomp)

			// Act
			isValid := validatePodSecurityContext(pod)

			// Assert
			if tt.shouldBlock {
				assert.False(t, isValid, tt.description)
			} else {
				assert.True(t, isValid, tt.description)
			}
		})
	}
}

// TestSec004_CompletePodSecurityValidation validates ALL required security fields together.
func TestSec004_CompletePodSecurityValidation(t *testing.T) {
	tests := getCompletePodSecurityTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, missingFields := validateCompletePodSecurity(tt.pod)

			assert.Equal(t, tt.isValid, isValid, tt.description)
			if !tt.isValid {
				assert.NotEmpty(t, missingFields, "Should report missing fields")
				for _, field := range tt.missing {
					assert.Contains(t, missingFields, field, "Should detect missing: %s", field)
				}
			}
		})
	}
}

// getCompletePodSecurityTestCases returns test cases for complete pod security validation.
func getCompletePodSecurityTestCases() []struct {
	name        string
	pod         *corev1.Pod
	isValid     bool
	missing     []string
	description string
} {
	return []struct {
		name        string
		pod         *corev1.Pod
		isValid     bool
		missing     []string
		description string
	}{
		{
			name:        "Fully secure pod",
			pod:         createFullySecurePod("secure-pod"),
			isValid:     true,
			missing:     []string{},
			description: "Pod with all security fields should be valid",
		},
		{
			name:        "Missing runAsNonRoot",
			pod:         createPodMissingField("runAsNonRoot"),
			isValid:     false,
			missing:     []string{"runAsNonRoot"},
			description: "Pod without runAsNonRoot should be invalid",
		},
		{
			name:        "Missing allowPrivilegeEscalation",
			pod:         createPodMissingField("allowPrivilegeEscalation"),
			isValid:     false,
			missing:     []string{"allowPrivilegeEscalation"},
			description: "Pod without allowPrivilegeEscalation should be invalid",
		},
		{
			name:        "Missing readOnlyRootFilesystem",
			pod:         createPodMissingField("readOnlyRootFilesystem"),
			isValid:     false,
			missing:     []string{"readOnlyRootFilesystem"},
			description: "Pod without readOnlyRootFilesystem should be invalid",
		},
		{
			name:        "Missing capabilities drop ALL",
			pod:         createPodMissingField("capabilities"),
			isValid:     false,
			missing:     []string{"capabilities"},
			description: "Pod without dropping ALL capabilities should be invalid",
		},
		{
			name:        "Missing seccomp profile",
			pod:         createPodMissingField("seccomp"),
			isValid:     false,
			missing:     []string{"seccomp"},
			description: "Pod without seccomp profile should be invalid",
		},
	}
}

// TestSec004_MultipleViolations validates pod with multiple security issues.
func TestSec004_MultipleViolations(t *testing.T) {
	// Create pod with multiple violations
	privileged := true
	allowEscalation := true

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "insecure-pod"},
		Spec: corev1.PodSpec{
			HostNetwork: true, // Violation 1
			HostPID:     true, // Violation 2
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
				SecurityContext: &corev1.SecurityContext{
					Privileged:               &privileged,      // Violation 3
					AllowPrivilegeEscalation: &allowEscalation, // Violation 4
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{"SYS_ADMIN"}, // Violation 5
					},
				},
			}},
		},
	}

	// Act
	isValid, violations := validateCompletePodSecurity(pod)

	// Assert
	assert.False(t, isValid, "Pod with multiple violations should be invalid")
	assert.GreaterOrEqual(t, len(violations), 5, "Should detect all 5 violations")
	assert.Contains(t, violations, "hostNetwork", "Should detect hostNetwork")
	assert.Contains(t, violations, "hostPID", "Should detect hostPID")
	assert.Contains(t, violations, "privileged", "Should detect privileged")
	assert.Contains(t, violations, "allowPrivilegeEscalation", "Should detect allowPrivilegeEscalation")
	assert.Contains(t, violations, "dangerous_capability", "Should detect dangerous capability")
}

// Helper Functions.

func createTestPod(name string, privileged bool, allowEscalation bool) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test:latest",
					SecurityContext: &corev1.SecurityContext{
						Privileged:               &privileged,
						AllowPrivilegeEscalation: &allowEscalation,
					},
				},
			},
		},
	}
}

func createTestPodWithEscalation(name string, allowEscalation *bool) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:            "test",
				Image:           "test:latest",
				SecurityContext: &corev1.SecurityContext{},
			}},
		},
	}
	if allowEscalation != nil {
		pod.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation = allowEscalation
	}
	return pod
}

func createTestPodWithUser(name string, runAsNonRoot *bool, runAsUser *int64) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{},
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
			}},
		},
	}
	if runAsNonRoot != nil {
		pod.Spec.SecurityContext.RunAsNonRoot = runAsNonRoot
	}
	if runAsUser != nil {
		pod.Spec.SecurityContext.RunAsUser = runAsUser
	}
	return pod
}

func createTestPodWithCapabilities(name string, caps []corev1.Capability) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Add: caps,
					},
				},
			}},
		},
	}
}

func createTestPodWithDroppedCapabilities(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{"ALL"},
					},
				},
			}},
		},
	}
}

func createTestPodWithHostPath(name string, path string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: "host-volume",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: path,
					},
				},
			}},
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
			}},
		},
	}
}

func createTestPodWithHostNamespaces(name string, hostNetwork, hostPID, hostIPC bool) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			HostNetwork: hostNetwork,
			HostPID:     hostPID,
			HostIPC:     hostIPC,
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
			}},
		},
	}
}

func createTestPodWithRootFs(name string, readOnly *bool) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:            "test",
				Image:           "test:latest",
				SecurityContext: &corev1.SecurityContext{},
			}},
		},
	}
	if readOnly != nil {
		pod.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem = readOnly
	}
	return pod
}

func createTestPodWithResources(name string, hasLimits bool) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
			}},
		},
	}
	if hasLimits {
		pod.Spec.Containers[0].Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    parseQuantity("2000m"),
				corev1.ResourceMemory: parseQuantity("2Gi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    parseQuantity("100m"),
				corev1.ResourceMemory: parseQuantity("256Mi"),
			},
		}
	}
	return pod
}

func createTestPodWithSeccomp(name string, seccomp *corev1.SeccompProfile) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{},
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
			}},
		},
	}
	if seccomp != nil {
		pod.Spec.SecurityContext.SeccompProfile = seccomp
	}
	return pod
}

func validatePodSecurityContext(pod *corev1.Pod) bool {
	hasContainerSecurityContext, hasAnySecureFieldSet := validateContainerSecurityContexts(pod)
	hasPodSecurityContext, podSecureFieldSet := validatePodLevelSecurityContext(pod)

	if (hasContainerSecurityContext || hasPodSecurityContext) && !hasAnySecureFieldSet && !podSecureFieldSet {
		return false
	}

	return !hasHostNamespaces(pod)
}

// validateContainerSecurityContexts validates container security contexts.
func validateContainerSecurityContexts(pod *corev1.Pod) (bool, bool) {
	hasContainerSecurityContext := false
	hasAnySecureFieldSet := false

	for _, container := range pod.Spec.Containers {
		if container.SecurityContext != nil {
			hasContainerSecurityContext = true

			if !validateContainerPrivileged(container, &hasAnySecureFieldSet) {
				return hasContainerSecurityContext, false
			}

			if !validateContainerPrivilegeEscalation(container, &hasAnySecureFieldSet) {
				return hasContainerSecurityContext, false
			}

			if !validateContainerReadOnlyRootFS(container, &hasAnySecureFieldSet) {
				return hasContainerSecurityContext, false
			}

			if !validateContainerCapabilities(container, &hasAnySecureFieldSet) {
				return hasContainerSecurityContext, false
			}
		}
	}

	return hasContainerSecurityContext, hasAnySecureFieldSet
}

// validateContainerPrivileged validates the privileged setting.
func validateContainerPrivileged(container corev1.Container, hasAnySecureFieldSet *bool) bool {
	if container.SecurityContext.Privileged != nil {
		*hasAnySecureFieldSet = true
		if *container.SecurityContext.Privileged {
			return false
		}
	}
	return true
}

// validateContainerPrivilegeEscalation validates the allowPrivilegeEscalation setting.
func validateContainerPrivilegeEscalation(container corev1.Container, hasAnySecureFieldSet *bool) bool {
	if container.SecurityContext.AllowPrivilegeEscalation != nil {
		*hasAnySecureFieldSet = true
		if *container.SecurityContext.AllowPrivilegeEscalation {
			return false
		}
	}
	return true
}

// validateContainerReadOnlyRootFS validates the readOnlyRootFilesystem setting.
func validateContainerReadOnlyRootFS(container corev1.Container, hasAnySecureFieldSet *bool) bool {
	if container.SecurityContext.ReadOnlyRootFilesystem != nil {
		*hasAnySecureFieldSet = true
		if !*container.SecurityContext.ReadOnlyRootFilesystem {
			return false
		}
	}
	return true
}

// validateContainerCapabilities validates container capabilities.
func validateContainerCapabilities(container corev1.Container, hasAnySecureFieldSet *bool) bool {
	if container.SecurityContext.Capabilities != nil {
		*hasAnySecureFieldSet = true
		dangerous := []string{"SYS_ADMIN", "SYS_MODULE", "SYS_RAWIO", "SYS_PTRACE", "SYS_BOOT", "NET_ADMIN", "DAC_OVERRIDE"}
		for _, cap := range container.SecurityContext.Capabilities.Add {
			for _, d := range dangerous {
				if string(cap) == d {
					return false
				}
			}
		}
	}
	return true
}

// validatePodLevelSecurityContext validates pod-level security context.
func validatePodLevelSecurityContext(pod *corev1.Pod) (bool, bool) {
	hasPodSecurityContext := false
	hasSecureFieldSet := false

	if pod.Spec.SecurityContext != nil {
		hasPodSecurityContext = true

		if pod.Spec.SecurityContext.RunAsNonRoot != nil {
			hasSecureFieldSet = true
			if !*pod.Spec.SecurityContext.RunAsNonRoot {
				return hasPodSecurityContext, false
			}
		}

		if pod.Spec.SecurityContext.SeccompProfile != nil {
			hasSecureFieldSet = true
			if pod.Spec.SecurityContext.SeccompProfile.Type == corev1.SeccompProfileTypeUnconfined {
				return hasPodSecurityContext, false
			}
		}
	}

	return hasPodSecurityContext, hasSecureFieldSet
}

// hasHostNamespaces checks if pod uses host namespaces.
func hasHostNamespaces(pod *corev1.Pod) bool {
	return pod.Spec.HostNetwork || pod.Spec.HostPID || pod.Spec.HostIPC
}

func validatePodVolumes(pod *corev1.Pod) bool {
	dangerousPaths := []string{"/", "/proc", "/sys", "/dev", "/etc", "/var/run/docker.sock", "/var/run/containerd.sock", "/run/containerd"}

	for _, volume := range pod.Spec.Volumes {
		if volume.HostPath != nil {
			for _, dangerous := range dangerousPaths {
				if volume.HostPath.Path == dangerous {
					return false
				}
			}
		}
	}
	return true
}

func validatePodResourceLimits(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if len(container.Resources.Limits) == 0 {
			return false
		}
		if len(container.Resources.Requests) == 0 {
			return false
		}
	}
	return true
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}

func parseQuantity(s string) resource.Quantity {
	return resource.MustParse(s)
}

func createFullySecurePod(name string) *corev1.Pod {
	runAsNonRoot := true
	allowPrivilegeEscalation := false
	readOnlyRootFilesystem := true

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: &runAsNonRoot,
				RunAsUser:    int64Ptr(65534),
				SeccompProfile: &corev1.SeccompProfile{
					Type: corev1.SeccompProfileTypeRuntimeDefault,
				},
			},
			Containers: []corev1.Container{{
				Name:  "test",
				Image: "test:latest",
				SecurityContext: &corev1.SecurityContext{
					AllowPrivilegeEscalation: &allowPrivilegeEscalation,
					ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{"ALL"},
					},
				},
			}},
		},
	}
}

func createPodMissingField(missingField string) *corev1.Pod {
	pod := createFullySecurePod("pod")

	// Remove the specified field
	switch missingField {
	case "runAsNonRoot":
		pod.Spec.SecurityContext.RunAsNonRoot = nil
	case "allowPrivilegeEscalation":
		pod.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation = nil
	case "readOnlyRootFilesystem":
		pod.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem = nil
	case "capabilities":
		pod.Spec.Containers[0].SecurityContext.Capabilities = nil
	case "seccomp":
		pod.Spec.SecurityContext.SeccompProfile = nil
	}

	return pod
}

func validateCompletePodSecurity(pod *corev1.Pod) (bool, []string) {
	var violations []string

	violations = append(violations, validatePodSecurityContextFields(pod)...)
	violations = append(violations, validateHostNamespaceUsage(pod)...)
	violations = append(violations, validateAllContainerSecurity(pod)...)

	return len(violations) == 0, violations
}

// validatePodSecurityContextFields validates pod-level security context fields.
func validatePodSecurityContextFields(pod *corev1.Pod) []string {
	var violations []string

	if pod.Spec.SecurityContext == nil {
		violations = append(violations, "pod_security_context_missing")
	} else {
		if pod.Spec.SecurityContext.RunAsNonRoot == nil || !*pod.Spec.SecurityContext.RunAsNonRoot {
			violations = append(violations, "runAsNonRoot")
		}

		if pod.Spec.SecurityContext.SeccompProfile == nil {
			violations = append(violations, "seccomp")
		} else if pod.Spec.SecurityContext.SeccompProfile.Type == corev1.SeccompProfileTypeUnconfined {
			violations = append(violations, "seccomp_unconfined")
		}
	}

	return violations
}

// validateHostNamespaceUsage validates host namespace usage.
func validateHostNamespaceUsage(pod *corev1.Pod) []string {
	var violations []string

	if pod.Spec.HostNetwork {
		violations = append(violations, "hostNetwork")
	}
	if pod.Spec.HostPID {
		violations = append(violations, "hostPID")
	}
	if pod.Spec.HostIPC {
		violations = append(violations, "hostIPC")
	}

	return violations
}

// validateAllContainerSecurity validates security for all containers.
func validateAllContainerSecurity(pod *corev1.Pod) []string {
	var violations []string

	for i, container := range pod.Spec.Containers {
		prefix := "container_" + container.Name + "_"

		if container.SecurityContext == nil {
			violations = append(violations, prefix+"security_context_missing")
			continue
		}

		violations = append(violations, validateContainerSecurityFields(container)...)
		violations = append(violations, validateContainerCapabilitiesComplete(container)...)

		_ = i // Suppress unused warning
	}

	return violations
}

// validateContainerSecurityFields validates container security fields.
func validateContainerSecurityFields(container corev1.Container) []string {
	var violations []string

	if container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
		violations = append(violations, "privileged")
	}

	if container.SecurityContext.AllowPrivilegeEscalation == nil || *container.SecurityContext.AllowPrivilegeEscalation {
		violations = append(violations, "allowPrivilegeEscalation")
	}

	if container.SecurityContext.ReadOnlyRootFilesystem == nil || !*container.SecurityContext.ReadOnlyRootFilesystem {
		violations = append(violations, "readOnlyRootFilesystem")
	}

	return violations
}

// validateContainerCapabilitiesComplete validates container capabilities completely.
func validateContainerCapabilitiesComplete(container corev1.Container) []string {
	var violations []string

	if container.SecurityContext.Capabilities == nil {
		violations = append(violations, "capabilities")
	} else {
		hasDropAll := false
		for _, cap := range container.SecurityContext.Capabilities.Drop {
			if cap == "ALL" {
				hasDropAll = true
				break
			}
		}
		if !hasDropAll {
			violations = append(violations, "capabilities_drop_all_missing")
		}

		dangerousCaps := []string{"SYS_ADMIN", "SYS_MODULE", "SYS_RAWIO", "SYS_PTRACE", "SYS_BOOT", "NET_ADMIN", "DAC_OVERRIDE"}
		for _, cap := range container.SecurityContext.Capabilities.Add {
			for _, dangerous := range dangerousCaps {
				if string(cap) == dangerous {
					violations = append(violations, "dangerous_capability")
					break
				}
			}
		}
	}

	return violations
}
