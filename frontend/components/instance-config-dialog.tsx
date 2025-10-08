"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Save, Trash2 } from "lucide-react";

interface InstanceData {
  id: string;
  name: string;
}

interface CellConfigDialogProps {
  open: boolean;
  onOpenChange: (open: boolean, reload: boolean) => void;
  instanceId?: string;
  environmentId: string;
  applicationId: string;
  environmentName: string;
  applicationName: string;
}

export function CellConfigDialog({
  open,
  onOpenChange,
  instanceId,
  environmentId,
  applicationId,
  environmentName,
  applicationName,
}: CellConfigDialogProps) {
  const [instanceData, setInstanceData] = useState<InstanceData>({
    id: "",
    name: "",
  });
  const [instanceExists, setInstanceExists] = useState(false);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (open) {
      fetchInstance();
    }
  }, [open, instanceId]);

  const fetchInstance = async () => {
    if (!instanceId) {
      setInstanceData({ id: "", name: "" });
      setInstanceExists(false);
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`/api/instances?id=${instanceId}`);
      if (response.ok) {
        const data = await response.json();
        setInstanceData({
          id: data.id,
          name: data.name,
        });
        setInstanceExists(true);
      } else {
        throw new Error(`Unexpected response status: ${response.status}`);
      }
    } catch (error) {
      console.error("Failed to fetch cell configuration:", error);
    } finally {
      setLoading(false);
    }
  };

  const createInstance = async () => {
    setSaving(true);
    try {
      const response = await fetch(`/api/instances`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: instanceData.name,
          environment_id: environmentId,
          application_id: applicationId,
        }),
      });

      if (response.ok) {
        onOpenChange(false, true);
      }
    } catch (error) {
      console.error("Failed to create instance:", error);
    } finally {
      setSaving(false);
    }
  };

  const updateInstance = async () => {
    setSaving(true);
    try {
      const response = await fetch(`/api/instances`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: instanceData.id,
          name: instanceData.name,
        }),
      });

      if (response.ok) {
        onOpenChange(false, true);
      }
    } catch (error) {
      console.error("Failed to update instance:", error);
    } finally {
      setSaving(false);
    }
  };

  const deleteInstance = async () => {
    setSaving(true);
    try {
      const response = await fetch(`/api/instances`, {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: instanceData.id,
        }),
      });

      if (response.ok) {
        onOpenChange(false, true);
      }
    } catch (error) {
      console.error("Failed to delete instance:", error);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(onOpen) => onOpenChange(onOpen, false)}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {instanceExists ? "Edit" : "Add"} {applicationName} in{" "}
            {environmentName}
          </DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder={"Instance Name"}
                value={instanceData.name || ""}
                onChange={(e) => {
                  setInstanceData({ ...instanceData, name: e.target.value });
                }}
              />
              <p className="text-sm text-muted-foreground">
                The name is used to identify this instance in Overseer.
              </p>
            </div>

            <div className="flex justify-between pt-4">
              {instanceExists ? (
                <Button
                  variant="outline"
                  onClick={deleteInstance}
                  disabled={saving}
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete Instance Tracking
                </Button>
              ) : (
                <div />
              )}
              <Button
                onClick={instanceExists ? updateInstance : createInstance}
                disabled={saving}
              >
                {saving ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <Save className="h-4 w-4 mr-2" />
                )}
                {instanceExists ? "Save Changes" : "Add Instance Tracking"}
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
