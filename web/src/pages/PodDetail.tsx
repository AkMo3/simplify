
import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Box, Circle, Network, Server, Plus } from 'lucide-react'
import { usePod } from '@/hooks/usePods'
import { useApplications } from '@/hooks/useApplications'
import { Button, Card, Skeleton } from '@/components/ui'
import { ContainerStatusBadge } from '@/components/ui/StatusBadge'
import { CreateApplicationForm } from '@/components/applications/CreateApplicationForm'

export default function PodDetail() {
    const { id } = useParams<{ id: string }>()
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
    const { data: pod, isLoading: isLoadingPod, error: podError } = usePod(id!)
    const { data: applications, isLoading: isLoadingApps } = useApplications()

    if (isLoadingPod || isLoadingApps) {
        return (
            <div className="space-y-6">
                <Skeleton className="h-8 w-1/3" />
                <Skeleton className="h-64 w-full" />
            </div>
        )
    }

    if (podError || !pod) {
        return (
            <div className="flex flex-col items-center justify-center p-12 text-center text-zinc-400">
                <Box className="mb-4 h-12 w-12 opacity-50" />
                <h3 className="mb-2 text-lg font-medium text-white">Pod not found</h3>
                <p className="mb-6 max-w-sm">
                    The pod you are looking for does not exist or has been deleted.
                </p>
                <Link to="/pods">
                    <Button variant="ghost">Back to Pods</Button>
                </Link>
            </div>
        )
    }

    // Filter applications belonging to this pod
    const podApps = applications?.filter((app) => app.pod_id === pod.id) || []

    return (
        <div className="space-y-8 animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="space-y-1">
                    <div className="flex items-center gap-3">
                        <Link
                            to="/pods"
                            className="text-zinc-400 transition-colors hover:text-white"
                        >
                            <ArrowLeft className="h-5 w-5" />
                        </Link>
                        <h1 className="text-2xl font-bold tracking-tight text-white">
                            {pod.name}
                        </h1>
                        <ContainerStatusBadge status={pod.status.toLowerCase() as any} />
                    </div>
                    <p className="ml-8 text-sm text-zinc-400 font-mono">ID: {pod.id}</p>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-3">
                {/* Main Content - Applications List */}
                <div className="md:col-span-2 space-y-6">
                    <section>
                        <div className="flex items-center justify-between mb-4">
                            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                                <Box className="h-5 w-5 text-zinc-400" />
                                Applications
                            </h2>
                            <Button size="sm" onClick={() => setIsCreateModalOpen(true)}>
                                <Plus className="h-4 w-4 mr-2" />
                                Add Application
                            </Button>
                        </div>

                        {podApps.length === 0 ? (
                            <Card className="flex flex-col items-center justify-center p-8 text-center border-dashed bg-transparent border-zinc-800">
                                <p className="text-zinc-400 mb-2">No applications in this pod</p>
                                <p className="text-xs text-zinc-500 max-w-xs">
                                    Applications assigned to this pod share its network namespace and localhost.
                                </p>
                            </Card>
                        ) : (
                            <div className="grid gap-4">
                                {podApps.map((app) => (
                                    <Link key={app.id} to={`/applications/${app.id}`}>
                                        <Card className="p-4 hover:bg-zinc-900/50 transition-colors group cursor-pointer border-zinc-800 bg-zinc-900/20">
                                            <div className="flex justify-between items-start">
                                                <div className="flex gap-3">
                                                    <div className="mt-1 h-8 w-8 rounded bg-blue-500/10 flex items-center justify-center text-blue-400">
                                                        <Box className="h-4 w-4" />
                                                    </div>
                                                    <div>
                                                        <h3 className="font-medium text-white group-hover:text-blue-400 transition-colors">
                                                            {app.name}
                                                        </h3>
                                                        <div className="flex items-center gap-2 text-xs text-zinc-500 mt-1">
                                                            <span>{app.image}</span>
                                                            <span>•</span>
                                                            <ContainerStatusBadge
                                                                status={app.status as any}
                                                            />
                                                        </div>
                                                    </div>
                                                </div>
                                                <div className="text-right">
                                                    {/* Assuming ports are internal if in pod, but showing them helps */}
                                                    {Object.keys(app.ports || {}).length > 0 && (
                                                        <div className="text-xs text-zinc-400 font-mono">
                                                            {Object.entries(app.ports).map(([container, host]) => (
                                                                <div key={container}>{container}{host ? ` -> ${host}` : '(internal)'}</div>
                                                            ))}
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        </Card>
                                    </Link>
                                ))}
                            </div>
                        )}
                    </section>
                </div>

                {/* Sidebar - Pod Info */}
                <div className="space-y-6">
                    <Card className="p-5 border-zinc-800 bg-zinc-900/20">
                        <h3 className="text-sm font-medium text-zinc-400 mb-4 uppercase tracking-wider">
                            Pod Info
                        </h3>

                        <div className="space-y-4">
                            <div>
                                <div className="text-xs text-zinc-500 mb-1 flex items-center gap-1.5">
                                    <Network className="h-3.5 w-3.5" />
                                    Ports (External)
                                </div>
                                <div className="space-y-1">
                                    {Object.entries(pod.ports || {}).map(([containerPort, hostPort]) => (
                                        <div
                                            key={containerPort}
                                            className="inline-flex items-center gap-2 rounded-md bg-zinc-800/50 px-2 py-1 text-xs font-mono text-zinc-300"
                                        >
                                            <span>{containerPort}</span>
                                            <span className="text-zinc-600">→</span>
                                            <span className="text-emerald-400">{hostPort}</span>
                                        </div>
                                    ))}
                                    {Object.keys(pod.ports || {}).length === 0 && (
                                        <span className="text-xs text-zinc-600 italic">No external ports exposed</span>
                                    )}
                                </div>
                            </div>

                            <div>
                                <div className="text-xs text-zinc-500 mb-1 flex items-center gap-1.5">
                                    <Circle className="h-3.5 w-3.5" />
                                    Status
                                </div>
                                <div className="text-sm text-white capitalize">{pod.status || 'Unknown'}</div>
                            </div>

                            <div>
                                <div className="text-xs text-zinc-500 mb-1 flex items-center gap-1.5">
                                    <Server className="h-3.5 w-3.5" />
                                    Created
                                </div>
                                <div className="text-sm text-zinc-300">
                                    {new Date(pod.created_at).toLocaleString()}
                                </div>
                            </div>
                        </div>
                    </Card>

                    <Card className="p-5 border-zinc-800 bg-blue-500/5">
                        <h4 className="flex items-center gap-2 text-sm font-medium text-blue-400 mb-2">
                            <Box className="h-4 w-4" />
                            Shared Context
                        </h4>
                        <p className="text-xs text-blue-200/70 leading-relaxed">
                            All applications in this pod share the same network namespace. They can communicate with each other over <code>localhost</code>.
                        </p>
                    </Card>
                </div>
            </div>

            <CreateApplicationForm
                isOpen={isCreateModalOpen}
                onClose={() => setIsCreateModalOpen(false)}
                initialPodId={pod.id}
            />
        </div>
    )
}
