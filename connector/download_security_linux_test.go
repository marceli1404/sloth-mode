//go:build linux

package main

import (
  "context"
  "crypto/sha256"
  "encoding/hex"
  "net/http"
  "net/http/httptest"
  "os"
  "path/filepath"
  "testing"
)

func TestDownloadDoesNotFollowPredictablePartialSymlink(t *testing.T) {
  payload := []byte("verified-model-chunk")
  sum := sha256.Sum256(payload)
  server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(payload) }))
  defer server.Close()
  previous := httpClient
  httpClient = server.Client()
  defer func() { httpClient = previous }()
  root := t.TempDir()
  sentinel := filepath.Join(root, "private-note")
  if err := os.WriteFile(sentinel, []byte("untouched"), 0600); err != nil { t.Fatal(err) }
  target := filepath.Join(root, "model.gguf")
  if err := os.Symlink(sentinel, target+".partial"); err != nil { t.Fatal(err) }
  if err := download(context.Background(), server.URL, target, 200, hex.EncodeToString(sum[:]), func(string) {}); err != nil { t.Fatal(err) }
  data, err := os.ReadFile(sentinel)
  if err != nil || string(data) != "untouched" { t.Fatalf("symlink target overwritten: %q %v", data, err) }
  data, err = os.ReadFile(target)
  if err != nil || string(data) != string(payload) { t.Fatalf("bad verified download: %q %v", data, err) }
}
