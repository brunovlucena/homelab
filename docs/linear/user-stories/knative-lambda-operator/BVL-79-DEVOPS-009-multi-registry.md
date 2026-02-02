# 🔄 DEVOPS-009: Multi-Registry Support

**Epic**: Container Registry Flexibility  
**Priority**: P1 | **Status**: Ready for Development**Points**: 8  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-241/devops-009-multi-registry-support

**Sprint**: v1.1.0  

---


## 📋 User Story

**As a** DevOps Engineer  
**I want to** multi-registry support  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] *AC1**: Can authenticate to GHCR using GitHub token
- [ ] *AC2**: Images push to `ghcr.io/brunovlucena/knative-lambda-builder`
- [ ] *AC3**: Makefile supports `REGISTRY_TYPE=ghcr` variable
- [ ] *AC4**: Helm values support GHCR registry configuration
- [ ] *AC5**: GHCR image pull secrets are created in templates
- [ ] *AC6**: Version tags follow semantic versioning
- [ ] *Tests:**
- [ ] u $GITHUB_USERNAME --password-stdin
- [ ] build docker-push REGISTRY_TYPE=ghcr
- [ ] lambda-builder:latest

---


## 🎯 Goal

Enable knative-lambda to push/pull images from multiple container registries (AWS ECR, GitHub GHCR, future GCR) with automatic selection based on environment.

---

## 📖 User Stories

### **Story 1: GitHub Container Registry (GHCR) Support**

**As a** developer  
**I want** to push images to GitHub Container Registry  
**So that** I can use free container hosting and integrate with GitHub workflows

**Acceptance Criteria:**

✅ **AC1**: Can authenticate to GHCR using GitHub token  
✅ **AC2**: Images push to `ghcr.io/brunovlucena/knative-lambda-builder`  
✅ **AC3**: Makefile supports `REGISTRY_TYPE=ghcr` variable  
✅ **AC4**: Helm values support GHCR registry configuration  
✅ **AC5**: GHCR image pull secrets are created in templates  
✅ **AC6**: Version tags follow semantic versioning  

**Tests:**

```bash
# Test 1: Build and push to GHCR
test_ghcr_build_push() {
  export REGISTRY_TYPE=ghcr
  export GITHUB_USERNAME=brunovlucena
  export GITHUB_TOKEN=$GITHUB_TOKEN
  
  # Login
  echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin
  
  # Build and push
  make docker-build docker-push REGISTRY_TYPE=ghcr
  
  # Verify image exists
  docker pull ghcr.io/brunovlucena/knative-lambda-builder:latest
  
  assert_success
}

# Test 2: Deploy with GHCR image
test_ghcr_deployment() {
  # Create secret
  kubectl create secret docker-registry ghcr-secret \
    --docker-server=ghcr.io \
    --docker-username=brunovlucena \
    --docker-password=$GITHUB_TOKEN \
    -n knative-lambda
  
  # Deploy
  helm upgrade --install knative-lambda-builder ./deploy \
    -f deploy/overlays/dev/values-dev.yaml \
    --set image.registry=ghcr.io/brunovlucena \
    --namespace knative-lambda
  
  # Verify pod is running
  kubectl wait --for=condition=ready pod -l app=knative-lambda-builder \
    -n knative-lambda --timeout=5m
  
  assert_success
}

# Test 3: Multi-component push (builder, sidecar, metrics-pusher)
test_ghcr_all_components() {
  make docker-build REGISTRY_TYPE=ghcr
  
  # Verify all images built
  docker images | grep ghcr.io/brunovlucena/knative-lambda-builder
  docker images | grep ghcr.io/brunovlucena/knative-lambda-sidecar
  docker images | grep ghcr.io/brunovlucena/knative-lambda-metrics-pusher
  
  # Push all
  make docker-push REGISTRY_TYPE=ghcr
  
  assert_success
}
```

---

### **Story 2: Environment-Based Registry Selection**

**As a** DevOps engineer  
**I want** automatic registry selection based on environment  
**So that** dev uses GHCR, production uses ECR without manual changes

**Acceptance Criteria:**

✅ **AC1**: `local` environment uses GHCR  
✅ **AC2**: `dev` environment uses GHCR  
✅ **AC3**: `prd` environment uses ECR  
✅ **AC4**: Registry config in environment-specific values files  
✅ **AC5**: Image pull secrets match registry type  
✅ **AC6**: CI/CD automatically selects registry from branch  

**Tests:**

```bash
# Test 1: Local environment defaults to GHCR
test_local_uses_ghcr() {
  grep "registry: ghcr.io/brunovlucena" deploy/overlays/local/values-local.yaml
  assert_success
}

# Test 2: Production uses ECR
test_prd_uses_ecr() {
  grep "registry: 339954290315.dkr.ecr.us-west-2.amazonaws.com" \
    deploy/overlays/prd/values-prd.yaml
  assert_success
}

# Test 3: Image pull secrets match registry
test_image_pull_secrets() {
  # Local should use ghcr-secret
  grep "ghcr-secret" deploy/overlays/local/values-local.yaml
  
  # Prd should use ecr-secret (or none for IAM)
  grep -E "ecr-secret | imagePullSecrets: \[\]" deploy/overlays/prd/values-prd.yaml
  
  assert_success
}

# Test 4: CI/CD registry selection
test_cicd_registry_selection() {
  # Simulate main branch
  export GITHUB_REF=refs/heads/main
  registry=$(./scripts/determine-registry.sh)
  assert_equals "ecr" "$registry"
  
  # Simulate develop branch
  export GITHUB_REF=refs/heads/develop
  registry=$(./scripts/determine-registry.sh)
  assert_equals "ghcr" "$registry"
  
  # Simulate feature branch
  export GITHUB_REF=refs/heads/feature/test
  registry=$(./scripts/determine-registry.sh)
  assert_equals "ghcr" "$registry"
}
```

---

### **Story 3: Version Tagging Strategy per Registry**

**As a** DevOps engineer  
**I want** consistent version tagging across registries  
**So that** I can track deployments and rollback easily

**Acceptance Criteria:**

✅ **AC1**: `main` branch → `v1.0.0` (semantic version)  
✅ **AC2**: `develop` branch → `v1.0.0-beta.TIMESTAMP` (beta tags)  
✅ **AC3**: Feature branches → `v1.0.0-dev.SHA` (dev tags)  
✅ **AC4**: All images tagged with: version + latest + commit SHA  
✅ **AC5**: Tags are pushed to Git repository  
✅ **AC6**: `version-manager.sh` generates correct tags  

**Tests:**

```bash
# Test 1: Main branch tagging
test_main_branch_tagging() {
  export GIT_BRANCH=main
  VERSION=1.0.0
  
  tag=$(./scripts/version-manager.sh generate-tag)
  assert_equals "v1.0.0" "$tag"
}

# Test 2: Develop branch tagging
test_develop_branch_tagging() {
  export GIT_BRANCH=develop
  VERSION=1.0.0
  
  tag=$(./scripts/version-manager.sh generate-tag)
  assert_matches "^v1.0.0-beta.[0-9]+$" "$tag"
}

# Test 3: Feature branch tagging
test_feature_branch_tagging() {
  export GIT_BRANCH=feature/test-feature
  VERSION=1.0.0
  SHA_SHORT=abc123
  
  tag=$(./scripts/version-manager.sh generate-tag)
  assert_equals "v1.0.0-dev.abc123" "$tag"
}

# Test 4: Multiple tags applied
test_multiple_tags() {
  make docker-build REGISTRY_TYPE=ghcr
  
  # Verify tags
  docker images ghcr.io/brunovlucena/knative-lambda-builder --format "{{.Tag}}" | \
    grep -E "^(v1.0.0 | latest | abc123)$"
  
  assert_success
}

# Test 5: Git tag creation
test_git_tag_creation() {
  ./scripts/version-manager.sh bump 1.1.0
  
  # Verify Chart.yaml updated
  grep "^version: 1.1.0" deploy/Chart.yaml
  
  # Verify git tag exists
  git tag | grep "knative-lambda-v1.1.0"
  
  assert_success
}
```

---

### **Story 4: GCR Support (Future)**

**As a** platform engineer  
**I want** to support Google Container Registry  
**So that** we can deploy to GCP environments

**Acceptance Criteria:**

✅ **AC1**: Makefile supports `REGISTRY_TYPE=gcr`  
✅ **AC2**: Can authenticate using gcloud CLI  
✅ **AC3**: Images push to `gcr.io/PROJECT/knative-lambda-builder`  
✅ **AC4**: Helm values support GCR configuration  
✅ **AC5**: GCR image pull secrets work in GKE  

**Tests:**

```bash
# Test 1: Build and push to GCR
test_gcr_build_push() {
  export REGISTRY_TYPE=gcr
  export GCP_PROJECT=my-project
  
  # Authenticate
  gcloud auth configure-docker gcr.io
  
  # Build and push
  make docker-build docker-push REGISTRY_TYPE=gcr
  
  # Verify image exists
  gcloud container images list --repository=gcr.io/$GCP_PROJECT | \
    grep knative-lambda-builder
  
  assert_success
}

# Test 2: GKE deployment with GCR
test_gcr_gke_deployment() {
  # Deploy to GKE
  helm upgrade --install knative-lambda-builder ./deploy \
    -f deploy/overlays/gcp/values-gcp.yaml \
    --namespace knative-lambda-gcp
  
  # Verify pod pulls from GCR
  kubectl describe pod -l app=knative-lambda-builder -n knative-lambda-gcp | \
    grep "gcr.io/$GCP_PROJECT"
  
  assert_success
}
```

---

## 🧪 Integration Tests

### Test Suite: Multi-Registry End-to-End

```bash
#!/bin/bash
# tests/integration/test-multi-registry.sh

set -e

echo "🧪 Testing Multi-Registry Support"

# Setup
export VERSION=1.0.0
export GIT_COMMIT=$(git rev-parse --short HEAD)

# Test 1: GHCR Flow
echo "Test 1: GHCR complete flow..."
export REGISTRY_TYPE=ghcr
make docker-login
make docker-build
make docker-push
docker pull ghcr.io/brunovlucena/knative-lambda-builder:latest
echo "✅ GHCR flow complete"

# Test 2: ECR Flow
echo "Test 2: ECR complete flow..."
export REGISTRY_TYPE=ecr
make docker-login
make docker-build
make docker-push
docker pull 339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambdas/knative-lambda-builder:latest
echo "✅ ECR flow complete"

# Test 3: Environment-based selection
echo "Test 3: Environment-based deployment..."
for env in local dev prd; do
  echo "Testing $env environment..."
  kubectl create namespace knative-lambda-$env --dry-run=client -o yaml | kubectl apply -f -
  
  helm upgrade --install knative-lambda-builder ./deploy \
    -f deploy/overlays/$env/values-$env.yaml \
    --namespace knative-lambda-$env \
    --dry-run
  
  echo "✅ $env environment valid"
done

# Test 4: Version tagging
echo "Test 4: Version tagging..."
for branch in main develop feature/test; do
  export GIT_BRANCH=$branch
  tag=$(./scripts/version-manager.sh generate-tag)
  echo "Branch $branch → Tag $tag"
done
echo "✅ Version tagging complete"

echo "🎉 All multi-registry tests passed!"
```

---

## 📝 Implementation Checklist

### Phase 1: GHCR Support (Week 1)

- [ ] Update Makefile with registry type selection
- [ ] Add GHCR authentication logic
- [ ] Update `deploy/values.yaml` with registry variables
- [ ] Create GHCR overlay: `deploy/overlays/dev/values-dev.yaml`
- [ ] Add GHCR image pull secret template
- [ ] Test GHCR build and push locally
- [ ] Test GHCR deployment to cluster
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Update documentation

### Phase 2: Environment-Based Selection (Week 2)

- [ ] Update all environment overlays with registry config
- [ ] Create `scripts/determine-registry.sh`
- [ ] Update Helm templates to use registry variables
- [ ] Add registry validation
- [ ] Test each environment configuration
- [ ] Write environment selection tests
- [ ] Update CI/CD workflows

### Phase 3: Version Tagging (Week 3)

- [ ] Enhance `scripts/version-manager.sh` with tag generation
- [ ] Add multi-tag support to Makefile
- [ ] Update CI/CD to apply all tags
- [ ] Test version tagging for all branch types
- [ ] Add Git tag creation automation
- [ ] Write tagging tests

### Phase 4: GCR Support (Future - Week 4)

- [ ] Add GCR support to Makefile
- [ ] Create GCR overlay: `deploy/overlays/gcp/values-gcp.yaml`
- [ ] Add GCR authentication
- [ ] Test GCR build and push
- [ ] Test GKE deployment
- [ ] Write GCR tests

---

## 🎯 Definition of Done

- ✅ All user stories implemented
- ✅ All acceptance criteria met
- ✅ All tests passing (unit + integration + e2e)
- ✅ Can build and push to GHCR
- ✅ Can build and push to ECR (existing)
- ✅ Environment-based selection works
- ✅ Version tagging works for all branches
- ✅ CI/CD workflows updated
- ✅ Zero production downtime
- ✅ Code reviewed and approved
- ✅ Documentation updated

---

**Estimated Effort**: 3 weeks (GHCR + Selection + Tagging)  
**Dependencies**: None  
**Risk Level**: Low (additive changes, no breaking changes)

