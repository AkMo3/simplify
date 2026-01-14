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
import { ContainerStatusBadge, HealthStatusBadge } from '@/components/ui/StatusBadge'
import type { Application } from '@/types/api'
import { useState, useEffect } from 'react'
import { PortMappingInput, type PortMappingItem } from './PortMappingInput'
import { inspectImage, updateApplication } from '@/lib/api'
import { Button } from '@/components/ui/Button'

interface ApplicationDrawerProps {
  application: Application | null
  isOpen: boolean
  onClose: () => void
  initialEditPorts?: boolean
}

export function ApplicationDrawer({
  application,
  isOpen,
  onClose,
  initialEditPorts = false,
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

  const [isEditingPorts, setIsEditingPorts] = useState(false)
  const [editedPorts, setEditedPorts] = useState<PortMappingItem[]>([])
  const [exposedPorts, setExposedPorts] = useState<string[]>([])
  const [isSaving, setIsSaving] = useState(false)

  const handleStartEdit = async () => {
    setIsEditingPorts(true)

    // Transform ports for editing:
    // API returns: { "80/tcp": "0.0.0.0:8080" } (Container -> Host)
    // We want unique ID for each row to maintain focus stability
    const rawPorts = application.ports || {}
    const items: PortMappingItem[] = []

    Object.entries(rawPorts).forEach(([containerKey, hostVal]) => {
      // Clean container port: "80/tcp" -> "80"
      const containerPort = containerKey.split('/')[0]

      // Clean host port: "0.0.0.0:8080" -> "8080"
      // If hostVal is empty string, we skip
      if (!hostVal) return

      const hostPort = hostVal.includes(':') ? hostVal.split(':').pop() || '' : hostVal

      if (hostPort && containerPort) {
        items.push({
          id: self.crypto.randomUUID(),
          hostPort,
          containerPort
        })
      }
    })

    setEditedPorts(items)

    // Fetch exposed ports
    try {
      const info = await inspectImage(application.image)
      console.log({ info })
      setExposedPorts(info.exposed_ports || [])
    } catch (error) {
      console.error('Failed to inspect image:', error)
    }
  }

  const handleSavePorts = async () => {
    setIsSaving(true)
    try {
      // Convert array back to Record for API
      const portsRecord: Record<string, string> = {}
      editedPorts.forEach(item => {
        if (item.hostPort && item.containerPort) {
          portsRecord[item.hostPort] = item.containerPort
        }
      })

      await updateApplication(application.id, {
        name: application.name,
        image: application.image,
        ports: portsRecord,
      })
      // We should ideally reload the application data here.
      // For now, we rely on the parent updating or a page refresh, 
      // but let's assume the mutation/query invalidation happens elsewhere or close resets it.
      setIsEditingPorts(false)
      onClose() // Close drawer to refresh (simple way) or we need a callback to refresh
    } catch (error) {
      console.error('Failed to update ports:', error)
    } finally {
      setIsSaving(false)
    }
  }

  const getPortUrl = (hostPort: string) => {
    return `http://${hostPort}`
  }

  // Reset/Initialize state when drawer opens
  useEffect(() => {
    if (isOpen) {
      if (initialEditPorts) {
        handleStartEdit()
      } else {
        setIsEditingPorts(false)
      }
    }
  }, [isOpen, initialEditPorts])

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
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-medium flex items-center gap-2">
              <ExternalLink className="h-4 w-4 text-muted-foreground" />
              Port Mappings
            </h3>
            {!isEditingPorts ? (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 text-xs"
                onClick={handleStartEdit}
              >
                Edit
              </Button>
            ) : (
              <div className="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 text-xs"
                  onClick={() => setIsEditingPorts(false)}
                  disabled={isSaving}
                >
                  Cancel
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  className="h-7 text-xs"
                  onClick={handleSavePorts}
                  isLoading={isSaving}
                >
                  Save
                </Button>
              </div>
            )}
          </div>

          {isEditingPorts ? (
            <div className="bg-[hsl(0_0%_12%)] p-4 rounded-lg border border-border/50">
              <PortMappingInput
                value={editedPorts}
                onChange={setEditedPorts}
                exposedPorts={exposedPorts}
              />
            </div>
          ) : (
            <>
              {application.ports && Object.keys(application.ports).length > 0 ? (
                <div className="space-y-2">
                  {Object.entries(application.ports).map(([containerPort, hostPort]) => (
                    <div
                      key={hostPort}
                      className="flex items-center justify-between py-2 px-3 rounded-md bg-[hsl(0_0%_12%)]"
                    >
                      <div className="flex items-center gap-3">
                        <span className="font-mono text-sm">
                          <span className="text-muted-foreground">Container:</span> {containerPort}
                        </span>
                        <span className="text-muted-foreground">→</span>
                        <span className="font-mono text-sm">
                          <span className="text-muted-foreground">Host:</span> {hostPort}
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
            </>
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
