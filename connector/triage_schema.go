package main

// llama.cpp converts this schema into a decoding grammar. Keep semantic and
// snapshot validation after generation: valid JSON alone cannot prove a useful
// answer, unique source IDs, or that a note has not changed during inference.
func triageSchema(notes []Task) map[string]any {
	ids := make([]int64, 0, len(notes))
	for _, note := range notes {
		ids = append(ids, note.ID)
	}
	return map[string]any{
		"type": "object", "additionalProperties": false, "required": []string{"items"},
		"properties": map[string]any{"items": map[string]any{
			"type": "array", "minItems": len(notes), "maxItems": len(notes),
			"items": map[string]any{
				"type": "object", "additionalProperties": false,
				"required": []string{"id", "priority", "category", "summary", "next_action"},
				"properties": map[string]any{
					"id":          map[string]any{"type": "integer", "enum": ids},
					"priority":    map[string]any{"type": "string", "enum": []string{"now", "later", "no_action"}},
					"category":    map[string]any{"type": "string", "enum": []string{"work", "personal", "learning", "admin", "other"}},
					"summary":     map[string]any{"type": "string", "minLength": 1, "maxLength": 200},
					"next_action": map[string]any{"type": "string", "minLength": 1, "maxLength": 220},
				},
			},
		}},
	}
}
