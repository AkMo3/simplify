import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listNetworks, deleteNetwork } from '@/lib/api'
import { Button } from '@/components/ui/Button'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/Table'
import { Plus, Trash2, Network as NetworkIcon } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { CreateNetworkModal } from '@/components/networks/CreateNetworkModal'
import { toast } from 'sonner'
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/AlertDialog'

export default function Networks() {
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
    const [networkToDelete, setNetworkToDelete] = useState<string | null>(null)

    const queryClient = useQueryClient()

    const { data: networks, isLoading } = useQuery({
        queryKey: ['networks'],
        queryFn: listNetworks,
    })

    const deleteMutation = useMutation({
        mutationFn: deleteNetwork,
        onSuccess: () => {
            toast.success('Network deleted successfully')
            queryClient.invalidateQueries({ queryKey: ['networks'] })
            setNetworkToDelete(null)
        },
        onError: () => {
            toast.error('Failed to delete network')
        },
    })

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Networks</h1>
                    <p className="text-muted-foreground">
                        Manage bridge networks for container communication.
                    </p>
                </div>
                <Button onClick={() => setIsCreateModalOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Network
                </Button>
            </div>

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Name</TableHead>
                            <TableHead>Driver</TableHead>
                            <TableHead>Subnet</TableHead>
                            <TableHead>Created</TableHead>
                            <TableHead className="w-[100px]"></TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                                    Loading networks...
                                </TableCell>
                            </TableRow>
                        ) : networks?.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center h-32 text-muted-foreground">
                                    No networks found. Create one to get started.
                                </TableCell>
                            </TableRow>
                        ) : (
                            networks?.map((network) => (
                                <TableRow key={network.id}>
                                    <TableCell className="font-medium">
                                        <div className="flex items-center gap-2">
                                            <NetworkIcon className="h-4 w-4 text-blue-500" />
                                            {network.name}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <span className="inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80">
                                            {network.driver}
                                        </span>
                                    </TableCell>
                                    <TableCell className="font-mono text-sm">
                                        {network.subnet || '-'}
                                    </TableCell>
                                    <TableCell>
                                        {formatDistanceToNow(new Date(network.created_at), {
                                            addSuffix: true,
                                        })}
                                    </TableCell>
                                    <TableCell>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="text-muted-foreground hover:text-destructive"
                                            onClick={() => setNetworkToDelete(network.id)}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>

            <CreateNetworkModal
                isOpen={isCreateModalOpen}
                onClose={() => setIsCreateModalOpen(false)}
            />

            <AlertDialog open={!!networkToDelete} onOpenChange={() => setNetworkToDelete(null)}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                        <AlertDialogDescription>
                            This action cannot be undone. This will permanently delete the network.
                            Ensure no containers are attached to this network before deleting.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            onClick={() => networkToDelete && deleteMutation.mutate(networkToDelete)}
                        >
                            Delete
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    )
}
