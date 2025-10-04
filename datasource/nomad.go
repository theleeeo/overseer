package datasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type NomadSource struct {
	address string
	token   string
	logger  *slog.Logger
}

type nomadEvent struct {
	Topic      string
	Type       string
	Key        string
	Namespace  string
	FilterKeys []string
	Index      uint64
	Payload    json.RawMessage // TODO: Type this
}

type nomadStreamItem struct {
	Index  int64
	Events []nomadEvent
}

// TODO: Should handle the initial sync of existing jobs.
func NewNomadSource(address, token string, logger *slog.Logger) *NomadSource {
	address = address + "/v1/event/stream?topic=Job"

	return &NomadSource{
		address: address,
		token:   token,
		logger:  logger,
	}
}

func (n *NomadSource) StreamEvents(ctx context.Context) (<-chan Event, error) {
	stream := make(chan Event, 10)

	go func() {
		defer close(stream)

		for {
			err := n.runStream(ctx, stream)
			if errors.Is(err, context.Canceled) {
				return
			}

			n.logger.Error("stream error", "error", err)

			// TODO: Sleep with backoff?
			time.Sleep(5 * time.Second)

			n.logger.Info("reconnecting to Nomad event stream...")
		}
	}()

	return stream, nil
}

func (n *NomadSource) runStream(ctx context.Context, stream chan Event) error {
	b, err := n.getNomadConnection(ctx)
	if err != nil {
		return err
	}
	defer b.Close()

	decoder := json.NewDecoder(b)

	for {
		var streamItem nomadStreamItem
		err := decoder.Decode(&streamItem)
		if err != nil {
			return err
		}

		for _, event := range streamItem.Events {
			n.logger.Debug("Received Nomad event", "topic", event.Topic, "type", event.Type, "key", event.Key, "index", event.Index)
			stream <- Event{
				Id: event.Key,
			}
		}
	}
}

func (n *NomadSource) getNomadConnection(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", n.address, nil)
	if err != nil {
		return nil, fmt.Errorf("creating the request: %w", err)
	}
	req.Header.Set("X-Nomad-Token", n.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connecting to Nomad: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from Nomad: %s", resp.Status)
	}

	return resp.Body, nil
}
