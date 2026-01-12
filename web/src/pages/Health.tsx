import { motion } from 'framer-motion'
import {
  Activity,
  Database,
  Container,
  CheckCircle,
  XCircle,
  RefreshCw,
  Clock
} from 'lucide-react'
import { useHealth, useReadiness } from '@/hooks/useApplications'
import { Button, Badge } from '@/components/ui'
import { cn } from '@/lib/utils'

interface HealthCardProps {
  title: string
  description: string
  icon: React.ElementType
  status: 'healthy' | 'unhealthy' | 'loading'
  message?: string
}

function HealthCard({ title, description, icon: Icon, status, message }: HealthCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className="card p-6"
    >
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-4">
          <div
            className={cn(
              'p-3 rounded-lg',
              status === 'healthy' && 'bg-success/10',
              status === 'unhealthy' && 'bg-destructive/10',
              status === 'loading' && 'bg-muted'
            )}
          >
            <Icon
              className={cn(
                'h-6 w-6',
                status === 'healthy' && 'text-success',
                status === 'unhealthy' && 'text-destructive',
                status === 'loading' && 'text-muted-foreground'
              )}
            />
          </div>
          <div>
            <h3 className="font-medium">{title}</h3>
            <p className="text-sm text-muted-foreground mt-1">{description}</p>
            {message && status === 'unhealthy' && (
              <p className="text-sm text-destructive mt-2 font-mono bg-destructive/5 px-2 py-1 rounded">
                {message}
              </p>
            )}
          </div>
        </div>
        <div>
          {status === 'loading' ? (
            <Badge variant="outline">
              <RefreshCw className="h-3 w-3 mr-1 animate-spin" />
              Checking
            </Badge>
          ) : status === 'healthy' ? (
            <Badge variant="success">
              <CheckCircle className="h-3 w-3 mr-1" />
              Healthy
            </Badge>
          ) : (
            <Badge variant="destructive">
              <XCircle className="h-3 w-3 mr-1" />
              Unhealthy
            </Badge>
          )}
        </div>
      </div>
    </motion.div>
  )
}

export function Health() {
  const { data: health, isLoading: healthLoading, refetch: refetchHealth, dataUpdatedAt: healthUpdatedAt } = useHealth()
  const { data: readiness, isLoading: readyLoading, refetch: refetchReadiness, dataUpdatedAt: readyUpdatedAt } = useReadiness()

  const isLoading = healthLoading || readyLoading

  const handleRefresh = () => {
    refetchHealth()
    refetchReadiness()
  }

  const formatLastUpdated = (timestamp: number) => {
    if (!timestamp) return 'Never'
    const date = new Date(timestamp)
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }

  const lastUpdated = Math.max(healthUpdatedAt || 0, readyUpdatedAt || 0)

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Health Status</h1>
          <p className="text-muted-foreground mt-1">
            Monitor the health of your Simplify cluster components
          </p>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Clock className="h-4 w-4" />
            <span>Last updated: {formatLastUpdated(lastUpdated)}</span>
          </div>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleRefresh}
            disabled={isLoading}
          >
            <RefreshCw className={cn('h-4 w-4 mr-2', isLoading && 'animate-spin')} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Overall Status */}
      <div className="card p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div
              className={cn(
                'h-4 w-4 rounded-full',
                isLoading && 'bg-muted animate-pulse',
                !isLoading && health?.status === 'healthy' && readiness?.status === 'healthy' && 'bg-success',
                !isLoading && (health?.status !== 'healthy' || readiness?.status !== 'healthy') && 'bg-destructive'
              )}
            />
            <div>
              <h2 className="text-lg font-medium">Overall Cluster Status</h2>
              <p className="text-sm text-muted-foreground">
                {isLoading
                  ? 'Checking cluster health...'
                  : health?.status === 'healthy' && readiness?.status === 'healthy'
                    ? 'All systems operational'
                    : 'Some components are experiencing issues'}
              </p>
            </div>
          </div>
          {!isLoading && (
            <Badge
              variant={
                health?.status === 'healthy' && readiness?.status === 'healthy'
                  ? 'success'
                  : 'destructive'
              }
              className="text-sm px-3 py-1"
            >
              {health?.status === 'healthy' && readiness?.status === 'healthy'
                ? 'Operational'
                : 'Degraded'}
            </Badge>
          )}
        </div>
      </div>

      {/* Component Health Cards */}
      <div className="grid gap-4">
        <HealthCard
          title="API Server"
          description="HTTP API endpoint liveness check"
          icon={Activity}
          status={healthLoading ? 'loading' : health?.status === 'healthy' ? 'healthy' : 'unhealthy'}
        />

        <HealthCard
          title="Database"
          description="BoltDB connection and read/write capability"
          icon={Database}
          status={
            readyLoading
              ? 'loading'
              : readiness?.checks?.database?.status === 'healthy'
                ? 'healthy'
                : 'unhealthy'
          }
          message={readiness?.checks?.database?.message}
        />

        <HealthCard
          title="Container Runtime"
          description="Podman socket connection and container management"
          icon={Container}
          status={
            readyLoading
              ? 'loading'
              : readiness?.checks?.podman?.status === 'healthy'
                ? 'healthy'
                : 'unhealthy'
          }
          message={readiness?.checks?.podman?.message}
        />
      </div>

      {/* Info Card */}
      <div className="card p-6 bg-muted/30">
        <h3 className="font-medium mb-2">About Health Checks</h3>
        <div className="text-sm text-muted-foreground space-y-2">
          <p>
            <strong>/healthz</strong> — Liveness probe. Returns healthy if the API server is running.
          </p>
          <p>
            <strong>/readyz</strong> — Readiness probe. Returns healthy only if all dependencies
            (database, container runtime) are accessible and ready to serve requests.
          </p>
          <p className="pt-2">
            Health checks are automatically refreshed every 10 seconds.
          </p>
        </div>
      </div>
    </div>
  )
}
