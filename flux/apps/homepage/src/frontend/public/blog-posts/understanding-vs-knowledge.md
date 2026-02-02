---
title: "Understanding vs. Knowledge: What AI Reveals About Human Cognition"
slug: "understanding-vs-knowledge"
authors:
  - Bruno Lucena
tags:
  - AI
  - Philosophy
  - Computer Science
  - Understanding
  - Cognitive Science
  - Observability
publishedDate: 2025-12-17
updatedDate: 2025-12-18
readTime: "12 min read"
---

**Central Thesis**: Large language models amplify and reveal patterns already present in data, though they occasionally produce genuine novelty through high-dimensional interpolation. The theoretical computational limits that constrain formal systems (Gödel, Turing) operate at a different scale than practical understanding, which relies on approximate, probabilistic computation.

There's a question that keeps coming up in AI development: If understanding is fundamentally computational, what are its limits, and do these limits matter for practical AI systems?

I adopt a **computationalist** position: mental states can be realized in computational systems. However, I distinguish between **formal deductive computation** (where Gödel and Turing's limits apply) and **approximate statistical computation** (where modern AI operates).

For this analysis:
- **Knowledge**: Possession of information (knowing *that*)
- **Understanding**: Comprehension of causal relationships (knowing *why* and *how*)

## What Large Language Models Actually Compute

Modern LLMs are autoregressive transformer architectures that learn statistical distributions through self-supervised learning. They:

1. Learn compressed representations in high-dimensional embedding spaces
2. Use self-attention to compute contextual relationships
3. Perform next-token prediction through probabilistic inference
4. Generate outputs by sampling from learned distributions

We understand the mathematics perfectly: gradient descent, backpropagation, softmax attention. What remains partially opaque is **mechanistic interpretability**—understanding which circuits emerge during training.

Key distinction: LLMs operate through **continuous optimization in high-dimensional spaces**, not discrete symbolic manipulation. This matters when evaluating whether theoretical limits apply.

## Theoretical Limits and Their Relevance

Two famous results are often invoked when discussing AI limitations:

**Gödel's Incompleteness Theorems (1931)**: In any consistent formal system powerful enough to encode arithmetic, there exist true statements that cannot be proven within that system.

**The Halting Problem (Turing, 1936)**: No general algorithm can determine whether arbitrary programs will halt or run forever.

These establish fundamental limits on **formal deductive systems** and **decision procedures**. But do they constrain modern AI?

Here's the thing: LLMs aren't formal proof systems or halting-problem solvers. They're **statistical models** performing **approximate inference** over learned patterns.

Think of fluid dynamics: theoretical limits exist (Navier-Stokes equations), but engineers model turbulence without exact solutions. Similarly, LLMs operate where:

- **No formal completeness required**: They approximate through pattern matching, not theorem proving
- **Halting guarantees unnecessary**: They use bounded computation (max tokens, timeouts)
- **"Good enough" suffices**: Probabilistic confidence replaces logical certainty

LLMs have failure modes (hallucination, context limitations), but the constraints come from their architecture, not from Gödel and Turing's theorems.

## Three Scales of Understanding

To understand whether theoretical limits constrain AI, I distinguish three computational scales:

**Scale 1: Formal/Symbolic Computation**
- Discrete symbols with explicit rules
- Examples: Theorem provers, expert systems
- Governed by formal logic (Gödel applies here)
- Strength: Verifiable, guarantees | Weakness: Brittle

**Scale 2: Approximate/Heuristic Computation**
- Optimization over continuous spaces
- Examples: Neural networks, LLMs
- Governed by statistical learning theory
- Strength: Handles ambiguity, scales | Weakness: Opaque, probabilistic

**Scale 3: Embodied/Contextual Computation**
- Grounded in physical/social interaction
- Examples: Human cognition, robotics
- Governed by evolutionary/ecological constraints
- Strength: Robust, adaptive | Weakness: Slow, resource-intensive

Modern AI operates primarily at Scale 2. The theoretical limits of Scale 1 don't directly constrain statistical learning systems, which face different limitations: sample complexity, approximation error, generalization bounds.

## Understanding Through Approximate Computation

Human understanding—and increasingly, AI understanding—operates through mechanisms that are computational but not formally deductive:

- **Pattern Recognition**: Identifying structural similarities without explicit logical rules
- **Analogical Reasoning**: Mapping relationships between domains via structural alignment
- **Causal Modeling**: Building probabilistic models of cause-effect relationships
- **Contextual Interpretation**: Adjusting understanding based on situation and goals

These mechanisms are **approximate**: they work "well enough" for practical purposes without requiring formal proof. This is Scale 2 understanding.

LLMs excel at this kind of computation. They recognize patterns, draw analogies, estimate probabilities, and adjust outputs to context. What they lack (so far) is robust grounding in physical causality and embodied experience (Scale 3).

## The Amplification Hypothesis

My central claim: **AI primarily amplifies and reveals patterns already present in training data and problem structures.**

Consider code review: An LLM trained on GitHub can identify anti-patterns because those patterns already existed in its training corpus—it makes them explicit and immediately available. The model didn't invent these patterns; it learned to recognize and articulate them.

This explains both AI's power and its limits:
- **Power**: Accelerates pattern recognition across vast datasets
- **Limit**: Cannot create fundamentally novel patterns outside its training distribution

Occasionally, LLMs produce something that looks like genuine novelty through **high-dimensional interpolation**—combining learned patterns in ways humans wouldn't have tried. But even this "novelty" is recombination of existing patterns, not creation ex nihilo.

## Case Study: AI-Amplified Problem Discovery

To test this theory empirically, I analyzed a production system load balancing issue using AI-assisted observability. The question: Does AI create insights, or does it amplify discovery of existing patterns?

### The Problem

My homelab runs `demo-notifi`, a Knative-based event processing system with 3 parser deployments autoscaling across a Kubernetes cluster. Under load testing, I wanted to verify workload distribution.

Initial analysis showed **perfect event distribution** (0.56% coefficient of variation across parsers). But when I asked AI to analyze pod-level metrics, it revealed a catastrophic imbalance:

**Traffic Distribution**: 8,815,487:1 max/min ratio (88% of traffic to 3 pods out of 23)  
**CPU Distribution**: 156% coefficient of variation (one pod at 18%, others at 1-2%)

This wasn't a theoretical problem—it was a **production performance disaster** hiding in plain sight.

![Before/After Comparison](./graphs/01_before_after_comparison.png?v=0.1.19)

**Figure 1**: The dramatic improvement across all metrics after applying fixes. Traffic ratio improved 57,275x (from 8.8M:1 to 154:1).

### Root Cause Analysis

The AI didn't invent the problem—it **revealed a pattern** in 174,000 Prometheus data points that would have taken hours to discover manually. The root causes:

1. **HTTP/2 Connection Pooling**: Clients reuse TCP connections, bypassing Kubernetes round-robin
2. **Knative Revision Routing**: Route-level load balancing before pod-level distribution
3. **Pod Readiness Races**: First-ready pods get disproportionate initial traffic
4. **No Topology Constraints**: Random pod placement allowed hotspots
5. **Service Mesh Affinity**: Connection-level stickiness amplified imbalance
6. **Autoscaler Lag**: HPA sees aggregates, misses per-pod saturation

### The Solution

Six architectural fixes:

1. **Topology Spread Constraints**: `maxSkew=1` across nodes/zones
2. **Pod Anti-Affinity**: Prefer different hosts for same-parser pods
3. **Knative Rollout Duration**: `0s` to eliminate readiness races
4. **CPU Limits**: Resource constraints for fair scheduling
5. **Knative ConfigMaps**: Tuned `container-concurrency-target`, `stable-window`
6. **HTTP/1.1 Preference**: Stream timeouts to break connection pooling

Implementation took 30 minutes. Result:

**Traffic**: 8,815,487:1 → **154:1** (57,275x improvement)  
**Pod Distribution**: 67.9% variance → **13.5% CV** (excellent)  
**Node Balance**: 1.29:1 max/min (near-perfect for 23 pods)

![Pod Traffic Distribution](./graphs/02_pod_traffic_distribution.png?v=0.1.19)

**Figure 2**: Per-pod network traffic showing all 12 active pods under load. While still showing some imbalance (154:1), this is dramatically better than the previous 8.8M:1 catastrophic imbalance.

![Node Distribution](./graphs/03_node_distribution.png?v=0.1.19)

**Figure 3**: Pod distribution across Kubernetes nodes showing 9:7:7 distribution with 13.5% coefficient of variation (Excellent).

### What This Proves

**The AI Contribution:**
- Parallel queries across 58 pods in 30 seconds (vs 2+ hours manually)
- Statistical analysis (max/min ratios, CV, percentiles) instantly
- Pattern recognition (identified "long tail" distribution anti-pattern)
- Root cause synthesis from architectural knowledge

**The Human Contribution:**
- Context: "We're load testing, expect high traffic"
- Judgment: "88% concentration is unacceptable"
- Prioritization: "Fix topology first, then tune autoscaling"
- Validation: Debugging, iteration, kubectl expertise

**Key Insight**: The 8.8M:1 imbalance was always there in the metrics. The AI just made it visible at human timescales.

This is **Scale 2 engineering**: approximate, heuristic problem-solving that guides toward solutions without requiring formal proof. The AI accelerated discovery by 10x, but the human engineer drove the process.

![Prediction Accuracy](./graphs/04_prediction_accuracy.png?v=0.1.19)

**Figure 4**: Comparison of predictions against actual results. The topology spread constraints exceeded optimistic predictions (13.5% actual vs 15% predicted).

![Improvement Timeline](./graphs/05_improvement_timeline.png?v=0.1.19)

**Figure 5**: Complete timeline from problem discovery to validated solution. Total time: 3 hours from discovery to validation, resulting in a 57,000x improvement in traffic distribution.

## Three Paradigms in Practice

This case study demonstrates how the three scales operate in real engineering:

**Scale 1 (Formal)**: Kubernetes scheduler guarantees `maxSkew` topology constraints—provably correct  
**Scale 2 (Heuristic)**: AI pattern recognition + statistical analysis—fast, approximate, "good enough"  
**Scale 3 (Embodied)**: Human judgment in production context—adaptation, trade-offs, iteration

The solution required all three: formal guarantees (topology spread), heuristic discovery (AI metrics analysis), and contextual judgment (human prioritization).

## Implications and Limitations

### What AI Excels At

- Pattern detection across massive datasets (174K points in 30 seconds)
- Statistical computation (CV, ratios, percentiles)
- Architectural synthesis (combining Kubernetes, Knative, service mesh knowledge)
- Code generation (topology constraints, ConfigMaps)

### What AI Cannot Replace

- Operational context (build systems might be flaky)
- Debugging skills (why did validation take 90 minutes?)
- Iteration patience (rebuild → test → measure → adjust)
- Production intuition (when to cut losses and try differently)

### The Symbol Grounding Problem

A philosophical note: Does querying real Prometheus metrics constitute a form of "grounding" for AI? The metrics represent actual system state, not just linguistic abstractions. This suggests a path toward more grounded AI: connecting models to real-world observability systems.

## Conclusions

**1. Understanding operates at multiple computational scales**

Formal mathematical understanding (Scale 1) is constrained by Gödel and Turing. Engineering understanding (Scale 2) and everyday understanding (Scale 3) operate through approximate, heuristic mechanisms where these limits rarely bind.

**2. AI amplifies pattern recognition, not creation**

LLMs primarily interpolate within their training distribution, making implicit patterns explicit. They excel at amplifying linguistic and pattern-recognition capabilities, but face limitations in robust causal reasoning and embodied cognition.

**3. Human-AI symbiosis is the path forward**

Neither human nor AI alone could achieve the efficiency demonstrated here. Together, they completed 3 hours of productive work that would have taken 12+ hours solo.

### Practical Implications

For AI practitioners:
- **Design for amplification**: Build systems that augment human pattern recognition
- **Embrace approximation**: "Good enough" at Scale 2 beats perfect at Scale 1
- **Maintain human oversight**: Domain expertise remains essential
- **Focus on robustness**: The real challenge is practical reliability, not theoretical limits

### Final Reflection

The question I began with—whether understanding has computational limits—has a nuanced answer: **Yes, but not the limits we typically imagine.**

The constraints that matter for AI aren't Gödel's theorems or the halting problem. They're the practical challenges of generalization, robustness, and alignment.

Understanding this distinction clarifies what AI can and cannot do. It can amplify our pattern recognition, accelerate our search through solution spaces, and reveal connections we'd miss alone. It can't replace human judgment, create understanding from nothing, or transcend the patterns in its training data.

The need to "babysit" AI systems isn't a temporary limitation—it reflects the complementary nature of human and machine cognition. We provide context, judgment, and goals. AI provides scale, speed, and pattern detection.

Maybe the deeper insight is that AI doesn't just amplify what we already have—it reveals the computational nature of human understanding itself.

---

## Key Takeaways

**For Engineers**
- Multi-layer observability is essential: event distribution was perfect (1.14% variance) while pod distribution was catastrophic (8.8M:1)
- Load balancing ≠ workload distribution: connection pooling breaks round-robin assumptions
- Topology matters: spread constraints and anti-affinity are critical for production performance

**For AI/ML Practitioners**
- AI amplifies, doesn't create: the 8.8M:1 imbalance existed in metrics, AI made it discoverable in 30 seconds vs 3 hours
- Pattern recognition at scale: AI excels at correlating 174K data points—tedious and error-prone for humans
- Human judgment is irreplaceable: AI suggested fixes, humans prioritized and adapted

**For Organizations**
- Observability powers AI: without Prometheus metrics, this analysis would be impossible
- AI ROI is real: 3 hours (with AI) vs. 12+ hours (without) = 4x productivity multiplier
- Invest in infrastructure: Grafana + Prometheus enabled this entire analysis

---

### References

[^1]: Putnam, H. (1967). "Psychological Predicates." *Art, Mind, and Religion*, University of Pittsburgh Press.

[^2]: Chalmers, D. (2011). "A Computational Foundation for the Study of Cognition." *Journal of Cognitive Science*, 12(4), 325-359.

[^3]: Zagzebski, L. (2001). "Recovering Understanding." *Knowledge, Truth, and Duty*, Oxford University Press.

[^4]: Vaswani, A., et al. (2017). "Attention Is All You Need." *NeurIPS*, 30.

[^5]: Brown, T., et al. (2020). "Language Models are Few-Shot Learners." *NeurIPS*, 33, 1877-1901.

[^6]: Olah, C., et al. (2020). "Zoom In: An Introduction to Circuits." *Distill*, 5(3), e00024.001.

[^7]: Gödel, K. (1931). "Über formal unentscheidbare Sätze." *Monatshefte für Mathematik*, 38, 173-198.

[^8]: Turing, A. M. (1936). "On Computable Numbers." *Proc. London Mathematical Society*, 42(2), 230-265.

[^9]: Ji, Z., et al. (2023). "Survey of Hallucination in Natural Language Generation." *ACM Computing Surveys*, 55(12), 1-38.

[^10]: Vapnik, V. (1998). *Statistical Learning Theory*. Wiley-Interscience.

[^11]: Shalev-Shwartz, S., & Ben-David, S. (2014). *Understanding Machine Learning*. Cambridge University Press.

[^12]: Fodor, J. A., & Pylyshyn, Z. W. (1988). "Connectionism and Cognitive Architecture." *Cognition*, 28(1-2), 3-71.

[^13]: Deletang, G., et al. (2023). "Neural Networks and the Chomsky Hierarchy." *ICLR*.

---

## Appendix: Technical Implementation Details

For those interested in the complete technical implementation, the full details are available in the [project repository](https://github.com/brunovlucena/homelab). Key implementation files:

- `knative-lambda-operator/src/operator/controllers/lambdaagent_controller.go`: Topology spread and anti-affinity implementation
- `knative-lambda-operator/k8s/base/knative-serving-config.yaml`: Knative ConfigMaps for load balancing tuning
- `demos/demo-notifi/`: Complete load testing infrastructure

The fix implementation took 30 minutes, deployment 20 minutes, validation 90 minutes. Total time from discovery to confirmed fix: 3 hours.

### Validation Metrics

Complete before/after metrics:

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Traffic Ratio (Max/Min) | 8,815,487:1 | 154:1 | 57,275x |
| Top-3 Pod Concentration | 88% | 75.2% | 12.8% reduction |
| Pod Distribution CV | 67.9% | 13.5% | 5x more even |
| Node Distribution | Unbalanced | 9:7:7 (1.29:1) | Near-optimal |

All metrics measured via direct Prometheus queries using PromQL across a 5-minute load test window with 23 autoscaled pods handling sustained traffic.
