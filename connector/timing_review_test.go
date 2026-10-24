package main

import (
	"strings"
	"testing"
)

func TestTimingOmissionWarnsWithoutRewriting(t *testing.T) {
	cases := []struct {
		source, summary, action, priority string
		warn                              bool
	}{
		{"Return library book tomorrow", "Return book", "Return the library book", "now", true},
		{"Call the manager today at 5 PM", "Call manager today", "Call the manager", "now", true},
		{"Pay bill tomorrow", "Pay bill TOMORROW", "Pay the bill", "now", false},
		{"Read next week", "Read next week", "Read a chapter", "later", false},
		{"Read when I have time", "Read a novel", "Read a chapter", "later", false},
		{"Finished desk yesterday", "Desk finished", "No action needed", "no_action", false},
		{"Not today; consider it next week", "Consider next week", "Review next week", "later", true},
	}
	for _, tc := range cases {
		item := TriageItem{ID: 1, Priority: tc.priority, Summary: tc.summary, NextAction: tc.action}
		before := item
		if got := needsTimingReview(item, tc.source); got != tc.warn {
			t.Errorf("%q warning=%v, want %v", tc.source, got, tc.warn)
		}
		if item != before {
			t.Fatal("timing checker rewrote model output")
		}
	}
}

func TestPlanSummaryPointsToSourceIDs(t *testing.T) {
	inputs := []Task{{ID: 42, Text: "Return book tomorrow"}}
	items := []TriageItem{{ID: 42, Priority: "now", Summary: "Return book", NextAction: "Return the book"}}
	message := planSummary(items, inputs)
	if !strings.Contains(message, "#42") || !strings.Contains(message, "original notes") {
		t.Fatalf("missing source review guidance: %s", message)
	}
}
