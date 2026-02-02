# 🇨🇳 法律指南 - 中国商业化 Homelab (Legal Guide - Commercializing Homelab in China)

> **Document Version**: 1.0  
> **Last Updated**: December 11, 2025  
> **Author**: Bruno Lucena  
> **Jurisdiction**: People's Republic of China (中华人民共和国)

---

## Executive Summary

This document provides comprehensive analysis of legal requirements for commercializing the Homelab platform in China. The Chinese market presents unique challenges including:

- Strict foreign ownership restrictions in telecom/internet services
- Complex licensing requirements (ICP, cybersecurity)
- Data localization requirements under PIPL
- Rapidly evolving regulatory landscape

### ⚠️ Critical Findings

| Risk Level | Issue | Action Required |
|------------|-------|-----------------|
| 🔴 **CRITICAL** | ICP License requires Chinese majority ownership | Use WFOE + VIE structure OR partner with local company |
| 🔴 **CRITICAL** | PIPL compliance for personal data | Data localization, DPO appointment |
| 🔴 **CRITICAL** | Cybersecurity Law compliance | Security assessments, certification |
| 🟡 **HIGH** | slither-analyzer (AGPLv3) | Commercial license or replacement |
| 🟡 **MEDIUM** | Trademark registration | Register with CNIPA early |

### ⚠️ Important Warning

**Operating internet services in China as a foreign company is highly complex and requires significant legal expertise. This guide provides an overview, but you MUST consult with qualified Chinese legal counsel before entering this market.**

---

## Table of Contents

1. [Regulatory Framework](#1-regulatory-framework)
2. [Business Structures for Foreign Companies](#2-business-structures-for-foreign-companies)
3. [Licensing Requirements](#3-licensing-requirements)
4. [Data Protection (PIPL)](#4-data-protection-pipl)
5. [Cybersecurity Law Compliance](#5-cybersecurity-law-compliance)
6. [Intellectual Property Protection](#6-intellectual-property-protection)
7. [Open Source License Compliance](#7-open-source-license-compliance)
8. [Taxation](#8-taxation)
9. [Compliance Checklist](#9-compliance-checklist)
10. [Risk Assessment](#10-risk-assessment)

---

## 1. Regulatory Framework

### 1.1 Key Laws and Regulations

| Law | Chinese Name | Effective Date | Scope |
|-----|--------------|----------------|-------|
| **Cybersecurity Law** | 网络安全法 | June 1, 2017 (Amended Jan 2026) | Network security, data protection |
| **Data Security Law** | 数据安全法 | September 1, 2021 | Data classification, cross-border transfer |
| **PIPL** | 个人信息保护法 | November 1, 2021 | Personal information protection |
| **E-Commerce Law** | 电子商务法 | January 1, 2019 | Online business operations |
| **Foreign Investment Law** | 外商投资法 | January 1, 2020 | Foreign investment rules |
| **Copyright Law** | 著作权法 | June 1, 2021 (Revised) | Software copyright |
| **Trademark Law** | 商标法 | November 1, 2019 (Revised) | Trademark protection |

### 1.2 Regulatory Authorities

| Authority | Chinese Name | Responsibility |
|-----------|--------------|----------------|
| **CAC** | 国家互联网信息办公室 | Cyberspace Administration - internet content, data |
| **MIIT** | 工业和信息化部 | Ministry of Industry and IT - telecom licenses |
| **CNIPA** | 国家知识产权局 | IP Office - patents, trademarks |
| **NCAC** | 国家版权局 | Copyright Administration - software copyright |
| **SAMR** | 国家市场监督管理总局 | Market Supervision - business registration |

### 1.3 Negative List Restrictions

The **2024 Negative List for Foreign Investment** restricts foreign ownership in:

| Sector | Restriction |
|--------|-------------|
| Value-Added Telecom Services | Foreign ownership ≤ 50% |
| Internet News Services | Prohibited for foreign investment |
| Online Publishing | Prohibited for foreign investment |
| Internet Cultural Services | Requires approval |
| Cloud Services | Requires partnership with local provider |

**Impact on Homelab**: As a SaaS platform, you may need an ICP license, which requires Chinese majority ownership.

---

## 2. Business Structures for Foreign Companies

### 2.1 Options Overview

| Structure | Chinese Name | Foreign Ownership | ICP Eligible | Complexity |
|-----------|--------------|-------------------|--------------|------------|
| **WFOE** | 外商独资企业 | 100% | ❌ No (for telecom) | Medium |
| **JV** | 中外合资企业 | ≤ 50% for telecom | ✅ Yes | High |
| **VIE** | 可变利益实体 | Indirect control | ✅ Via local entity | Very High |
| **Representative Office** | 代表处 | N/A | ❌ No | Low |

### 2.2 Wholly Foreign-Owned Enterprise (WFOE)

**Best for**: Software development, consulting, non-telecom services

**Advantages**:
- 100% foreign ownership and control
- Full profit repatriation
- No local partner required

**Limitations**:
- Cannot obtain ICP license directly
- Cannot operate public internet services

**Costs (Estimated)**:

| Item | Cost (RMB) | Cost (USD) |
|------|------------|------------|
| Registration capital | 100,000 - 1,000,000+ | $14,000 - $140,000+ |
| Setup fees | 30,000 - 80,000 | $4,200 - $11,200 |
| Annual audit | 20,000 - 50,000 | $2,800 - $7,000 |
| Office rental | 60,000 - 300,000/year | $8,400 - $42,000/year |

**Timeline**: 2-4 months

### 2.3 Variable Interest Entity (VIE) Structure

**What is VIE?**

A legal arrangement where a WFOE controls a Chinese domestic company through contractual agreements rather than equity ownership.

```
┌─────────────────────────────────────────────────────────────────┐
│                    OFFSHORE HOLDING COMPANY                      │
│                    (e.g., Cayman Islands)                        │
└───────────────────────────┬─────────────────────────────────────┘
                            │ 100% Ownership
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                         WFOE (China)                            │
│               Wholly Foreign-Owned Enterprise                    │
│         Provides services, receives fees from VIE               │
└───────────────────────────┬─────────────────────────────────────┘
                            │ Contractual Control
                            │ (Service Agreements, Loan Agreements,
                            │  Option Agreements, Voting Rights)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    VIE (Domestic Company)                        │
│           Chinese-owned, holds ICP license                       │
│           Operates internet services                             │
└─────────────────────────────────────────────────────────────────┘
```

**VIE Agreements**:
1. **Exclusive Service Agreement**: WFOE provides services, VIE pays fees
2. **Exclusive Option Agreement**: WFOE can acquire VIE when permitted
3. **Loan Agreement**: WFOE lends to VIE shareholders
4. **Equity Pledge Agreement**: VIE shares pledged to WFOE
5. **Power of Attorney**: VIE shareholders grant voting rights to WFOE

**⚠️ VIE Risks**:
- Legal uncertainty (not explicitly approved)
- Regulatory scrutiny increasing
- Enforceability of contracts unclear
- Chinese shareholders may not honor agreements

**Recommendation**: Consult specialized Chinese legal counsel before implementing VIE.

### 2.4 Joint Venture (JV)

**Best for**: Projects requiring Chinese partner expertise or licenses

| Requirement | Details |
|-------------|---------|
| Foreign ownership | ≤ 50% for telecom/internet |
| Chinese partner | Required, must hold ICP license |
| Board control | Negotiated |
| Profit sharing | Pro-rata or negotiated |

**Risks**:
- IP leakage to partner
- Loss of operational control
- Partner disputes

### 2.5 Recommended Structure for Homelab

**Option A: WFOE + Partnership (Lower Risk)**

- Establish WFOE for software development
- Partner with Chinese cloud provider (Alibaba Cloud, Tencent Cloud)
- Cloud provider handles ICP compliance
- You provide software, they handle infrastructure

**Option B: VIE Structure (Higher Control, Higher Risk)**

- For full market control
- Requires significant legal investment
- Risk of regulatory changes

**Option C: Cross-Border SaaS (Limited Market)**

- Serve Chinese customers from overseas
- Limited functionality (data must stay outside China)
- Compliance challenges for Chinese users

---

## 3. Licensing Requirements

### 3.1 ICP License (互联网内容提供商许可证)

**What is ICP?**

Internet Content Provider license, required for operating websites/services accessible in China.

**Types**:

| Type | Chinese | Purpose |
|------|---------|---------|
| **ICP Filing** | ICP备案 | Informational websites |
| **ICP License** | ICP经营许可证 | Commercial internet services |
| **EDI License** | EDI经营许可证 | E-commerce platforms |

**Requirements for ICP License**:

| Requirement | Details |
|-------------|---------|
| Chinese entity | ≥ 50% Chinese ownership for telecom |
| Registered capital | ≥ RMB 1,000,000 |
| Business scope | Must include "value-added telecom services" |
| Technical review | Server security assessment |
| Content review | Content compliance review |
| Domain | Must be registered in China |
| Servers | Must be hosted in mainland China |

**Timeline**: 3-6 months

**Cost**: RMB 50,000 - 200,000 (setup + consulting)

### 3.2 Cloud Service Licensing

If providing cloud/SaaS services:

| License | Purpose |
|---------|---------|
| **CDN License** | Content delivery |
| **ISP License** | Internet access services |
| **IDC License** | Data center services |

**Alternative**: Partner with licensed Chinese cloud providers who already have these licenses.

### 3.3 AI-Related Approvals

For AI components (LLM integration):

| Requirement | Authority | Notes |
|-------------|-----------|-------|
| Algorithm Registration | CAC | Required for AI recommendation algorithms |
| Generative AI Approval | CAC | For public-facing generative AI |
| Deep Synthesis Rules | CAC | For AI-generated content |

**As of 2025**: Generative AI services must:
- Register algorithms with CAC
- Implement content moderation
- Label AI-generated content
- Maintain training data records

---

## 4. Data Protection (PIPL)

### 4.1 Overview

The **Personal Information Protection Law (PIPL)** is China's comprehensive data protection law, similar to GDPR.

**Effective**: November 1, 2021

**Scope**: Applies to:
- Processing personal information in China
- Processing personal information of individuals in China (even from abroad)

### 4.2 Key Requirements

#### Legal Bases for Processing

| Legal Basis | When Applicable |
|-------------|-----------------|
| **Consent** | Default basis, must be informed and voluntary |
| **Contract Performance** | Necessary for contract execution |
| **Legal Obligation** | Required by law |
| **Public Interest** | Emergency, public health |
| **Legitimate Processing** | Within reasonable scope of disclosed info |

#### Sensitive Personal Information

| Category | Examples | Requirements |
|----------|----------|--------------|
| Biometrics | Fingerprints, facial recognition | Explicit consent + necessity |
| Religious beliefs | - | Explicit consent |
| Health information | Medical records | Explicit consent |
| Financial information | Bank accounts, credit | Explicit consent |
| Location data | GPS tracking | Explicit consent |
| Minors' data | Under 14 years old | Parental consent |

### 4.3 Data Localization

**General Rule**: No mandatory localization for most businesses.

**Exceptions** (data MUST stay in China):

| Category | Threshold |
|----------|-----------|
| Critical Information Infrastructure Operators (CIIOs) | All personal data |
| Processing personal info of ≥ 1 million individuals | All personal data |
| Cumulative transfer of ≥ 100,000 individuals' data | Requires security assessment |
| Cumulative transfer of ≥ 10,000 individuals' sensitive data | Requires security assessment |

### 4.4 Cross-Border Data Transfer

To transfer personal data outside China:

| Method | When to Use |
|--------|-------------|
| **CAC Security Assessment** | CIIOs, large-scale processing |
| **Standard Contract** | Regular processing, < thresholds |
| **Certification** | Groups of companies, specific scenarios |

**Standard Contract** (SCC) Requirements:
1. Sign CAC-approved standard contract
2. Conduct Personal Information Protection Impact Assessment (PIPIA)
3. File with provincial CAC

### 4.5 DPO Requirement

**When Required**:
- Processing personal info of ≥ 1 million individuals

**DPO Responsibilities**:
- Contact person for data subjects
- Supervise compliance
- Report to authorities

### 4.6 Compliance Audit

**Effective May 1, 2025**: PIPs must conduct compliance audits.

| Processor Size | Audit Frequency |
|----------------|-----------------|
| ≥ 10 million individuals' data | Every 2 years |
| CAC-mandated | As required |

### 4.7 Penalties

| Violation | Penalty |
|-----------|---------|
| General violations | Up to RMB 1 million |
| Serious violations | Up to RMB 50 million or 5% of annual revenue |
| Responsible individuals | Up to RMB 1 million personal liability |
| Additional | Service suspension, license revocation |

---

## 5. Cybersecurity Law Compliance

### 5.1 Key Obligations

| Obligation | Description |
|------------|-------------|
| **Network Security Protection** | Implement technical measures |
| **Multi-Level Protection Scheme (MLPS)** | Security grading system |
| **Incident Response** | 24-hour breach notification |
| **Log Retention** | ≥ 6 months of network logs |
| **Real-Name Registration** | Verify user identities |

### 5.2 Multi-Level Protection Scheme (MLPS 2.0)

Mandatory security certification for information systems.

| Level | System Type | Certification |
|-------|-------------|---------------|
| Level 1 | User-level systems | Self-assessment |
| Level 2 | Departmental systems | Authority review |
| Level 3 | City/industry systems | Third-party assessment |
| Level 4 | National systems | National authority |
| Level 5 | Critical national systems | Highest protection |

**For SaaS**: Typically Level 2 or Level 3 required.

**Cost**: RMB 50,000 - 500,000 for assessment and remediation.

### 5.3 2026 Amendments

Effective January 1, 2026:

- Enhanced AI-related provisions
- Stricter penalties (up to RMB 50 million)
- Expanded scope of violations
- Alignment with PIPL and Data Security Law

---

## 6. Intellectual Property Protection

### 6.1 Software Copyright

**Registration Authority**: National Copyright Administration of China (NCAC)

**Benefits of Registration**:
- Evidence of ownership
- Required for some government contracts
- Easier enforcement

**Process**:

| Step | Description | Timeline |
|------|-------------|----------|
| 1. Prepare materials | Source code, documentation | 1-2 weeks |
| 2. Submit application | Via China Copyright Protection Center | 1 day |
| 3. Review | Examination | 30-60 days |
| 4. Certificate | Registration certificate issued | ~3 months total |

**Costs**:

| Item | Fee (RMB) |
|------|-----------|
| Standard registration | 250 |
| Expedited (31-35 days) | 560 |
| Urgent (21-25 days) | 800 |
| Very urgent (11-15 days) | 1,200 |
| Super urgent (5-10 days) | 2,000 |

### 6.2 Trademark Registration

**Authority**: China National Intellectual Property Administration (CNIPA)

**Important**: China uses **first-to-file** system. Register early!

**Process**:

| Step | Description | Timeline |
|------|-------------|----------|
| 1. Search | Check for conflicts | 1-2 weeks |
| 2. Filing | Submit application | 1 day |
| 3. Examination | CNIPA review | 4-6 months |
| 4. Publication | Opposition period | 3 months |
| 5. Registration | Certificate issued | 12-18 months total |

**Costs**:

| Item | Fee (RMB) | Fee (USD) |
|------|-----------|-----------|
| Official fee | 300/class (≤10 items) | ~$42 |
| Additional items | 30/item | ~$4 |
| Agency fee | 1,000-3,000 | $140-420 |
| **Total per class** | ~1,500-3,500 | ~$210-490 |

**Relevant Classes**:

| Class | Description |
|-------|-------------|
| Class 9 | Downloadable software |
| Class 42 | SaaS, cloud computing, IT services |
| Class 35 | Advertising, business management |

### 6.3 Trade Secrets

Protected under Anti-Unfair Competition Law (2019 revised).

**Requirements**:
- Information not publicly known
- Has commercial value
- Subject to confidentiality measures

**Protection Measures**:
- NDAs with employees and partners
- Access controls
- "Confidential" markings
- Employee training

---

## 7. Open Source License Compliance

### 7.1 Legal Status in China

Open source licenses are **enforceable contracts** in China.

**Landmark Cases**:
- Beijing Hantao vs Zhongke Fangde (2021): AGPL enforcement
- Multiple GPL violation cases upheld by Chinese courts

### 7.2 AGPL Implications

Same as globally: AGPL requires source disclosure for network services.

**slither-analyzer Issue**: Same risk in China - need commercial license or alternative.

### 7.3 Government Preference

Chinese government has been promoting:
- Domestic open source ecosystems (Gitee, OpenEuler)
- Open source compliance in procurement
- Contributions to international open source

---

## 8. Taxation

### 8.1 Corporate Taxes

| Tax | Rate | Notes |
|-----|------|-------|
| **Corporate Income Tax (CIT)** | 25% | Standard rate |
| **High-Tech Enterprise** | 15% | Reduced rate with certification |
| **Small/Low-Profit** | 5-20% | Tiered based on profit |
| **VAT** | 6% | For software services |
| **Withholding Tax** | 10% | On dividends to foreign investors |

### 8.2 Software Industry Incentives

| Incentive | Benefit | Requirements |
|-----------|---------|--------------|
| High-Tech Enterprise | 15% CIT rate | R&D activities, IP ownership |
| Software Enterprise | 2-year exemption, 3-year 50% reduction | Software products |
| R&D Super Deduction | 200% deduction | Qualified R&D expenses |

### 8.3 Transfer Pricing

Transactions between WFOE and foreign parent must be at arm's length prices.

**Documentation Required**:
- Transfer pricing analysis
- Comparability study
- Annual reporting for related-party transactions

---

## 9. Compliance Checklist

### 9.1 Pre-Market Entry

#### Business Structure
- [ ] Decide on entry structure (WFOE, JV, VIE, partnership)
- [ ] Engage Chinese legal counsel
- [ ] Engage Chinese accounting firm
- [ ] Evaluate ICP license requirements
- [ ] Assess partnership with local cloud provider

#### Intellectual Property
- [ ] File trademark application with CNIPA (all relevant classes)
- [ ] Register software copyright with NCAC
- [ ] Document trade secrets
- [ ] Implement IP protection measures

#### Licensing
- [ ] Determine required licenses (ICP, cloud, etc.)
- [ ] Plan license acquisition strategy
- [ ] Budget for licensing costs and timeline

### 9.2 Data Compliance

#### PIPL
- [ ] Map personal data processing activities
- [ ] Identify legal bases for processing
- [ ] Prepare privacy policy (Chinese language)
- [ ] Implement consent mechanisms
- [ ] Assess data localization requirements
- [ ] Plan cross-border transfer mechanism (if needed)
- [ ] Appoint DPO (if required)
- [ ] Prepare for compliance audit

#### Cybersecurity
- [ ] Determine MLPS level requirement
- [ ] Plan security assessment
- [ ] Implement security controls
- [ ] Set up log retention (≥6 months)
- [ ] Implement incident response procedures

### 9.3 Operational

- [ ] Establish Chinese entity (if proceeding)
- [ ] Obtain business license
- [ ] Open bank account
- [ ] Register for taxes
- [ ] Set up local operations/support
- [ ] Localize product (Chinese language, cultural adaptation)
- [ ] Implement content moderation (if required)

---

## 10. Risk Assessment

### 10.1 Market Entry Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Regulatory changes | High | High | Flexible structure, local counsel |
| VIE enforceability | Medium | Critical | Careful structuring, alternatives |
| IP theft | Medium | High | Strong contracts, registration |
| Partner disputes | Medium | Medium | Careful partner selection, clear contracts |
| Data compliance failure | Medium | High | Robust compliance program |

### 10.2 Ongoing Operational Risks

| Risk | Description | Mitigation |
|------|-------------|------------|
| Content liability | Platform content violations | Moderation, filtering |
| Data breach | PIPL penalties | Security measures, insurance |
| Tax compliance | Transfer pricing disputes | Arm's length documentation |
| License revocation | Regulatory violations | Compliance monitoring |

### 10.3 Decision Framework

**Enter China market IF**:
- ✅ Significant market opportunity
- ✅ Willing to invest in local presence
- ✅ Accept regulatory complexity
- ✅ Have resources for compliance
- ✅ Can accept VIE/partnership risks

**Avoid China market IF**:
- ❌ Cannot commit significant resources
- ❌ Need full ownership/control
- ❌ Cannot accept regulatory uncertainty
- ❌ Limited tolerance for compliance costs

---

## Cost Summary

### Initial Entry Costs (Estimated)

| Item | Cost (RMB) | Cost (USD) |
|------|------------|------------|
| Legal counsel (setup) | 200,000 - 500,000 | $28,000 - $70,000 |
| WFOE registration | 50,000 - 100,000 | $7,000 - $14,000 |
| Registered capital | 1,000,000+ | $140,000+ |
| ICP license (if applicable) | 100,000 - 300,000 | $14,000 - $42,000 |
| Trademark registration | 5,000 - 10,000 | $700 - $1,400 |
| Software copyright | 500 - 2,000 | $70 - $280 |
| MLPS certification | 50,000 - 200,000 | $7,000 - $28,000 |
| **Total Minimum** | **~1,500,000** | **~$210,000** |
| **Total Maximum** | **~3,000,000+** | **~$420,000+** |

### Annual Operating Costs

| Item | Cost (RMB) | Cost (USD) |
|------|------------|------------|
| Office/registered address | 60,000 - 300,000 | $8,400 - $42,000 |
| Accounting/audit | 30,000 - 100,000 | $4,200 - $14,000 |
| Legal compliance | 50,000 - 200,000 | $7,000 - $28,000 |
| Local staff | 200,000 - 500,000+ | $28,000 - $70,000+ |
| **Total Annual** | **~340,000 - 1,100,000** | **~$48,000 - $154,000** |

---

## Resources

### Government Resources

| Resource | URL |
|----------|-----|
| CNIPA (Trademarks/Patents) | [cnipa.gov.cn](https://www.cnipa.gov.cn/) |
| NCAC (Copyright) | [ncac.gov.cn](http://www.ncac.gov.cn/) |
| CAC (Cybersecurity) | [cac.gov.cn](http://www.cac.gov.cn/) |
| MIIT (Telecom) | [miit.gov.cn](https://www.miit.gov.cn/) |
| SAMR (Business) | [samr.gov.cn](https://www.samr.gov.cn/) |

### Useful References

| Resource | Description |
|----------|-------------|
| China Briefing | Business guides and updates |
| China Law Blog | Legal commentary |
| NPC Observer | Legislative updates |

---

**Document prepared for**: Bruno Lucena / Homelab Project  
**Disclaimer**: This document is for informational purposes only and does not constitute legal advice. The Chinese regulatory environment is complex and rapidly changing. Consult with qualified Chinese legal counsel before making any business decisions regarding market entry.

**特别声明**: 本文件仅供参考，不构成法律建议。中国监管环境复杂且变化迅速。在做出任何市场进入决策前，请咨询合格的中国法律顾问。
