import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '../services/api'
import type { Project, Skill, Experience, Content } from '../services/api'

// =============================================================================
// 🎯 PROJECTS HOOKS
// =============================================================================

export const useProjects = () => {
  return useQuery({
    queryKey: ['projects'],
    queryFn: apiClient.getProjects,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  })
}

export const useProject = (id: number) => {
  return useQuery({
    queryKey: ['project', id],
    queryFn: () => apiClient.getProject(id),
    enabled: !!id,
    staleTime: 5 * 60 * 1000,
  })
}

export const useCreateProject = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.createProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })
}

export const useUpdateProject = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ id, project }: { id: number; project: Partial<Project> }) =>
      apiClient.updateProject(id, project),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      queryClient.invalidateQueries({ queryKey: ['project', id] })
    },
  })
}

export const useDeleteProject = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.deleteProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })
}

// =============================================================================
// 🛠️ SKILLS HOOKS
// =============================================================================

export const useSkills = () => {
  return useQuery({
    queryKey: ['skills'],
    queryFn: apiClient.getSkills,
    staleTime: 10 * 60 * 1000, // 10 minutes
    gcTime: 20 * 60 * 1000, // 20 minutes
  })
}

export const useSkill = (id: number) => {
  return useQuery({
    queryKey: ['skill', id],
    queryFn: () => apiClient.getSkill(id),
    enabled: !!id,
    staleTime: 10 * 60 * 1000,
  })
}

export const useCreateSkill = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.createSkill,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['skills'] })
    },
  })
}

export const useUpdateSkill = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ id, skill }: { id: number; skill: Partial<Skill> }) =>
      apiClient.updateSkill(id, skill),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['skills'] })
      queryClient.invalidateQueries({ queryKey: ['skill', id] })
    },
  })
}

export const useDeleteSkill = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.deleteSkill,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['skills'] })
    },
  })
}

// =============================================================================
// 💼 EXPERIENCES HOOKS
// =============================================================================

export const useExperiences = () => {
  return useQuery({
    queryKey: ['experiences'],
    queryFn: apiClient.getExperiences,
    staleTime: 10 * 60 * 1000,
    gcTime: 20 * 60 * 1000,
  })
}

export const useExperience = (id: number) => {
  return useQuery({
    queryKey: ['experience', id],
    queryFn: () => apiClient.getExperience(id),
    enabled: !!id,
    staleTime: 10 * 60 * 1000,
  })
}

export const useCreateExperience = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.createExperience,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['experiences'] })
    },
  })
}

export const useUpdateExperience = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ id, experience }: { id: number; experience: Partial<Experience> }) =>
      apiClient.updateExperience(id, experience),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['experiences'] })
      queryClient.invalidateQueries({ queryKey: ['experience', id] })
    },
  })
}

export const useDeleteExperience = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.deleteExperience,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['experiences'] })
    },
  })
}

// =============================================================================
// 📄 CONTENT HOOKS
// =============================================================================

export const useContent = () => {
  return useQuery({
    queryKey: ['content'],
    queryFn: apiClient.getContent,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  })
}

export const useContentByType = (type: string) => {
  return useQuery({
    queryKey: ['content', type],
    queryFn: () => apiClient.getContentByType(type),
    enabled: !!type,
    staleTime: 5 * 60 * 1000,
  })
}

export const useCreateContent = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.createContent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['content'] })
    },
  })
}

export const useUpdateContent = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ id, content }: { id: number; content: Partial<Content> }) =>
      apiClient.updateContent(id, content),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['content'] })
      queryClient.invalidateQueries({ queryKey: ['content', id] })
    },
  })
}

export const useDeleteContent = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.deleteContent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['content'] })
    },
  })
}

// =============================================================================
// 👤 ABOUT HOOKS
// =============================================================================

export const useAbout = () => {
  return useQuery({
    queryKey: ['about'],
    queryFn: apiClient.getAbout,
    staleTime: 10 * 60 * 1000,
    gcTime: 20 * 60 * 1000,
  })
}

export const useUpdateAbout = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.updateAbout,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['about'] })
    },
  })
}

// =============================================================================
// 📞 CONTACT HOOKS
// =============================================================================

export const useContact = () => {
  return useQuery({
    queryKey: ['contact'],
    queryFn: apiClient.getContact,
    staleTime: 10 * 60 * 1000,
    gcTime: 20 * 60 * 1000,
  })
}

export const useUpdateContact = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: apiClient.updateContact,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['contact'] })
    },
  })
}

// =============================================================================
// 🏥 HEALTH HOOKS
// =============================================================================

export const useHealthCheck = () => {
  return useQuery({
    queryKey: ['health'],
    queryFn: apiClient.healthCheck,
    refetchInterval: 30 * 1000, // Check every 30 seconds
    refetchIntervalInBackground: true,
  })
}

