package handlers

import "testing"

func TestFormatEventLabelStatusChanged(t *testing.T) {
	if got := formatEventLabel("status_changed", nil); got != "Status changed" {
		t.Fatalf("legacy label=%q", got)
	}
	got := formatEventLabel("status_changed", map[string]interface{}{"to": "In Progress"})
	if got != "Status changed · In Progress" {
		t.Fatalf("label=%q", got)
	}
}
