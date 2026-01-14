import type {
  Application,
  CreateApplicationRequest,
  UpdateApplicationRequest,
  HealthStatus,
  ReadinessStatus,
  ApiError,
  ImageInfo,
} from '@/types/api'

const API_BASE = '/api/v1'

/**
 * Custom error class for API errors
 */
export class ApiClientError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
    public field?: string
  ) {
    super(message)
    this.name = 'ApiClientError'
  }
}

/**
 * Generic fetch wrapper with error handling
 */
async function fetchApi<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const url = endpoint.startsWith('/api') || endpoint.startsWith('/health') || endpoint.startsWith('/ready')
    ? endpoint
    : `${API_BASE}${endpoint}`

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  // Handle no content responses
  if (response.status === 204) {
    return undefined as T
  }

  const data = await response.json()

  if (!response.ok) {
    const error = data as ApiError
    throw new ApiClientError(
      error.error?.message || 'An error occurred',
      response.status,
      error.error?.code,
      error.error?.field
    )
  }

  return data as T
}

// =============================================================================
// Health Endpoints
// =============================================================================

export async function getHealth(): Promise<HealthStatus> {
  return fetchApi<HealthStatus>('/healthz')
}

export async function getReadiness(): Promise<ReadinessStatus> {
  return fetchApi<ReadinessStatus>('/readyz')
}

// =============================================================================
// Application Endpoints
// =============================================================================

export async function listApplications(): Promise<Application[]> {
  return fetchApi<Application[]>('/applications')
}

export async function getApplication(id: string): Promise<Application> {
  return fetchApi<Application>(`/applications/${id}`)
}

export async function createApplication(
  data: CreateApplicationRequest
): Promise<Application> {
  return fetchApi<Application>('/applications', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateApplication(
  id: string,
  data: UpdateApplicationRequest
): Promise<Application> {
  return fetchApi<Application>(`/applications/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteApplication(id: string): Promise<void> {
  return fetchApi<void>(`/applications/${id}`, {
    method: 'DELETE',
  })
}

export async function inspectImage(image: string): Promise<ImageInfo> {
  return fetchApi<ImageInfo>(`/images/inspect?image=${encodeURIComponent(image)}`)
}
