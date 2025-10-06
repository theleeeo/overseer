package datasource

import (
	"context"
	"time"
)

type MockSource struct{}

func (s *MockSource) StreamEvents(ctx context.Context) (<-chan Event, error) {
	channel := make(chan Event)
	go func() {
		<-ctx.Done()
		close(channel)
	}()

	go func() {
		var mockEvents = []Event{
			{Id: "1", DeploymentName: "env1-app-1", DeployedAt: time.Now(), Version: "1.0.0"},
		}

		for _, event := range mockEvents {
			select {
			case <-ctx.Done():
				return
			case channel <- event:
			}
		}
	}()

	// Mock implementation
	return channel, nil
}
