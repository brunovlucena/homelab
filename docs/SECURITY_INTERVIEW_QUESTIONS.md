# Security Interview Questions for Homelab Projects

> **Based on**: [How Big Tech Interviews Security Engineers in 2026](https://secengweekly.substack.com/p/how-big-tech-interviews-security)  
> **Projects**: Homelab Infrastructure, Agents WhatsApp Rust, Knative Lambda Operator, Agent Medical, Agent Restaurant  
> **Date**: January 2025

---

## Overview

This document contains security interview questions tailored to specific projects in the homelab infrastructure. The questions follow the format from the referenced article, focusing on **judgment, tradeoff thinking, and real-world security maturity** rather than trivia.

Questions are organized into 6 categories:
1. Security Fundamentals & Core Concepts
2. Cloud Security & Infrastructure (Kubernetes/Knative)
3. Application Security & Secure SDLC
4. Threat Modeling & System Design
5. Detection Engineering & Incident Response
6. Project-Specific Security Scenarios

---

## A) Security Fundamentals & Core Concepts (1-10)

### 1) How does the CIA Triad apply to the agents-whatsapp-rust messaging platform?

**Focus**: Foundational security thinking in real systems  
**Core Idea**: Security is about tradeoffs, not absolutes

**Context**: The messaging platform requires E2EE (confidentiality), zero disconnections (availability), and exactly-once delivery (integrity).

**Strong Answers Cover**:
- Tradeoffs between E2EE complexity and message delivery latency
- Availability vs confidentiality when handling connection migrations
- Why message ordering (integrity) might require accepting some availability risk
- Business impact of message loss vs message exposure
- Long-term trust implications of security vs UX tradeoffs

---

### 2) Authentication vs Authorization in the knative-lambda-operator. Explain with examples.

**Focus**: Identity clarity in Kubernetes operators  
**Core Idea**: Most security bugs are access bugs

**Context**: The operator manages LambdaAgent CRDs, creates pods with service accounts, and handles RBAC.

**Strong Answers Cover**:
- Clear boundary between operator identity (who deploys) and function identity (who runs)
- Real production failures: VULN-003 (SA token theft), VULN-013 (receiver mode escalation)
- How RBAC mistakes scale silently across namespaces
- Developer ergonomics (easy deployments) vs safety (least privilege)
- Why `automountServiceAccountToken: false` matters for lambda functions

---

### 3) Explain how XSS, CSRF, and SSRF apply to the agent-restaurant web interface.

**Focus**: Web security reasoning in multi-agent systems  
**Core Idea**: Attacks exploit misplaced trust

**Context**: The restaurant agent has a Next.js web interface that communicates with backend agents via CloudEvents.

**Strong Answers Cover**:
- Threat models: where trust breaks between web UI → API → agents
- SSRF risks when agents make external API calls (menu APIs, payment systems)
- CSRF protection for state-changing operations (order creation, reservation booking)
- Prevention as design: input validation at API boundaries, not just UI
- Developer education: how to prevent prompt injection in agent interactions

---

### 4) What is SQL Injection and how does it relate to agent-medical's database access?

**Focus**: Secure coding fundamentals in HIPAA-compliant systems  
**Core Idea**: Input handling is architecture, not validation

**Context**: Agent-medical queries patient records from MongoDB/PostgreSQL with role-based access control.

**Strong Answers Cover**:
- Parameterization vs sanitization in medical record queries
- ORM false sense of safety (MongoDB ODM injection risks)
- Risk reduction at framework level (query builders, prepared statements)
- Long-term maintenance: why raw queries in medical systems are dangerous
- HIPAA implications of injection attacks (unauthorized PHI access)

---

### 5) What is a replay attack and how do you prevent it in CloudEvents-based agent communication?

**Focus**: Protocol reasoning in event-driven architecture  
**Core Idea**: Freshness matters as much as secrecy

**Context**: Agents communicate via CloudEvents through RabbitMQ/Knative Eventing.

**Strong Answers Cover**:
- Nonces, timestamps, token expiry in CloudEvents
- Tradeoffs with distributed systems (clock skew, event ordering)
- Failure modes: what happens when events are replayed after agent state changes
- Operational risks: idempotency keys vs replay prevention
- How message-processor agent handles duplicate events

---

### 6) Hashing vs Encryption vs Encoding in agent-medical's HIPAA compliance

**Focus**: Crypto fundamentals in regulated environments  
**Core Idea**: Wrong primitive equals broken security

**Context**: Agent-medical must encrypt PHI at rest, hash audit logs, and encode data for transmission.

**Strong Answers Cover**:
- Use cases: encryption for PHI, hashing for audit logs, encoding for JSON/Base64
- Irreversibility vs confidentiality: why you can't "decrypt" a hash
- Compliance expectations: HIPAA requires encryption, not hashing, for PHI
- Developer misuse patterns: encrypting passwords (wrong), hashing PHI (wrong)
- Key management: KMS integration for encryption keys

---

### 7) Symmetric vs Asymmetric encryption in agents-whatsapp-rust E2EE

**Focus**: Practical crypto usage in messaging systems  
**Core Idea**: Trust and performance tradeoffs

**Context**: The platform implements E2EE using Double Ratchet Protocol (Signal Protocol).

**Strong Answers Cover**:
- Key distribution challenges: how clients exchange keys without server access
- Hybrid models: asymmetric for key exchange, symmetric for message encryption
- Scalability concerns: performance of asymmetric crypto at message scale
- Real-world misuse: why you can't use RSA for every message
- Forward secrecy: automatic key rotation in Double Ratchet

---

### 8) How does TLS work end-to-end in the knative-lambda-operator's build pipeline?

**Focus**: Secure communication depth in CI/CD  
**Core Idea**: Validation failures break everything

**Context**: The operator builds images using Kaniko, pulls from registries, and deploys to Kubernetes.

**Strong Answers Cover**:
- Cert chains and trust anchors: registry certificate validation
- MITM risks: what happens if build jobs connect to malicious registries
- Misconfiguration risks: insecure registry settings, self-signed certs
- Long-term platform trust: why certificate pinning matters in build pipelines
- How to detect and prevent registry MITM attacks

---

### 9) What is Least Privilege and how do you enforce it in knative-lambda-operator RBAC?

**Focus**: Access design in Kubernetes operators  
**Core Idea**: Permissions decay silently

**Context**: The operator has cluster-wide permissions but functions should have minimal access.

**Strong Answers Cover**:
- Guardrails over reviews: automated RBAC generation vs manual configuration
- Automation vs manual enforcement: operator creates SAs, but who validates?
- Developer trust balance: easy deployments vs security controls
- Long-term blast radius reduction: why VULN-003 (SA token theft) was critical
- How security-rbac.yaml implements least privilege for different function types

---

### 10) What is Defense in Depth in the homelab infrastructure?

**Focus**: Security architecture across multiple systems  
**Core Idea**: Assume failure, not perfection

**Context**: Homelab runs multiple agents, operators, and services with varying security requirements.

**Strong Answers Cover**:
- Layered controls: network policies, RBAC, encryption, monitoring
- Cost vs protection tradeoffs: where to invest security resources
- Failure containment: how to prevent one compromised agent from affecting others
- Strategic resilience: recovery planning for different attack scenarios
- How agent-medical's HIPAA requirements differ from agent-restaurant's needs

---

## B) Cloud Security & Infrastructure (11-19)

### 11) Common IAM/RBAC misconfigurations you've seen in knative-lambda-operator

**Focus**: Real-world experience with Kubernetes RBAC  
**Core Idea**: Identity is the cloud perimeter

**Context**: The operator has had multiple RBAC vulnerabilities (VULN-003, VULN-004, VULN-013, BLUE-006).

**Strong Answers Cover**:
- Over-permissioned roles: why receiver mode had cluster-wide access (VULN-013)
- Lateral movement risk: how SA token theft enables privilege escalation
- Monitoring gaps: how to detect RBAC misconfigurations before exploitation
- Long-term permission creep: why security-rbac.yaml was created
- How to audit and reduce RBAC permissions over time

---

### 12) How do you design RBAC for multi-agent microservices (agent-restaurant, agent-medical)?

**Focus**: Service identity in event-driven architecture  
**Core Idea**: Humans shouldn't be in the auth path

**Context**: Multiple agents communicate via CloudEvents, each with different security requirements.

**Strong Answers Cover**:
- Service-to-service auth: how agents authenticate CloudEvents
- Short-lived credentials: token rotation for agent service accounts
- Operational complexity: managing RBAC across multiple namespaces
- Scalability impact: how RBAC scales with number of agents
- Cross-agent communication: how waiter-pierre talks to chef agent securely

---

### 13) How do you secure cloud networking in the homelab Kubernetes cluster?

**Focus**: Network isolation in multi-tenant environments  
**Core Idea**: Flat networks amplify damage

**Context**: Multiple agents, operators, and services share the same cluster.

**Strong Answers Cover**:
- Segmentation strategies: network policies for agent isolation
- Zero trust assumptions: why agents shouldn't trust each other by default
- Cost vs complexity: network policy management overhead
- Failure containment: preventing lateral movement between agents
- How to isolate agent-medical (HIPAA) from agent-restaurant (public-facing)

---

### 14) Securing object storage (MinIO) for agent-medical's PHI storage

**Focus**: Data exposure prevention in regulated environments  
**Core Idea**: Defaults are dangerous

**Context**: Agent-medical stores encrypted PHI in MinIO buckets.

**Strong Answers Cover**:
- Access policies: bucket policies, IAM roles for agent access
- Logging and alerting: detecting unauthorized bucket access
- Public exposure risks: why MinIO buckets must never be public
- Long-term data trust: encryption at rest, key rotation
- HIPAA compliance: audit logs for all PHI access

---

### 15) Secrets management in Kubernetes for knative-lambda-operator

**Focus**: Sensitive data handling in operators  
**Core Idea**: Secrets leak through convenience

**Context**: The operator needs registry credentials, API keys, and database passwords.

**Strong Answers Cover**:
- Rotation strategies: how to rotate secrets without downtime
- Developer UX: Sealed Secrets vs External Secrets Operator
- Automation vs risk: automated secret rotation vs manual processes
- Incident response readiness: how to revoke compromised secrets quickly
- VULN-004 implications: why cluster-wide secret access is dangerous

---

### 16) What is shared responsibility in Kubernetes security (homelab cluster)?

**Focus**: Accountability clarity in self-managed infrastructure  
**Core Idea**: Assumptions cause breaches

**Context**: Homelab runs on self-managed Kubernetes (not managed cloud).

**Strong Answers Cover**:
- Provider vs customer boundaries: what Kubernetes provides vs what you secure
- Misplaced trust: assuming Kubernetes defaults are secure
- Audit readiness: how to prove compliance in self-managed clusters
- Long-term ownership clarity: who owns security for operators vs applications
- How to handle security updates for control plane vs workloads

---

### 17) Cloud logging strategy for security in homelab infrastructure

**Focus**: Visibility across multiple systems  
**Core Idea**: You cannot defend what you cannot see

**Context**: Multiple agents, operators, and services generate logs (Loki, Prometheus, Tempo).

**Strong Answers Cover**:
- Control plane vs data plane: what to log for security
- Cost tradeoffs: log retention for compliance (HIPAA) vs storage costs
- Signal quality: how to reduce noise in security logs
- Detection maturity: what logs enable threat detection
- How to correlate logs across agents for security incidents

---

### 18) How do you think about cloud threat modeling for knative-lambda-operator?

**Focus**: Attacker mindset for Kubernetes operators  
**Core Idea**: Assume breach

**Context**: The operator has had multiple critical vulnerabilities (BLUE-001 SSRF, BLUE-002 template injection, VULN-003 code execution).

**Strong Answers Cover**:
- Identity abuse: how attackers exploit RBAC misconfigurations
- Lateral movement: how compromised functions access other resources
- Persistence: how attackers maintain access after initial compromise
- Recovery planning: how to detect and remediate operator compromises
- How the security assessment report (SECURITY_ASSESSMENT_REPORT.md) informs threat modeling

---

### 19) Kubernetes security vs traditional cloud security differences

**Focus**: Mental model shift for containerized workloads  
**Core Idea**: Speed increases risk

**Context**: Homelab uses Kubernetes operators, Knative, and serverless functions.

**Strong Answers Cover**:
- Automation impact: how operators increase attack surface
- Shared tooling risk: how one compromised operator affects all functions
- Control loss: how serverless reduces visibility
- Long-term governance: how to maintain security as infrastructure scales
- Why traditional IAM doesn't map directly to Kubernetes RBAC

---

## C) Application Security & Secure SDLC (20-28)

### 20) How do you secure APIs in agents-whatsapp-rust's microservices architecture?

**Focus**: AppSec design thinking in messaging platforms  
**Core Idea**: APIs are the most attacked surface today

**Context**: The platform has messaging-service, user-service, agent-gateway, and message-processor.

**Strong Answers Cover**:
- AuthN vs AuthZ at the API layer: JWT validation, WebSocket auth
- Rate limiting, abuse detection, and input validation
- Tradeoffs between security and developer velocity
- Long-term API versioning and backward compatibility risks
- How to prevent API abuse in a horizontally scalable system

---

### 21) What are the most common application security mistakes in agent implementations?

**Focus**: Pattern recognition across multiple agents  
**Core Idea**: Most bugs repeat, only contexts change

**Context**: Multiple agents (medical, restaurant, WhatsApp) share similar patterns.

**Strong Answers Cover**:
- Broken access control: how agents access resources they shouldn't
- Secrets exposure: environment variables, configmaps, logs
- Insecure defaults: why agents start with too many permissions
- Why teams repeat the same mistakes at scale
- How to prevent prompt injection in LLM-based agents

---

### 22) How do you integrate security into CI/CD pipelines for knative-lambda-operator?

**Focus**: Shift-left maturity in operator development  
**Core Idea**: Security that blocks pipelines gets bypassed

**Context**: The operator builds images, deploys functions, and manages RBAC.

**Strong Answers Cover**:
- SAST vs DAST vs dependency scanning tradeoffs
- False positive management: how to avoid alert fatigue
- Developer trust and adoption: security that helps, not hinders
- Measuring effectiveness over time: how to know security is working
- How to prevent BLUE-001 (SSRF) and BLUE-002 (template injection) in CI/CD

---

### 23) How do you think about dependency and supply chain security for Rust agents?

**Focus**: Modern attack vectors in compiled languages  
**Core Idea**: Your code is only as safe as your weakest dependency

**Context**: agents-whatsapp-rust uses Rust with Cargo dependencies.

**Strong Answers Cover**:
- SBOMs and dependency visibility: how to track Rust crate dependencies
- Risk-based patching: when to update dependencies vs accept risk
- Balancing velocity with exposure: how fast to update crates
- Long-term ecosystem risk: how to monitor Rust security advisories
- How to prevent supply chain attacks in Cargo registries

---

### 24) How do you handle secrets in application code for agent-medical?

**Focus**: Secure engineering discipline in HIPAA systems  
**Core Idea**: Secrets leak through convenience

**Context**: Agent-medical needs database credentials, API keys, and encryption keys.

**Strong Answers Cover**:
- Environment-based secret injection: Kubernetes Secrets, not hardcoded
- Rotation and revocation: how to rotate secrets without downtime
- Developer ergonomics: how to make secure practices easy
- Incident response readiness: how to revoke secrets quickly
- HIPAA implications of secret exposure

---

### 25) What is your approach to secure authentication flows in agents-whatsapp-rust?

**Focus**: Identity-first security in messaging platforms  
**Core Idea**: Auth bugs are catastrophic bugs

**Context**: The platform uses JWT tokens, WebSocket auth, and refresh tokens.

**Strong Answers Cover**:
- Token lifecycle management: generation, validation, rotation, revocation
- Session handling risks: how to prevent session hijacking
- Tradeoffs between UX and security: remember me, device trust
- Long-term identity trust: how to maintain security as users scale
- How to handle authentication in WebSocket connections

---

### 26) How do you test security controls in agent implementations?

**Focus**: Validation mindset for AI agents  
**Core Idea**: Controls that aren't tested don't exist

**Context**: Multiple agents with different security requirements (HIPAA, public-facing, internal).

**Strong Answers Cover**:
- Automated vs manual testing: k6 tests, penetration testing
- Regression prevention: how to prevent security regressions
- Coverage gaps: what security tests are missing
- Metrics for confidence: how to measure security test effectiveness
- How to test RBAC, encryption, and audit logging

---

### 27) How do you educate developers about security without slowing them down?

**Focus**: Influence and communication in security  
**Core Idea**: Security scales through people, not policies

**Context**: Multiple teams work on different agents with varying security maturity.

**Strong Answers Cover**:
- Just-in-time education: security guidance when developers need it
- Secure defaults: how to make secure choices the easy choices
- Feedback loops: how to learn from security incidents
- Long-term culture building: how to make security part of development
- How to prevent developers from bypassing security controls

---

### 28) How do you balance security requirements with product deadlines?

**Focus**: Judgment in security prioritization  
**Core Idea**: Security is prioritization, not absolutism

**Context**: Agents have different security requirements (HIPAA vs public-facing vs internal).

**Strong Answers Cover**:
- Risk acceptance vs mitigation: when to accept risk vs fix it
- Business alignment: how to explain security tradeoffs to stakeholders
- Documented tradeoffs: how to track security debt
- Trust with leadership: how to build credibility for security decisions
- How to prioritize security fixes (P0, P1, P2) based on risk

---

## D) Threat Modeling & System Design (29-35)

### 29) How do you approach threat modeling for a new agent (e.g., agent-medical)?

**Focus**: Structured thinking for HIPAA-compliant systems  
**Core Idea**: Ask before answering

**Context**: Agent-medical handles PHI, requires HIPAA compliance, and integrates with multiple systems.

**Strong Answers Cover**:
- Assets, actors, trust boundaries: PHI, doctors, patients, systems
- Clarifying assumptions: what do we trust, what don't we trust
- Threat prioritisation: which threats matter most for HIPAA
- Long-term design impact: how threat model informs architecture
- How to model threats for LLM-based agents (prompt injection, data leakage)

---

### 30) Threat model the agents-whatsapp-rust messaging platform

**Focus**: Applied reasoning for real-world systems  
**Core Idea**: Practical beats theoretical

**Context**: The platform handles E2EE messages, WebSocket connections, and AI agent routing.

**Strong Answers Cover**:
- Abuse cases: message spam, connection exhaustion, agent abuse
- Failure modes: what happens when encryption fails, connections drop
- Defense tradeoffs: security vs performance vs UX
- Business risk alignment: how threats affect user trust
- How to model threats for E2EE systems (key compromise, MITM, replay)

---

### 31) How do you prioritise security risks across multiple agents?

**Focus**: Risk management in multi-agent systems  
**Core Idea**: Not all risks deserve equal attention

**Context**: Homelab runs multiple agents with different risk profiles (HIPAA, public-facing, internal).

**Strong Answers Cover**:
- Impact vs likelihood: how to score risks across agents
- User trust implications: how security affects different user bases
- Cost of mitigation: where to invest security resources
- Long-term exposure reduction: how to reduce risk over time
- How to prioritize: agent-medical (HIPAA) vs agent-restaurant (public)

---

### 32) How do you think about blast radius in knative-lambda-operator?

**Focus**: Containment strategy for Kubernetes operators  
**Core Idea**: Fail safely, not perfectly

**Context**: The operator has cluster-wide permissions and manages functions across namespaces.

**Strong Answers Cover**:
- Isolation techniques: how to contain operator compromises
- Least privilege: how security-rbac.yaml reduces blast radius
- Recovery time objectives: how quickly to detect and remediate
- Organisational resilience: how to prevent single points of failure
- How VULN-003 (code execution) could have led to cluster compromise

---

### 33) How do you design systems assuming breach for agent-medical?

**Focus**: Defensive realism for HIPAA systems  
**Core Idea**: Breach is inevitable

**Context**: Agent-medical handles PHI and must comply with HIPAA breach notification requirements.

**Strong Answers Cover**:
- Lateral movement prevention: how to prevent agent compromises from spreading
- Detection-first mindset: how to detect breaches quickly
- Recovery planning: how to respond to PHI breaches
- Long-term survivability: how to maintain operations during incidents
- How audit logging enables breach detection and response

---

### 34) How do you evaluate new security tools or frameworks for homelab?

**Focus**: Tool judgment for infrastructure security  
**Core Idea**: Tools don't fix bad thinking

**Context**: Homelab uses multiple security tools (Trivy, Sealed Secrets, Network Policies, etc.).

**Strong Answers Cover**:
- Risk reduction vs complexity: when tools help vs hinder
- Integration cost: how to evaluate tool adoption overhead
- Developer trust: how to ensure tools are used, not bypassed
- Long-term maintainability: how to avoid tool sprawl
- How to evaluate: Trivy (scanning) vs Falco (runtime security) vs OPA (policy)

---

### 35) How do you communicate risk to non-technical stakeholders?

**Focus**: Leadership communication for security  
**Core Idea**: Security fails when it can't be explained

**Context**: Security decisions affect product timelines, costs, and user experience.

**Strong Answers Cover**:
- Business impact framing: how to explain technical risks in business terms
- Clear tradeoffs: security vs speed vs cost
- Metrics that matter: how to measure security effectiveness
- Executive trust: how to build credibility for security decisions
- How to explain HIPAA compliance requirements to non-technical stakeholders

---

## E) Detection Engineering & Incident Response (36-43)

### 36) How do you design a detection strategy for knative-lambda-operator?

**Focus**: Visibility for Kubernetes operators  
**Core Idea**: Detection beats prevention alone

**Context**: The operator has had multiple critical vulnerabilities that could lead to cluster compromise.

**Strong Answers Cover**:
- Signal-to-noise ratio: how to detect real attacks vs false positives
- Coverage gaps: what attacks are we missing
- Cost tradeoffs: log storage, alert processing
- Continuous improvement: how to refine detections over time
- How to detect: BLUE-001 (SSRF), BLUE-002 (template injection), VULN-003 (code execution)

---

### 37) What logs are most important for security in agent-medical?

**Focus**: Observability for HIPAA compliance  
**Core Idea**: Logs are your memory

**Context**: Agent-medical must maintain audit logs for all PHI access (HIPAA requirement).

**Strong Answers Cover**:
- Identity events: who accessed what PHI, when
- Privilege changes: role assignments, permission modifications
- Control-plane activity: API calls, configuration changes
- Long-term forensic value: how logs enable incident investigation
- How to balance: log detail vs storage costs vs compliance requirements

---

### 38) How do you reduce alert fatigue in homelab security monitoring?

**Focus**: Operational maturity for security teams  
**Core Idea**: Burned-out teams miss real attacks

**Context**: Multiple agents, operators, and services generate security alerts.

**Strong Answers Cover**:
- Alert quality metrics: how to measure alert effectiveness
- Context enrichment: how to make alerts actionable
- Triage automation: how to prioritize alerts automatically
- Feedback loops: how to learn from false positives
- How to prevent: alert fatigue from RBAC misconfigurations, failed scans

---

### 39) How do you design incident response for a HIPAA breach in agent-medical?

**Focus**: Response planning for regulated environments  
**Core Idea**: Preparation prevents panic

**Context**: Agent-medical must notify patients and regulators within 60 days of a PHI breach.

**Strong Answers Cover**:
- Detection: how to identify PHI breaches quickly
- Containment: how to stop breaches from spreading
- Notification: how to meet HIPAA breach notification requirements
- Recovery: how to restore operations securely
- Long-term improvements: how to prevent future breaches

---

### 40) How do you handle security incidents in knative-lambda-operator?

**Focus**: Response for critical infrastructure  
**Core Idea**: Operators are high-value targets

**Context**: The operator has cluster-wide permissions and manages all Lambda functions.

**Strong Answers Cover**:
- Detection: how to detect operator compromises
- Containment: how to prevent cluster-wide compromise
- Recovery: how to restore operator functionality securely
- Communication: how to notify affected teams
- Post-incident: how to prevent similar incidents

---

### 41) How do you test incident response procedures?

**Focus**: Validation of response capabilities  
**Core Idea**: Untested procedures fail under pressure

**Context**: Multiple agents with different security requirements need tested response procedures.

**Strong Answers Cover**:
- Tabletop exercises: how to practice incident response
- Red team exercises: how to test detection and response
- Metrics: how to measure response effectiveness
- Continuous improvement: how to refine procedures over time
- How to test: HIPAA breach response, operator compromise, agent abuse

---

### 42) How do you balance security monitoring costs with coverage?

**Focus**: Resource management for security  
**Core Idea**: Perfect security is unaffordable

**Context**: Homelab has limited resources but needs comprehensive security monitoring.

**Strong Answers Cover**:
- Cost tradeoffs: log retention, alert processing, tool licensing
- Coverage prioritization: where to focus monitoring efforts
- Risk-based approach: how to allocate resources based on risk
- Long-term optimization: how to reduce costs while maintaining coverage
- How to prioritize: agent-medical (HIPAA) vs agent-restaurant (public)

---

### 43) How do you measure security program effectiveness?

**Focus**: Metrics for security maturity  
**Core Idea**: You can't improve what you don't measure

**Context**: Multiple agents, operators, and services need security metrics.

**Strong Answers Cover**:
- Leading vs lagging indicators: how to measure security before incidents
- Business alignment: how to measure security in business terms
- Continuous improvement: how to use metrics to improve security
- Long-term trends: how to track security maturity over time
- How to measure: vulnerability reduction, incident response time, compliance

---

## F) Project-Specific Security Scenarios (44-50)

### 44) How would you secure the agents-whatsapp-rust E2EE implementation?

**Focus**: Applied cryptography in messaging systems  
**Core Idea**: E2EE is hard to get right

**Context**: The platform implements E2EE using Double Ratchet Protocol, but server cannot decrypt.

**Strong Answers Cover**:
- Key exchange: how clients exchange keys securely
- Forward secrecy: how to rotate keys automatically
- Key backup: how to handle key loss without compromising security
- Server role: how to route messages without decrypting
- Long-term trust: how to maintain security as users scale

---

### 45) How would you prevent the knative-lambda-operator vulnerabilities (BLUE-001, BLUE-002, VULN-003)?

**Focus**: Remediation of known vulnerabilities  
**Core Idea**: Learn from past mistakes

**Context**: The operator has had multiple critical vulnerabilities documented in SECURITY_ASSESSMENT_REPORT.md.

**Strong Answers Cover**:
- BLUE-001 (SSRF): how to validate git source URLs
- BLUE-002 (template injection): how to escape handler fields
- VULN-003 (code execution): how to sandbox inline source execution
- Prevention: how to prevent similar vulnerabilities in the future
- Long-term security: how to build security into operator design

---

### 46) How would you design RBAC for agent-medical to meet HIPAA requirements?

**Focus**: Compliance-driven access control  
**Core Idea**: HIPAA requires specific access controls

**Context**: Agent-medical must implement role-based access control (doctor, nurse, patient, admin).

**Strong Answers Cover**:
- Role design: how to define roles that meet HIPAA requirements
- Permission model: how to implement least privilege for PHI access
- Audit logging: how to log all PHI access for compliance
- Enforcement: how to enforce RBAC at application and database levels
- Long-term compliance: how to maintain HIPAA compliance as system evolves

---

### 47) How would you secure agent-restaurant's multi-agent communication?

**Focus**: Security in event-driven multi-agent systems  
**Core Idea**: Agents must trust each other, but verify

**Context**: Agent-restaurant has multiple agents (waiter, chef, sommelier, host) communicating via CloudEvents.

**Strong Answers Cover**:
- Event authentication: how to verify CloudEvents come from trusted agents
- Event authorization: how to ensure agents only receive authorized events
- Event integrity: how to prevent event tampering
- Failure handling: how to handle security failures in event processing
- Long-term security: how to maintain security as agents are added

---

### 48) How would you implement secrets rotation for knative-lambda-operator without downtime?

**Focus**: Operational security for critical infrastructure  
**Core Idea**: Secrets must rotate, but systems must stay up

**Context**: The operator needs registry credentials, API keys, and database passwords that must rotate regularly.

**Strong Answers Cover**:
- Rotation strategy: how to rotate secrets without service interruption
- Automation: how to automate secret rotation
- Validation: how to verify secrets work after rotation
- Rollback: how to rollback if rotation fails
- Long-term maintenance: how to make secret rotation routine

---

### 49) How would you design network isolation for homelab agents?

**Focus**: Network security in multi-tenant Kubernetes  
**Core Idea**: Network policies prevent lateral movement

**Context**: Multiple agents share the same cluster but have different security requirements.

**Strong Answers Cover**:
- Policy design: how to define network policies for agents
- Isolation levels: how to isolate agent-medical (HIPAA) from other agents
- Enforcement: how to enforce network policies consistently
- Operational complexity: how to manage network policies at scale
- Long-term maintenance: how to update policies as agents evolve

---

### 50) How would you respond to a security incident in agent-medical that exposed PHI?

**Focus**: HIPAA breach response  
**Core Idea**: Breach response is time-sensitive and regulated

**Context**: Agent-medical must notify patients and regulators within 60 days of a PHI breach.

**Strong Answers Cover**:
- Detection: how to detect PHI exposure quickly
- Assessment: how to determine breach scope and impact
- Containment: how to stop breach and prevent further exposure
- Notification: how to meet HIPAA breach notification requirements
- Recovery: how to restore operations and prevent future breaches
- Long-term improvements: how to learn from incidents

---

## Interview Evaluation Criteria

### Strong Answers Demonstrate:

1. **Tradeoff Thinking**: Can reason about security vs performance vs UX vs cost
2. **Business Alignment**: Can explain security in business terms
3. **Practical Experience**: References real vulnerabilities, incidents, or systems
4. **Defense in Depth**: Considers multiple layers of security
5. **Long-term Thinking**: Considers maintenance, scalability, and evolution
6. **Risk-Based Approach**: Prioritizes based on impact and likelihood
7. **Communication**: Can explain complex security concepts clearly

### Red Flags:

1. **Absolutism**: "Always encrypt everything" without considering tradeoffs
2. **Theoretical Only**: No practical experience or real-world examples
3. **No Business Context**: Can't explain security in business terms
4. **Single-Layer Thinking**: Only considers one security control
5. **No Long-term View**: Doesn't consider maintenance or evolution
6. **Poor Communication**: Can't explain security concepts clearly

---

## Notes for Interviewers

- These questions are designed to evaluate **judgment and maturity**, not trivia
- Strong candidates will reference specific vulnerabilities, incidents, or systems
- Look for candidates who can reason about tradeoffs, not just recite best practices
- Project-specific questions (44-50) test ability to apply security knowledge to real systems
- Use follow-up questions to probe deeper: "What would you do if...?", "How would you handle...?"

---

## References

- [How Big Tech Interviews Security Engineers in 2026](https://secengweekly.substack.com/p/how-big-tech-interviews-security)
- [Knative Lambda Operator Security Assessment Report](../knative/reports/SECURITY_ASSESSMENT_REPORT.md)
- [Agent Medical Architecture](../architecture/agent-medical-records.md)
- [Agents WhatsApp Rust Architecture](../agents-whatsapp-rust/ARCHITECTURE.md)
