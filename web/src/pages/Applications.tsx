import { useState, useMemo } from 'react'
import { Plus, Trash2, RefreshCw, Box, ExternalLink, Eye, Pencil } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui'
import { ContainerStatusBadge, HealthStatusBadge } from '@/components/ui/StatusBadge'
import { CreateApplicationForm } from '@/components/applications/CreateApplicationForm'
import { ApplicationDrawer } from '@/components/applications/ApplicationDrawer'
import { useApplications, useDeleteApplication } from '@/hooks/useApplications'
import { useSearch } from '@/contexts/SearchContext'
import type { Application } from '@/types/api'
import { Link } from 'react-router-dom'

const MAX_VISIBLE_APPS = 6

export function Applications() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [selectedApp, setSelectedApp] = useState<Application | null>(null)
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const [openInEditMode, setOpenInEditMode] = useState(false)

  const { data: applications, isLoading, isError, refetch } = useApplications()
  const deleteMutation = useDeleteApplication()
  const { query } = useSearch()

  // Filter, sort, and limit applications
  const displayedApplications = useMemo(() => {
    if (!applications) return []

    let filtered = applications

    // Client-side search filtering by name
    if (query.trim()) {
      const searchLower = query.toLowerCase()
      filtered = filtered.filter((app) =>
        app.name.toLowerCase().includes(searchLower)
      )
    }

    // Sort by updated_at descending (most recent first)
    const sorted = [...filtered].sort((a, b) => {
      const dateA = new Date(a.updated_at || a.created_at).getTime()
      const dateB = new Date(b.updated_at || b.created_at).getTime()
      return dateB - dateA
    })

    // Limit to MAX_VISIBLE_APPS
    return sorted.slice(0, MAX_VISIBLE_APPS)
  }, [applications, query])

  const handleDelete = async (e: React.MouseEvent, app: Application) => {
    e.stopPropagation() // Prevent row click
    if (!confirm(`Are you sure you want to delete "${app.name}"?`)) return

    setDeletingId(app.id)
    try {
      await deleteMutation.mutateAsync(app.id)
    } finally {
      setDeletingId(null)
    }
  }

  const handleRowClick = (app: Application) => {
    setSelectedApp(app)
    setOpenInEditMode(false)
    setIsDrawerOpen(true)
  }

  const handleDrawerClose = () => {
    setIsDrawerOpen(false)
    // Delay clearing selected app to allow exit animation
    setTimeout(() => {
      setSelectedApp(null)
      setOpenInEditMode(false)
    }, 200)
  }

  const formatDate = (dateString: string) => {
    if (!dateString) return '-'
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  // Generate port URL (localhost for now, will be domain later with Traefik)
  const getPortUrl = (hostPort: string) => {
    return `http://localhost:${hostPort}`
  }

  // Get display ports (show first 2, then +N more)
  const getPortsDisplay = (ports: Record<string, string> | undefined) => {
    if (!ports || Object.keys(ports).length === 0) {
      return null
    }

    const portEntries = Object.entries(ports)
    const visiblePorts = portEntries.slice(0, 2)
    const remainingCount = portEntries.length - 2

    return (
      <div className="flex items-center gap-1.5 flex-wrap">
        {visiblePorts.map(([_, hostVal]) => {
          // Parse host port (e.g. "0.0.0.0:8080" -> "8080")
          const hostPort = hostVal.includes(':') ? hostVal.split(':').pop() : hostVal
          if (!hostPort) return null

          return (
            <a
              key={hostVal}
              href={getPortUrl(hostPort)}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
              className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-[hsl(0_0%_15%)] text-xs font-mono text-[hsl(0_0%_70%)] hover:bg-[hsl(0_0%_20%)] hover:text-foreground transition-colors"
            >
              :{hostPort}
              <ExternalLink className="h-3 w-3" />
            </a>
          )
        })}
        {remainingCount > 0 && (
          <span className="text-xs text-muted-foreground">
            +{remainingCount} more
          </span>
        )}
      </div>
    )
  }

  return (
    <div className="animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Applications</h1>
          <p className="text-muted-foreground mt-1">
            Manage your containerized applications
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => refetch()}
            disabled={isLoading}
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
          </Button>
          <Button onClick={() => setIsCreateModalOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Create Application
          </Button>
        </div>
      </div>

      {/* Search indicator */}
      {query && (
        <div className="text-sm text-muted-foreground">
          Showing results for "<span className="text-foreground">{query}</span>"
          {displayedApplications.length === 0 && ' — No applications found'}
        </div>
      )}

      {/* Content */}
      {isLoading ? (
        <div className="card p-12 text-center">
          <RefreshCw className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
          <p className="mt-4 text-muted-foreground">Loading applications...</p>
        </div>
      ) : isError ? (
        <div className="card p-12 text-center">
          <p className="text-destructive">Failed to load applications</p>
          <Button variant="secondary" className="mt-4" onClick={() => refetch()}>
            Try Again
          </Button>
        </div>
      ) : applications?.length === 0 ? (
        <div className="card p-12 text-center">
          <Box className="h-12 w-12 mx-auto text-muted-foreground/50" />
          <h3 className="mt-4 text-lg font-medium">No applications yet</h3>
          <p className="mt-1 text-muted-foreground">
            Create your first application to get started
          </p>
          <Button className="mt-4" onClick={() => setIsCreateModalOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Create Application
          </Button>
        </div>
      ) : displayedApplications.length === 0 ? (
        <div className="card p-12 text-center">
          <Box className="h-12 w-12 mx-auto text-muted-foreground/50" />
          <h3 className="mt-4 text-lg font-medium">No matching applications</h3>
          <p className="mt-1 text-muted-foreground">
            Try adjusting your search query
          </p>
        </div>
      ) : (
        <div className="card overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border/50 bg-[hsl(0_0%_9%)]">
                <th className="table-header">Name</th>
                <th className="table-header">Status</th>
                <th className="table-header">Ports</th>
                <th className="table-header">Health</th>
                <th className="table-header">Created</th>
                <th className="table-header text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              <AnimatePresence>
                {displayedApplications.map((app) => (
                  <motion.tr
                    key={app.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0, height: 0 }}
                    onClick={() => handleRowClick(app)}
                    className="table-row cursor-pointer"
                  >
                    {/* Name */}
                    <td className="table-cell">
                      <div className="flex items-center gap-3 group">
                        <div className="h-9 w-9 rounded-lg bg-[hsl(0_0%_15%)] flex items-center justify-center">
                          <Box className="h-4 w-4 text-[hsl(0_0%_60%)]" />
                        </div>
                        <div>
                          <p className="font-medium group-hover:text-foreground transition-colors">
                            {app.name}
                          </p>
                          <p className="text-xs text-muted-foreground font-mono">
                            {app.image}
                          </p>
                        </div>
                      </div>
                    </td>

                    {/* Status */}
                    <td className="table-cell">
                      <ContainerStatusBadge status={app.status || 'stopped'} />
                    </td>

                    {/* Ports */}
                    <td className="table-cell">
                      {getPortsDisplay(app.ports) || (
                        <span className="text-xs text-muted-foreground">—</span>
                      )}
                    </td>

                    {/* Health */}
                    <td className="table-cell">
                      <HealthStatusBadge status={app.health_status || 'none'} />
                    </td>

                    {/* Created */}
                    <td className="table-cell text-sm text-muted-foreground">
                      {formatDate(app.created_at)}
                    </td>

                    {/* Actions */}
                    {/* Actions */}
                    <td className="table-cell text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Link to={`/applications/${app.id}`}>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-muted-foreground hover:text-foreground h-8 w-8 px-0"
                            title="View Details"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => handleDelete(e, app)}
                          disabled={deletingId === app.id}
                          className="text-muted-foreground hover:text-destructive h-8 w-8 px-0"
                          title="Delete Application"
                        >
                          {deletingId === app.id ? (
                            <RefreshCw className="h-4 w-4 animate-spin" />
                          ) : (
                            <Trash2 className="h-4 w-4" />
                          )}
                        </Button>
                      </div>
                    </td>
                  </motion.tr>
                ))}
              </AnimatePresence>
            </tbody>
          </table>

          {/* Show more indicator */}
          {applications && applications.length > MAX_VISIBLE_APPS && !query && (
            <div className="px-4 py-3 border-t border-border/50 text-center">
              <span className="text-sm text-muted-foreground">
                Showing {MAX_VISIBLE_APPS} of {applications.length} applications
              </span>
            </div>
          )}
        </div>
      )}

      {/* Create Modal */}
      <CreateApplicationForm
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />

      {/* Application Details Drawer */}
      <ApplicationDrawer
        application={selectedApp}
        isOpen={isDrawerOpen}
        onClose={handleDrawerClose}
        initialEditPorts={openInEditMode}
      />
    </div>
  )
}
