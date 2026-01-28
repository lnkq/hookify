package models

import (
	"encoding/json"
	"testing"
)

func TestRawEvent_UnmarshalJSON_HookIDAlias(t *testing.T) {
	var e RawEvent
	data := []byte(`{"id":1,"hook_id":42,"payload":"{}","status":"pending"}`)

	if err := json.Unmarshal(data, &e); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if e.WebhookID != 42 {
		t.Fatalf("expected WebhookID=42, got %d", e.WebhookID)
	}
	if e.ID != 1 {
		t.Fatalf("expected ID=1, got %d", e.ID)
	}
}
