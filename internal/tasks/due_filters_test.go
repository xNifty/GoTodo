package tasks

import (
	"strings"
	"testing"
)

func TestNormalizeDueFilter(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"overdue", "overdue"},
		{"TODAY", "today"},
		{" week ", "week"},
		{"through_week", "through_week"},
		{"THROUGH_WEEK", "through_week"},
		{"none", "none"},
		{"invalid", ""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := NormalizeDueFilter(tc.in); got != tc.want {
			t.Errorf("NormalizeDueFilter(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAppendDueDateCondition(t *testing.T) {
	where, args := appendDueDateCondition("user_id = $1", []interface{}{1}, "overdue", "America/New_York", "t")
	if where == "user_id = $1" {
		t.Fatal("expected overdue condition to be appended")
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[1] != "America/New_York" {
		t.Fatalf("timezone arg = %v", args[1])
	}

	throughWeek, throughArgs := appendDueDateCondition("user_id = $1", []interface{}{1}, "through_week", "UTC", "")
	if !strings.Contains(throughWeek, "date_trunc('week'") {
		t.Fatalf("expected calendar week bound, got %q", throughWeek)
	}
	if !strings.Contains(throughWeek, "completed") {
		t.Fatalf("expected incomplete constraint, got %q", throughWeek)
	}
	if len(throughArgs) != 2 || throughArgs[1] != "UTC" {
		t.Fatalf("through_week args = %v", throughArgs)
	}

	where2, args2 := appendDueDateCondition("x = 1", nil, "", "UTC", "")
	if where2 != "x = 1" || len(args2) != 0 {
		t.Fatalf("empty filter should be unchanged, got %q %v", where2, args2)
	}
}
