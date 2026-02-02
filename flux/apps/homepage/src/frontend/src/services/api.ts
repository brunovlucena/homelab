import axios, { AxiosInstance, AxiosResponse } from 'axios'
// Import shared types from central location
import type { Project, Skill, Experience, Content } from '../types'

// Re-export types for backward compatibility
export type { Project, Skill, Experience, Content }

// =============================================================================
// 📋 API-SPECIFIC TYPES
// =============================================================================

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

export interface SiteConfig {
  hero_title: string
  hero_subtitle: string
  resume_title: string
  resume_subtitle: string
  about_title: string
  about_description: string
}

// =============================================================================
// 🌐 API CLIENT
// =============================================================================

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_API_URL || '/api',
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        // Add auth token if available
        const token = localStorage.getItem('auth_token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        // Handle common errors
        if (error.response?.status === 401) {
          // Handle unauthorized
          localStorage.removeItem('auth_token')
        }
        return Promise.reject(error)
      }
    )
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
  // ⚙️ SITE CONFIG
  // =============================================================================

  async getSiteConfig(): Promise<SiteConfig> {
    const response: AxiosResponse<SiteConfig> = await this.client.get('/config')
    return response.data
  }

  async updateSiteConfig(config: SiteConfig): Promise<void> {
    await this.client.put('/config', config)
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