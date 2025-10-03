package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type NomadSource struct {
	address string
	token   string
	logger  slog.Logger
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
func NewNomadSource(address, token string, logger slog.Logger) *NomadSource {
	address = address + "/v1/event/stream?topic=Job"

	return &NomadSource{
		address: address,
		token:   token,
		logger:  logger,
	}
}

// TODO: What actually happens when the context is canceled?
func (n *NomadSource) StreamEvents(ctx context.Context) (<-chan Event, error) {
	stream := make(chan Event, 10)

	go func() {
		defer close(stream)

		var currentNomadConn io.ReadCloser

		go func() {
			<-ctx.Done()
			if currentNomadConn != nil {
				currentNomadConn.Close()
			}
		}()

		for {
			b, err := n.getNomadConnection()
			if err != nil {
				n.logger.Error("getting nomad connection", "error", err)
				continue
			}
			defer b.Close()

			currentNomadConn = b

			decoder := json.NewDecoder(currentNomadConn)

			for {
				var streamItem nomadStreamItem
				err := decoder.Decode(&streamItem)
				if err != nil {
					if err == io.EOF {
						n.logger.Info("connection closed by server, reconnecting...")
						break
					}

					n.logger.Error("decoding event", "error", err)
					continue // The decoder is still fine. It can continue to be used.
				}

				for _, event := range streamItem.Events {
					n.logger.Debug("Received Nomad event", "topic", event.Topic, "type", event.Type, "key", event.Key, "index", event.Index)
					stream <- Event{
						Id: event.Key,
					}
				}
			}

			// TODO: Sleep with backoff?
			n.logger.Info("Reconnecting to Nomad event stream...")
		}
	}()

	return stream, nil
}

func (n *NomadSource) getNomadConnection() (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", n.address, nil)
	if err != nil {
		return nil, fmt.Errorf("creating the request: %w", err)
	}
	req.Header.Set("X-Nomad-Token", n.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connecting to Nomad: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("connecting to Nomad: %s", resp.Status)
	}

	return resp.Body, nil
}
