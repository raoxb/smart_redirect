export interface User {
  id: number
  username: string
  email: string
  role: 'admin' | 'user'
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Link {
  id: number
  link_id: string
  business_unit: string
  network: string
  total_cap: number
  current_hits: number
  backup_url: string
  is_active: boolean
  created_at: string
  updated_at: string
  targets?: Target[]
}

export interface Target {
  id: number
  link_id: number
  url: string
  weight: number
  cap: number
  current_hits: number
  countries: string[]
  param_mapping: Record<string, string>
  static_params: Record<string, string>
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface LinkStats {
  link_id: string
  business_unit: string
  total_hits: number
  today_hits: number
  unique_ips: number
  countries: CountryStats[]
  targets: TargetStats[]
}

export interface CountryStats {
  country: string
  hits: number
}

export interface TargetStats {
  target_id: number
  url: string
  hits: number
}

export interface SystemStats {
  total_links: number
  total_hits: number
  today_hits: number
  unique_ips: number
  top_countries: CountryStats[]
}

export interface IPInfo {
  ip: string
  access_count: number
  last_access: string
  country: string
  is_blocked: boolean
  block_reason: string
  recent_logs: AccessLog[]
}

export interface AccessLog {
  id: number
  link_id: number
  target_id: number
  ip: string
  user_agent: string
  referer: string
  country: string
  created_at: string
}

export interface Template {
  id: number
  name: string
  description: string
  business_unit: string
  network: string
  total_cap: number
  backup_url: string
  target_config: string
  created_at: string
  updated_at: string
}

export interface BatchResponse {
  success: BatchResult[]
  errors: BatchError[]
}

export interface BatchResult {
  index: number
  link_id: string
  link_url: string
}

export interface BatchError {
  index: number
  message: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user_id: number
  username: string
  role: string
}

export interface CreateLinkRequest {
  business_unit: string
  network: string
  total_cap?: number
  backup_url?: string
}

export interface CreateTargetRequest {
  url: string
  weight: number
  cap?: number
  countries?: string[]
  param_mapping?: Record<string, string>
  static_params?: Record<string, string>
}

export interface PaginatedResponse<T> {
  total: number
  page: number
  size: number
  data: T[]
}

export interface ApiError {
  error: string
  code?: string
  details?: Record<string, any>
}