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
		// Process the event (e.g., update version info, trigger actions, etc.)
	}
	return nil
}
