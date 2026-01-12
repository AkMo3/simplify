export interface Application {
  id: string
  environment_id: string
  name: string
  image: string
  replicas: number
  ports: Record<string, string>
  env_vars: Record<string, string>
  created_at: string
  updated_at: string
}

export interface Team {
  id: string
  name: string
  slug: string
  created_at: string
}

export interface Project {
  id: string
  team_id: string
  name: string
  repo_url: string
  created_at: string
}

export interface Environment {
  id: string
  project_id: string
  name: string
  config: Record<string, string>
  created_at: string
}

// API request types
export interface CreateApplicationRequest {
  name: string
  image: string
  environment_id?: string
  replicas?: number
  ports?: Record<string, string>
  env_vars?: Record<string, string>
}

export interface UpdateApplicationRequest {
  name: string
  image: string
  environment_id?: string
  replicas?: number
  ports?: Record<string, string>
  env_vars?: Record<string, string>
}

// Health check types
export interface HealthStatus {
  status: 'healthy' | 'unhealthy'
}

export interface ReadinessStatus {
  status: 'healthy' | 'unhealthy'
  checks: Record<string, ComponentHealth>
}

export interface ComponentHealth {
  status: 'healthy' | 'unhealthy'
  message?: string
}

// API error response
export interface ApiError {
  error: {
    code: string
    message: string
    resource?: string
    id?: string
    field?: string
  }
}
