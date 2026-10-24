package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// The model already produced these fictional outputs. This regression checks
// the warning layer without running inference or downloading any assets.
func TestRecordedTimingOmissionsReachReviewWarning(t *testing.T) {
	root := filepath.Join("..", "docs", "evidence", "2026-10-24")
	var corpus struct {
		Cases []struct {
			ID   int64  `json:"id"`
			Text string `json:"text"`
		} `json:"cases"`
	}
	var result struct {
		Batches []struct {
			Items []TriageItem `json:"items"`
		} `json:"batches"`
	}
	for _, load := range []struct {
		name   string
		target any
	}{{"inputs.json", &corpus}, {"candidate-results.json", &result}} {
		data, err := os.ReadFile(filepath.Join(root, load.name))
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, load.target); err != nil {
			t.Fatal(err)
		}
	}
	sources := map[int64]string{}
	for _, note := range corpus.Cases {
		sources[note.ID] = note.Text
	}
	warnings := []int64{}
	for _, batch := range result.Batches {
		for _, item := range batch.Items {
			if needsTimingReview(item, sources[item.ID]) {
				warnings = append(warnings, item.ID)
			}
		}
	}
	sort.Slice(warnings, func(i, j int) bool { return warnings[i] < warnings[j] })
	want := []int64{201, 202, 203, 206, 209, 210, 212}
	if !reflect.DeepEqual(warnings, want) {
		t.Fatalf("recorded omission warnings=%v, want %v", warnings, want)
	}
}
