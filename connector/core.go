package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID         int64  `json:"id"`
	Text       string `json:"text"`
	Created    string `json:"created"`
	Priority   string `json:"priority,omitempty"`
	Category   string `json:"category,omitempty"`
	Summary    string `json:"summary,omitempty"`
	NextAction string `json:"next_action,omitempty"`
	Completed  bool   `json:"completed"`
}
type dataFile struct {
	Version int    `json:"version"`
	NextID  int64  `json:"next_id"`
	Tasks   []Task `json:"tasks"`
}
type Store struct {
	mu    sync.Mutex
	path  string
	state dataFile
}

func openStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "tasks.json"), state: dataFile{Version: 1, NextID: 1, Tasks: []Task{}}}
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &s.state); err != nil {
		// Preserve the damaged file for diagnosis and try the last validated snapshot.
		backup, backupErr := os.ReadFile(s.path + ".bak")
		var previous dataFile
		if backupErr != nil || json.Unmarshal(backup, &previous) != nil || previous.Version != 1 {
			return nil, fmt.Errorf("invalid local task file; no valid recovery snapshot: %w", err)
		}
		damaged := s.path + ".corrupt-" + time.Now().UTC().Format("20060102T150405")
		if renameErr := os.Rename(s.path, damaged); renameErr != nil {
			return nil, fmt.Errorf("could not preserve damaged task file: %w", renameErr)
		}
		if restoreErr := os.WriteFile(s.path, backup, 0600); restoreErr != nil {
			return nil, fmt.Errorf("could not restore saved task backup: %w", restoreErr)
		}
		s.state = previous
		fmt.Fprintf(os.Stderr, "Sloth Mode recovered tasks from backup; corrupted file preserved as %s\n", damaged)
	}
	if s.state.Version != 1 {
		return nil, errors.New("unsupported task file version")
	}
	if s.state.NextID <= 0 {
		s.state.NextID = 1
	}
	return s, nil
}

// commitLocked writes a validated new snapshot before publishing it in memory.
// The last known-good version is backed up for recovery after an interrupted write.
func (s *Store) commitLocked(next dataFile) error {
	payload, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	if next.Version != 1 || next.NextID <= 0 {
		return errors.New("invalid task snapshot")
	}
	tmp := s.path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err = f.Write(payload); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	// Keep a valid fallback before replacing the live file.
	if old, readErr := os.ReadFile(s.path); readErr == nil {
		var validated dataFile
		if json.Unmarshal(old, &validated) == nil && validated.Version == 1 {
			backTmp := s.path + ".bak.tmp"
			if writeErr := os.WriteFile(backTmp, old, 0600); writeErr != nil {
				os.Remove(tmp)
				return writeErr
			}
			if renameErr := os.Rename(backTmp, s.path+".bak"); renameErr != nil {
				os.Remove(backTmp)
				os.Remove(tmp)
				return renameErr
			}
		}
	}
	if err = os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return err
	}
	s.state = next
	return nil
}

func (s *Store) All() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Task{}, s.state.Tasks...)
}
func (s *Store) Add(text string) error {
	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > 3000 {
		return errors.New("enter a task (1–3,000 characters)")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{ID: s.state.NextID, Text: text, Created: time.Now().Format(time.RFC3339)}
	next := s.state
	next.NextID++
	next.Tasks = append(append([]Task{}, s.state.Tasks...), t)
	return s.commitLocked(next)
}

// Edit resets stale model annotations so that a changed source note cannot
// retain an earlier AI decision presented as though it describes the new text.
func (s *Store) Edit(id int64, text string) error {
	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > 3000 {
		return errors.New("enter a task (1–3,000 characters)")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.state
	next.Tasks = append([]Task(nil), s.state.Tasks...)
	for i := range next.Tasks {
		if next.Tasks[i].ID == id {
			next.Tasks[i].Text = text
			next.Tasks[i].Priority, next.Tasks[i].Category = "", ""
			next.Tasks[i].Summary, next.Tasks[i].NextAction = "", ""
			return s.commitLocked(next)
		}
	}
	return errors.New("task not found")
}

// AddBatch validates every note before touching disk. All additions share one
// atomic save, or none are imported (including if persistence fails).
func (s *Store) AddBatch(lines []string) (int, error) {
	if len(lines) == 0 || len(lines) > 40 {
		return 0, errors.New("import between 1 and 40 notes")
	}
	normalized := make([]string, len(lines))
	for i, line := range lines {
		value := strings.TrimSpace(line)
		if value == "" || len([]rune(value)) > 3000 {
			return 0, fmt.Errorf("note %d must contain 1–3,000 characters", i+1)
		}
		normalized[i] = value
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.state
	next.Tasks = append([]Task(nil), s.state.Tasks...)
	for _, text := range normalized {
		next.Tasks = append(next.Tasks, Task{ID: next.NextID, Text: text, Created: time.Now().Format(time.RFC3339)})
		next.NextID++
	}
	if err := s.commitLocked(next); err != nil {
		return 0, err
	}
	return len(normalized), nil
}

func (s *Store) Complete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.state
	next.Tasks = append([]Task(nil), s.state.Tasks...)
	for i := range next.Tasks {
		if next.Tasks[i].ID == id {
			next.Tasks[i].Completed = !next.Tasks[i].Completed
			return s.commitLocked(next)
		}
	}
	return errors.New("task not found")
}
func (s *Store) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.state
	next.Tasks = append([]Task(nil), s.state.Tasks...)
	for i, t := range next.Tasks {
		if t.ID == id {
			next.Tasks = append(next.Tasks[:i], next.Tasks[i+1:]...)
			return s.commitLocked(next)
		}
	}
	return errors.New("task not found")
}
func (s *Store) Pending(max int) []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Task{}
	for _, t := range s.state.Tasks {
		if !t.Completed && t.Priority == "" {
			out = append(out, t)
			if len(out) >= max {
				break
			}
		}
	}
	return out
}
func (s *Store) Apply(items []TriageItem, inputs []Task) error {
	if err := validateItems(items, inputs); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.state
	next.Tasks = append([]Task(nil), s.state.Tasks...)
	for _, r := range items {
		for i := range next.Tasks {
			if next.Tasks[i].ID == r.ID && next.Tasks[i].Priority == "" && !next.Tasks[i].Completed {
				t := &next.Tasks[i]
				t.Priority = r.Priority
				t.Category = r.Category
				t.Summary = r.Summary
				t.NextAction = r.NextAction
			}
		}
	}
	return s.commitLocked(next)
}
func (s *Store) Export(dir string) (string, error) {
	tasks := s.All()
	sort.SliceStable(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	var b strings.Builder
	b.WriteString("# Sloth Mode — private briefing\n\n")
	for _, bucket := range []struct{ key, title string }{{"now", "Do next"}, {"later", "Save for later"}, {"no_action", "No action needed"}, {"", "Not yet reviewed"}} {
		b.WriteString("## " + bucket.title + "\n")
		count := 0
		for _, t := range tasks {
			if !t.Completed && t.Priority == bucket.key {
				count++
				label := t.Summary
				if label == "" {
					label = t.Text
				}
				fmt.Fprintf(&b, "- %s\n  - Original #%d: %s\n", escapeMarkdown(label), t.ID, escapeMarkdown(t.Text))
				if t.NextAction != "" {
					fmt.Fprintf(&b, "  - Next action: %s\n", escapeMarkdown(t.NextAction))
				}
			}
		}
		if count == 0 {
			b.WriteString("- Nothing here.\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("---\nGenerated on your computer. Model inference uses localhost only.\n")
	out := filepath.Join(dir, "briefing.md")
	err := os.WriteFile(out, []byte(b.String()), 0600)
	return out, err
}
func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

type TriageItem struct {
	ID         int64  `json:"id"`
	Priority   string `json:"priority"`
	Category   string `json:"category"`
	Summary    string `json:"summary"`
	NextAction string `json:"next_action"`
}

func validateItems(got []TriageItem, inputs []Task) error {
	if len(got) != len(inputs) {
		return fmt.Errorf("AI returned %d tasks, expected %d (nothing saved)", len(got), len(inputs))
	}
	want := map[int64]bool{}
	for _, x := range inputs {
		want[x.ID] = true
	}
	seen := map[int64]bool{}
	for _, x := range got {
		if !want[x.ID] || seen[x.ID] {
			return fmt.Errorf("invalid or duplicate AI source ID %d", x.ID)
		}
		seen[x.ID] = true
		if x.Priority != "now" && x.Priority != "later" && x.Priority != "no_action" {
			return fmt.Errorf("invalid AI priority %q", x.Priority)
		}
		if x.Category != "work" && x.Category != "personal" && x.Category != "learning" && x.Category != "admin" && x.Category != "other" {
			return fmt.Errorf("invalid category %q", x.Category)
		}
		if len(strings.TrimSpace(x.Summary)) == 0 || len([]rune(x.Summary)) > 200 || len(strings.TrimSpace(x.NextAction)) == 0 || len([]rune(x.NextAction)) > 220 {
			return errors.New("AI returned invalid summary or action")
		}
	}
	return nil
}
func parseTriage(text string, inputs []Task) ([]TriageItem, error) {
	s := strings.TrimSpace(text)
	// Some compatible llama.cpp versions wrap JSON in markdown despite instructions.
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	}
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start < 0 || end <= start {
		return nil, errors.New("Gemma did not return JSON")
	}
	var data struct {
		Items []TriageItem `json:"items"`
	}
	dec := json.NewDecoder(bytes.NewBufferString(s[start : end+1]))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("Gemma JSON error: %w", err)
	}
	if err := validateItems(data.Items, inputs); err != nil {
		return nil, err
	}
	return data.Items, nil
}
