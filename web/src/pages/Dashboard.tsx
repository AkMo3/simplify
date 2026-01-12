import { Link } from 'react-router-dom'
import { Activity, Box, CheckCircle, XCircle, ArrowRight } from 'lucide-react'
import { useHealth, useReadiness, useApplications } from '@/hooks/useApplications'
import { cn } from '@/lib/utils'

export function Dashboard() {
  const healthQuery = useHealth()
  const readyQuery = useReadiness()
  const appsQuery = useApplications()

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground mt-1">
          Overview of your Simplify cluster
        </p>
      </div>

      {/* Status cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Server Health */}
        <div className="card p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={cn(
                'p-2 rounded-md',
                healthQuery.data?.status === 'healthy' ? 'bg-success/10' : 'bg-destructive/10'
              )}>
                <Activity className={cn(
                  'h-5 w-5',
                  healthQuery.data?.status === 'healthy' ? 'text-success' : 'text-destructive'
                )} />
              </div>
              <div>
                <p className="text-sm font-medium">Server Health</p>
                <p className="text-xs text-muted-foreground">Liveness probe</p>
              </div>
            </div>
            {healthQuery.isLoading ? (
              <span className="text-sm text-muted-foreground">Loading...</span>
            ) : healthQuery.data?.status === 'healthy' ? (
              <CheckCircle className="h-5 w-5 text-success" />
            ) : (
              <XCircle className="h-5 w-5 text-destructive" />
            )}
          </div>
        </div>

        {/* Database Status */}
        <div className="card p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={cn(
                'p-2 rounded-md',
                readyQuery.data?.checks?.database?.status === 'healthy'
                  ? 'bg-success/10'
                  : 'bg-destructive/10'
              )}>
                <Box className={cn(
                  'h-5 w-5',
                  readyQuery.data?.checks?.database?.status === 'healthy'
                    ? 'text-success'
                    : 'text-destructive'
                )} />
              </div>
              <div>
                <p className="text-sm font-medium">Database</p>
                <p className="text-xs text-muted-foreground">BoltDB connection</p>
              </div>
            </div>
            {readyQuery.isLoading ? (
              <span className="text-sm text-muted-foreground">Loading...</span>
            ) : readyQuery.data?.checks?.database?.status === 'healthy' ? (
              <CheckCircle className="h-5 w-5 text-success" />
            ) : (
              <XCircle className="h-5 w-5 text-destructive" />
            )}
          </div>
        </div>

        {/* Application Count */}
        <Link to="/applications" className="card p-4 group hover:border-primary/50 transition-colors">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-md bg-primary/10">
                <Box className="h-5 w-5 text-primary" />
              </div>
              <div>
                <p className="text-sm font-medium">Applications</p>
                <p className="text-xs text-muted-foreground">Total deployed</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">
                {appsQuery.isLoading ? '...' : appsQuery.data?.length ?? 0}
              </span>
              <ArrowRight className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
            </div>
          </div>
        </Link>
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Link to="/applications" className="card p-6 group hover:border-primary/50 transition-colors">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-medium">Manage Applications</h3>
              <p className="text-sm text-muted-foreground mt-1">
                Create, view, and delete containerized applications
              </p>
            </div>
            <ArrowRight className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
          </div>
        </Link>
        <Link to="/health" className="card p-6 group hover:border-primary/50 transition-colors">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-medium">Health Status</h3>
              <p className="text-sm text-muted-foreground mt-1">
                View detailed health checks for all components
              </p>
            </div>
            <ArrowRight className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
          </div>
        </Link>
      </div>
    </div>
  )
}
