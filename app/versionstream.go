package app

import (
	"log/slog"
	"overseer/datasource"
)

func (a *App) RunVersionStream(source datasource.EventStream) error {
	for event := range source {
		slog.Info("Received event", "id", event.Id)
		// Process the event (e.g., update version info, trigger actions, etc.)
	}
	return nil
}
