import { useState } from 'react'
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/Dialog'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Label } from '@/components/ui/Label'
import { createNetwork } from '@/lib/api'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

interface CreateNetworkModalProps {
    isOpen: boolean
    onClose: () => void
}

export function CreateNetworkModal({ isOpen, onClose }: CreateNetworkModalProps) {
    const [name, setName] = useState('')
    const [isSubmitting, setIsSubmitting] = useState(false)

    const queryClient = useQueryClient()

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setIsSubmitting(true)

        try {
            await createNetwork({ name })
            toast.success('Network created successfully')
            queryClient.invalidateQueries({ queryKey: ['networks'] })
            onClose()
            setName('')
        } catch (error) {
            toast.error('Failed to create network')
            console.error(error)
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Create Network</DialogTitle>
                </DialogHeader>

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="name">Network Name</Label>
                        <Input
                            id="name"
                            placeholder="my-bridge-network"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            required
                        />
                        <p className="text-sm text-muted-foreground">
                            A bridge network allows containers to communicate via DNS names.
                        </p>
                    </div>

                    <div className="flex justify-end gap-2">
                        <Button type="button" variant="outline" onClick={onClose}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={isSubmitting}>
                            {isSubmitting ? 'Creating...' : 'Create Network'}
                        </Button>
                    </div>
                </form>
            </DialogContent>
        </Dialog>
    )
}
