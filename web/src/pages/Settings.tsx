import { Settings as SettingsIcon } from 'lucide-react'

export function Settings() {
  return (
    <div className="space-y-6 animate-fade-in">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
        <p className="text-muted-foreground mt-1">
          Configure your Simplify instance
        </p>
      </div>

      {/* Placeholder content */}
      <div className="card p-12 text-center">
        <div className="inline-flex items-center justify-center h-16 w-16 rounded-full bg-muted mb-4">
          <SettingsIcon className="h-8 w-8 text-muted-foreground" />
        </div>
        <h2 className="text-lg font-medium">Coming Soon</h2>
        <p className="text-muted-foreground mt-2 max-w-md mx-auto">
          Settings and configuration options will be available here in a future update.
          This will include cluster settings, user preferences, and more.
        </p>
      </div>

      {/* Future sections preview */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <div className="card p-6 opacity-50">
          <h3 className="font-medium">General</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Instance name, timezone, and locale settings
          </p>
        </div>
        <div className="card p-6 opacity-50">
          <h3 className="font-medium">Security</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Authentication, API keys, and access control
          </p>
        </div>
        <div className="card p-6 opacity-50">
          <h3 className="font-medium">Notifications</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Email, Slack, and webhook integrations
          </p>
        </div>
        <div className="card p-6 opacity-50">
          <h3 className="font-medium">Registry</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Configure private container registries
          </p>
        </div>
        <div className="card p-6 opacity-50">
          <h3 className="font-medium">Backup</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Automated backups and disaster recovery
          </p>
        </div>
        <div className="card p-6 opacity-50">
          <h3 className="font-medium">Advanced</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Debug logs, feature flags, and experimental options
          </p>
        </div>
      </div>
    </div>
  )
}
