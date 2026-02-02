import { apiClient } from './api';

export interface ChatbotResponse {
  text: string;
  suggestions?: string[];
  action?: 'navigate' | 'show_project' | 'show_contact';
  data?: any;
}

export interface LLMChatRequest {
  message: string;
  context?: string;
}

export interface LLMChatResponse {
  response: string;
  sources?: string[];
  model: string;
  timestamp: string;
}

export class ChatbotService {
  private static instance: ChatbotService;
  private projects: any[] = [];
  private skills: any[] = [];
  private experience: any[] = [];
  private about: any = null;
  private useLLM: boolean = true; // Toggle between LLM and rule-based responses

  private constructor() {}

  static getInstance(): ChatbotService {
    if (!ChatbotService.instance) {
      ChatbotService.instance = new ChatbotService();
    }
    return ChatbotService.instance;
  }

  async initialize(): Promise<void> {
    try {
      console.log('🔄 ChatbotService: Starting initialization...');
      
      // Load data in parallel with better error handling
      console.log('📡 ChatbotService: Loading projects...');
      const [projectsData, skillsData, experienceData, aboutData] = await Promise.allSettled([
        apiClient.getProjects(),
        apiClient.getSkills(),
        apiClient.getAbout(),
        apiClient.getExperiences()
      ]);

      console.log('📊 ChatbotService: Processing results...');
      console.log('  - Projects status:', projectsData.status, 'value:', projectsData.status === 'fulfilled' ? projectsData.value : 'rejected');
      console.log('  - Skills status:', skillsData.status, 'value:', skillsData.status === 'fulfilled' ? skillsData.value : 'rejected');
      console.log('  - Experience status:', experienceData.status, 'value:', experienceData.status === 'fulfilled' ? experienceData.value : 'rejected');
      console.log('  - About status:', aboutData.status, 'value:', aboutData.status === 'fulfilled' ? aboutData.value : 'rejected');

      this.projects = projectsData.status === 'fulfilled' ? (projectsData.value as unknown as any[]) || [] : [];
      this.skills = skillsData.status === 'fulfilled' ? (skillsData.value as unknown as any[]) || [] : [];
      this.experience = experienceData.status === 'fulfilled' ? (experienceData.value as unknown as any[]) || [] : [];
      this.about = aboutData.status === 'fulfilled' ? (aboutData.value as unknown as any) || { key: 'about', value: { description: '' } } : { key: 'about', value: { description: '' } };

      console.log('✅ ChatbotService: Data loaded successfully:', {
        projects: this.projects?.length || 0,
        skills: this.skills?.length || 0,
        experience: this.experience?.length || 0,
        hasAbout: !!this.about?.value
      });

      if (this.projects?.length > 0) {
        console.log('📋 ChatbotService: Sample project:', this.projects[0]);
      }
    } catch (error) {
      console.error('❌ ChatbotService: Failed to load data:', error);
      // Initialize with empty data to prevent errors
      this.projects = [];
      this.skills = [];
      this.experience = [];
      this.about = { value: '' };
    }
  }

  async processMessage(userInput: string): Promise<ChatbotResponse> {
    const input = userInput.toLowerCase().trim();

    // Use LLM for responses if enabled
    if (this.useLLM) {
      try {
        console.log('🤖 Using LLM for response generation...');
        const llmResponse = await this.processWithLLM(userInput);
        return {
          text: llmResponse.response,
          suggestions: this.getContextualSuggestions(input),
          data: {
            model: llmResponse.model,
            timestamp: llmResponse.timestamp,
            sources: llmResponse.sources
          }
        };
      } catch (error) {
        console.error('❌ Agent-Bruno processing failed:', error);
        // Return error message instead of falling back to rule-based
        return {
          text: "I'm sorry, but the AI service (agent-bruno) is currently unavailable. Please try again later or contact Bruno directly for assistance.",
          suggestions: ['Try again', 'Contact Bruno directly', 'Check back later']
        };
      }
    }
    
    // Use rule-based responses only when explicitly disabled
    return this.processWithRules(input);
  }

  private async processWithLLM(userInput: string): Promise<LLMChatResponse> {
    const response = await fetch('/api/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        message: userInput
      } as LLMChatRequest),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(`LLM API error: ${errorData.error || response.statusText}`);
    }

    return response.json();
  }

  private processWithRules(input: string): ChatbotResponse {
    // Experience and work history
    if (this.matchesKeywords(input, ['experience', 'work', 'job', 'career', 'background'])) {
      return this.handleExperienceQuery();
    }

    // Projects
    if (this.matchesKeywords(input, ['project', 'work', 'site', 'github'])) {
      return this.handleProjectsQuery();
    }

    // Skills and technologies
    if (this.matchesKeywords(input, ['skill', 'technology', 'tech', 'stack', 'tools', 'languages'])) {
      return this.handleSkillsQuery();
    }

    // Contact information
    if (this.matchesKeywords(input, ['contact', 'email', 'reach', 'get in touch', 'hire', 'available'])) {
      return this.handleContactQuery();
    }
    
    // Resume
    if (this.matchesKeywords(input, ['resume', 'cv', 'education', 'certification'])) {
      return this.handleResumeQuery();
    }
    
    // About
    if (this.matchesKeywords(input, ['about', 'who', 'introduce', 'tell me about'])) {
      return this.handleAboutQuery();
    }
    
    // Greetings
    if (this.matchesKeywords(input, ['hello', 'hi', 'hey', 'good morning', 'good afternoon'])) {
      return {
        text: "I can help you learn more about Bruno's experience, projects, skills, and how to get in touch. What would you like to know?",
        suggestions: ['Tell me about his experience', 'Show me his projects', 'What are his skills?', 'How can I contact him?']
      };
    }
    
    // Help
    if (this.matchesKeywords(input, ['help', 'what can you do', 'commands', 'options'])) {
      return {
        text: "I can help you with information about Bruno's professional background. Here are some things you can ask me about:",
        suggestions: ['Experience & Work History', 'Projects & Site', 'Skills & Technologies', 'Contact Information', 'Resume & Education']
      };
    }
    
    // Default response
    return {
      text: "That's an interesting question! Bruno has a diverse background in cloud infrastructure, AI/ML, and DevOps. Could you be more specific about what you'd like to know? I can help with his experience, projects, skills, or contact information.",
      suggestions: ['Tell me about his experience', 'Show me his projects', 'What are his skills?', 'How can I contact him?']
    };
  }

  private matchesKeywords(input: string, keywords: string[]): boolean {
    return keywords.some(keyword => input.includes(keyword));
  }

  private handleExperienceQuery(): ChatbotResponse {
    if (!this.experience || this.experience.length === 0) {
      return {
        text: "Bruno has 12+ years of experience in SRE, DevSecOps, and AI Engineering. He's worked with AWS, Kubernetes, and has extensive experience in infrastructure automation and AI/ML technologies.",
        suggestions: ['Tell me about specific roles', 'What companies has he worked for?', 'Show me his skills']
      };
    }

    const recentExperience = (this.experience as any[]).slice(0, 3);
    const experienceText = recentExperience.map(exp => 
      `${exp.title} at ${exp.company} (${exp.period})`
    ).join(', ');

    return {
      text: `Bruno has extensive experience including: ${experienceText}. He specializes in cloud-native infrastructure, AI/ML, and DevOps automation. Would you like to know about specific roles or technologies?`,
      suggestions: ['Tell me about specific roles', 'What are his key skills?', 'Show me his projects']
    };
  }

  private handleProjectsQuery(): ChatbotResponse {
    if (!this.projects || this.projects.length === 0) {
      return {
        text: "Bruno has worked on several interesting projects including cloud-native infrastructure, AI/ML implementations, and DevOps automation. Some highlights include Kubernetes cluster management, CI/CD pipelines, and AI model deployment.",
        suggestions: ['Tell me about his experience', 'What are his skills?', 'How can I contact him?']
      };
    }

    const activeProjects = (this.projects as any[]).filter(p => p.active).slice(0, 3);
    const projectNames = activeProjects.map(p => p.title).join(', ');

    return {
      text: `Bruno has worked on various projects including: ${projectNames}. These cover areas like cloud infrastructure, AI/ML, and automation. Which area interests you most?`,
      suggestions: ['Tell me about specific projects', 'What technologies does he use?', 'Show me his experience']
    };
  }

  private handleSkillsQuery(): ChatbotResponse {
    if (!this.skills || this.skills.length === 0) {
      return {
        text: "Bruno's key skills include Kubernetes, Docker, AWS, Terraform, Python, Go, AI/ML, CI/CD, monitoring, and observability. He's also experienced with AI agents, Knative, and cloud-native technologies.",
        suggestions: ['Tell me about his experience', 'Show me his projects', 'How can I contact him?']
      };
    }

    const skillCategories = (this.skills as any[]).reduce((acc, skill) => {
      if (!acc[skill.category]) {
        acc[skill.category] = [];
      }
      acc[skill.category].push(skill.name);
      return acc;
    }, {} as Record<string, string[]>);

    const skillText = Object.entries(skillCategories)
      .map(([category, skills]) => `${category}: ${(skills as string[]).slice(0, 3).join(', ')}`)
      .join('; ');

    return {
      text: `Bruno's skills include: ${skillText}. He has expertise across cloud infrastructure, AI/ML, and DevOps. What specific technology would you like to know more about?`,
      suggestions: ['Tell me about his experience', 'Show me his projects', 'What are his certifications?']
    };
  }

  private handleContactQuery(): ChatbotResponse {
    return {
      text: "You can reach Bruno through LinkedIn, GitHub, or email. He's currently available for new opportunities and consulting work. Would you like me to provide specific contact information or discuss his availability?",
      action: 'show_contact',
      suggestions: ['Tell me about his experience', 'Show me his projects', 'What are his skills?']
    };
  }

  private handleResumeQuery(): ChatbotResponse {
    return {
      text: "Bruno's resume includes his extensive experience in cloud infrastructure, his work with major tech companies, and his expertise in AI/ML. You can view his detailed resume on the resume page, or I can highlight specific aspects of his background.",
      action: 'navigate',
      data: { path: '/resume' },
      suggestions: ['Tell me about his experience', 'Show me his projects', 'What are his skills?']
    };
  }

  private handleAboutQuery(): ChatbotResponse {
    if (this.about?.value?.description) {
      const aboutText = this.about.value.description?.length > 200 
        ? this.about.value.description.substring(0, 200) + '...'
        : this.about.value.description;
      
      return {
        text: aboutText,
        suggestions: ['Tell me about his experience', 'Show me his projects', 'What are his skills?']
      };
    }

    return {
      text: "Bruno is a Senior SRE/DevSecOps/AI Engineer with 12+ years of experience in cloud-native infrastructure, Kubernetes, and AI/ML technologies. He's passionate about building scalable, secure, and efficient systems.",
      suggestions: ['Tell me about his experience', 'Show me his projects', 'What are his skills?']
    };
  }

  getQuickSuggestions(): string[] {
    return [
      'Tell me about his experience',
      'Show me his projects', 
      'What are his skills?',
      'How can I contact him?',
      'Tell me about his background'
    ];
  }

  private getContextualSuggestions(input: string): string[] {
    // Return contextual suggestions based on the input
    if (this.matchesKeywords(input, ['experience', 'work', 'job'])) {
      return [
        'What companies has he worked for?',
        'Tell me about his current role',
        'What are his key achievements?'
      ];
    }

    if (this.matchesKeywords(input, ['skills', 'technology', 'tech'])) {
      return [
        'What cloud platforms does he use?',
        'Tell me about his programming skills',
        'What DevOps tools does he know?'
      ];
    }

    if (this.matchesKeywords(input, ['projects', 'github', 'work'])) {
      return [
        'Show me his featured projects',
        'What technologies does he use?',
        'Tell me about Homepage project'
      ];
    }

    if (this.matchesKeywords(input, ['contact', 'hire', 'available'])) {
      return [
        'Is he available for new opportunities?',
        'How can I reach him?',
        'What\'s his LinkedIn profile?'
      ];
    }

    // Default suggestions
    return [
      'Tell me about his experience',
      'What are his skills?',
      'Show me his projects',
      'How can I contact him?'
    ];
  }

  // Method to toggle between LLM and rule-based responses
  setUseLLM(useLLM: boolean): void {
    this.useLLM = useLLM;
    console.log(`🔄 Chatbot mode switched to: ${useLLM ? 'LLM' : 'Rule-based'}`);
  }

  // Method to check LLM health
  async checkLLMHealth(): Promise<boolean> {
    try {
      const response = await fetch('/api/chat/health');
      const data = await response.json();
      const isHealthy = response.ok && data.status === 'healthy';
      
      if (!isHealthy) {
        console.warn('⚠️ LLM health check failed:', data);
      }
      
      return isHealthy;
    } catch (error) {
      console.error('❌ LLM health check failed:', error);
      return false;
    }
  }

  // Method to get LLM status
  async getLLMStatus(): Promise<any> {
    try {
      const response = await fetch('/api/chat/health');
      const data = await response.json();
      
      if (!response.ok) {
        console.warn('⚠️ Agent-Bruno status check failed:', data);
        return {
          status: 'unhealthy',
          error: data.error || 'Unknown error',
          model: 'llama3.2:3b',
          provider: 'agent-bruno'
        };
      }
      
      return data;
    } catch (error: any) {
      console.error('❌ Failed to get Agent-Bruno status:', error);
      return { 
        status: 'error', 
        error: error?.message || 'Unknown error',
        model: 'llama3.2:3b',
        provider: 'agent-bruno'
      };
    }
  }
}

export default ChatbotService.getInstance(); 