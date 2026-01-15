import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2, RefreshCw, ExternalLink, Cuboid } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui'
import { ContainerStatusBadge } from '@/components/ui/StatusBadge'
import { CreatePodModal } from '@/components/pods/CreatePodModal'
import { usePods, useDeletePod } from '@/hooks/usePods'
import { useSearch } from '@/contexts/SearchContext'
import type { Pod } from '@/types/api'

const MAX_VISIBLE_PODS = 6

export function Pods() {
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
    const [deletingId, setDeletingId] = useState<string | null>(null)

    const { data: pods, isLoading, isError, refetch } = usePods()
    const deleteMutation = useDeletePod()
    const { query } = useSearch()
    const navigate = useNavigate()

    // Filter, sort, and limit pods
    const displayedPods = useMemo(() => {
        if (!pods) return []

        let filtered = pods

        // Client-side search filtering by name
        if (query.trim()) {
            const searchLower = query.toLowerCase()
            filtered = filtered.filter((pod) =>
                pod.name.toLowerCase().includes(searchLower)
            )
        }

        // Sort by created_at descending (most recent first)
        const sorted = [...filtered].sort((a, b) => {
            const dateA = new Date(a.created_at).getTime()
            const dateB = new Date(b.created_at).getTime()
            return dateB - dateA
        })

        // Limit to MAX_VISIBLE_PODS
        return sorted.slice(0, MAX_VISIBLE_PODS)
    }, [pods, query])

    const handleDelete = async (e: React.MouseEvent, pod: Pod) => {
        e.stopPropagation() // Prevent row click
        if (!confirm(`Are you sure you want to delete pod "${pod.name}"?`)) return

        setDeletingId(pod.id)
        try {
            await deleteMutation.mutateAsync(pod.id)
        } finally {
            setDeletingId(null)
        }
    }

    const formatDate = (dateString: string) => {
        if (!dateString) return '-'
        const date = new Date(dateString)
        return date.toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        })
    }

    // Generate port URL
    const getPortUrl = (hostPort: string) => {
        return `http://localhost:${hostPort}`
    }

    // Get display ports (show first 2, then +N more)
    const getPortsDisplay = (ports: Record<string, string> | undefined) => {
        if (!ports || Object.keys(ports).length === 0) {
            return null
        }

        const portEntries = Object.entries(ports)
        const visiblePorts = portEntries.slice(0, 2)
        const remainingCount = portEntries.length - 2

        return (
            <div className="flex items-center gap-1.5 flex-wrap">
                {visiblePorts.map(([hostVal, containerVal]) => {
                    return (
                        <a
                            key={hostVal}
                            href={getPortUrl(hostVal)}
                            target="_blank"
                            rel="noopener noreferrer"
                            onClick={(e) => e.stopPropagation()}
                            className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-[hsl(0_0%_15%)] text-xs font-mono text-[hsl(0_0%_70%)] hover:bg-[hsl(0_0%_20%)] hover:text-foreground transition-colors"
                            title={`Host: ${hostVal} -> Pod: ${containerVal}`}
                        >
                            :{hostVal}
                            <ExternalLink className="h-3 w-3" />
                        </a>
                    )
                })}
                {remainingCount > 0 && (
                    <span className="text-xs text-muted-foreground">
                        +{remainingCount} more
                    </span>
                )}
            </div>
        )
    }

    return (
        <div className="animate-fade-in">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-2xl font-semibold tracking-tight">Pods</h1>
                    <p className="text-muted-foreground mt-1">
                        Manage shared network namespaces for your applications
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => refetch()}
                        disabled={isLoading}
                    >
                        <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                    </Button>
                    <Button onClick={() => setIsCreateModalOpen(true)}>
                        <Plus className="h-4 w-4 mr-2" />
                        Create Pod
                    </Button>
                </div>
            </div>

            {/* Search indicator */}
            {query && (
                <div className="text-sm text-muted-foreground">
                    Showing results for "<span className="text-foreground">{query}</span>"
                    {displayedPods.length === 0 && ' — No pods found'}
                </div>
            )}

            {/* Content */}
            {isLoading ? (
                <div className="card p-12 text-center">
                    <RefreshCw className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
                    <p className="mt-4 text-muted-foreground">Loading pods...</p>
                </div>
            ) : isError ? (
                <div className="card p-12 text-center">
                    <p className="text-destructive">Failed to load pods</p>
                    <Button variant="secondary" className="mt-4" onClick={() => refetch()}>
                        Try Again
                    </Button>
                </div>
            ) : pods?.length === 0 ? (
                <div className="card p-12 text-center">
                    <Cuboid className="h-12 w-12 mx-auto text-muted-foreground/50" />
                    <h3 className="mt-4 text-lg font-medium">No pods yet</h3>
                    <p className="mt-1 text-muted-foreground">
                        Create your first pod to get started
                    </p>
                    <Button className="mt-4" onClick={() => setIsCreateModalOpen(true)}>
                        <Plus className="h-4 w-4 mr-2" />
                        Create Pod
                    </Button>
                </div>
            ) : displayedPods.length === 0 ? (
                <div className="card p-12 text-center">
                    <Cuboid className="h-12 w-12 mx-auto text-muted-foreground/50" />
                    <h3 className="mt-4 text-lg font-medium">No matching pods</h3>
                    <p className="mt-1 text-muted-foreground">
                        Try adjusting your search query
                    </p>
                </div>
            ) : (
                <div className="card overflow-hidden">
                    <table className="w-full">
                        <thead>
                            <tr className="border-b border-border/50 bg-[hsl(0_0%_9%)]">
                                <th className="table-header">Name</th>
                                <th className="table-header">Status</th>
                                <th className="table-header">Ports</th>
                                <th className="table-header">Created</th>
                                <th className="table-header text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            <AnimatePresence>
                                {displayedPods.map((pod) => (
                                    <motion.tr
                                        key={pod.id}
                                        initial={{ opacity: 0 }}
                                        animate={{ opacity: 1 }}
                                        exit={{ opacity: 0, height: 0 }}
                                        className="table-row cursor-pointer hover:bg-zinc-800/10"
                                        onClick={() => navigate(`/pods/${pod.id}`)}
                                    >
                                        {/* Name */}
                                        <td className="table-cell">
                                            <div className="flex items-center gap-3 group">
                                                <div className="h-9 w-9 rounded-lg bg-[hsl(0_0%_15%)] flex items-center justify-center">
                                                    <Cuboid className="h-4 w-4 text-[hsl(0_0%_60%)]" />
                                                </div>
                                                <div>
                                                    <p className="font-medium group-hover:text-foreground transition-colors">
                                                        {pod.name}
                                                    </p>
                                                    <p className="text-xs text-muted-foreground font-mono">
                                                        {pod.id.substring(0, 12)}
                                                    </p>
                                                </div>
                                            </div>
                                        </td>

                                        {/* Status */}
                                        <td className="table-cell">
                                            {/* Reuse ContainerStatusBadge as Pod statuses are similar (Created, Running, etc.) */}
                                            {/* Or we might need a specific PodStatusBadge if statuses differ significantly */}
                                            <ContainerStatusBadge status={pod.status.toLowerCase() as any || 'stopped'} />
                                        </td>

                                        {/* Ports */}
                                        <td className="table-cell">
                                            {getPortsDisplay(pod.ports) || (
                                                <span className="text-xs text-muted-foreground">—</span>
                                            )}
                                        </td>

                                        {/* Created */}
                                        <td className="table-cell text-sm text-muted-foreground">
                                            {formatDate(pod.created_at)}
                                        </td>

                                        {/* Actions */}
                                        <td className="table-cell text-right">
                                            <div className="flex items-center justify-end gap-1">
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={(e) => handleDelete(e, pod)}
                                                    disabled={deletingId === pod.id}
                                                    className="text-muted-foreground hover:text-destructive h-8 w-8 px-0"
                                                    title="Delete Pod"
                                                >
                                                    {deletingId === pod.id ? (
                                                        <RefreshCw className="h-4 w-4 animate-spin" />
                                                    ) : (
                                                        <Trash2 className="h-4 w-4" />
                                                    )}
                                                </Button>
                                            </div>
                                        </td>
                                    </motion.tr>
                                ))}
                            </AnimatePresence>
                        </tbody>
                    </table>

                    {/* Show more indicator */}
                    {pods && pods.length > MAX_VISIBLE_PODS && !query && (
                        <div className="px-4 py-3 border-t border-border/50 text-center">
                            <span className="text-sm text-muted-foreground">
                                Showing {MAX_VISIBLE_PODS} of {pods.length} pods
                            </span>
                        </div>
                    )}
                </div>
            )}

            {/* Create Modal */}
            <CreatePodModal
                isOpen={isCreateModalOpen}
                onClose={() => setIsCreateModalOpen(false)}
            />
        </div>
    )
}
