import { Plus, Trash2, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/Button'

export interface PortMappingItem {
    id: string
    hostPort: string
    containerPort: string
}

interface PortMappingInputProps {
    value: PortMappingItem[]
    onChange: (value: PortMappingItem[]) => void
    exposedPorts?: string[]
}

export function PortMappingInput({
    value,
    onChange,
    exposedPorts = [],
}: PortMappingInputProps) {
    const handleAddPort = () => {
        onChange([
            ...value,
            { id: self.crypto.randomUUID(), hostPort: '', containerPort: '' }
        ])
    }

    const handleRemovePort = (id: string) => {
        onChange(value.filter((item) => item.id !== id))
    }

    const handlePortChange = (
        id: string,
        field: 'hostPort' | 'containerPort',
        newValue: string
    ) => {
        // Validation: only allow numbers
        if (newValue && !/^\d*$/.test(newValue)) return

        onChange(value.map((item) => {
            if (item.id === id) {
                return { ...item, [field]: newValue }
            }
            return item
        }))
    }

    // Helper for adding exposed ports
    const addExposedPort = (port: string) => {
        // Strip protocol like /tcp or /udp
        let cleanPort = port.split('/')[0]

        // Add new mapping
        onChange([
            ...value,
            {
                id: self.crypto.randomUUID(),
                hostPort: cleanPort, // Suggest same host port
                containerPort: cleanPort
            }
        ])
    }

    return (
        <div className="space-y-3">
            <div className="space-y-2">
                {value.map((item) => (
                    <div key={item.id} className="flex items-center gap-2">
                        <div className="grid grid-cols-[1fr,auto,1fr] gap-2 items-center flex-1">
                            <input
                                type="text"
                                value={item.containerPort}
                                onChange={(e) => handlePortChange(item.id, 'containerPort', e.target.value)}
                                placeholder="Container Port"
                                className="w-full px-3 py-2 bg-[hsl(0_0%_12%)] border border-border/50 rounded-md text-sm font-mono focus:outline-none focus:border-primary/50"
                            />
                            <ArrowRight className="h-4 w-4 text-muted-foreground" />
                            <input
                                type="text"
                                value={item.hostPort}
                                onChange={(e) => handlePortChange(item.id, 'hostPort', e.target.value)}
                                placeholder="Host Port"
                                className="w-full px-3 py-2 bg-[hsl(0_0%_12%)] border border-border/50 rounded-md text-sm font-mono focus:outline-none focus:border-primary/50"
                            />
                        </div>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRemovePort(item.id)}
                            className="h-10 w-10 p-0 text-muted-foreground hover:text-destructive"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
                ))}
            </div>

            <div className="flex items-center gap-4">
                <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={handleAddPort}
                    className="h-8"
                >
                    <Plus className="h-4 w-4 mr-2" />
                    Add Port Mapping
                </Button>

                {exposedPorts.length > 0 && (
                    <div className="flex items-center gap-2 pl-2 border-l border-border/30">
                        <span className="text-xs text-muted-foreground">Suggestions:</span>
                        {exposedPorts.map((port) => {
                            const cleanPort = port.split('/')[0]
                            // Check if this container port is already mapped
                            const isMapped = value.some(item => item.containerPort === cleanPort)

                            if (isMapped) return null

                            return (
                                <button
                                    key={port}
                                    type="button"
                                    onClick={() => addExposedPort(port)}
                                    className="px-2 py-1 rounded bg-primary/10 text-primary hover:bg-primary/20 text-xs font-mono transition-colors"
                                >
                                    +{cleanPort}
                                </button>
                            )
                        })}
                    </div>
                )}
            </div>
        </div>
    )
}
