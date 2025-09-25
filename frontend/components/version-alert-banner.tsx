"use client"

import { AlertTriangle, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useState } from "react"

interface VersionAlertBannerProps {
  criticalApps: Array<{
    name: string
    environments: string[]
  }>
}

export function VersionAlertBanner({ criticalApps }: VersionAlertBannerProps) {
  const [dismissed, setDismissed] = useState(false)

  if (dismissed || criticalApps.length === 0) {
    return null
  }

  return (
    <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 mb-6">
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3">
          <AlertTriangle className="h-5 w-5 text-destructive mt-0.5 flex-shrink-0" />
          <div>
            <h3 className="font-semibold text-destructive mb-1">Critical Updates Required</h3>
            <p className="text-sm text-destructive/80 mb-2">
              {criticalApps.length} application{criticalApps.length > 1 ? "s" : ""} have major version updates
              available:
            </p>
            <ul className="text-sm text-destructive/80 space-y-1">
              {criticalApps.map((app) => (
                <li key={app.name}>
                  <strong>{app.name}</strong> in {app.environments.join(", ")}
                </li>
              ))}
            </ul>
          </div>
        </div>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setDismissed(true)}
          className="text-destructive hover:text-destructive/80"
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
