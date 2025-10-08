"use client";

import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  AlertTriangle,
  RefreshCw,
  Server,
  AlertCircleIcon,
  Clock,
} from "lucide-react";
import { getVersionStatus, getEnvironmentRisk } from "@/lib/version-utils";
import { AdminConfigPanel } from "@/components/admin-config-panel";
import { CellConfigDialog } from "@/components/instance-config-dialog";

interface Application {
  id: string;
  name: string;
  description: string;
  versionType: "semantic" | "commit" | "timestamp";
  order: number;
}

interface Environment {
  id: string;
  name: string;
  description: string;
  category: "production" | "test";
  order: number;
}

interface VersionCell {
  instance: {
    id: string;
    environment_id: string;
    application_id: string;
  };
  deployment?: {
    version: string;
    deployed_at: string;
  };
}

interface LatestVersion {
  application_id: string;
  version: string;
}

interface AppVersionInfo {
  id: string;
  name: string;
  description: string;
  versionType: string;
  versions: Record<string, { version: string; deployedAt?: string } | null>;
  isOutdated: boolean;
  latestVersion: string;
}

export default function VersionDashboard() {
  const [applications, setApplications] = useState<Application[]>([]);
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [versionCells, setVersionCells] = useState<VersionCell[]>([]);
  const [latestVersions, setLatestVersions] = useState<LatestVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [spinning, setSpinning] = useState(false);

  const [cellConfigDialog, setCellConfigDialog] = useState<{
    open: boolean;
    instanceId?: string;
    environmentId: string;
    applicationId: string;
    environmentName: string;
    applicationName: string;
  }>({
    open: false,
    instanceId: undefined,
    environmentId: "",
    applicationId: "",
    environmentName: "",
    applicationName: "",
  });

  const fetchAllData = async () => {
    try {
      setLoading(true);

      // Fetch all data in parallel
      const [appsResponse, envsResponse, versionsResponse] = await Promise.all([
        fetch("/api/applications"),
        fetch("/api/environments"),
        fetch("/api/versions"),
      ]);

      const [appsData, envsData, versionsData] = await Promise.all([
        appsResponse.json(),
        envsResponse.json(),
        versionsResponse.json(),
      ]);

      setApplications(appsData.applications || []);
      setEnvironments(envsData.environments || []);
      setVersionCells(versionsData.cells || []);
      setLatestVersions(versionsData.latest || []);
      setLastUpdated(new Date());
    } catch (error) {
      console.error("Failed to fetch data:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAllData();
    const interval = setInterval(fetchAllData, 30000);
    return () => clearInterval(interval);
  }, []);

  if (applications.length === 0 || environments.length === 0) {
    if (loading) {
      return (
        <div className="min-h-screen bg-background flex items-center justify-center">
          <div className="flex items-center gap-2">
            <RefreshCw className="h-4 w-4 animate-spin" />
            <span>Loading version data...</span>
          </div>
        </div>
      );
    }

    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <AlertTriangle className="h-8 w-8 text-destructive mx-auto mb-2" />
          <p>Failed to load configuration data</p>
          <Button onClick={fetchAllData} className="mt-4">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const openCellConfig = (envId: string, appId: string) => {
    const environment = environments.find((env) => env.id === envId);
    const application = applications.find((app) => app.id === appId);

    if (!environment || !application) return;

    const instance = versionCells.find(
      (cell) =>
        cell.instance.environment_id === envId &&
        cell.instance.application_id === appId
    )?.instance;

    setCellConfigDialog({
      open: true,
      instanceId: instance?.id,
      environmentId: environment.id,
      applicationId: application.id,
      environmentName: environment.name,
      applicationName: application.name,
    });
  };

  const appVersions: AppVersionInfo[] = applications.map((app) => {
    const versions: Record<
      string,
      { version: string; deployedAt?: string } | null
    > = {};
    let isOutdated = false;

    // Initialize all environments as null (empty cells)
    environments.forEach((env) => {
      versions[env.id] = null;
    });

    // Fill in actual version data where it exists
    versionCells
      .filter((cell) => cell.instance.application_id === app.id)
      .forEach((cell) => {
        if (cell.deployment) {
          versions[cell.instance.environment_id] = {
            version: cell.deployment.version,
            deployedAt: cell.deployment.deployed_at,
          };
        } else {
          versions[cell.instance.environment_id] = {
            version: "",
          };
        }
      });

    // Find latest version for this app
    const latestVersion =
      latestVersions.find((lv) => lv.application_id === app.id)?.version ||
      "unknown";

    // Check if any environment is outdated
    Object.values(versions).forEach((versionData) => {
      if (versionData) {
        const status = getVersionStatus(versionData.version, latestVersion);
        if (status.status !== "current") {
          isOutdated = true;
        }
      }
    });

    return {
      id: app.id,
      name: app.name,
      description: app.description,
      versionType: app.versionType,
      versions,
      isOutdated,
      latestVersion,
    };
  });

  const criticalApps = appVersions
    .map((app) => {
      const criticalEnvs = environments
        .filter((env) => {
          const versionData = app.versions[env.id];
          if (!versionData) return false;
          const status = getVersionStatus(
            versionData.version,
            app.latestVersion
          );
          return status.severity === "critical";
        })
        .map((env) => env.id);
      return criticalEnvs.length > 0
        ? { name: app.name, environments: criticalEnvs }
        : null;
    })
    .filter(Boolean) as Array<{ name: string; environments: string[] }>;

  const sortedApps = appVersions.sort((a, b) => {
    // First, try to sort by the original application order
    const appA = applications.find((app) => app.id === a.id);
    const appB = applications.find((app) => app.id === b.id);

    if (appA?.order !== undefined && appB?.order !== undefined) {
      return appA.order - appB.order;
    }

    // Fallback to risk-based sorting if no order is defined
    const aVersions = Object.fromEntries(
      Object.entries(a.versions)
        .filter(([, v]) => v !== null)
        .map(([k, v]) => [k, v!.version])
    );
    const bVersions = Object.fromEntries(
      Object.entries(b.versions)
        .filter(([, v]) => v !== null)
        .map(([k, v]) => [k, v!.version])
    );

    const aRisk = getEnvironmentRisk(aVersions, a.latestVersion);
    const bRisk = getEnvironmentRisk(bVersions, b.latestVersion);

    const severityOrder = { critical: 0, warning: 1, good: 2 };
    return severityOrder[aRisk.level] - severityOrder[bRisk.level];
  });

  // Sort environments by order
  const sortedEnvironments = [...environments].sort(
    (a, b) => a.order - b.order
  );

  return (
    <div className="min-h-screen bg-background p-4">
      <header className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              Version Monitor
            </h1>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span>
                {applications.length} apps • {environments.length} environments
              </span>
              {criticalApps.length > 0 && (
                <span className="text-destructive font-medium">
                  <AlertCircleIcon className="h-3 w-3 inline mr-1" />
                  {criticalApps.length} critical
                </span>
              )}
              {lastUpdated && (
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  Updated {lastUpdated.toLocaleTimeString()}
                </span>
              )}
            </div>
          </div>
          <div className="flex gap-2">
            <AdminConfigPanel />
            <Button
              onClick={() => {
                fetchAllData();
                setSpinning(true);
                setTimeout(() => setSpinning(false), 500);
              }}
              variant="outline"
              size="sm"
            >
              <RefreshCw
                className={`h-4 w-4 ${spinning ? "animate-spinOnce" : ""}`}
              />
            </Button>
          </div>
        </div>
      </header>

      <div className="space-y-2">
        <div className="grid grid-cols-[200px_1fr] gap-4 mb-2">
          <div className="font-medium text-sm text-muted-foreground">
            Application
          </div>
          <div
            className="grid gap-2"
            style={{
              gridTemplateColumns: `repeat(${sortedEnvironments.length}, 1fr)`,
            }}
          >
            {sortedEnvironments.map((env) => (
              <div
                key={env.id}
                className="text-center text-sm font-medium text-muted-foreground"
              >
                {env.name}
              </div>
            ))}
          </div>
        </div>

        {sortedApps.map((app) => {
          const populatedVersions = Object.fromEntries(
            Object.entries(app.versions)
              .filter(([, v]) => v !== null)
              .map(([k, v]) => [k, v!.version])
          );

          const environmentRisk = getEnvironmentRisk(
            populatedVersions,
            app.latestVersion
          );

          return (
            <Card
              key={app.id}
              className={`transition-all duration-200 ${
                environmentRisk.level === "critical"
                  ? "border-destructive bg-destructive/5"
                  : environmentRisk.level === "warning"
                  ? "border-yellow-500 bg-yellow-50/50"
                  : ""
              }`}
            >
              <CardContent className="p-3">
                <div className="grid grid-cols-[200px_1fr] gap-4 items-center">
                  <div className="min-w-0">
                    <div className="font-medium text-sm truncate">
                      {app.name}
                    </div>
                    <div className="text-xs text-muted-foreground font-mono">
                      Latest: {app.latestVersion}
                    </div>
                    {/* {environmentRisk.level === "critical" && (
                      <Badge variant="destructive" className="text-xs mt-1">
                        URGENT
                      </Badge>
                    )} */}
                  </div>

                  <div
                    className="grid gap-2"
                    style={{
                      gridTemplateColumns: `repeat(${sortedEnvironments.length}, 1fr)`,
                    }}
                  >
                    {sortedEnvironments.map((env) => {
                      const versionData = app.versions[env.id];

                      if (!versionData) {
                        return (
                          <NotTrackedCard
                            key={env.id}
                            envId={env.id}
                            appId={app.id}
                            openConfig={openCellConfig}
                          />
                        );
                      }

                      const versionStatus = getVersionStatus(
                        versionData.version,
                        app.latestVersion
                      );

                      if (versionStatus.status === "not deployed") {
                        return <NotDeployedCard key={env.id} />;
                      }

                      return (
                        <InstanceCard
                          key={env.id}
                          version={versionData.version}
                          deployedAt={versionData.deployedAt}
                          severity={versionStatus.severity}
                        />
                      );
                    })}
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {sortedApps.length === 0 && (
        <div className="text-center py-12">
          <Server className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-medium mb-2">No applications found</h3>
          <p className="text-muted-foreground">No applications to monitor.</p>
        </div>
      )}

      <CellConfigDialog
        open={cellConfigDialog.open}
        onOpenChange={(open, reload) => {
          reload ? fetchAllData() : null;
          setCellConfigDialog({ ...cellConfigDialog, open });
        }}
        instanceId={cellConfigDialog.instanceId}
        environmentId={cellConfigDialog.environmentId}
        applicationId={cellConfigDialog.applicationId}
        environmentName={cellConfigDialog.environmentName}
        applicationName={cellConfigDialog.applicationName}
      />
    </div>
  );
}

const getSeverityColor = (severity: string) => {
  switch (severity) {
    case "critical":
      return "bg-destructive text-destructive-foreground";
    case "warning":
      return "bg-yellow-500 text-yellow-950";
    case "info":
      return "bg-blue-500 text-blue-950";
    default:
      return "bg-green-500 text-green-950";
  }
};

function NotDeployedCard() {
  return (
    <div
      className="p-2 rounded text-center text-xs flex items-center justify-center transition-all duration-200"
      style={{
        background:
          "repeating-linear-gradient(135deg, #fff, #fff 10px, #ff0000 10px, #ff0000 30px)",
      }}
    >
      <div className="font-mono font-bold text-center bg-white p-1 rounded">
        Not Deployed
      </div>
    </div>
  );
}

function InstanceCard({
  version,
  deployedAt,
  severity,
}: {
  version: string;
  deployedAt?: string;
  severity: string;
}) {
  return (
    <div
      className={`p-2 rounded text-center text-xs transition-all duration-200 ${getSeverityColor(
        severity
      )}`}
    >
      <div className="font-mono font-medium truncate">{version}</div>
      {deployedAt && (
        <div className="text-xs opacity-75 mt-1">
          {new Date(deployedAt).toLocaleDateString()}
        </div>
      )}
    </div>
  );
}

const NotTrackedCard = ({
  envId,
  appId,
  openConfig,
}: {
  envId: string;
  appId: string;
  openConfig: (envId: string, appId: string) => void;
}) => {
  return (
    <div
      className="p-2 rounded text-center text-xs border border-dashed border-muted-foreground/20 cursor-pointer relative"
      onClick={() => openConfig(envId, appId)}
    >
      <div className="text-muted-foreground">—</div>
      <div className="text-xs opacity-50 mt-1">Not Tracked</div>
      <div className="absolute rounded inset-0 bg-black/5 opacity-0 hover:opacity-100 transition-opacity" />
    </div>
  );
};
