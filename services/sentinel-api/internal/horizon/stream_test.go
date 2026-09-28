package horizon

import (
	"context"
	"strings"
	"testing"
)

func TestConsumeSSE(t *testing.T) {
	stream := strings.NewReader("event: open\ndata: hello\n\nevent: open\ndata: \"hello\"\n\nevent: message\ndata: {\"id\":\"1\",\"paging_token\":\"10\",\"type\":\"set_options\"}\n\nevent: message\ndata: {\"id\":\"2\",\"paging_token\":\"11\",\"type\":\"change_trust\"}\n\n")
	operations := make([]Operation, 0, 2)
	err := consumeSSE(context.Background(), stream, func(operation Operation) error {
		operations = append(operations, operation)
		return nil
	})
	if err != nil {
		t.Fatalf("consumeSSE returned error: %v", err)
	}
	if len(operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(operations))
	}
	if operations[0].ID != "1" || operations[1].PagingToken != "11" {
		t.Fatalf("unexpected operations: %#v", operations)
	}
}
