import { useState, useEffect, type FormEvent } from 'react'
import { Loader2 } from 'lucide-react'
import { Button, Input, Modal } from '@/components/ui'
import { useCreateApplication } from '@/hooks/useApplications'
import { usePods } from '@/hooks/usePods'
import { useNetworks } from '@/hooks/useNetworks'
import { ApiClientError, inspectImage } from '@/lib/api'
import { PortMappingInput, type PortMappingItem } from './PortMappingInput'

interface CreateApplicationFormProps {
  isOpen: boolean
  onClose: () => void
  initialPodId?: string
}

interface FormData {
  name: string
  image: string
  replicas: string
  podId: string
  networkId: string
}

interface FormErrors {
  name?: string
  image?: string
  replicas?: string
  ports?: string
  general?: string
}

export function CreateApplicationForm({ isOpen, onClose, initialPodId }: CreateApplicationFormProps) {
  const [formData, setFormData] = useState<FormData>({
    name: '',
    image: '',
    replicas: '1',
    podId: initialPodId || '',
    networkId: '',
  })
  const [ports, setPorts] = useState<PortMappingItem[]>([])
  const [errors, setErrors] = useState<FormErrors>({})
  const [inspecting, setInspecting] = useState(false)
  const [exposedPorts, setExposedPorts] = useState<string[]>([])

  const { data: pods } = usePods()
  const { data: networks } = useNetworks()
  const createMutation = useCreateApplication()

  // Debounced image inspection
  useEffect(() => {
    const timer = setTimeout(async () => {
      if (!formData.image.trim()) {
        setExposedPorts([])
        return
      }

      setInspecting(true)
      try {
        const info = await inspectImage(formData.image.trim())
        setExposedPorts(info.exposed_ports || [])
      } catch (error) {
        // Ignore errors during typing (image might not exist yet)
        setExposedPorts([])
      } finally {
        setInspecting(false)
      }
    }, 1000)

    return () => clearTimeout(timer)
  }, [formData.image])

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required'
    } else if (!/^[a-z0-9-]+$/.test(formData.name)) {
      newErrors.name = 'Name must contain only lowercase letters, numbers, and hyphens'
    }

    if (!formData.image.trim()) {
      newErrors.image = 'Image is required'
    }

    const replicas = parseInt(formData.replicas, 10)
    if (isNaN(replicas) || replicas < 1 || replicas > 10) {
      newErrors.replicas = 'Replicas must be between 1 and 10'
    }

    // Validate ports
    ports.forEach(item => {
      if (!item.hostPort && !item.containerPort) return

      if (item.hostPort && (isNaN(parseInt(item.hostPort)) || parseInt(item.hostPort) < 1 || parseInt(item.hostPort) > 65535)) {
        newErrors.ports = 'Host ports must be valid numbers (1-65535)'
      }
      if (item.containerPort && (isNaN(parseInt(item.containerPort)) || parseInt(item.containerPort) < 1 || parseInt(item.containerPort) > 65535)) {
        newErrors.ports = 'Container ports must be valid numbers (1-65535)'
      }
    })

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()

    if (!validateForm()) return

    // Clean up ports (remove empty)
    const validPorts: Record<string, string> = {}
    ports.forEach(item => {
      if (item.hostPort && item.containerPort) {
        validPorts[item.hostPort] = item.containerPort
      }
    })

    try {
      await createMutation.mutateAsync({
        name: formData.name.trim(),
        image: formData.image.trim(),
        replicas: parseInt(formData.replicas, 10),
        pod_id: formData.podId || undefined,
        network_id: formData.networkId || undefined,
        ports: validPorts,
      })

      handleClose()
    } catch (error) {
      if (error instanceof ApiClientError) {
        if (error.field) {
          setErrors({ [error.field]: error.message })
        } else {
          setErrors({ general: error.message })
        }
      } else {
        setErrors({ general: 'An unexpected error occurred' })
      }
    }
  }

  const handleClose = () => {
    setFormData({ name: '', image: '', replicas: '1', podId: initialPodId || '', networkId: '' })
    setPorts([])
    setExposedPorts([])
    setErrors({})
    onClose()
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Create Application"
      description="Deploy a new containerized application"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {errors.general && (
          <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            {errors.general}
          </div>
        )}

        <Input
          label="Name"
          name="name"
          placeholder="my-app"
          value={formData.name}
          onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
          error={errors.name}
          hint="Lowercase letters, numbers, and hyphens only"
        />

        <div className="space-y-1">
          <div className="flex items-center justify-between">
            <Input
              label="Image"
              name="image"
              placeholder="nginx:latest"
              value={formData.image}
              onChange={(e) => setFormData((prev) => ({ ...prev, image: e.target.value }))}
              error={errors.image}
              className="flex-1"
            />
            {inspecting && (
              <div className="ml-2 mt-8 text-muted-foreground animate-pulse">
                <Loader2 className="h-4 w-4 animate-spin" />
              </div>
            )}
          </div>


        </div>

        {/* Pod Selection */}
        <div className="space-y-1">
          <label className="text-sm font-medium mb-1.5 block">Pod (Optional)</label>
          <select
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
            value={formData.podId}
            onChange={(e) => {
              const val = e.target.value
              setFormData((prev) => ({
                ...prev,
                podId: val,
                networkId: val ? '' : prev.networkId // Clear network if pod selected
              }))
            }}
            disabled={!!formData.networkId}
          >
            <option value="">None (Standalone)</option>
            {pods?.map((pod) => (
              <option key={pod.id} value={pod.id}>
                {pod.name}
              </option>
            ))}
          </select>
          {formData.networkId && (
            <p className="text-xs text-muted-foreground mt-1">
              Pod selection is disabled because a Network is selected.
            </p>
          )}
          {formData.podId && (
            <p className="text-xs text-muted-foreground mt-1">
              Application will run inside the selected pod's network namespace.
              External ports should be managed at the Pod level.
            </p>
          )}
        </div>

        {/* Network Selection */}
        <div className="space-y-1">
          <label className="text-sm font-medium mb-1.5 block">Network (Optional)</label>
          <select
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
            value={formData.networkId}
            onChange={(e) => {
              const val = e.target.value
              setFormData((prev) => ({
                ...prev,
                networkId: val,
                podId: val ? '' : prev.podId // Clear pod if network selected
              }))
            }}
            disabled={!!formData.podId}
          >
            <option value="">None (Default Bridge)</option>
            {networks?.map((net) => (
              <option key={net.id} value={net.id}>
                {net.name} ({net.driver})
              </option>
            ))}
          </select>
          {formData.podId && (
            <p className="text-xs text-muted-foreground mt-1">
              Network selection is disabled because a Pod is selected.
            </p>
          )}
        </div>

        <div>
          <label className="text-sm font-medium mb-1.5 block">Port Mappings</label>
          <PortMappingInput
            value={ports}
            onChange={setPorts}
            exposedPorts={exposedPorts}
          />
          {errors.ports && (
            <p className="text-xs text-destructive mt-1">{errors.ports}</p>
          )}
        </div>

        <Input
          label="Replicas"
          name="replicas"
          type="number"
          min="1"
          max="10"
          value={formData.replicas}
          onChange={(e) => setFormData((prev) => ({ ...prev, replicas: e.target.value }))}
          error={errors.replicas}
        />

        <div className="flex justify-end gap-3 pt-4">
          <Button type="button" variant="secondary" onClick={handleClose}>
            Cancel
          </Button>
          <Button type="submit" isLoading={createMutation.isPending}>
            Create Application
          </Button>
        </div>
      </form>
    </Modal >
  )
}
