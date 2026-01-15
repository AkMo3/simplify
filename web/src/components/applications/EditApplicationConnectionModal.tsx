import { useState, useEffect } from 'react'
import { Modal, Button } from '@/components/ui'
import { Label } from '@/components/ui/Label'
import { useNetworks } from '@/hooks/useNetworks'
import { usePods } from '@/hooks/usePods'
import { updateApplication } from '@/lib/api'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import type { Application } from '@/types/api'
import { clsx } from 'clsx'
import { Network as NetworkIcon, Box } from 'lucide-react'

interface EditApplicationConnectionModalProps {
    application: Application
    isOpen: boolean
    onClose: () => void
}

type ConnectionType = 'network' | 'pod'

export function EditApplicationConnectionModal({
    application,
    isOpen,
    onClose,
}: EditApplicationConnectionModalProps) {
    const [connectionType, setConnectionType] = useState<ConnectionType>('network')
    const [networkId, setNetworkId] = useState('')
    const [podId, setPodId] = useState('')
    const [isSubmitting, setIsSubmitting] = useState(false)

    const { data: networks } = useNetworks()
    const { data: pods } = usePods()
    const queryClient = useQueryClient()

    useEffect(() => {
        if (isOpen) {
            if (application.pod_id) {
                setConnectionType('pod')
                setPodId(application.pod_id)
                setNetworkId('')
            } else {
                setConnectionType('network')
                setNetworkId(application.network_id || '')
                setPodId('')
            }
        }
    }, [isOpen, application])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setIsSubmitting(true)

        try {
            const payload = {
                name: application.name || '',
                image: application.image || '',
                replicas: application.replicas,
                ports: application.ports,
                env_vars: application.env_vars,
                network_id: undefined as string | undefined,
                pod_id: undefined as string | undefined,
            }

            if (connectionType === 'network') {
                payload.network_id = networkId || undefined
                payload.pod_id = undefined
            } else {
                payload.pod_id = podId || undefined
                payload.network_id = undefined
            }

            // Hack: our updateApplication types might not like explicit undefined for clearing?
            // let's check if the API handles it. 
            // JSON encoding 'undefined' fields omits them usually.
            // We might need to send empty string if the backend expects it, but UUID usually implies strict parsing.
            // However, previous attempts used undefined. Let's stick with that.
            // The issue is how to CLEAR a field.
            // If I send network_id: undefined, it won't be in JSON.
            // If backend checks for presence, it won't see it.
            // To strictly clear, maybe empty string?
            // But Go backend: "json.NewDecoder(r.Body).Decode(&app)".
            // If field missing, struct field is zero value (empty string) IF it's a new struct.
            // But update uses the struct.
            // "s.store.UpdateApplication(&app)"
            // `app` is populated from JSON.
            // If JSON is {name: "foo", network_id: ""} -> network_id is "" -> DB updates to "".
            // If JSON is {name: "foo"} (missing network_id) -> network_id is "" (Go default) -> DB updates to "".
            // SO, sending undefined (omitted) OR empty string should work fine for clearing, 
            // ASSUMING `UpdateApplication` REPLACES the whole record or fields.
            // Current implementation REPLACES the whole struct. So omitting a field means it becomes empty.

            await updateApplication(application.id, payload)

            toast.success('Connection updated')
            queryClient.invalidateQueries({ queryKey: ['applications'] })
            queryClient.invalidateQueries({ queryKey: ['application', application.id] })
            onClose()
        } catch (error) {
            toast.error('Failed to update connection')
            console.error(error)
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <Modal
            isOpen={isOpen}
            onClose={onClose}
            title={`Edit Connection for ${application.name}`}
            description="Connect this application to a Network or attach it to a Pod."
        >
            <form onSubmit={handleSubmit} className="space-y-4">
                {/* Type Selection */}
                <div className="flex p-1 bg-muted rounded-lg">
                    <button
                        type="button"
                        className={clsx(
                            "flex-1 flex items-center justify-center gap-2 py-1.5 text-sm font-medium rounded-md transition-all",
                            connectionType === 'network'
                                ? "bg-background shadow-sm text-foreground"
                                : "text-muted-foreground hover:text-foreground"
                        )}
                        onClick={() => {
                            setConnectionType('network')
                            setPodId('')
                        }}
                    >
                        <NetworkIcon className="h-4 w-4" />
                        Network
                    </button>
                    <button
                        type="button"
                        className={clsx(
                            "flex-1 flex items-center justify-center gap-2 py-1.5 text-sm font-medium rounded-md transition-all",
                            connectionType === 'pod'
                                ? "bg-background shadow-sm text-foreground"
                                : "text-muted-foreground hover:text-foreground"
                        )}
                        onClick={() => {
                            setConnectionType('pod')
                            setNetworkId('')
                        }}
                    >
                        <Box className="h-4 w-4" />
                        Pod
                    </button>
                </div>

                {connectionType === 'network' ? (
                    <div className="space-y-2 animate-in fade-in slide-in-from-top-1 duration-200">
                        <Label>Network</Label>
                        <select
                            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                            value={networkId}
                            onChange={(e) => setNetworkId(e.target.value)}
                        >
                            <option value="">None (Default Bridge)</option>
                            {networks?.map((net) => (
                                <option key={net.id} value={net.id}>
                                    {net.name} ({net.driver})
                                </option>
                            ))}
                        </select>
                        <p className="text-xs text-muted-foreground">
                            Connecting to a custom network enables DNS resolution by container name.
                        </p>
                    </div>
                ) : (
                    <div className="space-y-2 animate-in fade-in slide-in-from-top-1 duration-200">
                        <Label>Pod</Label>
                        <select
                            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                            value={podId}
                            onChange={(e) => setPodId(e.target.value)}
                        >
                            <option value="">Select a Pod</option>
                            {pods?.map((pod) => (
                                <option key={pod.id} value={pod.id}>
                                    {pod.name} ({pod.status})
                                </option>
                            ))}
                        </select>
                        <p className="text-xs text-muted-foreground">
                            Application will run inside the selected pod's network namespace.
                            External ports should be managed at the Pod level.
                        </p>
                    </div>
                )}

                <div className="flex justify-end gap-2 pt-2">
                    <Button type="button" variant="secondary" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button type="submit" isLoading={isSubmitting}>
                        Save Changes
                    </Button>
                </div>
            </form>
        </Modal>
    )
}
