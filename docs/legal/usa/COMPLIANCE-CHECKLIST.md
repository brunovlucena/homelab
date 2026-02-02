# ✅ Legal Compliance Checklist

> **Document Version**: 1.0  
> **Last Updated**: December 11, 2025  
> **Author**: Bruno Lucena  
> **Purpose**: Pre-launch legal compliance verification

---

## Quick Reference: Critical Items

| Priority | Item | Status | Deadline |
|----------|------|--------|----------|
| 🔴 CRITICAL | Resolve Slither AGPL license | ⬜ TODO | Before launch |
| 🔴 CRITICAL | Add LICENSE file | ⬜ TODO | Before release |
| 🔴 CRITICAL | Create NOTICE file with attributions | ⬜ TODO | Before release |
| 🔴 CRITICAL | Form business entity (LLC) | ⬜ TODO | Before first sale |
| 🟡 HIGH | Register copyright | ⬜ TODO | Within 90 days |
| 🟡 HIGH | File trademark application | ⬜ TODO | Before marketing |
| 🟢 MEDIUM | Draft Terms of Service | ⬜ TODO | Before SaaS launch |
| 🟢 MEDIUM | Draft Privacy Policy | ⬜ TODO | Before data collection |

---

## Section 1: Open Source License Compliance

### 1.1 MIT License Dependencies

| Requirement | Status | Notes |
|-------------|--------|-------|
| Include MIT license text in distribution | ⬜ | |
| Include copyright notices | ⬜ | |
| No warranty disclaimer included | ⬜ | |

**Action Items**:
- [ ] Create `NOTICE` file listing all MIT dependencies
- [ ] Verify copyright years are correct

### 1.2 Apache 2.0 License Dependencies

| Requirement | Status | Notes |
|-------------|--------|-------|
| Include APACHE-2.0 license text | ⬜ | |
| Include NOTICE file if present in dependency | ⬜ | |
| State changes if modified | ⬜ | |
| Retain copyright notices | ⬜ | |

**Action Items**:
- [ ] Collect NOTICE files from all Apache 2.0 dependencies
- [ ] Document any modifications to Apache 2.0 code
- [ ] Create centralized attribution file

### 1.3 AGPL License Dependencies (⚠️ CRITICAL)

| Component | License | Status | Remediation |
|-----------|---------|--------|-------------|
| slither-analyzer | AGPLv3 | 🔴 ISSUE | Purchase license OR replace |
| Grafana | AGPLv3 | ⬜ OK | Using unmodified |
| Loki | AGPLv3 | ⬜ OK | Using unmodified |
| Tempo | AGPLv3 | ⬜ OK | Using unmodified |
| MinIO | AGPLv3 | ⬜ OK | Using unmodified |

**Slither Resolution Options**:
- [ ] Option A: Contact Trail of Bits for commercial license quote
- [ ] Option B: Replace with Mythril (MIT) or Securify2 (Apache 2.0)
- [ ] Option C: Keep Agent-Contracts fully open source under AGPL
- [ ] Option D: Remove smart contract scanning feature

**Decision**: _________________ Date: _________

### 1.4 BSD License Dependencies

| Requirement | Status |
|-------------|--------|
| Include BSD license text | ⬜ |
| Include copyright notice | ⬜ |
| No endorsement without permission | ⬜ |

---

## Section 2: API Terms Compliance

### 2.1 Anthropic Claude API

| Requirement | Status | Verification |
|-------------|--------|--------------|
| Not building competing AI product | ⬜ | Review product positioning |
| Not reverse engineering | ⬜ | Confirm implementation |
| Not training competing models | ⬜ | No model training |
| Terms accepted | ⬜ | Date: _________ |

**Documentation**:
- [ ] Save copy of current Anthropic Terms of Service
- [ ] Document how Claude is used in product
- [ ] Implement fallback for API unavailability

### 2.2 LLM Model Licenses

| Model | License | Commercial OK | Usage Documented |
|-------|---------|---------------|------------------|
| Llama 3.1 | Meta License | ✅ (<700M MAU) | ⬜ |
| Mistral | Apache 2.0 | ✅ | ⬜ |
| DeepSeek-Coder | DeepSeek License | ⬜ Verify | ⬜ |

**Action Items**:
- [ ] Document which models are deployed
- [ ] Verify commercial use terms for each model
- [ ] Create model selection guide for customers

---

## Section 3: Business Formation Compliance

### 3.1 Entity Formation

| Task | Status | Date |
|------|--------|------|
| Choose entity type (LLC/C-Corp) | ⬜ | |
| Choose state (Delaware recommended) | ⬜ | |
| Name availability search | ⬜ | |
| File formation documents | ⬜ | |
| Obtain EIN from IRS | ⬜ | |
| Open business bank account | ⬜ | |
| Register as foreign entity in operating state | ⬜ | |

### 3.2 Operating Documents

| Document | Status | Attorney Review |
|----------|--------|-----------------|
| Operating Agreement (LLC) / Bylaws (Corp) | ⬜ | ⬜ |
| IP Assignment to Company | ⬜ | ⬜ |
| Founder Agreement (if multiple) | ⬜ | ⬜ |
| Employee IP Agreement template | ⬜ | ⬜ |
| Contractor Agreement template | ⬜ | ⬜ |

---

## Section 4: Intellectual Property Compliance

### 4.1 Copyright

| Task | Status | Date |
|------|--------|------|
| Add copyright headers to all source files | ⬜ | |
| Prepare source code deposit | ⬜ | |
| File copyright registration (copyright.gov) | ⬜ | |
| Receive registration certificate | ⬜ | |

### 4.2 Trademark

| Task | Status | Date |
|------|--------|------|
| Conduct comprehensive trademark search | ⬜ | |
| Clear product name "Homelab" or alternative | ⬜ | |
| File trademark application (USPTO TEAS) | ⬜ | |
| Monitor application status | ⬜ | |
| Respond to Office Actions (if any) | ⬜ | |
| Registration granted | ⬜ | |

### 4.3 Trade Secrets

| Task | Status |
|------|--------|
| Identify trade secrets | ⬜ |
| Create trade secret register | ⬜ |
| Implement access controls | ⬜ |
| Add DTSA notice to employment agreements | ⬜ |
| Train employees on confidentiality | ⬜ |

### 4.4 Patent (Optional)

| Task | Status | Date |
|------|--------|------|
| Identify patentable inventions | ⬜ | |
| Conduct prior art search | ⬜ | |
| File provisional patent (if proceeding) | ⬜ | |
| 12-month deadline for non-provisional | | |

---

## Section 5: Product Legal Documents

### 5.1 Terms of Service

| Section | Status | Last Updated |
|---------|--------|--------------|
| Acceptance of terms | ⬜ | |
| Account registration | ⬜ | |
| Permitted use | ⬜ | |
| Prohibited use | ⬜ | |
| Intellectual property rights | ⬜ | |
| Third-party services | ⬜ | |
| Payment terms | ⬜ | |
| Termination | ⬜ | |
| Disclaimer of warranties | ⬜ | |
| Limitation of liability | ⬜ | |
| Indemnification | ⬜ | |
| Governing law (Delaware) | ⬜ | |
| Dispute resolution | ⬜ | |
| Modifications | ⬜ | |

### 5.2 Privacy Policy

| Section | Status | Last Updated |
|---------|--------|--------------|
| Information collected | ⬜ | |
| How information is used | ⬜ | |
| Information sharing | ⬜ | |
| Data security | ⬜ | |
| Data retention | ⬜ | |
| User rights (CCPA/GDPR) | ⬜ | |
| Cookies and tracking | ⬜ | |
| Children's privacy (COPPA) | ⬜ | |
| Changes to policy | ⬜ | |
| Contact information | ⬜ | |

### 5.3 End User License Agreement (EULA) - For Downloadable Software

| Section | Status |
|---------|--------|
| License grant | ⬜ |
| License restrictions | ⬜ |
| Ownership | ⬜ |
| Updates and support | ⬜ |
| Term and termination | ⬜ |

---

## Section 6: Open Source Release Compliance

### 6.1 Repository Setup

| Task | Status |
|------|--------|
| Create LICENSE file (MIT) | ⬜ |
| Create NOTICE file with attributions | ⬜ |
| Create CONTRIBUTING.md | ⬜ |
| Create CODE_OF_CONDUCT.md | ⬜ |
| Create SECURITY.md (vulnerability reporting) | ⬜ |
| Create .github/FUNDING.yml (optional) | ⬜ |

### 6.2 Contributor License Agreement

| Task | Status |
|------|--------|
| Draft CLA document | ⬜ |
| Set up CLA-bot or CLA Assistant | ⬜ |
| Document CLA requirements in CONTRIBUTING.md | ⬜ |

### 6.3 Attribution and NOTICE File

Template for NOTICE file:

```
NOTICE

Homelab
Copyright (c) 2025 Bruno Lucena

This product includes software developed at:
- The Apache Software Foundation (https://www.apache.org/)
- The Kubernetes Authors
- The Knative Authors

Third-Party Licenses
====================

[List all dependencies with their licenses]

```

---

## Section 7: Data Protection Compliance

### 7.1 CCPA Compliance (California)

| Requirement | Status | Implementation |
|-------------|--------|----------------|
| Privacy policy discloses data practices | ⬜ | |
| "Do Not Sell My Info" link (if applicable) | ⬜ | |
| Respond to access/deletion requests | ⬜ | |
| Verify requestor identity | ⬜ | |
| 45-day response window | ⬜ | |

### 7.2 GDPR Compliance (If serving EU customers)

| Requirement | Status | Implementation |
|-------------|--------|----------------|
| Lawful basis for processing | ⬜ | |
| Privacy policy in plain language | ⬜ | |
| Right to access | ⬜ | |
| Right to erasure | ⬜ | |
| Right to data portability | ⬜ | |
| Data Processing Agreement (DPA) | ⬜ | |
| Data breach notification process | ⬜ | |

---

## Section 8: Enterprise Customer Requirements

### 8.1 Security Documentation

| Document | Status | Date |
|----------|--------|------|
| Security whitepaper | ⬜ | |
| Architecture diagram | ⬜ | |
| Penetration test results | ⬜ | |
| Vulnerability management policy | ⬜ | |

### 8.2 Compliance Certifications (Future)

| Certification | Status | Timeline |
|---------------|--------|----------|
| SOC 2 Type I | ⬜ Not started | Year 2 |
| SOC 2 Type II | ⬜ Not started | Year 2-3 |
| ISO 27001 | ⬜ Not started | Year 3 |
| HIPAA (if healthcare) | ⬜ Not started | If needed |

### 8.3 Enterprise Agreements

| Document | Status |
|----------|--------|
| Master Services Agreement (MSA) | ⬜ |
| Service Level Agreement (SLA) | ⬜ |
| Data Processing Agreement (DPA) | ⬜ |
| Business Associate Agreement (BAA) | ⬜ |
| Non-Disclosure Agreement (NDA) | ⬜ |

---

## Section 9: Export Control Compliance

### 9.1 EAR/ITAR Review

| Question | Answer | Notes |
|----------|--------|-------|
| Does software contain encryption? | Yes | |
| Is encryption >64-bit key length? | Yes | |
| EAR Classification | Likely 5D002 | |
| License Exception TSR eligible? | ⬜ Review | |

**Action Items**:
- [ ] Determine ECCN classification
- [ ] File encryption commodity classification (if required)
- [ ] Document export compliance procedures

---

## Section 10: Insurance

### 10.1 Recommended Coverage

| Insurance Type | Status | Coverage |
|----------------|--------|----------|
| General Liability | ⬜ | $1M+ |
| Professional Liability (E&O) | ⬜ | $1M+ |
| Cyber Liability | ⬜ | $1M+ |
| Directors & Officers (D&O) | ⬜ | When raising |

---

## Pre-Launch Certification

### Sign-Off

I certify that all critical compliance items have been addressed:

**Founder/CEO**: _____________________ Date: _________

**Legal Counsel Review**: _____________________ Date: _________

---

## Appendix: Resources

### Legal Templates (Free/Low-Cost)

- [Cooley GO](https://www.cooleygo.com/documents/) - Startup legal docs
- [Y Combinator Series A Docs](https://www.ycombinator.com/documents/)
- [Indie Hackers Legal Guide](https://www.indiehackers.com/)

### Compliance Tools

- [FOSSA](https://fossa.com/) - License compliance scanning
- [Snyk](https://snyk.io/) - Dependency security
- [WhiteSource](https://www.mend.io/) - Open source management

### Legal Services

- [Clerky](https://www.clerky.com/) - Startup legal automation
- [Stripe Atlas](https://stripe.com/atlas) - Delaware incorporation
- [LegalZoom](https://www.legalzoom.com/) - Business formation

---

**Document prepared for**: Bruno Lucena / Homelab Project  
**Review Required**: This checklist should be reviewed by qualified legal counsel before commercial launch.
