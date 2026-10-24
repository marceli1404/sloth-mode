package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// This is a limited English phrase-presence check, not a date parser or an
// urgency classifier. It warns about omissions without changing Gemma's
// suggestions, copying new deadlines into them, or interpreting calendar dates.
var timingPhrase = regexp.MustCompile(`(?i)\b(?:today|tomorrow|tonight|yesterday|(?:this|next)\s+(?:week|weekend|month|monday|tuesday|wednesday|thursday|friday|saturday|sunday)|(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday)|(?:[0-1]?[0-9]|2[0-3]):[0-5][0-9](?:\s*[ap]m)?|(?:1[0-2]|0?[1-9])\s*[ap]m)\b`)

func needsTimingReview(item TriageItem, original string) bool {
	if item.Priority == "no_action" {
		return false
	}
	output := strings.ToLower(strings.Join(strings.Fields(item.Summary+" "+item.NextAction), " "))
	for _, phrase := range timingPhrase.FindAllString(original, -1) {
		phrase = strings.ToLower(strings.Join(strings.Fields(phrase), " "))
		if !strings.Contains(output, phrase) {
			return true
		}
	}
	return false
}

func planSummary(items []TriageItem, inputs []Task) string {
	sources := make(map[int64]string, len(inputs))
	for _, note := range inputs {
		sources[note.ID] = note.Text
	}
	ids := []string{}
	for _, item := range items {
		if needsTimingReview(item, sources[item.ID]) {
			ids = append(ids, "#"+strconv.FormatInt(item.ID, 10))
		}
	}
	message := fmt.Sprintf("Gemma drafted %d tasks locally. Review the suggestions before acting.", len(items))
	if len(ids) > 0 {
		message += " Timing was omitted from suggestions for " + strings.Join(ids, ", ") + "; check their original notes."
	}
	return message
}
