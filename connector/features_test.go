package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestBulkImportIsAllOrNothing(t *testing.T) {
	b := testBridge(t)
	w := invoke(t, b, "POST", "/v1/tasks/bulk", bridgeOrigin, b.secret, `{"lines":["Schedule a call","", "Pay a bill"]}`)
	if w.Code != 400 || len(b.store.All()) != 0 {
		t.Fatalf("partial import: %d %s", w.Code, w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/tasks/bulk", bridgeOrigin, b.secret, `{"lines":["Schedule a call","Pay a bill"]}`)
	if w.Code != 201 || len(b.store.All()) != 2 {
		t.Fatalf("bulk import failed %d %s", w.Code, w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/tasks/bulk", "https://different.example", b.secret, `{"lines":["evil"]}`)
	if w.Code != 403 || len(b.store.All()) != 2 {
		t.Fatal("bad origin accepted")
	}
	var response struct {
		Tasks []Task `json:"tasks"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(invoke(t, b, "POST", "/v1/tasks/bulk", bridgeOrigin, b.secret, `{"lines":["Read tonight"]}`).Body.String()), &response); err != nil || response.Count != 1 || len(response.Tasks) != 3 {
		t.Fatal("batch response invalid")
	}
}

func TestEditResetsAIAndRetainsID(t *testing.T) {
	b := testBridge(t)
	if err := b.store.Add("Earlier text"); err != nil {
		t.Fatal(err)
	}
	b.store.mu.Lock()
	b.store.state.Tasks[0].Priority = "now"
	b.store.state.Tasks[0].Category = "admin"
	b.store.state.Tasks[0].Summary = "Old summary"
	b.store.state.Tasks[0].NextAction = "Old action"
	b.store.mu.Unlock()
	w := invoke(t, b, "POST", "/v1/tasks/edit", bridgeOrigin, b.secret, `{"id":1,"text":" Rewritten text "}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	got := b.store.All()[0]
	if got.ID != 1 || got.Text != "Rewritten text" || got.Priority != "" || got.Summary != "" || got.NextAction != "" || got.Category != "" {
		t.Fatalf("stale result after edit: %#v", got)
	}
	w = invoke(t, b, "POST", "/v1/tasks/edit", bridgeOrigin, b.secret, `{"id":1,"text":" "}`)
	if w.Code != 400 || b.store.All()[0].Text != "Rewritten text" {
		t.Fatal("invalid edit changed task")
	}
}

func TestBatchPersistsAcrossReopen(t *testing.T) {
	b := testBridge(t)
	_, err := b.store.AddBatch([]string{"one", "two"})
	if err != nil {
		t.Fatal(err)
	}
	s, err := openStore(b.base)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.All()) != 2 {
		t.Fatal("reopened data lost")
	}
	if _, err = s.AddBatch([]string{strings.Repeat("a", 3001)}); err == nil {
		t.Fatal("oversized line accepted")
	}
	if len(s.All()) != 2 {
		t.Fatal("invalid import persisted")
	}
	info, err := os.Stat(s.path)
	if err != nil || info.Size() == 0 {
		t.Fatal("store missing")
	}
}
