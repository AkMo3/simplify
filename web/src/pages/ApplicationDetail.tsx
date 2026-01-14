import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  ArrowLeft,
  Box,
  Trash2,
  RefreshCw,
  Calendar,
  Image,
  Layers,
  Terminal,
  Settings,
  ExternalLink,
} from 'lucide-react'
import { Button, Badge, Modal } from '@/components/ui'
import { useApplication, useDeleteApplication } from '@/hooks/useApplications' // Assuming we might need update hook but using direct api for now or maybe add it 
import { PortMappingInput, type PortMappingItem } from '@/components/applications/PortMappingInput'
import { inspectImage, updateApplication } from '@/lib/api'

export function ApplicationDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [showLogs, setShowLogs] = useState(false)

  const { data: app, isLoading, isError, refetch } = useApplication(id || '')
  const deleteMutation = useDeleteApplication()

  const handleDelete = async () => {
    if (!id) return
    try {
      await deleteMutation.mutateAsync(id)
      navigate('/applications')
    } catch (error) {
      console.error('Failed to delete application:', error)
    }
  }

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
    if (!app) return
    setIsEditingPorts(true)

    // Transform ports for editing:
    // API returns: { "80/tcp": "0.0.0.0:8080" } (Container -> Host)
    // We want unique ID for each row to maintain focus stability
    const rawPorts = app.ports || {}
    const items: PortMappingItem[] = []

    Object.entries(rawPorts).forEach(([hostVal, containerKey]) => {
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
      const info = await inspectImage(app.image)
      setExposedPorts(info.exposed_ports || [])
    } catch (error) {
      console.error('Failed to inspect image:', error)
    }
  }

  const handleSavePorts = async () => {
    if (!app) return
    setIsSaving(true)
    try {
      // Convert array back to Record for API
      const portsRecord: Record<string, string> = {}
      editedPorts.forEach(item => {
        if (item.hostPort && item.containerPort) {
          portsRecord[item.hostPort] = item.containerPort
        }
      })

      await updateApplication(app.id, {
        name: app.name,
        image: app.image,
        ports: portsRecord,
      })

      await refetch()
      setIsEditingPorts(false)
    } catch (error) {
      console.error('Failed to update ports:', error)
    } finally {
      setIsSaving(false)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (isError || !app) {
    return (
      <div className="space-y-6 animate-fade-in">
        <Link
          to="/applications"
          className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Applications
        </Link>
        <div className="card p-12 text-center">
          <Box className="h-12 w-12 mx-auto text-muted-foreground/50" />
          <h3 className="mt-4 text-lg font-medium">Application not found</h3>
          <p className="mt-1 text-muted-foreground">
            The application you're looking for doesn't exist or has been deleted.
          </p>
          <Button className="mt-4" onClick={() => navigate('/applications')}>
            View All Applications
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Breadcrumb */}
      <Link
        to="/applications"
        className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft className="h-4 w-4 mr-2" />
        Back to Applications
      </Link>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-4">
          <motion.div
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="h-14 w-14 rounded-lg bg-primary/10 flex items-center justify-center"
          >
            <Box className="h-7 w-7 text-primary" />
          </motion.div>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{app.name}</h1>
            <p className="text-muted-foreground font-mono text-sm mt-1">{app.id}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4" />
          </Button>
          <Button
            variant="destructive"
            size="sm"
            onClick={() => setShowDeleteModal(true)}
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left Column - Details */}
        <div className="lg:col-span-2 space-y-6">
          {/* Container Info */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="card p-6"
          >
            <h2 className="text-lg font-medium mb-4 flex items-center gap-2">
              <Settings className="h-5 w-5 text-muted-foreground" />
              Container Configuration
            </h2>
            <dl className="grid gap-4 sm:grid-cols-2">
              <div>
                <dt className="text-sm text-muted-foreground flex items-center gap-2">
                  <Image className="h-4 w-4" />
                  Image
                </dt>
                <dd className="mt-1">
                  <code className="text-sm bg-muted px-2 py-1 rounded">
                    {app.image}
                  </code>
                </dd>
              </div>
              <div>
                <dt className="text-sm text-muted-foreground flex items-center gap-2">
                  <Layers className="h-4 w-4" />
                  Replicas
                </dt>
                <dd className="mt-1">
                  <Badge variant="outline">{app.replicas || 1}</Badge>
                </dd>
              </div>
              <div>
                <dt className="text-sm text-muted-foreground flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  Created
                </dt>
                <dd className="mt-1 text-sm">{formatDate(app.created_at)}</dd>
              </div>
              <div>
                <dt className="text-sm text-muted-foreground flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  Updated
                </dt>
                <dd className="mt-1 text-sm">{formatDate(app.updated_at)}</dd>
              </div>
            </dl>
          </motion.div>

          {/* Ports */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="card p-6"
          >
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-medium flex items-center gap-2">
                <ExternalLink className="h-5 w-5 text-muted-foreground" />
                Port Mappings
              </h2>
              {!isEditingPorts ? (
                <Button variant="ghost" size="sm" onClick={handleStartEdit}>
                  Edit
                </Button>
              ) : (
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setIsEditingPorts(false)}
                    disabled={isSaving}
                  >
                    Cancel
                  </Button>
                  <Button
                    size="sm"
                    onClick={handleSavePorts}
                    isLoading={isSaving}
                  >
                    Save
                  </Button>
                </div>
              )}
            </div>

            {isEditingPorts ? (
              <PortMappingInput
                value={editedPorts}
                onChange={setEditedPorts}
                exposedPorts={exposedPorts}
              />
            ) : (
              app.ports && Object.keys(app.ports).length > 0 ? (
                <div className="space-y-2">
                  {Object.entries(app.ports).map(([hostPort, containerPort]) => (
                    <div
                      key={hostPort}
                      className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-md"
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
                        href={`http://localhost:${hostPort}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-1 px-2 py-1 rounded bg-background/50 text-xs hover:bg-background transition-colors"
                      >
                        Open
                        <ExternalLink className="h-3 w-3" />
                      </a>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No port mappings configured</p>
              )
            )}
          </motion.div>

          {/* Environment Variables */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className="card p-6"
          >
            <h2 className="text-lg font-medium mb-4 flex items-center gap-2">
              <Terminal className="h-5 w-5 text-muted-foreground" />
              Environment Variables
            </h2>
            {app.env_vars && Object.keys(app.env_vars).length > 0 ? (
              <div className="space-y-2">
                {Object.entries(app.env_vars).map(([key, value]) => (
                  <div
                    key={key}
                    className="flex items-center gap-2 py-2 px-3 bg-muted/50 rounded-md font-mono text-sm"
                  >
                    <span className="text-primary">{key}</span>
                    <span className="text-muted-foreground">=</span>
                    <span className="truncate">{value}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No environment variables configured</p>
            )}
          </motion.div>
        </div>

        {/* Right Column - Actions & Quick Info */}
        <div className="space-y-6">
          {/* Quick Actions */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.15 }}
            className="card p-6"
          >
            <h2 className="text-lg font-medium mb-4">Quick Actions</h2>
            <div className="space-y-2">
              <Button
                variant="secondary"
                className="w-full justify-start"
                onClick={() => setShowLogs(true)}
              >
                <Terminal className="h-4 w-4 mr-2" />
                View Logs
              </Button>
              <Button variant="secondary" className="w-full justify-start" disabled>
                <RefreshCw className="h-4 w-4 mr-2" />
                Restart Container
                <Badge variant="outline" className="ml-auto text-xs">
                  Soon
                </Badge>
              </Button>
            </div>
          </motion.div>

          {/* Container Name */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.25 }}
            className="card p-6"
          >
            <h2 className="text-sm font-medium text-muted-foreground mb-2">
              Container Name
            </h2>
            <code className="text-sm bg-muted px-2 py-1 rounded block">
              simplify-{app.id}
            </code>
            <p className="text-xs text-muted-foreground mt-2">
              Used by the reconciler to manage the container
            </p>
          </motion.div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        title="Delete Application"
        description="Are you sure you want to delete this application? This action cannot be undone."
      >
        <div className="space-y-4">
          <div className="p-4 bg-destructive/10 rounded-lg">
            <p className="text-sm">
              Deleting <strong>{app.name}</strong> will:
            </p>
            <ul className="mt-2 text-sm text-muted-foreground list-disc list-inside space-y-1">
              <li>Remove the application from the database</li>
              <li>Stop and remove the running container</li>
              <li>Delete all associated configuration</li>
            </ul>
          </div>
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={() => setShowDeleteModal(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              isLoading={deleteMutation.isPending}
            >
              Delete Application
            </Button>
          </div>
        </div>
      </Modal>

      {/* Logs Modal */}
      <Modal
        isOpen={showLogs}
        onClose={() => setShowLogs(false)}
        title={`Logs: ${app.name}`}
        className="max-w-4xl"
      >
        <div className="space-y-4">
          <div className="bg-black rounded-lg p-4 h-96 overflow-auto font-mono text-sm text-green-400">
            <p className="text-muted-foreground">
              Log streaming is not yet implemented in the API.
            </p>
            <p className="text-muted-foreground mt-2">
              To view logs, use the CLI:
            </p>
            <code className="block mt-2 text-white">
              simplify logs simplify-{app.id} --follow
            </code>
          </div>
          <div className="flex justify-end">
            <Button variant="secondary" onClick={() => setShowLogs(false)}>
              Close
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
