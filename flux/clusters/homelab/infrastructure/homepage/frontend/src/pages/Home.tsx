import React, { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { apiClient } from '../services/api'
import { Project, Skill } from '../types'
import { getAssetUrl } from '../utils/assets'
import { 
  SiReact, 
  SiTypescript, 
  SiVite, 
  SiTailwindcss, 
  SiReactrouter,
  SiGo,
  SiPostgresql,
  SiRedis,
  SiKubernetes,
  SiFlux,
  SiArgo,
  SiHelm,
  SiNginx,
  SiDocker,
  SiGithub,
  SiGithubactions,
  SiPulumi,
  SiPrometheus,
  SiGrafana,
  SiAmazon,
  SiGooglecloud,
  SiTerraform,
  SiOpentelemetry,
  SiKnative,
  SiRabbitmq
} from 'react-icons/si'
import { FaSearch, FaClock, FaDatabase, FaStream, FaEye } from 'react-icons/fa'
import { FaShieldAlt, FaLock, FaChartBar, FaCloud, FaRobot, FaRocket } from 'react-icons/fa'
import { BiCertification } from 'react-icons/bi'

const Home: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchParams] = useSearchParams()

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        setLoading(true)
        const fetchedProjects = await apiClient.getProjects()
        setProjects(fetchedProjects || [])
        setError(null)
      } catch (err) {
        console.error('Failed to fetch projects:', err)
        setError('Failed to load projects')
        setProjects([]) // Ensure projects is always an array
      } finally {
        setLoading(false)
      }
    }

    fetchProjects()
  }, [])

  // Handle scrolling to projects section when section parameter is present
  useEffect(() => {
    const section = searchParams.get('section')
    if (section === 'projects') {
      const projectsElement = document.getElementById('projects')
      if (projectsElement) {
        projectsElement.scrollIntoView({ behavior: 'smooth' })
      }
    }
  }, [searchParams])

  const getIconComponent = (iconName: string) => {
    const iconMap: { [key: string]: React.ComponentType<any> } = {
      'react': SiReact,
      'typescript': SiTypescript,
      'vite': SiVite,
      'tailwind': SiTailwindcss,
      'router': SiReactrouter,
      'go': SiGo,
      'postgresql': SiPostgresql,
      'redis': SiRedis,
      'kubernetes': SiKubernetes,
      'knative': SiKnative,
      'flux': SiFlux,
      'argocd': SiArgo,
      'helm': SiHelm,
      'nginx': SiNginx,
      'docker': SiDocker,
      'github': SiGithub,
      'githubactions': SiGithubactions,
      'pulumi': SiPulumi,
      'prometheus': SiPrometheus,
      'grafana': SiGrafana,
      'loki': FaStream,
      'tempo': FaDatabase,
      'opentelemetry': FaChartBar,
      'terraform': SiTerraform,
      'aws': SiAmazon,
      'gcp': SiGooglecloud,
      'rabbitmq': SiRabbitmq,
      'shield': FaShieldAlt,
      'certification': BiCertification,
      'pydantic': FaRobot,
      'logfire': FaRobot,
      'vertexai': FaCloud,
      'langchain': FaRobot,
      'langgraph': FaRocket,
      'flyte': FaStream,
      'wandb': FaChartBar,
      'kamaji': SiKubernetes,
      'mcp': FaRobot,
    }
    return iconMap[iconName.toLowerCase()] || SiGithub
  }

  const isVideoUrl = (url: string): boolean => {
    const videoDomains = [
      'youtube.com',
      'youtu.be',
      'vimeo.com',
      'dailymotion.com',
      'twitch.tv'
    ]
    try {
      const urlObj = new URL(url)
      return videoDomains.some(domain => urlObj.hostname.includes(domain))
    } catch {
      return false
    }
  }

  const getVideoEmbedUrl = (url: string): string => {
    try {
      const urlObj = new URL(url)
      
      // YouTube
      if (urlObj.hostname.includes('youtube.com') || urlObj.hostname.includes('youtu.be')) {
        const videoId = urlObj.searchParams.get('v') || urlObj.pathname.slice(1)
        return `https://www.youtube.com/embed/${videoId}`
      }
      
      // Vimeo
      if (urlObj.hostname.includes('vimeo.com')) {
        const videoId = urlObj.pathname.slice(1)
        return `https://player.vimeo.com/video/${videoId}`
      }
      
      // Dailymotion
      if (urlObj.hostname.includes('dailymotion.com')) {
        const videoId = urlObj.pathname.split('/').pop() || ''
        return `https://www.dailymotion.com/embed/video/${videoId}`
      }
      
      // Twitch
      if (urlObj.hostname.includes('twitch.tv')) {
        const videoId = urlObj.pathname.split('/').pop() || ''
        return `https://player.twitch.tv/?video=v${videoId}&parent=${window.location.hostname}`
      }
      
      return url
    } catch {
      return url
    }
  }

  return (
    <div className="home">
      <section className="hero">
        <div className="container">
          <h1>Agentic Team Lead</h1>
          <p>SRE • DevSecOps • AI/ML Infrastructure • Platform Engineering</p>
        </div>
      </section>
      
      <section id="about" className="section">
        <div className="container">
          <div className="about-header">
            <div className="about-intro">
              <div className="about-image-container">
                <picture>
                  <source 
                    srcSet={getAssetUrl('eu.webp')} 
                    type="image/webp" 
                  />
                  <img 
                    src={getAssetUrl('eu.webp')} 
                    alt="Bruno Lucena" 
                    className="about-image"
                    loading="eager"
                    decoding="async"
                    fetchPriority="high"
                    onLoad={(e) => {
                      e.currentTarget.style.opacity = '1';
                    }}
                    style={{
                      opacity: '0',
                      transition: 'opacity 0.3s ease-in-out'
                    }}
                  />
                </picture>
              </div>
              <h2>About Me</h2>
              <p>Senior Agentic Team Lead with 10+ years of experience designing, building, and scaling mission-critical cloud-native platforms. Expert in Kubernetes ecosystem, multi-cloud architectures (AWS/GCP), and modern observability stacks. Proven track record in Site Reliability Engineering (SRE), DevSecOps practices, and leading high-performing infrastructure teams. Specialized in AI/ML infrastructure, LLMOps, and building resilient systems that handle millions of requests. Passionate about automation, security-first approaches, and driving innovation in cloud-native technologies. Delivers enterprise-grade solutions with 99.9%+ uptime, optimized performance, and comprehensive security postures.</p>
            </div>
          </div>
          
          <div className="skills-grid">
            <div className="skill-tag">
              <SiGo className="skill-icon" style={{ color: '#00ADD8' }} />
              <span>Go</span>
            </div>
            <div className="skill-tag">
              <SiKubernetes className="skill-icon" style={{ color: '#326CE5' }} />
              <span>Kubernetes</span>
            </div>
            <div className="skill-tag">
              <SiDocker className="skill-icon" style={{ color: '#2496ED' }} />
              <span>Docker</span>
            </div>
            <div className="skill-tag">
              <SiPostgresql className="skill-icon" style={{ color: '#336791' }} />
              <span>PostgreSQL</span>
            </div>
            <div className="skill-tag">
              <SiRedis className="skill-icon" style={{ color: '#DC382D' }} />
              <span>Redis</span>
            </div>
            <div className="skill-tag">
              <SiPrometheus className="skill-icon" style={{ color: '#E6522C' }} />
              <span>Prometheus</span>
            </div>
            <div className="skill-tag">
              <SiGrafana className="skill-icon" style={{ color: '#F46800' }} />
              <span>Grafana</span>
            </div>
            <div className="skill-tag">
              <SiPulumi className="skill-icon" style={{ color: '#00B4D8' }} />
              <span>Pulumi</span>
            </div>
            <div className="skill-tag">
              <SiAmazon className="skill-icon" style={{ color: '#FF9900' }} />
              <span>AWS</span>
            </div>
            <div className="skill-tag">
              <SiGooglecloud className="skill-icon" style={{ color: '#4285F4' }} />
              <span>Google Cloud</span>
            </div>
            <div className="skill-tag">
              <SiGithub className="skill-icon" style={{ color: '#181717' }} />
              <span>GitHub</span>
            </div>
            <div className="skill-tag">
              <SiGithubactions className="skill-icon" style={{ color: '#2088FF' }} />
              <span>GitHub Actions</span>
            </div>
            <div className="skill-tag">
              <SiTerraform className="skill-icon" style={{ color: '#7B42BC' }} />
              <span>Terraform</span>
            </div>
            <div className="skill-tag">
              <FaChartBar className="skill-icon" style={{ color: '#F5A800' }} />
              <span>OpenTelemetry</span>
            </div>
            <div className="skill-tag">
              <FaStream className="skill-icon" style={{ color: '#F15922' }} />
              <span>Loki</span>
            </div>
            <div className="skill-tag">
              <FaDatabase className="skill-icon" style={{ color: '#E6522C' }} />
              <span>Tempo</span>
            </div>
            <div className="skill-tag">
              <SiKnative className="skill-icon" style={{ color: '#0865AD' }} />
              <span>Knative</span>
            </div>
            <div className="skill-tag">
              <SiRabbitmq className="skill-icon" style={{ color: '#FF6600' }} />
              <span>RabbitMQ</span>
            </div>
            <div className="skill-tag">
              <SiFlux className="skill-icon" style={{ color: '#0B122A' }} />
              <span>Flux</span>
            </div>
            <div className="skill-tag">
              <SiArgo className="skill-icon" style={{ color: '#326CE5' }} />
              <span>ArgoCD</span>
            </div>
            <div className="skill-tag">
              <FaRobot className="skill-icon" style={{ color: '#FF6B35' }} />
              <span>Pydantic Logfire</span>
            </div>
            <div className="skill-tag">
              <FaCloud className="skill-icon" style={{ color: '#4285F4' }} />
              <span>Vertex AI</span>
            </div>
            <div className="skill-tag">
              <FaRobot className="skill-icon" style={{ color: '#1C3C3C' }} />
              <span>Langchain</span>
            </div>
            <div className="skill-tag">
              <FaRocket className="skill-icon" style={{ color: '#FF6B35' }} />
              <span>Langgraph</span>
            </div>
            <div className="skill-tag">
              <FaStream className="skill-icon" style={{ color: '#E6522C' }} />
              <span>Flyte</span>
            </div>
            <div className="skill-tag">
              <FaChartBar className="skill-icon" style={{ color: '#FF6B35' }} />
              <span>Wandb</span>
            </div>
            <div className="skill-tag">
              <SiKubernetes className="skill-icon" style={{ color: '#326CE5' }} />
              <span>Kamaji</span>
            </div>
            <div className="skill-tag">
              <FaRobot className="skill-icon" style={{ color: '#00B4D8' }} />
              <span>MCP-Servers</span>
            </div>
          </div>
        </div>
      </section>

      <section id="projects" className="section">
        <div className="container">
          <h2>Homelab</h2>
          {loading && (
            <div className="loading">
              <p>Loading homelab projects...</p>
            </div>
          )}
          
          {error && (
            <div className="error">
              <p>Error: {error}</p>
            </div>
          )}
          
          {!loading && !error && (!projects || projects.length === 0) && (
            <div className="no-projects">
              <p>No homelab projects available at the moment.</p>
            </div>
          )}
          
          {!loading && !error && projects && projects.length > 0 && (
            <div className="projects-grid">
              {projects.map((project) => {
                // Skip rendering if project is malformed
                if (!project || !project.id || !project.title) {
                  return null;
                }
                
                const IconComponent = getIconComponent((project.technologies && project.technologies[0]) || 'github')
                return (
                  <div key={project.id} className="project-card">
                    <div className="project-header">
                      <IconComponent className="project-icon" />
                      <h3>{project.title}</h3>
                    </div>
                    <p className="project-description">{project.description || ''}</p>
                    
                    {/* Video Embed - if live_url is a video URL */}
                    {project.live_url && isVideoUrl(project.live_url) && (
                      <div className="project-video">
                        <iframe
                          src={getVideoEmbedUrl(project.live_url)}
                          title={`${project.title} Video`}
                          frameBorder="0"
                          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                          allowFullScreen
                          loading="lazy"
                        ></iframe>
                      </div>
                    )}
                    
                    <div className="project-meta">
                      <span className="project-type">{project.type || ''}</span>
                      {project.github_url && (
                        (project.github_active !== undefined ? project.github_active : true) ? (
                          <a 
                            href={project.github_url} 
                            target="_blank" 
                            rel="noopener noreferrer"
                            className="project-link"
                          >
                            <SiGithub className="project-link-icon" />
                            <span>GitHub</span>
                          </a>
                        ) : (
                          <span className="project-link project-link-disabled">
                            <SiGithub className="project-link-icon" />
                            <span>GitHub</span>
                          </span>
                        )
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </section>

      <footer className="footer">
        <div className="container">
          <div className="footer-content">
            <div className="footer-tech">
              <p>This site was built with:</p>
              <div className="footer-icons">
                <SiReact className="footer-icon" style={{ color: '#61DAFB' }} />
                <SiTypescript className="footer-icon" style={{ color: '#3178C6' }} />
                <SiVite className="footer-icon" style={{ color: '#646CFF' }} />
                <SiTailwindcss className="footer-icon" style={{ color: '#06B6D4' }} />
              </div>
            </div>
            <div className="footer-links">
              <a href="https://github.com/brunovlucena" target="_blank" rel="noopener noreferrer">
                <SiGithub className="footer-link-icon" />
                GitHub
              </a>
              <a href="https://www.linkedin.com/in/bvlucena" target="_blank" rel="noopener noreferrer">
                <SiGithub className="footer-link-icon" />
                LinkedIn
              </a>
            </div>
          </div>
        </div>
      </footer>

    </div>
  )
}

export default Home
