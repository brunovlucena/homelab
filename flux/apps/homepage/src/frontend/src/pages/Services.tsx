import React from 'react'

const Services: React.FC = () => {
  return (
    <div className="services-page">
      <section className="hero">
        <div className="container">
          <h1>AI Agent Automation Services</h1>
          <p>Enterprise-grade AI automation solutions using intelligent agents</p>
        </div>
      </section>

      <section id="services" className="section">
        <div className="container">
          <p className="services-intro">
            Enterprise-grade AI automation solutions using intelligent agents. Production-ready, domain-specific agents 
            that transform business operations and reduce costs by up to 60%.
          </p>
          <div className="services-grid">
            <div className="service-card">
              <div className="service-icon">🏥</div>
              <h3>Healthcare AI Agents</h3>
              <p>HIPAA-compliant AI agents for medical records management, reducing documentation time by 60%. Built-in 
                 compliance, audit logging, and role-based access control.</p>
              <ul className="service-features">
                <li>HIPAA-compliant by design</li>
                <li>Natural language medical records access</li>
                <li>Drug interaction checking</li>
                <li>Comprehensive audit logging</li>
              </ul>
            </div>
            
            <div className="service-card">
              <div className="service-icon">🍽️</div>
              <h3>Restaurant AI Agents</h3>
              <p>AI-powered fine dining experience with specialized agents for hosts, waiters, sommeliers, and chefs. 
                 Increase average check size by 15-25% through intelligent upselling and personalized service.</p>
              <ul className="service-features">
                <li>Multi-agent restaurant system</li>
                <li>Real-time kitchen coordination</li>
                <li>Wine pairing recommendations</li>
                <li>Table optimization</li>
              </ul>
            </div>
            
            <div className="service-card">
              <div className="service-icon">🛒</div>
              <h3>E-Commerce AI Agents</h3>
              <p>Multi-brand e-commerce platform with AI sellers, WhatsApp integration, and 24/7 customer service. 
                 Increase conversion rates by 20-30% with intelligent product recommendations.</p>
              <ul className="service-features">
                <li>Brand-specific AI personalities</li>
                <li>WhatsApp Business API integration</li>
                <li>24/7 automated customer service</li>
                <li>Multi-channel sales management</li>
              </ul>
            </div>
            
            <div className="service-card">
              <div className="service-icon">⚙️</div>
              <h3>DevOps/SRE Automation</h3>
              <p>AI-powered SRE/DevOps automation agents for infrastructure automation, incident response, and cost 
                 optimization. Reduce MTTR by 50% and automate routine tasks.</p>
              <ul className="service-features">
                <li>Kubernetes automation</li>
                <li>Automated incident response</li>
                <li>Infrastructure cost optimization</li>
                <li>Observability integration</li>
              </ul>
            </div>
            
            <div className="service-card">
              <div className="service-icon">🔒</div>
              <h3>Security Automation</h3>
              <p>Comprehensive security agent suite including Red Team, Blue Team, and DevSecOps automation. Proactive 
                 threat detection and automated security testing.</p>
              <ul className="service-features">
                <li>Automated penetration testing</li>
                <li>Threat response automation</li>
                <li>Security compliance validation</li>
                <li>24/7 vulnerability scanning</li>
              </ul>
            </div>
            
            <div className="service-card">
              <div className="service-icon">💎</div>
              <h3>Smart Contract Security</h3>
              <p>AI-powered smart contract vulnerability detection. Automated security audits at 99% cost reduction 
                 compared to manual audits. Multi-chain support for Ethereum, Polygon, BSC, and more.</p>
              <ul className="service-features">
                <li>Automated vulnerability detection</li>
                <li>Multi-chain support</li>
                <li>Static analysis + LLM-based detection</li>
                <li>Continuous contract monitoring</li>
              </ul>
            </div>
          </div>
          <div className="services-cta">
            <p>Ready to transform your business with AI automation?</p>
            <a href="/#contact" className="cta-button">Schedule a Free Consultation</a>
          </div>
        </div>
      </section>
    </div>
  )
}

export default Services
