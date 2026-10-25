package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTriageRequestConstrainsOutputAndStillRejectsMalformedPriority(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ResponseFormat struct {
				Type   string `json:"type"`
				Schema struct {
					Properties struct {
						Items struct {
							MinItems int `json:"minItems"`
							MaxItems int `json:"maxItems"`
							Items    struct {
								Properties map[string]struct {
									Enum json.RawMessage `json:"enum"`
								} `json:"properties"`
							} `json:"items"`
						} `json:"items"`
					} `json:"properties"`
				} `json:"schema"`
			} `json:"response_format"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		items := request.ResponseFormat.Schema.Properties.Items
		if request.ResponseFormat.Type != "json_object" || items.MinItems != 1 || items.MaxItems != 1 {
			t.Error("request must constrain one source result")
		}
		for field, want := range map[string]string{"id": "[7]", "priority": `["now","later","no_action"]`, "category": `["work","personal","learning","admin","other"]`} {
			if string(items.Items.Properties[field].Enum) != want {
				t.Errorf("%s enum = %s", field, items.Items.Properties[field].Enum)
			}
		}
		if len(request.Messages) != 1 || strings.Contains(request.Messages[0].Content, "-> now/work") {
			t.Error("ambiguous shorthand example remains")
		}
		// Simulate an incompatible runtime that ignores the decoding schema.
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": `{"items":[{"id":7,"priority":"now/work","category":"work","summary":"Poster","next_action":"Submit the poster"}]}`}}}})
	}))
	defer server.Close()
	engine := inferEngine{host: server.URL, client: server.Client()}
	if _, err := engine.triage(context.Background(), []Task{{ID: 7, Text: "Submit the poster tomorrow"}}); err == nil || !strings.Contains(err.Error(), "invalid AI priority") {
		t.Fatalf("expected priority rejection, got %v", err)
	}
}
