import { useIsFetching, useIsMutating } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'

export function SyncIndicator() {
    const isFetching = useIsFetching()
    const isMutating = useIsMutating()

    if (!isFetching && !isMutating) return null

    return (
        <div className="flex items-center gap-1.5 px-2 py-1 rounded bg-[hsl(0_0%_15%)] border border-[hsl(0_0%_20%)] transition-all duration-200">
            <RefreshCw className="h-3 w-3 animate-spin text-muted-foreground" />
            <span className="text-xs text-muted-foreground font-medium hidden md:block">
                {isMutating ? 'Updating...' : 'Syncing'}
            </span>
        </div>
    )
}
