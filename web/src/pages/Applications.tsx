import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Trash2, RefreshCw, Box, Eye } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button, Badge } from '@/components/ui'
import { CreateApplicationForm } from '@/components/applications/CreateApplicationForm'
import { useApplications, useDeleteApplication } from '@/hooks/useApplications'
import { cn } from '@/lib/utils'
import type { Application } from '@/types/api'

export function Applications() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const { data: applications, isLoading, isError, refetch } = useApplications()
  const deleteMutation = useDeleteApplication()

  const handleDelete = async (app: Application) => {
    if (!confirm(`Are you sure you want to delete "${app.name}"?`)) return

    setDeletingId(app.id)
    try {
      await deleteMutation.mutateAsync(app.id)
    } finally {
      setDeletingId(null)
    }
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

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
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
            <RefreshCw className={cn('h-4 w-4', isLoading && 'animate-spin')} />
          </Button>
          <Button onClick={() => setIsCreateModalOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Create Application
          </Button>
        </div>
      </div>

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
      ) : (
        <div className="card overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left text-xs font-medium text-muted-foreground uppercase tracking-wider px-4 py-3">
                  Name
                </th>
                <th className="text-left text-xs font-medium text-muted-foreground uppercase tracking-wider px-4 py-3">
                  Image
                </th>
                <th className="text-left text-xs font-medium text-muted-foreground uppercase tracking-wider px-4 py-3">
                  Replicas
                </th>
                <th className="text-left text-xs font-medium text-muted-foreground uppercase tracking-wider px-4 py-3">
                  Created
                </th>
                <th className="text-right text-xs font-medium text-muted-foreground uppercase tracking-wider px-4 py-3">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              <AnimatePresence>
                {applications?.map((app) => (
                  <motion.tr
                    key={app.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0, height: 0 }}
                    className="border-b last:border-0 hover:bg-muted/30 transition-colors"
                  >
                    <td className="px-4 py-3">
                      <Link to={`/applications/${app.id}`} className="flex items-center gap-3 group">
                        <div className="h-8 w-8 rounded-md bg-primary/10 flex items-center justify-center">
                          <Box className="h-4 w-4 text-primary" />
                        </div>
                        <div>
                          <p className="font-medium group-hover:text-primary transition-colors">{app.name}</p>
                          <p className="text-xs text-muted-foreground font-mono">
                            {app.id.slice(0, 8)}
                          </p>
                        </div>
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      <code className="text-sm bg-muted px-2 py-1 rounded">
                        {app.image}
                      </code>
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant="outline">{app.replicas || 1}</Badge>
                    </td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">
                      {formatDate(app.created_at)}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Link to={`/applications/${app.id}`}>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-muted-foreground hover:text-foreground"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(app)}
                          disabled={deletingId === app.id}
                          className="text-muted-foreground hover:text-destructive"
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
        </div>
      )}

      {/* Create Modal */}
      <CreateApplicationForm
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />
    </div>
  )
}
