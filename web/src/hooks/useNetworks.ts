import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
    listNetworks,
    createNetwork,
    deleteNetwork,
} from '@/lib/api'
import type { CreateNetworkRequest } from '@/types/api'

// Query keys for cache management
export const queryKeys = {
    networks: ['networks'] as const,
}

// =============================================================================
// Network Hooks
// =============================================================================

export function useNetworks() {
    return useQuery({
        queryKey: queryKeys.networks,
        queryFn: listNetworks,
    })
}

export function useCreateNetwork() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (data: CreateNetworkRequest) => createNetwork(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: queryKeys.networks })
        },
    })
}

export function useDeleteNetwork() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (id: string) => deleteNetwork(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: queryKeys.networks })
        },
    })
}
