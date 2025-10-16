import axios, { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { env } from '../utils/env'
import { metricsCollector } from '../utils/metrics'

// =============================================================================
// 📋 TYPES
// =============================================================================

// Extend AxiosRequestConfig to include metadata
interface RequestConfig extends InternalAxiosRequestConfig {
  metadata?: { startTime: number }
}

// =============================================================================
// 📋 TYPES
// =============================================================================

export interface Project {
  id: number
  title: string
  description: string
  short_description: string
  type: string
  icon: string
  github_url: string
  live_url: string
  technologies: string[]
  active: boolean
  github_active: boolean
}

export interface Skill {
  id: number
  name: string
  category: string
  proficiency: number
  icon: string
  order: number
}

export interface Experience {
  id: number
  title: string
  company: string
  start_date: string
  end_date?: string
  current: boolean
  description: string
  technologies: string[]
  order: number
  active: boolean
}

export interface AboutData {
  description: string
  highlights: Array<{
    icon: string
    text: string
  }>
}

export interface ContactData {
  email: string
  location: string
  linkedin: string
  github: string
  availability: string
}

export interface Content {
  id: number
  type: string
  value: string
}

// =============================================================================
// 🌐 API CLIENT
// =============================================================================

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: env.API_URL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor
    this.client.interceptors.request.use(
      (config: RequestConfig) => {
        // Add auth token if available
        const token = localStorage.getItem('auth_token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        
        // 📊 Add metrics tracking metadata
        config.metadata = { startTime: Date.now() }
        
        return config
      },
      (error) => {
        // Track request errors
        metricsCollector.recordError(
          'api_request_error',
          error?.message || 'Request setup failed',
          error?.stack
        )
        return Promise.reject(error)
      }
    )

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => {
        // 📊 Track successful API calls
        const config = response.config as RequestConfig
        if (config.metadata?.startTime) {
          const duration = Date.now() - config.metadata.startTime
          const endpoint = this.getEndpointPath(config.url || '')
          const method = config.method?.toUpperCase() || 'GET'
          
          metricsCollector.recordAPICall(
            endpoint,
            method,
            response.status,
            duration,
            true
          )
        }
        
        return response
      },
      (error) => {
        // 📊 Track failed API calls
        const config = error.config as RequestConfig
        if (config?.metadata?.startTime) {
          const duration = Date.now() - config.metadata.startTime
          const endpoint = this.getEndpointPath(config.url || '')
          const method = config.method?.toUpperCase() || 'GET'
          const status = error.response?.status || 0
          
          metricsCollector.recordAPICall(
            endpoint,
            method,
            status,
            duration,
            false
          )
        }
        
        // Handle common errors
        if (error.response?.status === 401) {
          // Handle unauthorized
          localStorage.removeItem('auth_token')
        }
        
        return Promise.reject(error)
      }
    )
  }

  // Helper method to extract endpoint path from URL
  private getEndpointPath(url: string): string {
    try {
      // Remove base URL if present
      const path = url.replace(env.API_URL, '')
      // Remove query parameters
      return path.split('?')[0]
    } catch {
      return url
    }
  }

  // =============================================================================
  // 🎯 PROJECTS
  // =============================================================================

  async getProjects(): Promise<Project[]> {
    const response: AxiosResponse<Project[]> = await this.client.get('/projects')
    return response.data
  }

  async getProject(id: number): Promise<Project> {
    const response: AxiosResponse<Project> = await this.client.get(`/projects/${id}`)
    return response.data
  }

  async createProject(project: Omit<Project, 'id'>): Promise<Project> {
    const response: AxiosResponse<Project> = await this.client.post('/projects', project)
    return response.data
  }

  async updateProject(id: number, project: Partial<Project>): Promise<void> {
    await this.client.put(`/projects/${id}`, project)
  }

  async deleteProject(id: number): Promise<void> {
    await this.client.delete(`/projects/${id}`)
  }

  // =============================================================================
  // 🛠️ SKILLS
  // =============================================================================

  async getSkills(): Promise<Skill[]> {
    const response: AxiosResponse<Skill[]> = await this.client.get('/skills')
    return response.data
  }

  async getSkill(id: number): Promise<Skill> {
    const response: AxiosResponse<Skill> = await this.client.get(`/skills/${id}`)
    return response.data
  }

  async createSkill(skill: Omit<Skill, 'id'>): Promise<Skill> {
    const response: AxiosResponse<Skill> = await this.client.post('/skills', skill)
    return response.data
  }

  async updateSkill(id: number, skill: Partial<Skill>): Promise<void> {
    await this.client.put(`/skills/${id}`, skill)
  }

  async deleteSkill(id: number): Promise<void> {
    await this.client.delete(`/skills/${id}`)
  }

  // =============================================================================
  // 💼 EXPERIENCES
  // =============================================================================

  async getExperiences(): Promise<Experience[]> {
    const response: AxiosResponse<Experience[]> = await this.client.get('/experiences')
    return response.data
  }

  async getExperience(id: number): Promise<Experience> {
    const response: AxiosResponse<Experience> = await this.client.get(`/experiences/${id}`)
    return response.data
  }

  async createExperience(experience: Omit<Experience, 'id'>): Promise<Experience> {
    const response: AxiosResponse<Experience> = await this.client.post('/experiences', experience)
    return response.data
  }

  async updateExperience(id: number, experience: Partial<Experience>): Promise<void> {
    await this.client.put(`/experiences/${id}`, experience)
  }

  async deleteExperience(id: number): Promise<void> {
    await this.client.delete(`/experiences/${id}`)
  }

  // =============================================================================
  // 📄 CONTENT
  // =============================================================================

  async getContent(): Promise<Content[]> {
    const response: AxiosResponse<Content[]> = await this.client.get('/content')
    return response.data
  }

  async getContentByType(type: string): Promise<Content[]> {
    const response: AxiosResponse<Content[]> = await this.client.get(`/content/${type}`)
    return response.data
  }

  async createContent(content: Omit<Content, 'id'>): Promise<Content> {
    const response: AxiosResponse<Content> = await this.client.post('/content', content)
    return response.data
  }

  async updateContent(id: number, content: Partial<Content>): Promise<void> {
    await this.client.put(`/content/${id}`, content)
  }

  async deleteContent(id: number): Promise<void> {
    await this.client.delete(`/content/${id}`)
  }

  // =============================================================================
  // 👤 ABOUT
  // =============================================================================

  async getAbout(): Promise<AboutData> {
    const response: AxiosResponse<AboutData> = await this.client.get('/about')
    return response.data
  }

  async updateAbout(about: AboutData): Promise<void> {
    await this.client.put('/about', about)
  }

  // =============================================================================
  // 📞 CONTACT
  // =============================================================================

  async getContact(): Promise<ContactData> {
    const response: AxiosResponse<ContactData> = await this.client.get('/contact')
    return response.data
  }

  async updateContact(contact: ContactData): Promise<void> {
    await this.client.put('/contact', contact)
  }

  // =============================================================================
  // 🏥 HEALTH
  // =============================================================================

  async healthCheck(): Promise<{ status: string; timestamp: string; service: string }> {
    const response = await this.client.get('/health')
    return response.data
  }
}

// =============================================================================
// 📤 EXPORTS
// =============================================================================

export const apiClient = new ApiClient()
export default apiClient 