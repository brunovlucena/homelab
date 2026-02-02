# Agent-SRE Test Manifests

This directory contains Kubernetes manifests for creating problematic pods and resources to test LambdaFunction remediation capabilities.

## Test Scenarios

### Pod Issues

1. **pod-crashloopbackoff.yaml**
   - Pod that crashes immediately
   - Tests: `pod-restart` LambdaFunction
   - Expected: Pod should be restarted

2. **pod-imagepullbackoff.yaml**
   - Pod with invalid image
   - Tests: `pod-check-status` LambdaFunction
   - Expected: Status check should identify ImagePullBackOff

3. **pod-pending.yaml**
   - Pod that cannot be scheduled (node selector mismatch)
   - Tests: `pod-check-status` LambdaFunction
   - Expected: Status check should identify Pending status

4. **pod-resource-constrained.yaml**
   - Pod requesting excessive resources
   - Tests: `pod-check-status` LambdaFunction
   - Expected: Status check should identify resource constraints

5. **pod-oomkilled.yaml**
   - Pod that gets OOMKilled
   - Tests: `pod-restart` LambdaFunction
   - Expected: Pod should be restarted (may need memory increase)

### Deployment Issues

6. **deployment-unhealthy.yaml**
   - Deployment with crashing pods
   - Tests: `pod-restart` LambdaFunction (type: deployment)
   - Expected: Deployment should be restarted

7. **deployment-needs-scaling.yaml**
   - Deployment with insufficient replicas
   - Tests: `scale-deployment` LambdaFunction
   - Expected: Deployment should be scaled up

8. **deployment-high-cpu.yaml**
   - Deployment with high CPU usage
   - Tests: `scale-deployment` LambdaFunction
   - Expected: Deployment should be scaled horizontally

### Storage Issues

9. **pvc-filling-up.yaml**
   - PVC that is filling up
   - Tests: `check-pvc-status` LambdaFunction
   - Expected: Status check should show PVC usage

### Flux Issues

10. **flux-kustomization-broken.yaml**
    - Kustomization with invalid path
    - Tests: `flux-reconcile-kustomization` LambdaFunction
    - Expected: Reconciliation should be attempted (will fail due to invalid path)

## Usage

### Apply All Test Manifests

```bash
kubectl apply -k flux/ai/agent-sre/k8s/test
```

### Apply Specific Test Scenario

```bash
kubectl apply -f flux/ai/agent-sre/k8s/test/pod-crashloopbackoff.yaml
```

### Check Test Pod Status

```bash
kubectl get pods -n ai -l test.agent-sre.io/scenario
```

### Trigger Alert for Testing

Create a PrometheusRule that fires alerts for these test scenarios:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: agent-sre-test-alerts
  namespace: prometheus
spec:
  groups:
  - name: agent-sre-tests
    rules:
    - alert: TestPodCrashLoopBackOff
      expr: kube_pod_status_phase{phase="Failed"} == 1
      for: 1m
      labels:
        test_scenario: crashloopbackoff
      annotations:
        lambda_function: "pod-restart"
        lambda_parameters: '{"name": "{{ $labels.pod }}", "namespace": "{{ $labels.namespace }}", "type": "pod"}'
```

### Clean Up Test Resources

```bash
# Remove all test resources
kubectl delete -k flux/ai/agent-sre/k8s/test

# Remove specific test
kubectl delete -f flux/ai/agent-sre/k8s/test/pod-crashloopbackoff.yaml
```

## Test Labels

All test resources are labeled with:
- `test.agent-sre.io/scenario: <scenario-name>` - Identifies the test scenario
- `test.agent-sre.io/lambda-function: <function-name>` - Indicates which LambdaFunction should be tested

## LambdaFunction Mapping

| Test Scenario | LambdaFunction | Parameters |
|--------------|----------------|------------|
| crashloopbackoff | pod-restart | `{"name": "<pod>", "namespace": "ai", "type": "pod"}` |
| unhealthy-deployment | pod-restart | `{"name": "<deployment>", "namespace": "ai", "type": "deployment"}` |
| needs-scaling | scale-deployment | `{"name": "<deployment>", "namespace": "ai", "replicas": 3}` |
| pvc-filling-up | check-pvc-status | `{"namespace": "ai"}` |
| broken-kustomization | flux-reconcile-kustomization | `{"name": "test-broken-kustomization", "namespace": "flux-system"}` |

## Notes

- All test resources are created in the `ai` namespace (except Flux resources in `flux-system`)
- Test resources are designed to be problematic to trigger alerts
- Clean up test resources after testing to avoid resource waste
- Some scenarios (like OOMKilled) may require manual intervention beyond LambdaFunction capabilities

