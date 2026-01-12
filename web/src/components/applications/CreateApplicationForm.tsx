import { useState, type FormEvent } from 'react'
import { Button, Input, Modal } from '@/components/ui'
import { useCreateApplication } from '@/hooks/useApplications'
import { ApiClientError } from '@/lib/api'

interface CreateApplicationFormProps {
  isOpen: boolean
  onClose: () => void
}

interface FormData {
  name: string
  image: string
  replicas: string
}

interface FormErrors {
  name?: string
  image?: string
  replicas?: string
  general?: string
}

export function CreateApplicationForm({ isOpen, onClose }: CreateApplicationFormProps) {
  const [formData, setFormData] = useState<FormData>({
    name: '',
    image: '',
    replicas: '1',
  })
  const [errors, setErrors] = useState<FormErrors>({})

  const createMutation = useCreateApplication()

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

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()

    if (!validateForm()) return

    try {
      await createMutation.mutateAsync({
        name: formData.name.trim(),
        image: formData.image.trim(),
        replicas: parseInt(formData.replicas, 10),
      })

      // Reset form and close modal on success
      setFormData({ name: '', image: '', replicas: '1' })
      setErrors({})
      onClose()
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
    setFormData({ name: '', image: '', replicas: '1' })
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

        <Input
          label="Image"
          name="image"
          placeholder="nginx:latest"
          value={formData.image}
          onChange={(e) => setFormData((prev) => ({ ...prev, image: e.target.value }))}
          error={errors.image}
          hint="Docker image name with optional tag"
        />

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
    </Modal>
  )
}
