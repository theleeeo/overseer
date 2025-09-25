"use client";

import type React from "react";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Trash2, Edit, Settings, GripVertical } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface Environment {
  id: string;
  name: string;
  type: "production" | "test";
  order: number;
}

interface Application {
  id: string;
  name: string;
  versionType: "semver" | "commit" | "timestamp";
  repository: string;
  order?: number;
}

export function AdminConfigPanel() {
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [applications, setApplications] = useState<Application[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [editingEnv, setEditingEnv] = useState<Environment | null>(null);
  const [editingApp, setEditingApp] = useState<Application | null>(null);
  const [activeTab, setActiveTab] = useState<"environments" | "applications">(
    "environments"
  );
  const [draggedItem, setDraggedItem] = useState<string | null>(null);
  const [deleteConfirmation, setDeleteConfirmation] = useState<{
    type: "environment" | "application" | null;
    id: string | null;
    name: string | null;
  }>({ type: null, id: null, name: null });
  const { toast } = useToast();

  const [envForm, setEnvForm] = useState({
    name: "",
    type: "production" as "production" | "test",
  });
  const [appForm, setAppForm] = useState({
    name: "",
    versionType: "semver" as "semver" | "commit" | "timestamp",
    repository: "",
  });

  useEffect(() => {
    fetchEnvironments();
    fetchApplications();
  }, []);

  const fetchEnvironments = async () => {
    try {
      const response = await fetch("/api/environments");
      const data = await response.json();
      setEnvironments(
        data.environments.sort(
          (a: Environment, b: Environment) => a.order - b.order
        )
      );
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to fetch environments",
        variant: "destructive",
      });
    }
  };

  const fetchApplications = async () => {
    try {
      const response = await fetch("/api/applications");
      const data = await response.json();
      setApplications(
        data.applications.sort(
          (a: Application, b: Application) => (a.order || 0) - (b.order || 0)
        )
      );
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to fetch applications",
        variant: "destructive",
      });
    }
  };

  const handleEnvDragStart = (e: React.DragEvent, envId: string) => {
    setDraggedItem(envId);
    e.dataTransfer.effectAllowed = "move";
  };

  const handleEnvDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
  };

  const handleEnvDrop = async (e: React.DragEvent, targetEnvId: string) => {
    e.preventDefault();
    if (!draggedItem || draggedItem === targetEnvId) return;

    const draggedIndex = environments.findIndex(
      (env) => env.id === draggedItem
    );
    const targetIndex = environments.findIndex((env) => env.id === targetEnvId);

    if (draggedIndex === -1 || targetIndex === -1) return;

    const newEnvironments = [...environments];
    const [draggedEnv] = newEnvironments.splice(draggedIndex, 1);
    newEnvironments.splice(targetIndex, 0, draggedEnv);

    const updatedEnvironments = newEnvironments.map((env, index) => ({
      ...env,
      order: index + 1,
    }));

    setEnvironments(updatedEnvironments);
    setDraggedItem(null);

    try {
      await fetch("/api/environments/reorder", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify([...updatedEnvironments.map((env) => env.id)]),
      });
      toast({ title: "Success", description: "Environment order updated" });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update environment order",
        variant: "destructive",
      });
      fetchEnvironments();
    }
  };

  const handleAppDragStart = (e: React.DragEvent, appId: string) => {
    setDraggedItem(appId);
    e.dataTransfer.effectAllowed = "move";
  };

  const handleAppDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
  };

  const handleAppDrop = async (e: React.DragEvent, targetAppId: string) => {
    e.preventDefault();
    if (!draggedItem || draggedItem === targetAppId) return;

    const draggedIndex = applications.findIndex(
      (app) => app.id === draggedItem
    );
    const targetIndex = applications.findIndex((app) => app.id === targetAppId);

    if (draggedIndex === -1 || targetIndex === -1) return;

    const newApplications = [...applications];
    const [draggedApp] = newApplications.splice(draggedIndex, 1);
    newApplications.splice(targetIndex, 0, draggedApp);

    const updatedApplications = newApplications.map((app, index) => ({
      ...app,
      order: index + 1,
    }));

    setApplications(updatedApplications);
    setDraggedItem(null);

    try {
      await fetch("/api/applications/reorder", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify([...updatedApplications.map((app) => app.id)]),
      });
      toast({ title: "Success", description: "Application order updated" });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update application order",
        variant: "destructive",
      });
      fetchApplications();
    }
  };

  const handleSaveEnvironment = async () => {
    try {
      const method = editingEnv ? "PUT" : "POST";
      const body = editingEnv ? { ...envForm, id: editingEnv.id } : envForm;

      const response = await fetch("/api/config/environments", {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (response.ok) {
        toast({
          title: "Success",
          description: `Environment ${
            editingEnv ? "updated" : "created"
          } successfully`,
        });
        fetchEnvironments();
        resetEnvForm();
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save environment",
        variant: "destructive",
      });
    }
  };

  const handleSaveApplication = async () => {
    try {
      const method = editingApp ? "PUT" : "POST";
      const body = editingApp ? { ...appForm, id: editingApp.id } : appForm;

      const response = await fetch("/api/config/applications", {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (response.ok) {
        toast({
          title: "Success",
          description: `Application ${
            editingApp ? "updated" : "created"
          } successfully`,
        });
        fetchApplications();
        resetAppForm();
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save application",
        variant: "destructive",
      });
    }
  };

  const handleDeleteEnvironment = async (id: string) => {
    try {
      const response = await fetch(`/api/config/environments?id=${id}`, {
        method: "DELETE",
      });
      if (response.ok) {
        toast({
          title: "Success",
          description: "Environment deleted successfully",
        });
        fetchEnvironments();
        setDeleteConfirmation({ type: null, id: null, name: null });
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete environment",
        variant: "destructive",
      });
    }
  };

  const handleDeleteApplication = async (id: string) => {
    try {
      const response = await fetch(`/api/config/applications?id=${id}`, {
        method: "DELETE",
      });
      if (response.ok) {
        toast({
          title: "Success",
          description: "Application deleted successfully",
        });
        fetchApplications();
        setDeleteConfirmation({ type: null, id: null, name: null });
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete application",
        variant: "destructive",
      });
    }
  };

  const resetEnvForm = () => {
    setEnvForm({ name: "", type: "production" });
    setEditingEnv(null);
  };

  const resetAppForm = () => {
    setAppForm({ name: "", versionType: "semver", repository: "" });
    setEditingApp(null);
  };

  const startEditEnvironment = (env: Environment) => {
    setEditingEnv(env);
    setEnvForm({ name: env.name, type: env.type });
    setActiveTab("environments");
  };

  const startEditApplication = (app: Application) => {
    setEditingApp(app);
    setAppForm({
      name: app.name,
      versionType: app.versionType,
      repository: app.repository,
    });
    setActiveTab("applications");
  };

  const showDeleteEnvironmentConfirmation = (env: Environment) => {
    setDeleteConfirmation({ type: "environment", id: env.id, name: env.name });
  };

  const showDeleteApplicationConfirmation = (app: Application) => {
    setDeleteConfirmation({ type: "application", id: app.id, name: app.name });
  };

  const confirmDelete = () => {
    if (deleteConfirmation.type === "environment" && deleteConfirmation.id) {
      handleDeleteEnvironment(deleteConfirmation.id);
    } else if (
      deleteConfirmation.type === "application" &&
      deleteConfirmation.id
    ) {
      handleDeleteApplication(deleteConfirmation.id);
    }
  };

  const cancelDelete = () => {
    setDeleteConfirmation({ type: null, id: null, name: null });
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger asChild>
          <Button variant="outline" size="sm">
            <Settings className="h-4 w-4 mr-2" />
            Configure
          </Button>
        </DialogTrigger>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>System Configuration</DialogTitle>
          </DialogHeader>

          <div className="space-y-6">
            <div className="flex space-x-1 bg-muted p-1 rounded-lg">
              <Button
                variant={activeTab === "environments" ? "default" : "ghost"}
                size="sm"
                onClick={() => setActiveTab("environments")}
                className="flex-1"
              >
                Environments ({environments.length})
              </Button>
              <Button
                variant={activeTab === "applications" ? "default" : "ghost"}
                size="sm"
                onClick={() => setActiveTab("applications")}
                className="flex-1"
              >
                Applications ({applications.length})
              </Button>
            </div>

            {activeTab === "environments" && (
              <div className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      {editingEnv ? "Edit Environment" : "Add New Environment"}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="env-name">Environment Name</Label>
                        <Input
                          id="env-name"
                          value={envForm.name}
                          onChange={(e) =>
                            setEnvForm({ ...envForm, name: e.target.value })
                          }
                          placeholder="e.g., Production, Staging"
                        />
                      </div>
                      <div>
                        <Label htmlFor="env-type">Environment Type</Label>
                        <Select
                          value={envForm.type}
                          onValueChange={(value: "production" | "test") =>
                            setEnvForm({ ...envForm, type: value })
                          }
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="production">
                              Production
                            </SelectItem>
                            <SelectItem value="test">Test</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        onClick={handleSaveEnvironment}
                        disabled={!envForm.name}
                      >
                        {editingEnv ? "Update" : "Add"} Environment
                      </Button>
                      {editingEnv && (
                        <Button variant="outline" onClick={resetEnvForm}>
                          Cancel
                        </Button>
                      )}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      Existing Environments
                    </CardTitle>
                    <p className="text-sm text-muted-foreground">
                      Drag and drop to reorder environments
                    </p>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {environments.map((env) => (
                        <div
                          key={env.id}
                          className={`flex items-center justify-between p-3 border rounded-lg cursor-move transition-colors ${
                            draggedItem === env.id
                              ? "opacity-50"
                              : "hover:bg-muted/50"
                          }`}
                          draggable
                          onDragStart={(e) => handleEnvDragStart(e, env.id)}
                          onDragOver={handleEnvDragOver}
                          onDrop={(e) => handleEnvDrop(e, env.id)}
                        >
                          <div className="flex items-center gap-3">
                            <GripVertical className="h-4 w-4 text-muted-foreground" />
                            <div>
                              <div className="font-medium">{env.name}</div>
                              <div className="text-sm text-muted-foreground">
                                ID: {env.id}
                              </div>
                            </div>
                            <Badge
                              variant={
                                env.type === "production"
                                  ? "default"
                                  : "secondary"
                              }
                            >
                              {env.type}
                            </Badge>
                          </div>
                          <div className="flex gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => startEditEnvironment(env)}
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                showDeleteEnvironmentConfirmation(env)
                              }
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}

            {activeTab === "applications" && (
              <div className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      {editingApp ? "Edit Application" : "Add New Application"}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-1 gap-4">
                      <div>
                        <Label htmlFor="app-name">Application Name</Label>
                        <Input
                          id="app-name"
                          value={appForm.name}
                          onChange={(e) =>
                            setAppForm({ ...appForm, name: e.target.value })
                          }
                          placeholder="e.g., User Service, Payment Gateway"
                        />
                      </div>
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <Label htmlFor="app-version-type">Version Type</Label>
                          <Select
                            value={appForm.versionType}
                            onValueChange={(
                              value: "semver" | "commit" | "timestamp"
                            ) => setAppForm({ ...appForm, versionType: value })}
                          >
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="semver">
                                Semantic Version (v1.2.3)
                              </SelectItem>
                              <SelectItem value="commit">
                                Commit Hash (abc123f)
                              </SelectItem>
                              <SelectItem value="timestamp">
                                Unix Timestamp
                              </SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                        <div>
                          <Label htmlFor="app-repository">
                            Repository (Optional)
                          </Label>
                          <Input
                            id="app-repository"
                            value={appForm.repository}
                            onChange={(e) =>
                              setAppForm({
                                ...appForm,
                                repository: e.target.value,
                              })
                            }
                            placeholder="github.com/org/repo"
                          />
                        </div>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        onClick={handleSaveApplication}
                        disabled={!appForm.name}
                      >
                        {editingApp ? "Update" : "Add"} Application
                      </Button>
                      {editingApp && (
                        <Button variant="outline" onClick={resetAppForm}>
                          Cancel
                        </Button>
                      )}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      Existing Applications
                    </CardTitle>
                    <p className="text-sm text-muted-foreground">
                      Drag and drop to reorder applications
                    </p>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {applications.map((app) => (
                        <div
                          key={app.id}
                          className={`flex items-center justify-between p-3 border rounded-lg cursor-move transition-colors ${
                            draggedItem === app.id
                              ? "opacity-50"
                              : "hover:bg-muted/50"
                          }`}
                          draggable
                          onDragStart={(e) => handleAppDragStart(e, app.id)}
                          onDragOver={handleAppDragOver}
                          onDrop={(e) => handleAppDrop(e, app.id)}
                        >
                          <div className="flex items-center gap-3">
                            <GripVertical className="h-4 w-4 text-muted-foreground" />
                            <div>
                              <div className="font-medium">{app.name}</div>
                              <div className="text-sm text-muted-foreground">
                                ID: {app.id} • Type: {app.versionType}
                              </div>
                              {app.repository && (
                                <div className="text-xs text-muted-foreground">
                                  {app.repository}
                                </div>
                              )}
                            </div>
                            <Badge variant="outline">{app.versionType}</Badge>
                          </div>
                          <div className="flex gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => startEditApplication(app)}
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                showDeleteApplicationConfirmation(app)
                              }
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={deleteConfirmation.type !== null}
        onOpenChange={(open) => !open && cancelDelete()}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm Deletion</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the {deleteConfirmation.type} "
              {deleteConfirmation.name}"? This action cannot be undone and may
              affect your version tracking.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={cancelDelete}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
