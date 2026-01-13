import { Link } from 'react-router-dom'
import {
  Box,
  ExternalLink,
  Calendar,
  Image,
  Layers,
  Terminal,
  ArrowRight,
} from 'lucide-react'
import { Drawer } from '@/components/ui/Drawer'
import { Button } from '@/components/ui/Button'
import { ContainerStatusBadge, HealthStatusBadge } from '@/components/ui/StatusBadge'
import type { Application } from '@/types/api'

interface ApplicationDrawerProps {
  application: Application | null
  isOpen: boolean
  onClose: () => void
}

export function ApplicationDrawer({
  application,
  isOpen,
  onClose,
}: ApplicationDrawerProps) {
  if (!application) return null

  const formatDate = (dateString: string) => {
    if (!dateString) return '-'
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      weekday: 'short',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const getPortUrl = (hostPort: string) => {
    return `http://localhost:${hostPort}`
  }

  return (
    <Drawer
      isOpen={isOpen}
      onClose={onClose}
      title={application.name}
      description={application.id}
      width="lg"
    >
      <div className="space-y-6">
        {/* Status Section */}
        <div className="flex items-center gap-3">
          <ContainerStatusBadge status={application.status || 'stopped'} />
          <HealthStatusBadge status={application.health_status || 'none'} />
        </div>

        {/* Quick Info Cards */}
        <div className="grid grid-cols-2 gap-3">
          <div className="rounded-lg bg-[hsl(0_0%_12%)] p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-sm">
              <Image className="h-4 w-4" />
              Image
            </div>
            <p className="mt-1.5 font-mono text-sm truncate" title={application.image}>
              {application.image}
            </p>
          </div>
          <div className="rounded-lg bg-[hsl(0_0%_12%)] p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-sm">
              <Layers className="h-4 w-4" />
              Replicas
            </div>
            <p className="mt-1.5 text-sm font-medium">
              {application.replicas || 1}
            </p>
          </div>
        </div>

        {/* Timestamps */}
        <div className="space-y-2">
          <div className="flex items-center justify-between py-2 border-b border-border/30">
            <span className="text-sm text-muted-foreground flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              Created
            </span>
            <span className="text-sm">{formatDate(application.created_at)}</span>
          </div>
          <div className="flex items-center justify-between py-2 border-b border-border/30">
            <span className="text-sm text-muted-foreground flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              Updated
            </span>
            <span className="text-sm">{formatDate(application.updated_at)}</span>
          </div>
        </div>

        {/* Port Mappings */}
        <div>
          <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
            <ExternalLink className="h-4 w-4 text-muted-foreground" />
            Port Mappings
          </h3>
          {application.ports && Object.keys(application.ports).length > 0 ? (
            <div className="space-y-2">
              {Object.entries(application.ports).map(([hostPort, containerPort]) => (
                <div
                  key={hostPort}
                  className="flex items-center justify-between py-2 px-3 rounded-md bg-[hsl(0_0%_12%)]"
                >
                  <div className="flex items-center gap-3">
                    <span className="font-mono text-sm">
                      <span className="text-muted-foreground">Host:</span> {hostPort}
                    </span>
                    <span className="text-muted-foreground">→</span>
                    <span className="font-mono text-sm">
                      <span className="text-muted-foreground">Container:</span> {containerPort}
                    </span>
                  </div>
                  <a
                    href={getPortUrl(hostPort)}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 px-2 py-1 rounded bg-[hsl(0_0%_18%)] text-xs hover:bg-[hsl(0_0%_22%)] transition-colors"
                  >
                    Open
                    <ExternalLink className="h-3 w-3" />
                  </a>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground py-3 px-3 rounded-md bg-[hsl(0_0%_10%)]">
              No port mappings configured
            </p>
          )}
        </div>

        {/* Environment Variables */}
        <div>
          <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
            <Terminal className="h-4 w-4 text-muted-foreground" />
            Environment Variables
          </h3>
          {application.env_vars && Object.keys(application.env_vars).length > 0 ? (
            <div className="space-y-2 max-h-48 overflow-y-auto">
              {Object.entries(application.env_vars).map(([key, value]) => (
                <div
                  key={key}
                  className="flex items-center gap-2 py-2 px-3 rounded-md bg-[hsl(0_0%_12%)] font-mono text-sm"
                >
                  <span className="text-[hsl(0_0%_70%)]">{key}</span>
                  <span className="text-muted-foreground">=</span>
                  <span className="truncate text-[hsl(0_0%_55%)]" title={value}>
                    {value}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground py-3 px-3 rounded-md bg-[hsl(0_0%_10%)]">
              No environment variables configured
            </p>
          )}
        </div>

        {/* Health Check Config */}
        {application.health_check && (
          <div>
            <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
              <Box className="h-4 w-4 text-muted-foreground" />
              Health Check Configuration
            </h3>
            <div className="grid grid-cols-2 gap-2">
              <div className="py-2 px-3 rounded-md bg-[hsl(0_0%_12%)]">
                <span className="text-xs text-muted-foreground">Path</span>
                <p className="font-mono text-sm">{application.health_check.path}</p>
              </div>
              <div className="py-2 px-3 rounded-md bg-[hsl(0_0%_12%)]">
                <span className="text-xs text-muted-foreground">Interval</span>
                <p className="text-sm">{application.health_check.interval}s</p>
              </div>
              <div className="py-2 px-3 rounded-md bg-[hsl(0_0%_12%)]">
                <span className="text-xs text-muted-foreground">Timeout</span>
                <p className="text-sm">{application.health_check.timeout}s</p>
              </div>
              <div className="py-2 px-3 rounded-md bg-[hsl(0_0%_12%)]">
                <span className="text-xs text-muted-foreground">Retries</span>
                <p className="text-sm">{application.health_check.retries}</p>
              </div>
            </div>
          </div>
        )}

        {/* View Full Details Button */}
        <div className="pt-4 border-t border-border/30">
          <Link to={`/applications/${application.id}`} onClick={onClose}>
            <Button className="w-full justify-center">
              View Full Details
              <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          </Link>
        </div>
      </div>
    </Drawer>
  )
}
