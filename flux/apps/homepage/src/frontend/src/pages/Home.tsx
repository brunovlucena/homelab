import React, { useState, useEffect, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { apiClient } from '../services/api'
import type { SiteConfig } from '../services/api'
import type { Project, Skill } from '../types'
import SkillsGrid from '../components/SkillsGrid'
import { getAssetUrl } from '../utils'
// Icons used in project cards and footer
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
  SiTerraform,
  SiKnative,
  SiRabbitmq,
  SiPython
} from 'react-icons/si'
import { FaDatabase, FaStream, FaShieldAlt, FaChartBar, FaCloud, FaRobot, FaRocket } from 'react-icons/fa'
import { BiCertification } from 'react-icons/bi'

const Home: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([])
  const [skills, setSkills] = useState<Skill[]>([])
  const [loading, setLoading] = useState(true)
  const [skillsLoading, setSkillsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchParams] = useSearchParams()
  const [siteConfig, setSiteConfig] = useState<SiteConfig | null>(null)

  const fetchData = useCallback(async () => {
    try {
      setLoading(true)
      setSkillsLoading(true)
      
      // Use Promise.allSettled to handle individual failures gracefully
      const [projectsResult, configResult, skillsResult] = await Promise.allSettled([
        apiClient.getProjects().catch(err => {
          console.warn('[Home] Failed to fetch projects:', err)
          return []
        }),
        apiClient.getSiteConfig().catch(err => {
          console.warn('[Home] Failed to fetch site config:', err)
          return null
        }),
        apiClient.getSkills().catch(err => {
          console.warn('[Home] Failed to fetch skills:', err)
          return []
        })
      ])
      
      setProjects(projectsResult.status === 'fulfilled' ? (projectsResult.value ?? []) : [])
      setSiteConfig(configResult.status === 'fulfilled' ? configResult.value : null)
      setSkills(skillsResult.status === 'fulfilled' ? (skillsResult.value ?? []) : [])
      
      // Only set error if ALL requests failed
      const allFailed = projectsResult.status === 'rejected' && 
                       configResult.status === 'rejected' && 
                       skillsResult.status === 'rejected'
      if (allFailed) {
        setError('Failed to load data from API')
      } else {
        setError(null)
      }
    } catch (err) {
      console.error('[Home] Unexpected error fetching data:', err)
      setError('Failed to load data')
      setProjects([])
    } finally {
      setLoading(false)
      setSkillsLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

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
      'python': SiPython,
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
      'rabbitmq': SiRabbitmq,
      'shield': FaShieldAlt,
      'certification': BiCertification,
      // New additions from homelab projects
      'flagger': FaRocket,
      'linkerd': FaRobot,
      'cloudevents': FaCloud,
      'minio': FaDatabase,
      'ollama': FaRobot,
      'ai': FaRobot,
      'aiagents': FaRobot,
      'smartcontracts': FaShieldAlt,
      'slither': FaShieldAlt,
      'ethereum': FaDatabase,
      'solidity': FaRobot,
      'defi': FaDatabase,
      'kaniko': SiDocker,
      'kustomize': SiKubernetes,
      'kind': SiKubernetes,
      'k3s': SiKubernetes,
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

  const isYouTubeUrl = (url: string): boolean => {
    try {
      const urlObj = new URL(url)
      return urlObj.hostname.includes('youtube.com') || urlObj.hostname.includes('youtu.be')
    } catch {
      return false
    }
  }

  const getYouTubeVideoId = (url: string): string | null => {
    try {
      const urlObj = new URL(url)
      if (urlObj.hostname.includes('youtube.com')) {
        return urlObj.searchParams.get('v')
      }
      if (urlObj.hostname.includes('youtu.be')) {
        return urlObj.pathname.slice(1)
      }
      return null
    } catch {
      return null
    }
  }

  const getVideoEmbedUrl = (url: string): string => {
    try {
      const urlObj = new URL(url)
      
      // YouTube - handled by lite-youtube-embed
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
          <h1>IT Engineer</h1>
          <p>SRE • DevSecOps • AI Infrastructure • Platform Engineering</p>
        </div>
      </section>
      
      <section id="about" className="section">
        <div className="container">
          <h2>{siteConfig?.about_title || 'About Me'}</h2>
          <div className="about-content">
            <div className="about-image-container">
              <picture>
                <source srcSet={getAssetUrl('assets/eu.webp')} type="image/webp" />
                <img 
                  src={getAssetUrl('assets/eu.png')} 
                  alt="Bruno Lucena" 
                  className="about-image"
                  loading="lazy"
                />
              </picture>
            </div>
            <p className="about-description" style={{}}>IT Engineer with 15+ years of diverse experience spanning Computer and Network Technician, IT Security Analyst, Project Manager, DevOps Engineer, and SRE Lead roles. I have a proven track record of architecting and operating production-grade systems and observability infrastructure using AWS, Baremetal, GCP, Prometheus, Loki, Tempo, Alloy, Mimir, OpenTelemetry, and Grafana.

I also have extensive experience establishing systems from the ground up - from prototyping on Raspberry Pi to production multi-region Kubernetes clusters, from mobile applications to distributed cloud infrastructure. I've built comprehensive observability platforms through sophisticated automation using both traditional Terraform and modern Infrastructure-as-Code tools like Pulumi. Currently, I'm developing agent-sre, an AI-powered system that automatically responds to alerts by following runbooks, significantly reducing manual toil and enabling faster incident resolution.</p>
          </div>
        </div>
      </section>

      <section id="projects" className="section">
        <div className="container">
          <h2>Projects</h2>
          {loading && (
            <div className="loading">
              <p>Loading projects...</p>
            </div>
          )}
          
          {error && (
            <div className="error">
              <p>Error: {error}</p>
            </div>
          )}
          
          {!loading && !error && (!projects || projects.length === 0) && (
            <div className="no-projects">
              <p>No projects available at the moment.</p>
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
                    
                    {/* Video Embed - YouTube uses lite-youtube-embed for 77% faster loading */}
                    {project.live_url && isVideoUrl(project.live_url) && (
                      <div className="project-video">
                        {isYouTubeUrl(project.live_url) ? (
                          // @ts-expect-error - lite-youtube is a web component
                          <lite-youtube 
                            videoid={getYouTubeVideoId(project.live_url)}
                            playlabel={`Play ${project.title}`}
                            style={{ backgroundImage: 'none' }}
                          />
                        ) : (
                          <iframe
                            src={getVideoEmbedUrl(project.live_url)}
                            title={`${project.title} Video`}
                            frameBorder="0"
                            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                            allowFullScreen
                            loading="lazy"
                          ></iframe>
                        )}
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

      <section id="technologies" className="section">
        <div className="container">
          <h2>Technologies</h2>
          <SkillsGrid skills={skills} loading={skillsLoading} />
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
