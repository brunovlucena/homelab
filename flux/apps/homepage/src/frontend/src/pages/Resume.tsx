import { useState, useEffect } from 'react'
import { apiClient } from '../services/api'
import type { SiteConfig, Experience } from '../services/api'
import { 
  SiTypescript,
  SiGo,
  SiPostgresql,
  SiRedis,
  SiKubernetes,
  SiFlux,
  SiDocker,
  SiGithubactions,
  SiPulumi,
  SiPrometheus,
  SiGrafana,
  SiAmazon,
  SiGooglecloud,
  SiTerraform,
  SiKnative,
  SiRabbitmq
} from 'react-icons/si'
import { FaSearch, FaDatabase, FaStream, FaEye } from 'react-icons/fa'
import { FaShieldAlt, FaLock, FaChartBar, FaCloud, FaRobot, FaRocket } from 'react-icons/fa'
import { BiCertification } from 'react-icons/bi'

// Technology URLs mapping
const getTechnologyUrl = (techName: string): string => {
  const urlMap: { [key: string]: string } = {
    // Cloud Platforms
    'AWS': 'https://aws.amazon.com/',
    
    // Kubernetes & Containerization
    'Kubernetes': 'https://kubernetes.io/',
    'Docker': 'https://www.docker.com/',
    'EKS': 'https://aws.amazon.com/eks/',
    'Kops': 'https://kops.sigs.k8s.io/',
    'Bare-metal': 'https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/',
    'Kind': 'https://kind.sigs.k8s.io/',
    'k3s': 'https://k3s.io/',
    'Kaniko': 'https://github.com/GoogleContainerTools/kaniko',
    'Kustomize': 'https://kustomize.io/',
    'Kubernetes Operators': 'https://kubernetes.io/docs/concepts/extend-kubernetes/operator/',
    
    // Infrastructure as Code
    'Terraform': 'https://www.terraform.io/',
    'Pulumi': 'https://www.pulumi.com/',
    'Ansible': 'https://www.ansible.com/',
    'Chef': 'https://www.chef.io/',
    'Saltstack': 'https://www.saltproject.io/',
    
    // CI/CD & DevOps
    'CI/CD': 'https://www.atlassian.com/continuous-delivery',
    'GitHub Actions': 'https://github.com/features/actions',
    'ArgoCD': 'https://argoproj.github.io/argo-cd/',
    'Flux': 'https://fluxcd.io/',
    'GitOps': 'https://www.gitops.tech/',
    'Flagger': 'https://flagger.app/',
    'Canary Deployments': 'https://flagger.app/',
    
    // Monitoring & Observability
    'Prometheus': 'https://prometheus.io/',
    'Grafana': 'https://grafana.com/',
    'Loki': 'https://grafana.com/oss/loki/',
    'Tempo': 'https://grafana.com/oss/tempo/',
    'Thanos': 'https://thanos.io/',
    'ELK': 'https://www.elastic.co/elk-stack',
    'EFK': 'https://www.elastic.co/elk-stack',
    'OpenTelemetry': 'https://opentelemetry.io/',
    'Monitoring': 'https://prometheus.io/',
    'Logging': 'https://grafana.com/oss/loki/',
    'Tracing': 'https://grafana.com/oss/tempo/',
    'Alerting': 'https://prometheus.io/docs/alerting/latest/',
    'Metrics': 'https://prometheus.io/',
    'k6': 'https://k6.io/',
    
    // Programming Languages
    'Go': 'https://golang.org/',
    'Golang': 'https://golang.org/',
    'Python': 'https://www.python.org/',
    'Bash': 'https://www.gnu.org/software/bash/',
    'JavaScript': 'https://developer.mozilla.org/en-US/docs/Web/JavaScript',
    'TypeScript': 'https://www.typescriptlang.org/',
    
    // Databases & Messaging
    'PostgreSQL': 'https://www.postgresql.org/',
    'Redis': 'https://redis.io/',
    'RabbitMQ': 'https://www.rabbitmq.com/',
    'MongoDB': 'https://www.mongodb.com/',
    'Kafka': 'https://kafka.apache.org/',
    'MinIO': 'https://min.io/',
    
    // Distributed Systems
    'Mesos': 'http://mesos.apache.org/',
    'Consul': 'https://www.consul.io/',
    'Linkerd': 'https://linkerd.io/',
    'Distributed Systems': 'https://en.wikipedia.org/wiki/Distributed_computing',
    'Service Mesh': 'https://linkerd.io/',
    'Multi-Cluster': 'https://linkerd.io/2/features/multicluster/',
    
    // Serverless & Platforms
    'Serverless': 'https://www.serverless.com/',
    'AWS Lambda': 'https://aws.amazon.com/lambda/',
    'Knative': 'https://knative.dev/',
    'Knative Serving': 'https://knative.dev/docs/serving/',
    'Knative Eventing': 'https://knative.dev/docs/eventing/',
    'CloudEvents': 'https://cloudevents.io/',
    'FaaS': 'https://www.ibm.com/topics/faas',
    'Scale-to-Zero': 'https://knative.dev/docs/serving/autoscaling/scale-to-zero/',
    
    // Security
    'Security': 'https://owasp.org/',
    'Compliance': 'https://www.iso.org/iso-27001-information-security.html',
    'Network Security': 'https://www.cisco.com/c/en/us/solutions/enterprise-networks/network-security.html',
    'VPN': 'https://www.openvpn.org/',
    'Sealed Secrets': 'https://sealed-secrets.netlify.app/',
    'cert-manager': 'https://cert-manager.io/',
    'Falco': 'https://falco.org/',
    'mTLS': 'https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/',
    
    // AI/ML & LLM
    'RAG': 'https://www.ibm.com/topics/retrieval-augmented-generation',
    'Vertex AI': 'https://cloud.google.com/vertex-ai',
    'Machine Learning': 'https://www.tensorflow.org/',
    'Ollama': 'https://ollama.ai/',
    'VLLM': 'https://docs.vllm.ai/',
    'LLM': 'https://en.wikipedia.org/wiki/Large_language_model',
    'AI Agents': 'https://en.wikipedia.org/wiki/Intelligent_agent',
    'DeepSeek': 'https://www.deepseek.com/',
    
    // Blockchain & Smart Contracts
    'Ethereum': 'https://ethereum.org/',
    'Smart Contracts': 'https://ethereum.org/en/developers/docs/smart-contracts/',
    'Solidity': 'https://soliditylang.org/',
    'DeFi': 'https://ethereum.org/en/defi/',
    'Slither': 'https://github.com/crytic/slither',
    'Anvil': 'https://book.getfoundry.sh/anvil/',
    'Foundry': 'https://book.getfoundry.sh/',
    'Web3': 'https://ethereum.org/en/web3/',
    'EVM': 'https://ethereum.org/en/developers/docs/evm/',
    
    // Networking
    'Load Balancing': 'https://kubernetes.io/docs/concepts/services-networking/service/',
    'API Gateway': 'https://kubernetes.io/docs/concepts/services-networking/ingress/',
    
    // Management
    'Team Leadership': 'https://www.atlassian.com/agile',
    'People Management': 'https://www.atlassian.com/agile',
    'Project Management': 'https://www.atlassian.com/agile',
    'Agile/Scrum': 'https://www.scrum.org/',
    
    // Operations
    'Operations': 'https://sre.google/',
    'Infrastructure': 'https://www.hashicorp.com/',
    'Automation': 'https://www.ansible.com/',
    'Cloud Operations': 'https://cloud.google.com/architecture/framework/operational-excellence',
    'Infrastructure as Code': 'https://www.terraform.io/',
    'Site Reliability Engineering': 'https://sre.google/',
    'DevSecOps': 'https://www.redhat.com/en/topics/devops/what-is-devsecops',
    'SRE': 'https://sre.google/',
    
    // Observability
    'Observability': 'https://opentelemetry.io/',
    'Problem-Solving': 'https://kubernetes.io/docs/tasks/debug/',
    'Troubleshooting': 'https://kubernetes.io/docs/tasks/debug/',
    'Collaboration': 'https://github.com/',
    
    // General
    'Cloud Migration': 'https://aws.amazon.com/migration/',
    'VMware ESXi': 'https://www.vmware.com/products/esxi-and-esx.html',
    'Cloud Native Infrastructure': 'https://kubernetes.io/',
    'Cloud-Native Infrastructure': 'https://kubernetes.io/',
    
    // Default
    'default': `https://www.google.com/search?q=${encodeURIComponent(techName)}`
  }

  return urlMap[techName] ?? urlMap['default'] ?? `https://www.google.com/search?q=${encodeURIComponent(techName)}`
}

// Technology icons mapping with React Icons and colors
const getTechnologyIcon = (techName: string) => {
  const iconMap: { [key: string]: { component: React.ComponentType<any>, color: string } } = {
    // Cloud Platforms
    'AWS': { component: SiAmazon, color: '#FF9900' },
    'GCP': { component: SiGooglecloud, color: '#4285F4' },
    'Google Cloud Platform': { component: SiGooglecloud, color: '#4285F4' },
    'Azure': { component: FaCloud, color: '#0078D4' },
    'OpenStack': { component: FaCloud, color: '#ED1944' },
    
    // Kubernetes & Containerization
    'Kubernetes': { component: SiKubernetes, color: '#326CE5' },
    'Docker': { component: SiDocker, color: '#2496ED' },
    'EKS': { component: SiKubernetes, color: '#326CE5' },
    'Kops': { component: SiKubernetes, color: '#326CE5' },
    'Bare-metal': { component: FaRobot, color: '#666666' },
    'Kind': { component: SiKubernetes, color: '#326CE5' },
    'k3s': { component: SiKubernetes, color: '#FFC61C' },
    'Kaniko': { component: SiDocker, color: '#F46800' },
    'Kustomize': { component: SiKubernetes, color: '#326CE5' },
    'Kubernetes Operators': { component: SiKubernetes, color: '#326CE5' },
    
    // Infrastructure as Code
    'Terraform': { component: SiTerraform, color: '#7B42BC' },
    'Pulumi': { component: SiPulumi, color: '#00B4D8' },
    'Ansible': { component: FaRobot, color: '#EE0000' },
    'Chef': { component: FaRobot, color: '#F09820' },
    'Saltstack': { component: FaRobot, color: '#FF6600' },
    
    // CI/CD & DevOps
    'CI/CD': { component: FaRocket, color: '#FF6B6B' },
    'GitHub Actions': { component: SiGithubactions, color: '#2088FF' },
    'GitLab CI/CD': { component: FaRocket, color: '#FC6D26' },
    'Jenkins': { component: FaRobot, color: '#D33833' },
    'ArgoCD': { component: FaRocket, color: '#326CE5' },
    'Flux': { component: SiFlux, color: '#0B122A' },
    'GitOps': { component: FaRocket, color: '#326CE5' },
    'Flagger': { component: FaRocket, color: '#326CE5' },
    'Canary Deployments': { component: FaRocket, color: '#28A745' },
    
    // Monitoring & Observability
    'Prometheus': { component: SiPrometheus, color: '#E6522C' },
    'Grafana': { component: SiGrafana, color: '#F46800' },
    'Loki': { component: FaStream, color: '#F15922' },
    'Tempo': { component: FaDatabase, color: '#E6522C' },
    'Thanos': { component: FaDatabase, color: '#326CE5' },
    'ELK': { component: FaDatabase, color: '#F15922' },
    'EFK': { component: FaDatabase, color: '#F15922' },
    'OpenTelemetry': { component: FaChartBar, color: '#F5A800' },
    'Monitoring': { component: FaChartBar, color: '#F5A800' },
    'Logging': { component: FaStream, color: '#F15922' },
    'Tracing': { component: FaSearch, color: '#326CE5' },
    'Alerting': { component: FaEye, color: '#FF6B6B' },
    'Metrics': { component: FaChartBar, color: '#F5A800' },
    'k6': { component: FaChartBar, color: '#7D64FF' },
    
    // Programming Languages
    'Go': { component: SiGo, color: '#00ADD8' },
    'Golang': { component: SiGo, color: '#00ADD8' },
    'Python': { component: FaRobot, color: '#3776AB' },
    'Bash': { component: FaRobot, color: '#4EAA25' },
    'JavaScript': { component: FaRobot, color: '#F7DF1E' },
    'TypeScript': { component: SiTypescript, color: '#3178C6' },
    
    // Databases & Messaging
    'PostgreSQL': { component: SiPostgresql, color: '#336791' },
    'Redis': { component: SiRedis, color: '#DC382D' },
    'RabbitMQ': { component: SiRabbitmq, color: '#FF6600' },
    'MongoDB': { component: FaDatabase, color: '#47A248' },
    'Kafka': { component: FaDatabase, color: '#231F20' },
    'MinIO': { component: FaDatabase, color: '#C72C48' },
    
    // Distributed Systems
    'Mesos': { component: FaRobot, color: '#E23F2E' },
    'Consul': { component: FaRobot, color: '#DC477D' },
    'Linkerd': { component: FaRobot, color: '#2BEDA7' },
    'Distributed Systems': { component: FaCloud, color: '#326CE5' },
    'Service Mesh': { component: FaRobot, color: '#2BEDA7' },
    'Multi-Cluster': { component: SiKubernetes, color: '#326CE5' },
    
    // Serverless & Platforms
    'Serverless': { component: FaRocket, color: '#FF6B6B' },
    'AWS Lambda': { component: FaRocket, color: '#FF9900' },
    'Knative': { component: SiKnative, color: '#0865AD' },
    'Knative Serving': { component: SiKnative, color: '#0865AD' },
    'Knative Eventing': { component: SiKnative, color: '#0865AD' },
    'CloudEvents': { component: FaCloud, color: '#326CE5' },
    'FaaS': { component: FaRocket, color: '#FF6B6B' },
    'Scale-to-Zero': { component: FaRocket, color: '#28A745' },
    
    // Security
    'Security': { component: FaShieldAlt, color: '#FF6B6B' },
    'Compliance': { component: BiCertification, color: '#28A745' },
    'Network Security': { component: FaShieldAlt, color: '#FF6B6B' },
    'VPN': { component: FaLock, color: '#FF6B6B' },
    'Sealed Secrets': { component: FaLock, color: '#326CE5' },
    'cert-manager': { component: FaLock, color: '#28A745' },
    'Falco': { component: FaShieldAlt, color: '#00AEC7' },
    'mTLS': { component: FaLock, color: '#28A745' },
    
    // AI/ML & LLM
    'RAG': { component: FaRobot, color: '#FF6B6B' },
    'Vertex AI': { component: FaRobot, color: '#4285F4' },
    'Machine Learning': { component: FaRobot, color: '#FF6B6B' },
    'Ollama': { component: FaRobot, color: '#000000' },
    'VLLM': { component: FaRobot, color: '#7C3AED' },
    'LLM': { component: FaRobot, color: '#7C3AED' },
    'AI Agents': { component: FaRobot, color: '#10B981' },
    'DeepSeek': { component: FaRobot, color: '#0066FF' },
    
    // Blockchain & Smart Contracts
    'Ethereum': { component: FaDatabase, color: '#627EEA' },
    'Smart Contracts': { component: FaShieldAlt, color: '#627EEA' },
    'Solidity': { component: FaRobot, color: '#363636' },
    'DeFi': { component: FaDatabase, color: '#627EEA' },
    'Slither': { component: FaShieldAlt, color: '#FF6B6B' },
    'Anvil': { component: FaRobot, color: '#363636' },
    'Foundry': { component: FaRobot, color: '#363636' },
    'Web3': { component: FaCloud, color: '#F16822' },
    'EVM': { component: FaDatabase, color: '#627EEA' },
    
    // Networking
    'Load Balancing': { component: FaRobot, color: '#326CE5' },
    'API Gateway': { component: FaRobot, color: '#326CE5' },
    
    // Management
    'Team Leadership': { component: FaRobot, color: '#FF6B6B' },
    'People Management': { component: FaRobot, color: '#FF6B6B' },
    'Project Management': { component: FaRobot, color: '#FF6B6B' },
    'Agile/Scrum': { component: FaRocket, color: '#FF6B6B' },
    
    // Operations
    'Operations': { component: FaRobot, color: '#666666' },
    'Infrastructure': { component: FaRobot, color: '#666666' },
    'Automation': { component: FaRobot, color: '#666666' },
    'Cloud Operations': { component: FaCloud, color: '#666666' },
    'Infrastructure as Code': { component: FaRobot, color: '#666666' },
    'Site Reliability Engineering': { component: FaRobot, color: '#666666' },
    'DevSecOps': { component: FaShieldAlt, color: '#FF6B6B' },
    'SRE': { component: FaRobot, color: '#666666' },
    
    // Observability
    'Observability': { component: FaEye, color: '#326CE5' },
    'Problem-Solving': { component: FaSearch, color: '#326CE5' },
    'Troubleshooting': { component: FaSearch, color: '#326CE5' },
    'Collaboration': { component: FaRobot, color: '#FF6B6B' },
    
    // General
    'Cloud Migration': { component: FaCloud, color: '#666666' },
    'VMware ESXi': { component: FaRobot, color: '#666666' },
    'Cloud Native Infrastructure': { component: SiKubernetes, color: '#326CE5' },
    'Cloud-Native Infrastructure': { component: SiKubernetes, color: '#326CE5' },
    
    // Default
    'default': { component: FaRobot, color: '#666666' }
  }

  const tech = iconMap[techName] ?? iconMap['default']
  if (!tech) {
    return <FaRobot className="skill-icon" style={{ color: '#666666' }} />
  }
  const IconComponent = tech.component
  return <IconComponent className="skill-icon" style={{ color: tech.color }} />
}

const Resume: React.FC = () => {
  const [experience, setExperience] = useState<Experience[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [siteConfig, setSiteConfig] = useState<SiteConfig | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [data, config] = await Promise.all([
          apiClient.getExperiences(),
          apiClient.getSiteConfig()
        ])
        
        setSiteConfig(config)
        
        // Remove duplicates based on unique combination of title, company, and start_date
        const uniqueData = data.reduce((acc, current) => {
          const key = `${current.title}-${current.company}-${current.start_date}`
          const exists = acc.find(item => 
            `${item.title}-${item.company}-${item.start_date}` === key
          )
          if (!exists) {
            acc.push(current)
          }
          return acc
        }, [] as Experience[])
        
        const sortedData = uniqueData.sort((a, b) => {
          // Sort by order first (highest first), then by start_date (most recent first)
          if (a.order !== b.order) {
            return b.order - a.order
          }
          const dateA = new Date(a.start_date)
          const dateB = new Date(b.start_date)
          return dateB.getTime() - dateA.getTime()
        })
        
        setExperience(sortedData)
        setError(null)
      } catch (err) {
        console.error('Failed to fetch data:', err)
        setError('Failed to fetch experience data')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'long' 
    })
  }

  const formatPeriod = (startDate: string, endDate: string | null, current: boolean) => {
    const start = formatDate(startDate)
    if (current) {
      return `${start} - Present`
    }
    if (endDate) {
      const end = formatDate(endDate)
      return `${start} - ${end}`
    }
    return start
  }



  if (loading) {
    return (
      <div className="resume">
        <div className="container">
          <h1>{siteConfig?.resume_title || 'Bruno Lucena'}</h1>
          <h2>{siteConfig?.resume_subtitle || 'Senior IT Engineer'}</h2>
          <div className="loading">
            <p>Loading professional experience from database...</p>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="resume">
        <div className="container">
          <h1>{siteConfig?.resume_title || 'Bruno Lucena'}</h1>
          <h2>{siteConfig?.resume_subtitle || 'Senior IT Engineer'}</h2>
          <div className="error">
            <p>Error loading experience: {error}</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="resume">
      <div className="container">
        <h1>{siteConfig?.resume_title || 'Bruno Lucena'}</h1>
        <h2>{siteConfig?.resume_subtitle || 'Senior IT Engineer'}</h2>
        
        <section className="resume-section">
          <h3>Professional Experience</h3>
          {experience.length === 0 ? (
            <div className="no-experience">
              <p>No experience data available.</p>
            </div>
          ) : (
            <div className="experience-items">
              {experience.map((exp) => (
                <div key={exp.id} className="experience-item">
                  <div className="experience-header">
                    <h4 className="experience-title">{exp.title}</h4>
                    <span className="experience-company">
                      {exp.company === 'Crealytics' && (
                        <a href="https://www.crealytics.com/" target="_blank" rel="noopener noreferrer" className="company-link">
                          {exp.company}
                        </a>
                      )}
                      {exp.company === 'Tempest Security Intelligence' && (
                        <a href="https://www.tempest.com.br/" target="_blank" rel="noopener noreferrer" className="company-link">
                          {exp.company}
                        </a>
                      )}
                      {exp.company === 'Mobimeo' && (
                        <a href="https://mobimeo.com/en/home-page/" target="_blank" rel="noopener noreferrer" className="company-link">
                          {exp.company}
                        </a>
                      )}
                      {exp.company === 'Notifi' && (
                        <a href="http://notifi.network/" target="_blank" rel="noopener noreferrer" className="company-link">
                          {exp.company}
                        </a>
                      )}
                      {exp.company === 'Namecheap, Inc' && (
                        <a href="https://www.namecheap.com/" target="_blank" rel="noopener noreferrer" className="company-link">
                          {exp.company}
                        </a>
                      )}
                      {exp.company === 'Lesara' && (
                        <a href="https://www.linkedin.com/company/lesara/" target="_blank" rel="noopener noreferrer" className="company-link">
                          {exp.company}
                        </a>
                      )}
                      {!['Crealytics', 'Tempest Security Intelligence', 'Mobimeo', 'Notifi', 'Namecheap, Inc', 'Lesara'].includes(exp.company) && exp.company}
                    </span>
                    <span className="experience-period">
                      {formatPeriod(exp.start_date, exp.end_date ?? null, exp.current)}
                    </span>
                  </div>
                  {exp.company_description && (
                    <div className="experience-company-description">
                      <p className="company-about">
                        {exp.company_description.replace(/^Key Responsibilities:?\s*/i, '').trim()}
                      </p>
                    </div>
                  )}
                  <div className="experience-description">
                    <h5 className="responsibilities-title">Key Responsibilities:</h5>
                    <p>{exp.description}</p>
                  </div>
                  {exp.technologies && exp.technologies.length > 0 && (
                    <div className="experience-technologies">
                      <div className="technology-icons">
                        {exp.technologies.map((tech, index) => (
                          <a 
                            key={index} 
                            href={getTechnologyUrl(tech)}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="technology-icon" 
                            title={tech}
                          >
                            {getTechnologyIcon(tech)}
                          </a>
                        ))}
                      </div>
                    </div>
                                    )}
                  {(!exp.technologies || exp.technologies.length === 0) && (
                    <div className="experience-technologies">
                      <strong>DEBUG: No technologies for {exp.company}</strong>
                      <div>Technologies: {JSON.stringify(exp.technologies)}</div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Comprehensive Technology List */}
        <section className="resume-section">
          <h3>Technologies & Tools</h3>
          <div className="technology-list">
            {(() => {
              // Get all unique technologies from all experiences
              const allTechnologies = Array.from(new Set(
                experience.flatMap(exp => exp.technologies || [])
              )).sort()
              
              return (
                <div className="technology-names">
                  {allTechnologies.map((tech, index) => (
                    <span key={index}>
                      <a 
                        href={getTechnologyUrl(tech)}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="keyword-link"
                        style={{
                          color: '#2563eb',
                          textDecoration: 'underline',
                          fontWeight: '500',
                          backgroundColor: '#f8fafc',
                          padding: '2px 4px',
                          borderRadius: '4px',
                          margin: '0 2px'
                        }}
                      >
                        {tech}
                      </a>
                      {index < allTechnologies.length - 1 && ', '}
                    </span>
                  ))}
                </div>
              )
            })()}
          </div>
        </section>
      </div>
    </div>
  )
}

export default Resume
