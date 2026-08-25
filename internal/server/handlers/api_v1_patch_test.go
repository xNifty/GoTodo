package handlers

import (
	"encoding/json"
	"testing"
)

func TestAPITaskPatchSprintIDJSON(t *testing.T) {
	t.Run("omitted leaves unchanged", func(t *testing.T) {
		var req apiTaskPatchRequest
		if err := json.Unmarshal([]byte(`{"title":"Keep sprint"}`), &req); err != nil {
			t.Fatal(err)
		}
		if req.SprintID.Set {
			t.Fatal("omitted sprint_id should not be set")
		}
		if got := req.SprintID.toPatchInt(true); got != nil {
			t.Fatalf("omitted sprint_id patch=%#v, want nil", got)
		}
	})

	t.Run("null clears to backlog", func(t *testing.T) {
		var req apiTaskPatchRequest
		if err := json.Unmarshal([]byte(`{"sprint_id":null}`), &req); err != nil {
			t.Fatal(err)
		}
		if !req.SprintID.Set || !req.SprintID.Null {
			t.Fatalf("null sprint_id: set=%v null=%v", req.SprintID.Set, req.SprintID.Null)
		}
		got := req.SprintID.toPatchInt(true)
		if got == nil || *got != nil {
			t.Fatalf("null sprint_id patch=%#v, want explicit clear", got)
		}
	})

	t.Run("zero clears to backlog", func(t *testing.T) {
		var req apiTaskPatchRequest
		if err := json.Unmarshal([]byte(`{"sprint_id":0}`), &req); err != nil {
			t.Fatal(err)
		}
		if !req.SprintID.Set || req.SprintID.Null || req.SprintID.Value != 0 {
			t.Fatalf("zero sprint_id: %+v", req.SprintID)
		}
		got := req.SprintID.toPatchInt(true)
		if got == nil || *got != nil {
			t.Fatalf("zero sprint_id patch=%#v, want explicit clear", got)
		}
	})

	t.Run("value assigns sprint", func(t *testing.T) {
		var req apiTaskPatchRequest
		if err := json.Unmarshal([]byte(`{"sprint_id":42}`), &req); err != nil {
			t.Fatal(err)
		}
		if !req.SprintID.Set || req.SprintID.Null || req.SprintID.Value != 42 {
			t.Fatalf("value sprint_id: %+v", req.SprintID)
		}
		got := req.SprintID.toPatchInt(true)
		if got == nil || *got == nil || **got != 42 {
			t.Fatalf("value sprint_id patch=%#v, want 42", got)
		}
	})
}

func TestAPITaskPatchEstimatePointsJSONNullClears(t *testing.T) {
	var req apiTaskPatchRequest
	if err := json.Unmarshal([]byte(`{"estimate_points":null}`), &req); err != nil {
		t.Fatal(err)
	}
	if !req.EstimatePoints.Set || !req.EstimatePoints.Null {
		t.Fatalf("null estimate_points: %+v", req.EstimatePoints)
	}
	got := req.EstimatePoints.toPatchInt(false)
	if got == nil || *got != nil {
		t.Fatalf("null estimate_points patch=%#v, want explicit clear", got)
	}

	var keep apiTaskPatchRequest
	if err := json.Unmarshal([]byte(`{"title":"Keep estimate"}`), &keep); err != nil {
		t.Fatal(err)
	}
	if keep.EstimatePoints.toPatchInt(false) != nil {
		t.Fatal("omitted estimate_points should leave the field unchanged")
	}
}

func TestAPISprintPatchLockDateJSON(t *testing.T) {
	t.Run("omitted leaves unchanged", func(t *testing.T) {
		var req apiSprintPatchRequest
		if err := json.Unmarshal([]byte(`{"name":"Keep lock"}`), &req); err != nil {
			t.Fatal(err)
		}
		if req.LockDate.Set {
			t.Fatal("omitted lock_date should not be set")
		}
		if got := req.LockDate.toPatchString(); got != nil {
			t.Fatalf("omitted lock_date patch=%#v, want nil", got)
		}
	})

	t.Run("null clears lock", func(t *testing.T) {
		var req apiSprintPatchRequest
		if err := json.Unmarshal([]byte(`{"lock_date":null}`), &req); err != nil {
			t.Fatal(err)
		}
		if !req.LockDate.Set || !req.LockDate.Null {
			t.Fatalf("null lock_date: set=%v null=%v", req.LockDate.Set, req.LockDate.Null)
		}
		got := req.LockDate.toPatchString()
		if got == nil || *got != "" {
			t.Fatalf("null lock_date patch=%#v, want empty string", got)
		}
	})

	t.Run("value sets lock date", func(t *testing.T) {
		var req apiSprintPatchRequest
		if err := json.Unmarshal([]byte(`{"lock_date":"2026-08-25"}`), &req); err != nil {
			t.Fatal(err)
		}
		if !req.LockDate.Set || req.LockDate.Null || req.LockDate.Value != "2026-08-25" {
			t.Fatalf("value lock_date: %+v", req.LockDate)
		}
		got := req.LockDate.toPatchString()
		if got == nil || *got != "2026-08-25" {
			t.Fatalf("value lock_date patch=%#v, want 2026-08-25", got)
		}
	})
}
