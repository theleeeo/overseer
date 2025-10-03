package datasource

import "context"

type MockSource struct{}

func (s *MockSource) StreamEvents(ctx context.Context) (<-chan Event, error) {
	channel := make(chan Event)
	go func() {
		<-ctx.Done()
		close(channel)
	}()
	// Mock implementation
	return channel, nil
}
