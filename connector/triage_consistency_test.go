package main

import "testing"

func TestRejectContradictoryNextActionWithoutSaving(t *testing.T) {
	inputs := []Task{{ID: 1, Text: "Pay electricity bill tomorrow."}}
	for _, priority := range []string{"now", "later"} {
		for _, action := range []string{"No action needed", "none", " No action needed. "} {
			item := TriageItem{ID: 1, Priority: priority, Category: "admin", Summary: "Pay bill", NextAction: action}
			if err := validateItems([]TriageItem{item}, inputs); err == nil {
				t.Errorf("accepted contradictory priority=%s action=%q", priority, action)
			}
		}
	}
	for _, item := range []TriageItem{
		{ID: 1, Priority: "now", Category: "admin", Summary: "Pay bill", NextAction: "Pay the bill"},
		{ID: 1, Priority: "no_action", Category: "admin", Summary: "Bill already paid", NextAction: "No action needed"},
	} {
		if err := validateItems([]TriageItem{item}, inputs); err != nil {
			t.Errorf("rejected consistent item: %v", err)
		}
	}
}
