import { cn } from '@/lib/utils'
import type { ContainerStatus, HealthCheckStatus } from '@/types/api'

// =============================================================================
// Container Status Badge
// =============================================================================

interface ContainerStatusBadgeProps {
  status: ContainerStatus
  className?: string
}

const containerStatusConfig: Record<ContainerStatus, { label: string; className: string }> = {
  creating: {
    label: 'Creating',
    className: 'bg-blue-500/15 text-blue-400',
  },
  created: {
    label: 'Created',
    className: 'bg-emerald-500/15 text-emerald-400',
  },
  running: {
    label: 'Running',
    className: 'bg-emerald-500/15 text-emerald-400',
  },
  restarting: {
    label: 'Restarting',
    className: 'bg-yellow-500/15 text-yellow-400',
  },
  stopping: {
    label: 'Stopping',
    className: 'bg-orange-500/15 text-orange-400',
  },
  stopped: {
    label: 'Stopped',
    className: 'bg-[hsl(0_0%_25%)] text-[hsl(0_0%_55%)]',
  },
  error: {
    label: 'Error',
    className: 'bg-red-500/15 text-red-400',
  },
}

export function ContainerStatusBadge({ status, className }: ContainerStatusBadgeProps) {
  const config = containerStatusConfig[status] || containerStatusConfig.stopped

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium',
        config.className,
        className
      )}
    >
      <span
        className={cn(
          'h-1.5 w-1.5 rounded-full',
          status === 'running' && 'bg-emerald-400 animate-pulse',
          status === 'created' && 'bg-emerald-400 animate-pulse',
          status === 'creating' && 'bg-blue-400 animate-pulse',
          status === 'restarting' && 'bg-yellow-400 animate-pulse',
          status === 'stopping' && 'bg-orange-400',
          status === 'stopped' && 'bg-[hsl(0_0%_45%)]',
          status === 'error' && 'bg-red-400'
        )}
      />
      {config.label}
    </span>
  )
}

// =============================================================================
// Health Status Badge
// =============================================================================

interface HealthStatusBadgeProps {
  status: HealthCheckStatus
  className?: string
}

const healthStatusConfig: Record<HealthCheckStatus, { label: string; className: string }> = {
  healthy: {
    label: 'Healthy',
    className: 'bg-emerald-500/15 text-emerald-400',
  },
  unhealthy: {
    label: 'Unhealthy',
    className: 'bg-red-500/15 text-red-400',
  },
  starting: {
    label: 'Starting',
    className: 'bg-yellow-500/15 text-yellow-400',
  },
  none: {
    label: 'No Check',
    className: 'bg-[hsl(0_0%_20%)] text-[hsl(0_0%_50%)]',
  },
}

export function HealthStatusBadge({ status, className }: HealthStatusBadgeProps) {
  const config = healthStatusConfig[status] || healthStatusConfig.none

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        config.className,
        className
      )}
    >
      {config.label}
    </span>
  )
}
