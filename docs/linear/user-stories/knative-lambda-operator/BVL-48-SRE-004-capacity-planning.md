# ⚡ SRE-004: Capacity Planning

**Status**: Done
**Linear URL**: https://linear.app/bvlucena/issue/BVL-222/sre-004-capacity-planning
**Priority**: P1
**Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-170/sre-004-capacity-planning  
**Created**: 2025-10-29  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---

## 📋 User Story

**As an** SRE Engineer  
**I want** to forecast resource requirements and plan capacity  
**So that** the platform handles traffic spikes without degradation

---


## 🎯 Acceptance Criteria

- [ ] [ ] Capacity model predicts resource needs 30 days ahead
- [ ] [ ] Headroom maintained at 30% for unexpected spikes
- [ ] [ ] Load tests validate capacity before major events
- [ ] [ ] Cost per build tracked and optimized
- [ ] [ ] Auto-scaling handles 3x traffic without manual intervention
- [ ] --

---


## 📊 Acceptance Criteria

- [ ] Capacity model predicts resource needs 30 days ahead
- [ ] Headroom maintained at 30% for unexpected spikes
- [ ] Load tests validate capacity before major events
- [ ] Cost per build tracked and optimized
- [ ] Auto-scaling handles 3x traffic without manual intervention

---

## 📈 Current Capacity | Resource | Current | Peak Usage | Headroom | Limit | |---------- | --------- | ------------ | ---------- | ------- | | **Kaniko Jobs** | 15 concurrent | 45 | 10% | 50 | | **Builder CPU** | 0.8 cores | 1.5 cores | 25% | 2 cores | | **Builder Memory** | 1.2Gi | 2.8Gi | 30% | 4Gi | | **RabbitMQ Queue** | 150 msgs | 800 msgs | 20% | 1000 | | **ECR Push Rate** | 2/s | 8/s | 20% | 10/s | **Action Needed**: Scale up Kaniko job limit to 100 (current 90% utilization at peak)

---

## 🎯 Forecasting Model

### Historical Growth

```
Month | Builds/Day | Peak Concurrent | Trend
--------- | ------------ | ----------------- | -------
Oct 2024 | 1,200 | 25 | +15%
Nov 2024 | 1,380 | 32 | +18%
Dec 2024 | 1,656 | 42 | +20%
Jan 2025 | 2,040 | 58 | +23%
```

**Forecast (Mar 2025)**: 3,000 builds/day, 85 peak concurrent

### Resource Requirements

```python
# Capacity calculation
builds_per_day = 3000
avg_build_duration = 45  # seconds
peak_multiplier = 3  # peak is 3x average

# Peak concurrent builds
peak_concurrent = (builds_per_day / 86400) * avg_build_duration * peak_multiplier
# = (3000/86400) * 45 * 3 ≈ 4.7 builds/s * 45s ≈ 212 concurrent

# Kaniko job limit needed (with 30% headroom)
kaniko_limit = ceil(peak_concurrent * 1.3) = 276 jobs

# Builder replicas needed (10 builds/replica)
builder_replicas = ceil(peak_concurrent / 10) = 22 replicas
```

**Recommendation**: 
- Increase Kaniko job limit: 50 → 300
- Enable HPA: 2-25 replicas (currently 2-10)
- Add dedicated build node pool

---

## 🧪 Load Testing

### Test Scenario: Black Friday (3x traffic)

```bash
# Generate 300 concurrent builds
for i in {1..300}; do
  make trigger-build-prd PARSER_ID=loadtest-$i &
  sleep 0.1  # 10 builds/s
done

# Monitor metrics
watch -n 5 'kubectl top nodes | grep build-node'
watch -n 2 'kubectl get jobs -n knative-lambda | grep -c Running'
```

**Results**:
- ✅ All 300 builds completed in 8 minutes
- ✅ No OOMKilled pods
- ❌ ECR rate limited after 250 builds (needs sharding)
- ❌ Kaniko job queue depth hit 180 (needs higher limit)

---

## 💰 Cost Analysis

### Current Costs (monthly) | Resource | Cost | Utilization | Waste | |---------- | ------ | ------------- | ------- | | Build Nodes (3x m5.2xlarge) | $450 | 65% | $157 | | ECR Storage (500GB) | $50 | - | $0 | | RabbitMQ (t3.medium) | $35 | 40% | $21 | | **Total** | **$535** | **60%** | **$178** | ### Optimizations

1. **Use Spot Instances for Build Nodes**: Save 60% ($270/month)
2. **Lifecycle Policy for ECR**: Delete images >90 days old (save $20/month)
3. **Right-size RabbitMQ**: t3.medium → t3.small (save $18/month)

**Total Savings**: $308/month (58%)

---

