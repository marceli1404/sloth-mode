package main

import (
	"os"
	"strings"
	"testing"
)

func TestPlanRejectsChangedInputWithoutPartialSave(t *testing.T) {
	for _, mutation := range []string{"edit", "complete", "delete", "other-plan"} {
		t.Run(mutation, func(t *testing.T) {
			store, err := openStore(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{"Pay the bill tomorrow", "Read the novel later"} {
				if err := store.Add(text); err != nil {
					t.Fatal(err)
				}
			}
			inputs := store.Pending(10)
			items := []TriageItem{
				{ID: 1, Priority: "now", Category: "admin", Summary: "Pay bill", NextAction: "Pay the bill"},
				{ID: 2, Priority: "later", Category: "learning", Summary: "Read novel", NextAction: "Read a chapter"},
			}
			switch mutation {
			case "edit":
				err = store.Edit(2, "Call the dentist today")
			case "complete":
				err = store.Complete(2)
			case "delete":
				err = store.Delete(2)
			case "other-plan":
				err = store.Apply(items[1:], inputs[1:])
			}
			if err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(store.path)
			if err != nil {
				t.Fatal(err)
			}
			if err = store.Apply(items, inputs); err == nil || !strings.Contains(err.Error(), "changed while Gemma") {
				t.Fatalf("stale plan not rejected clearly: %v", err)
			}
			after, err := os.ReadFile(store.path)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("rejected batch changed stored notes")
			}
			if store.All()[0].Priority != "" {
				t.Fatal("unchanged first note received a partial result")
			}
		})
	}
}

func TestPlanAllowsUnrelatedNewNotes(t *testing.T) {
	store, err := openStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Add("Read the novel later"); err != nil {
		t.Fatal(err)
	}
	inputs := store.Pending(10)
	if err = store.Add("A new thought captured during inference"); err != nil {
		t.Fatal(err)
	}
	items := []TriageItem{{ID: 1, Priority: "later", Category: "learning", Summary: "Read novel", NextAction: "Read a chapter"}}
	if err = store.Apply(items, inputs); err != nil {
		t.Fatal(err)
	}
	got := store.All()
	if got[0].Priority != "later" || got[1].Priority != "" {
		t.Fatal("unrelated note changed")
	}
}
