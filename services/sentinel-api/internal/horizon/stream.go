package horizon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) StreamOperations(
	ctx context.Context,
	publicKey string,
	network string,
	cursor string,
	onConnected func(),
	onOperation func(Operation) error,
) error {
	endpoint, err := c.OperationsURL(publicKey, network, cursor)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("crear stream Horizon: %w", err)
	}
	request.Header.Set("Accept", "text/event-stream")
	request.Header.Set("Cache-Control", "no-cache")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("abrir stream Horizon: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("stream Horizon respondió %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	onConnected()
	return consumeSSE(ctx, response.Body, onOperation)
}

func consumeSSE(ctx context.Context, reader io.Reader, onOperation func(Operation) error) error {
	scanner := bufio.NewScanner(reader)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 2*1024*1024)
	var data strings.Builder

	flush := func() error {
		if data.Len() == 0 {
			return nil
		}
		payload := strings.TrimSpace(data.String())
		data.Reset()
		if payload == "" || payload == "hello" {
			return nil
		}

		var operation Operation
		if err := json.Unmarshal([]byte(payload), &operation); err != nil {
			return fmt.Errorf("decodificar operación SSE: %w", err)
		}
		operation.Raw = json.RawMessage(payload)
		if operation.ID == "" {
			return nil
		}
		return onOperation(operation)
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("leer stream Horizon: %w", err)
	}
	if err := flush(); err != nil {
		return err
	}
	return io.EOF
}
