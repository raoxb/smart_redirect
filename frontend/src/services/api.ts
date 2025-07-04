import axios, { AxiosResponse } from 'axios'
import type {
  User,
  Link,
  Target,
  LinkStats,
  SystemStats,
  IPInfo,
  Template,
  BatchResponse,
  LoginRequest,
  LoginResponse,
  CreateLinkRequest,
  CreateTargetRequest,
  PaginatedResponse
} from '@/types/api'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth API
export const authApi = {
  login: (credentials: LoginRequest): Promise<AxiosResponse<LoginResponse>> =>
    api.post('/auth/login', credentials),
    
  register: (data: {
    username: string
    email: string
    password: string
  }): Promise<AxiosResponse<User>> =>
    api.post('/auth/register', data),
    
  getProfile: (): Promise<AxiosResponse<User>> =>
    api.get('/auth/profile'),
}

// Links API
export const linksApi = {
  getLinks: (params?: {
    page?: number
    page_size?: number
  }): Promise<AxiosResponse<PaginatedResponse<Link>>> =>
    api.get('/links', { params }),
    
  getLink: (linkId: string): Promise<AxiosResponse<Link>> =>
    api.get(`/links/${linkId}`),
    
  createLink: (data: CreateLinkRequest): Promise<AxiosResponse<Link>> =>
    api.post('/links', data),
    
  updateLink: (linkId: string, data: Partial<CreateLinkRequest>): Promise<AxiosResponse<Link>> =>
    api.put(`/links/${linkId}`, data),
    
  deleteLink: (linkId: string): Promise<AxiosResponse<{ message: string }>> =>
    api.delete(`/links/${linkId}`),
    
  // Targets
  getTargets: (linkId: string): Promise<AxiosResponse<Target[]>> =>
    api.get(`/links/${linkId}/targets`),
    
  createTarget: (linkId: string, data: CreateTargetRequest): Promise<AxiosResponse<Target>> =>
    api.post(`/links/${linkId}/targets`, data),
    
  updateTarget: (targetId: number, data: Partial<CreateTargetRequest>): Promise<AxiosResponse<Target>> =>
    api.put(`/targets/${targetId}`, data),
    
  deleteTarget: (targetId: number): Promise<AxiosResponse<{ message: string }>> =>
    api.delete(`/targets/${targetId}`),
}

// Batch API
export const batchApi = {
  createLinks: (data: {
    links: Array<CreateLinkRequest & { targets: CreateTargetRequest[] }>
  }): Promise<AxiosResponse<BatchResponse>> =>
    api.post('/batch/links', data),
    
  updateLinks: (data: {
    updates: Array<{ link_id: string } & Partial<CreateLinkRequest>>
  }): Promise<AxiosResponse<BatchResponse>> =>
    api.put('/batch/links', data),
    
  deleteLinks: (data: {
    link_ids: string[]
  }): Promise<AxiosResponse<BatchResponse>> =>
    api.delete('/batch/links', { data }),
    
  importCSV: (file: File): Promise<AxiosResponse<BatchResponse>> => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/batch/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  
  exportCSV: (): Promise<AxiosResponse<Blob>> =>
    api.get('/batch/export', { responseType: 'blob' }),
}

// Templates API
export const templatesApi = {
  getTemplates: (params?: {
    page?: number
    page_size?: number
  }): Promise<AxiosResponse<PaginatedResponse<Template>>> =>
    api.get('/templates', { params }),
    
  getTemplate: (id: number): Promise<AxiosResponse<Template>> =>
    api.get(`/templates/${id}`),
    
  createTemplate: (data: {
    name: string
    description: string
    business_unit: string
    network: string
    total_cap?: number
    backup_url?: string
    targets: CreateTargetRequest[]
  }): Promise<AxiosResponse<Template>> =>
    api.post('/templates', data),
    
  updateTemplate: (id: number, data: Partial<{
    name: string
    description: string
    business_unit: string
    network: string
    total_cap: number
    backup_url: string
    targets: CreateTargetRequest[]
  }>): Promise<AxiosResponse<Template>> =>
    api.put(`/templates/${id}`, data),
    
  deleteTemplate: (id: number): Promise<AxiosResponse<{ message: string }>> =>
    api.delete(`/templates/${id}`),
    
  createFromTemplate: (data: {
    template_id: number
    count: number
    overrides?: Record<string, any>
  }): Promise<AxiosResponse<BatchResponse>> =>
    api.post('/templates/create-links', data),
}

// Statistics API
export const statsApi = {
  getLinkStats: (linkId: string): Promise<AxiosResponse<LinkStats>> =>
    api.get(`/stats/links/${linkId}`),
    
  getHourlyStats: (linkId: string, hours?: number): Promise<AxiosResponse<Array<{
    hour: string
    hits: number
  }>>> =>
    api.get(`/stats/links/${linkId}/hourly`, { params: { hours } }),
    
  getSystemStats: (): Promise<AxiosResponse<SystemStats>> =>
    api.get('/stats/system'),
    
  getIPInfo: (ip: string): Promise<AxiosResponse<IPInfo>> =>
    api.get(`/stats/ip/${ip}`),
    
  blockIP: (ip: string, data: {
    reason: string
    duration?: number
  }): Promise<AxiosResponse<{ message: string }>> =>
    api.post(`/stats/ip/${ip}/block`, data),
    
  unblockIP: (ip: string): Promise<AxiosResponse<{ message: string }>> =>
    api.delete(`/stats/ip/${ip}/block`),
}

// Users API (Admin only)
export const usersApi = {
  getUsers: (params?: {
    page?: number
    page_size?: number
  }): Promise<AxiosResponse<PaginatedResponse<User>>> =>
    api.get('/users', { params }),
    
  getUser: (id: number): Promise<AxiosResponse<User>> =>
    api.get(`/users/${id}`),
    
  createUser: (data: {
    username: string
    email: string
    password: string
    role: 'admin' | 'user'
  }): Promise<AxiosResponse<User>> =>
    api.post('/users', data),
    
  updateUser: (id: number, data: Partial<{
    email: string
    password: string
    role: 'admin' | 'user'
    is_active: boolean
  }>): Promise<AxiosResponse<User>> =>
    api.put(`/users/${id}`, data),
    
  deleteUser: (id: number): Promise<AxiosResponse<{ message: string }>> =>
    api.delete(`/users/${id}`),
    
  assignLink: (userId: number, data: {
    link_id: number
    can_edit: boolean
    can_delete: boolean
  }): Promise<AxiosResponse<{ message: string }>> =>
    api.post(`/users/${userId}/links`, data),
    
  getUserLinks: (userId: number): Promise<AxiosResponse<any[]>> =>
    api.get(`/users/${userId}/links`),
}

export { api }
export default api