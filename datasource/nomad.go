package datasource

import (
	"context"
	"fmt"
)

type NomadSource struct {
	Address string
	Token   string
}

func NewNomadSource(address, token string) *NomadSource {
	return &NomadSource{
		Address: address,
		Token:   token,
	}
}

func (n *NomadSource) StreamEvents(ctx context.Context) (EventStream, error) {
	stream := make(chan Event, 10)
	go func() {
		// Simulate streaming events
		for i := range 5 {
			select {
			case stream <- Event{Id: fmt.Sprintf("%d", i)}:
			case <-ctx.Done():
				close(stream)
				return
			}
		}
		close(stream)
	}()

	return stream, nil
}
