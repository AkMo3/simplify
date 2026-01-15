import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
    listPods,
    createPod,
    deletePod,
    getPod,
} from '@/lib/api'
import type { CreatePodRequest } from '@/types/api'

// Query keys for cache management
export const queryKeys = {
    pods: ['pods'] as const,
    pod: (id: string) => ['pods', id] as const,
}

// =============================================================================
// Pod Hooks
// =============================================================================

export function usePods() {
    return useQuery({
        queryKey: queryKeys.pods,
        queryFn: listPods,
    })
}

export function usePod(id: string) {
    return useQuery({
        queryKey: queryKeys.pod(id),
        queryFn: () => getPod(id),
        enabled: !!id,
    })
}

export function useCreatePod() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (data: CreatePodRequest) => createPod(data),
        onSuccess: () => {
            // Invalidate pods list to refetch
            queryClient.invalidateQueries({ queryKey: queryKeys.pods })
        },
    })
}

export function useDeletePod() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (id: string) => deletePod(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: queryKeys.pods })
        },
    })
}
