import { useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button, Input, Modal } from '@/components/ui'
import { useCreatePod } from '@/hooks/usePods'
import { ApiClientError } from '@/lib/api'

interface CreatePodModalProps {
    isOpen: boolean
    onClose: () => void
}

interface PortMapping {
    id: string
    host: string
    container: string
}

export function CreatePodModal({ isOpen, onClose }: CreatePodModalProps) {
    const [name, setName] = useState('')
    const [ports, setPorts] = useState<PortMapping[]>([])
    const [error, setError] = useState<string | null>(null)

    const createMutation = useCreatePod()

    const resetForm = () => {
        setName('')
        setPorts([])
        setError(null)
    }

    const handleClose = () => {
        resetForm()
        onClose()
    }

    const addPort = () => {
        setPorts([
            ...ports,
            {
                id: crypto.randomUUID(),
                host: '',
                container: '',
            },
        ])
    }

    const removePort = (id: string) => {
        setPorts(ports.filter((p) => p.id !== id))
    }

    const updatePort = (id: string, field: 'host' | 'container', value: string) => {
        setPorts(
            ports.map((p) => (p.id === id ? { ...p, [field]: value } : p))
        )
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError(null)

        if (!name.trim()) {
            setError('Pod name is required')
            return
        }

        try {
            // Convert ports array to object
            const portsRecord: Record<string, string> = {}
            ports.forEach((p) => {
                if (p.host && p.container) {
                    portsRecord[p.host] = p.container
                }
            })

            await createMutation.mutateAsync({
                name: name.trim(),
                ports: portsRecord,
            })

            handleClose()
        } catch (err) {
            if (err instanceof ApiClientError) {
                setError(err.message)
            } else {
                setError('Failed to create pod')
            }
        }
    }

    return (
        <Modal isOpen={isOpen} onClose={handleClose} title="Create Pod">
            <form onSubmit={handleSubmit} className="space-y-6">
                {error && (
                    <div className="p-3 text-sm text-destructive-foreground bg-destructive/10 border border-destructive/20 rounded-md">
                        {error}
                    </div>
                )}

                {/* Name */}
                <div className="space-y-2">
                    <label htmlFor="name" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">Pod Name</label>
                    <Input
                        id="name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="e.g. web-stack"
                        required
                    />
                    <p className="text-xs text-muted-foreground">
                        A unique name for your pod
                    </p>
                </div>

                {/* Ports */}
                <div className="space-y-3">
                    <div className="flex items-center justify-between">
                        <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">Port Mapping</label>
                        <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={addPort}
                            className="text-xs h-7"
                        >
                            <Plus className="h-3 w-3 mr-1" />
                            Add Port
                        </Button>
                    </div>

                    <div className="space-y-2">
                        <AnimatePresence>
                            {ports.map((port) => (
                                <motion.div
                                    key={port.id}
                                    initial={{ opacity: 0, height: 0 }}
                                    animate={{ opacity: 1, height: 'auto' }}
                                    exit={{ opacity: 0, height: 0 }}
                                    className="flex items-start gap-2"
                                >
                                    <div className="flex-1">
                                        <Input
                                            placeholder="Host Port (e.g. 8080)"
                                            value={port.host}
                                            onChange={(e) => updatePort(port.id, 'host', e.target.value)}
                                            className="font-mono text-sm"
                                        />
                                    </div>
                                    <div className="flex items-center justify-center pt-2 text-muted-foreground">
                                        :
                                    </div>
                                    <div className="flex-1">
                                        <Input
                                            placeholder="Pod Port (e.g. 80)"
                                            value={port.container}
                                            onChange={(e) => updatePort(port.id, 'container', e.target.value)}
                                            className="font-mono text-sm"
                                        />
                                    </div>
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => removePort(port.id)}
                                        className="text-muted-foreground hover:text-destructive shrink-0 h-9 w-9 p-0"
                                    >
                                        <Trash2 className="h-4 w-4" />
                                    </Button>
                                </motion.div>
                            ))}
                        </AnimatePresence>
                        {ports.length === 0 && (
                            <div className="text-sm text-muted-foreground text-center py-4 border border-dashed border-border rounded-md">
                                No ports exposed. Applications inside this pod can communicate via localhost.
                            </div>
                        )}
                    </div>
                </div>

                {/* Footer */}
                <div className="flex justify-end gap-3 pt-2">
                    <Button type="button" variant="ghost" onClick={handleClose}>
                        Cancel
                    </Button>
                    <Button type="submit" disabled={createMutation.isPending}>
                        {createMutation.isPending ? 'Creating...' : 'Create Pod'}
                    </Button>
                </div>
            </form>
        </Modal>
    )
}
