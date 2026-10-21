package main

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"
)

// TestRealGemmaTriageOptIn exercises actual local llama.cpp inference, not the
// deterministic website demo or a mocked model. It never downloads anything.
// Run explicitly on a test machine with verified model/runtime already installed.
func TestRealGemmaTriageOptIn(t *testing.T) {
	if os.Getenv("SLOTH_REAL_GEMMA_TEST") != "1" {
		t.Skip("set SLOTH_REAL_GEMMA_TEST=1 on a machine with installed Gemma and llama.cpp")
	}
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("real inference smoke test currently supports Windows and Linux")
	}
	backend := os.Getenv("SLOTH_TEST_BACKEND")
	if backend == "" {
		backend = "cpu"
	}
	if backend != "cpu" && backend != "vulkan" {
		t.Fatalf("invalid SLOTH_TEST_BACKEND %q (expected cpu or vulkan)", backend)
	}
	base, err := appDataDir()
	if err != nil {
		t.Fatal(err)
	}
	if !detectAI(base, backend) {
		t.Fatalf("verified model and %s runtime must be installed first; test will not download anything", backend)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	bootStart := time.Now()
	engine, err := startEngine(ctx, base, backend, func(message string) {
		t.Log("runtime:", message)
	})
	if err != nil {
		t.Fatalf("real %s startup failed: %v", backend, err)
	}
	defer engine.stop()
	bootDuration := time.Since(bootStart)

	// Synthetic tasks prevent test reports accidentally exposing private notes.
	inputs := []Task{
		{ID: 1, Text: "Pay electricity bill tomorrow."},
		{ID: 2, Text: "Read a chapter of a novel when I have time."},
		{ID: 3, Text: "I finished tidying my desk yesterday; no more action is needed."},
	}
	start := time.Now()
	items, err := engine.triage(ctx, inputs)
	triageDuration := time.Since(start)
	if err != nil {
		t.Fatalf("real Gemma response was not valid task triage: %v", err)
	}
	if err := validateItems(items, inputs); err != nil {
		t.Fatalf("Gemma result omitted, duplicated, or changed task IDs: %v", err)
	}
	t.Logf("REAL GEMMA RESULT: OS=%s arch=%s backend=%s quant=%s tasks=%d startup=%s triage=%s",
		runtime.GOOS, runtime.GOARCH, backend, preferredQuant(base, profileHardware()),
		len(items), bootDuration.Round(time.Millisecond), triageDuration.Round(time.Millisecond))
	for _, item := range items {
		t.Logf("task #%d priority=%s category=%s", item.ID, item.Priority, item.Category)
	}
	t.Log("This test verifies local inference while online; repeat with the network disconnected for the offline release gate. No manual-effort savings are inferred from these timings.")
}
