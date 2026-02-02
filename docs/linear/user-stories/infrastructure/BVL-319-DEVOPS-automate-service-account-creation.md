# 🔄 DEVOPS: Automate Service Account Creation

**Linear ID**: BVL-319  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-319/devops-automate-service-account-creation

---

## 📋 User Story

**As a** Site Reliability Engineer  
**I want** service accounts to be created automatically through GitOps  
**So that** I can reduce manual configuration errors and ensure consistent service account management across the cluster

---

## 🎯 Feature Description

Currently, service accounts are created manually, which leads to:
- Inconsistent configurations
- Manual errors
- Lack of version control
- Difficulty tracking service account lifecycle
- Time-consuming manual processes

This feature will automate service account creation through GitOps workflows, ensuring all service accounts are:
- Defined as code
- Version controlled
- Automatically applied to the cluster
- Consistent across environments
- Easily auditable

---

## 🎯 Acceptance Criteria

- [ ] Service accounts defined in Git repository (YAML manifests)
- [ ] GitOps workflow automatically applies service accounts to cluster
- [ ] Service accounts created with consistent naming conventions
- [ ] RoleBindings/ClusterRoleBindings automatically created with service accounts
- [ ] Service account secrets automatically generated and mounted
- [ ] Support for multiple namespaces
- [ ] Validation of service account configurations before applying
- [ ] Documentation for creating new service accounts
- [ ] Integration with existing GitOps tooling (Flux/Kustomize)
- [ ] Service account lifecycle management (create, update, delete)
- [ ] Monitoring/alerting for service account creation failures

---

## 🔧 Technical Requirements

**Implementation Approach:**
1. Define service accounts as Kubernetes manifests in Git
2. Use GitOps tool (Flux/Kustomize) to apply manifests
3. Create templates/patterns for common service account configurations
4. Implement validation checks
5. Add documentation and examples

**Service Account Structure:**
- Standard Kubernetes ServiceAccount resource
- Associated RoleBindings/ClusterRoleBindings
- ImagePullSecrets (if needed)
- Annotations and labels for tracking

**GitOps Integration:**
- Service accounts stored in Git repository
- Automatic sync via Flux/Kustomize
- Environment-specific configurations (dev/staging/prod)
- Rollback capabilities

**Current State:**
Service accounts are currently defined manually in various locations:
- `flux/ai/agent-*/k8s/kustomize/base/rbac.yaml`
- `flux/infrastructure/*/k8s/base/rbac.yaml`
- Individual service account YAML files

**Proposed Structure:**
```
flux/infrastructure/service-accounts/
├── base/
│   ├── kustomization.yaml
│   ├── serviceaccount-template.yaml
│   └── rbac-template.yaml
├── namespaces/
│   ├── default/
│   ├── ai/
│   └── ...
└── README.md
```

---

## 🔄 GitOps Workflow

```
┌────────────────────────────────────────────────────────────────┐
│          SERVICE ACCOUNT AUTOMATION WORKFLOW                   │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  1. DEVELOPER CREATES SERVICE ACCOUNT MANIFEST                │
│     Developer → Create YAML in Git → git commit → git push     │
│                                                                │
│  2. FLUX DETECTS CHANGE                                       │
│     Flux (5min interval) → Poll Git repository                 │
│     └─ Detect new service account manifest                    │
│                                                                │
│  3. FLUX APPLIES SERVICE ACCOUNT                              │
│     Flux → kubectl apply -f service-accounts/                  │
│     ├─ ServiceAccount                                         │
│     ├─ Role/ClusterRole                                       │
│     ├─ RoleBinding/ClusterRoleBinding                         │
│     └─ ImagePullSecrets (if configured)                        │
│                                                                │
│  4. VALIDATION                                                │
│     ├─ Check naming conventions                               │
│     ├─ Verify RBAC permissions                                │
│     └─ Validate namespace exists                              │
│                                                                │
│  5. MONITORING                                                │
│     ├─ Alert on creation failures                             │
│     ├─ Track service account lifecycle                        │
│     └─ Audit log for compliance                               │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## 📊 Benefits

- **Consistency**: All service accounts follow the same patterns
- **Version Control**: Full history of service account changes
- **Automation**: No manual kubectl commands needed
- **Auditability**: Easy to track who created what and when
- **Scalability**: Easy to create multiple service accounts
- **Compliance**: Service accounts defined as code for compliance
- **Self-Service**: Developers can create service accounts via PR
- **Rollback**: Easy rollback via Git revert

---

## 🔗 Related Work

- GitOps workflow improvements
- RBAC automation
- Cluster configuration management
- Flux Kustomization enhancements

---

## 📝 Implementation Notes

**Phase 1: Foundation**
- Create service account template structure
- Set up Flux Kustomization for service accounts
- Document naming conventions

**Phase 2: Automation**
- Implement validation checks
- Add monitoring/alerting
- Create developer documentation

**Phase 3: Enhancement**
- Add service account rotation policies
- Integrate with secret management
- Multi-cluster support

---

**Last Updated**: January 13, 2026  
**Status**: Backlog  
**Priority**: Normal  
**Labels**: SRE, Feature, DevOps
