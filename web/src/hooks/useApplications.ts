import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listApplications,
  getApplication,
  createApplication,
  updateApplication,
  deleteApplication,
  getHealth,
  getReadiness,
} from '@/lib/api'
import type { CreateApplicationRequest, UpdateApplicationRequest } from '@/types/api'

// Query keys for cache management
export const queryKeys = {
  applications: ['applications'] as const,
  application: (id: string) => ['applications', id] as const,
  health: ['health'] as const,
  readiness: ['readiness'] as const,
}

// =============================================================================
// Health Hooks
// =============================================================================

export function useHealth() {
  return useQuery({
    queryKey: queryKeys.health,
    queryFn: getHealth,
    refetchInterval: 10000, // Poll every 10 seconds
  })
}

export function useReadiness() {
  return useQuery({
    queryKey: queryKeys.readiness,
    queryFn: getReadiness,
    refetchInterval: 10000,
  })
}

// =============================================================================
// Application Hooks
// =============================================================================

export function useApplications() {
  return useQuery({
    queryKey: queryKeys.applications,
    queryFn: listApplications,
  })
}

export function useApplication(id: string) {
  return useQuery({
    queryKey: queryKeys.application(id),
    queryFn: () => getApplication(id),
    enabled: !!id,
  })
}

export function useCreateApplication() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateApplicationRequest) => createApplication(data),
    onSuccess: () => {
      // Invalidate applications list to refetch
      queryClient.invalidateQueries({ queryKey: queryKeys.applications })
    },
  })
}

export function useUpdateApplication() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateApplicationRequest }) =>
      updateApplication(id, data),
    onSuccess: (_, variables) => {
      // Invalidate both list and specific application
      queryClient.invalidateQueries({ queryKey: queryKeys.applications })
      queryClient.invalidateQueries({ queryKey: queryKeys.application(variables.id) })
    },
  })
}

export function useDeleteApplication() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteApplication(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications })
    },
  })
}
