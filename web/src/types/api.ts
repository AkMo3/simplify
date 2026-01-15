export interface Application {
  id: string
  environment_id: string
  name: string
  image: string
  status: ContainerStatus
  replicas: number
  pod_id?: string
  network_id?: string
  health_check?: HealthCheckConfig
  health_status: HealthCheckStatus
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
  pod_id?: string
  network_id?: string
  ports?: Record<string, string>
  env_vars?: Record<string, string>
  health_check?: HealthCheckConfig
}

export interface UpdateApplicationRequest {
  name: string
  image: string
  environment_id?: string
  replicas?: number
  pod_id?: string
  network_id?: string
  ports?: Record<string, string>
  env_vars?: Record<string, string>
  health_check?: HealthCheckConfig
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

export type ContainerStatus =
  | 'creating'
  | 'running'
  | 'restarting'
  | 'stopping'
  | 'stopped'
  | 'error'
  | 'created'

export type HealthCheckStatus = 'healthy' | 'unhealthy' | 'starting' | 'none'

export interface HealthCheckConfig {
  path: string
  interval: number
  timeout: number
  retries: number
}

export interface ImageInfo {
  id: string
  exposed_ports: string[]
}

export interface Pod {
  id: string
  created_at: string
  name: string
  ports: Record<string, string>
  status: string
}

export interface CreatePodRequest {
  name: string
  ports: Record<string, string>
}

export interface Network {
  id: string
  created_at: string
  name: string
  subnet: string
  driver: string
}

export interface CreateNetworkRequest {
  name: string
}

