# BVL-132-SRE-001-reduce-eks-main-workers-from-4-to-3

**Status**: Backlog  
**Priority**: 🟡 High  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-132/reduce-eks-main-workers-from-4-to-3-nodes  
**Created**: 2026-01-08T18:11:20.231Z  
**Updated**: 2026-01-08T18:11:20.231Z  
**Project**: notifi

---

## Objective

Reduce the EKS production cluster main worker node group from 4 nodes to 3 nodes to optimize costs while maintaining adequate capacity and availability.

## Current State

* **Node Group**: `main-2b` (notifi-uw2-prd-eks-cluster)
* **Current Nodes**: 4x r8g.2xlarge (ARM64)
* **Current Utilization**: ~28.55% CPU, ~28.75% Memory across 4 nodes
* **Projected Utilization on 3 Nodes**: ~38% CPU, ~38.3% Memory per node

## Configuration Changes Required

Update `platform/stacks/uw2-prd.yaml`:

* Change `main-2b.desired_group_size` from 4 to 3
* Change `main-2b.max_group_size` from 4 to 3
* Keep `main-2b.min_group_size` at 3 (already correct)

## Key Tasks

1. Pre-change validation (verify metrics, PDBs, node affinity)
2. Update Terraform configuration
3. Apply changes via `atmos terraform apply eks`
4. Monitor scale-down and pod rescheduling
5. Post-change validation (24-48 hour monitoring)

## Acceptance Criteria

- [ ] Node group successfully reduced from 4 to 3 nodes
- [ ] All pods rescheduled and running healthy
- [ ] No service disruptions
- [ ] Resource utilization remains below 70% per node
- [ ] Configuration changes committed

## Cost Impact

Estimated ~25% reduction in main worker node costs.

**User Story**: See `docs/user-stories/SRE-001-reduce-eks-main-workers-from-4-to-3.md`
