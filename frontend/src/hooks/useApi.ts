import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { message } from 'antd'
import {
  linksApi,
  statsApi,
  templatesApi,
  usersApi,
  batchApi,
  authApi
} from '@/services/api'

// Links hooks
export const useLinks = (page = 1, pageSize = 20) => {
  return useQuery({
    queryKey: ['links', page, pageSize],
    queryFn: () => linksApi.getLinks({ page, page_size: pageSize }).then(res => res.data),
  })
}

export const useLink = (linkId: string) => {
  return useQuery({
    queryKey: ['link', linkId],
    queryFn: () => linksApi.getLink(linkId).then(res => res.data),
    enabled: !!linkId,
  })
}

export const useCreateLink = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: linksApi.createLink,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['links'] })
      message.success('Link created successfully')
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to create link')
    },
  })
}

export const useUpdateLink = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ linkId, data }: { linkId: string; data: any }) =>
      linksApi.updateLink(linkId, data),
    onSuccess: (_, { linkId }) => {
      queryClient.invalidateQueries({ queryKey: ['links'] })
      queryClient.invalidateQueries({ queryKey: ['link', linkId] })
      message.success('Link updated successfully')
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to update link')
    },
  })
}

export const useDeleteLink = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: linksApi.deleteLink,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['links'] })
      message.success('Link deleted successfully')
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to delete link')
    },
  })
}

// Targets hooks
export const useTargets = (linkId: string) => {
  return useQuery({
    queryKey: ['targets', linkId],
    queryFn: () => linksApi.getTargets(linkId).then(res => res.data),
    enabled: !!linkId,
  })
}

export const useCreateTarget = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ linkId, data }: { linkId: string; data: any }) =>
      linksApi.createTarget(linkId, data),
    onSuccess: (_, { linkId }) => {
      queryClient.invalidateQueries({ queryKey: ['targets', linkId] })
      queryClient.invalidateQueries({ queryKey: ['link', linkId] })
      message.success('Target created successfully')
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to create target')
    },
  })
}

export const useUpdateTarget = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: ({ targetId, data }: { targetId: number; data: any }) =>
      linksApi.updateTarget(targetId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['targets'] })
      message.success('Target updated successfully')
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to update target')
    },
  })
}

export const useDeleteTarget = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: linksApi.deleteTarget,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['targets'] })
      message.success('Target deleted successfully')
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to delete target')
    },
  })
}

// Statistics hooks
export const useLinkStats = (linkId: string) => {
  return useQuery({
    queryKey: ['linkStats', linkId],
    queryFn: () => statsApi.getLinkStats(linkId).then(res => res.data),
    enabled: !!linkId,
  })
}

export const useSystemStats = () => {
  return useQuery({
    queryKey: ['systemStats'],
    queryFn: () => statsApi.getSystemStats().then(res => res.data),
    refetchInterval: 30000, // Refresh every 30 seconds
  })
}

export const useHourlyStats = (linkId: string, hours = 24) => {
  return useQuery({
    queryKey: ['hourlyStats', linkId, hours],
    queryFn: () => statsApi.getHourlyStats(linkId, hours).then(res => res.data),
    enabled: !!linkId,
  })
}

// Templates hooks
export const useTemplates = (page = 1, pageSize = 20) => {
  return useQuery({
    queryKey: ['templates', page, pageSize],
    queryFn: () => templatesApi.getTemplates({ page, page_size: pageSize }).then(res => res.data),
  })
}

export const useCreateTemplate = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: templatesApi.createTemplate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['templates'] })
      message.success('Template created successfully')
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to create template')
    },
  })
}

// Batch operations hooks
export const useBatchImport = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: batchApi.importCSV,
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['links'] })
      const { success, errors } = response.data
      message.success(`Imported ${success.length} links successfully`)
      if (errors.length > 0) {
        message.warning(`${errors.length} imports failed`)
      }
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Import failed')
    },
  })
}

// Users hooks (Admin only)
export const useUsers = (page = 1, pageSize = 20) => {
  return useQuery({
    queryKey: ['users', page, pageSize],
    queryFn: () => usersApi.getUsers({ page, page_size: pageSize }).then(res => res.data),
  })
}

export const useCreateUser = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: usersApi.createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      message.success('User created successfully')
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Failed to create user')
    },
  })
}

// Auth hooks
export const useLogin = () => {
  return useMutation({
    mutationFn: authApi.login,
    onError: (error: any) => {
      message.error(error.response?.data?.error || 'Login failed')
    },
  })
}

export const useProfile = () => {
  return useQuery({
    queryKey: ['profile'],
    queryFn: () => authApi.getProfile().then(res => res.data),
    retry: false,
  })
}