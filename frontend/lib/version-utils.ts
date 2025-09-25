export type VersionType = "semver" | "timestamp" | "commit"

export function detectVersionType(version: string): VersionType {
  // Check if it's a semantic version (e.g., v1.2.3, 1.2.3)
  if (/^v?\d+\.\d+\.\d+/.test(version)) {
    return "semver"
  }

  // Check if it's a timestamp (ISO format or Unix timestamp)
  if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/.test(version) || /^\d{10,13}$/.test(version)) {
    return "timestamp"
  }

  // Check if it's a commit hash (7-40 hex characters)
  if (/^[a-f0-9]{7,40}$/i.test(version)) {
    return "commit"
  }

  // Default to commit hash for other formats
  return "commit"
}

export function parseVersion(version: string): { major: number; minor: number; patch: number } {
  // Remove 'v' prefix if present
  const cleanVersion = version.replace(/^v/, "")
  const parts = cleanVersion.split(".").map(Number)

  return {
    major: parts[0] || 0,
    minor: parts[1] || 0,
    patch: parts[2] || 0,
  }
}

export function compareVersions(version1: string, version2: string): number {
  const type1 = detectVersionType(version1)
  const type2 = detectVersionType(version2)

  // If types don't match, we can't compare meaningfully
  if (type1 !== type2) {
    return 0 // Consider them equal if we can't compare
  }

  switch (type1) {
    case "semver":
      return compareSemver(version1, version2)
    case "timestamp":
      return compareTimestamps(version1, version2)
    case "commit":
      return compareCommits(version1, version2)
    default:
      return 0
  }
}

function compareSemver(version1: string, version2: string): number {
  const v1 = parseVersion(version1)
  const v2 = parseVersion(version2)

  if (v1.major !== v2.major) {
    return v1.major - v2.major
  }
  if (v1.minor !== v2.minor) {
    return v1.minor - v2.minor
  }
  return v1.patch - v2.patch
}

function compareTimestamps(version1: string, version2: string): number {
  let timestamp1: number
  let timestamp2: number

  // Handle ISO format
  if (version1.includes("T")) {
    timestamp1 = new Date(version1).getTime()
  } else {
    // Handle Unix timestamp
    timestamp1 = Number.parseInt(version1) * (version1.length === 10 ? 1000 : 1)
  }

  if (version2.includes("T")) {
    timestamp2 = new Date(version2).getTime()
  } else {
    timestamp2 = Number.parseInt(version2) * (version2.length === 10 ? 1000 : 1)
  }

  return timestamp1 - timestamp2
}

function compareCommits(version1: string, version2: string): number {
  // For commit hashes, we can't determine which is "newer" without git history
  // So we'll consider them equal unless they're exactly the same
  return version1 === version2 ? 0 : -1 // Always consider different commits as "behind"
}

export function getVersionStatus(currentVersion: string, latestVersion: string) {
  const currentType = detectVersionType(currentVersion)
  const latestType = detectVersionType(latestVersion)

  // If we can't compare (different types), show as unknown
  if (currentType !== latestType) {
    return {
      status: "unknown" as const,
      severity: "info" as const,
      message: "Different version format",
      color: "bg-secondary text-secondary-foreground",
    }
  }

  const comparison = compareVersions(currentVersion, latestVersion)

  if (comparison === 0) {
    return {
      status: "current" as const,
      severity: "none" as const,
      message: "Up to date",
      color: "bg-primary text-primary-foreground",
    }
  }

  if (comparison < 0) {
    switch (currentType) {
      case "semver":
        return getSemverStatus(currentVersion, latestVersion)
      case "timestamp":
        return getTimestampStatus(currentVersion, latestVersion)
      case "commit":
        return getCommitStatus(currentVersion, latestVersion)
      default:
        return {
          status: "outdated" as const,
          severity: "warning" as const,
          message: "Behind latest",
          color: "bg-accent text-accent-foreground",
        }
    }
  }

  // Ahead of latest
  return {
    status: "ahead" as const,
    severity: "warning" as const,
    message: "Ahead of latest",
    color: "bg-chart-2 text-foreground",
  }
}

function getSemverStatus(currentVersion: string, latestVersion: string) {
  const current = parseVersion(currentVersion)
  const latest = parseVersion(latestVersion)

  // Major version behind - critical
  if (current.major < latest.major) {
    return {
      status: "outdated" as const,
      severity: "critical" as const,
      message: "Major version behind",
      color: "bg-destructive text-destructive-foreground",
    }
  }

  // Minor version behind - warning
  if (current.minor < latest.minor) {
    return {
      status: "outdated" as const,
      severity: "warning" as const,
      message: "Minor version behind",
      color: "bg-accent text-accent-foreground",
    }
  }

  // Patch version behind - info
  return {
    status: "outdated" as const,
    severity: "info" as const,
    message: "Patch version behind",
    color: "bg-secondary text-secondary-foreground",
  }
}

function getTimestampStatus(currentVersion: string, latestVersion: string) {
  const currentTime = currentVersion.includes("T")
    ? new Date(currentVersion).getTime()
    : Number.parseInt(currentVersion) * 1000
  const latestTime = latestVersion.includes("T")
    ? new Date(latestVersion).getTime()
    : Number.parseInt(latestVersion) * 1000

  const diffHours = (latestTime - currentTime) / (1000 * 60 * 60)

  if (diffHours > 168) {
    // More than a week behind
    return {
      status: "outdated" as const,
      severity: "critical" as const,
      message: "Over a week behind",
      color: "bg-destructive text-destructive-foreground",
    }
  }

  if (diffHours > 24) {
    // More than a day behind
    return {
      status: "outdated" as const,
      severity: "warning" as const,
      message: "Over a day behind",
      color: "bg-accent text-accent-foreground",
    }
  }

  return {
    status: "outdated" as const,
    severity: "info" as const,
    message: "Behind latest",
    color: "bg-secondary text-secondary-foreground",
  }
}

function getCommitStatus(currentVersion: string, latestVersion: string) {
  return {
    status: "outdated" as const,
    severity: "warning" as const,
    message: "Different commit",
    color: "bg-accent text-accent-foreground",
  }
}

export function getEnvironmentRisk(environments: Record<string, string>, latestVersion: string) {
  const risks = Object.entries(environments).map(([env, version]) => {
    const status = getVersionStatus(version, latestVersion)
    return {
      environment: env,
      ...status,
    }
  })

  const criticalCount = risks.filter((r) => r.severity === "critical").length
  const warningCount = risks.filter((r) => r.severity === "warning").length

  if (criticalCount > 0) {
    return {
      level: "critical" as const,
      message: `${criticalCount} environment(s) with critical updates needed`,
      color: "text-destructive",
    }
  }

  if (warningCount > 0) {
    return {
      level: "warning" as const,
      message: `${warningCount} environment(s) with updates available`,
      color: "text-foreground",
    }
  }

  return {
    level: "good" as const,
    message: "All environments up to date",
    color: "text-primary",
  }
}
