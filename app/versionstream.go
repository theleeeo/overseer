package app

import (
	"context"
	"fmt"
	"log/slog"
	"overseer/datasource"
)

type EventSource interface {
	StreamEvents(ctx context.Context) (<-chan datasource.Event, error)
}

func (a *App) RunVersionStream(ctx context.Context, source EventSource) error {
	s, err := source.StreamEvents(ctx)
	if err != nil {
		return fmt.Errorf("starting the event stream: %w", err)
	}

	for event := range s {
		slog.Info("Received event", "id", event.Id, "name", event.DeploymentName, "version", event.Version, "deployedAt", event.DeployedAt)

		instanceResp, err := a.ListInstances(ctx, ListInstancesParameters{
			Name: event.DeploymentName,
		})
		if err != nil {
			slog.Error("listing instances", "error", err)
			continue
		}
		if len(instanceResp) == 0 {
			slog.Warn("no instance found for deployment", "deployment", event.DeploymentName)
			continue
		}

		if len(instanceResp) > 1 {
			panic("multiple instances found for deployment, this should never happen")
		}

		if err = a.RegisterDeployment(ctx, RegisterDeploymentParams{
			InstanceId: instanceResp[0].Id,
			Version:    event.Version,
			DeployedAt: event.DeployedAt,
		}); err != nil {
			slog.Error("registering deployment", "error", err)
		}
	}
	return nil
}
