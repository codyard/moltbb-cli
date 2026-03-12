package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSignalRInvocation_EmptyArgs_EncodesArgumentsArray(t *testing.T) {
	msg := signalrMsg{
		Type:         1,
		InvocationId: "1",
		Target:       "JoinPipeline",
		Arguments:    make([]json.RawMessage, 0),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal invocation: %v", err)
	}

	encoded := string(data)
	if !strings.Contains(encoded, `"arguments":[]`) {
		t.Fatalf("expected empty arguments array, got %s", encoded)
	}
}
