package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"GoTodo/internal/server/utils"
)

func TestAPIV1UsersSearchMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/search", nil)
	req = utils.SetAPIUserID(req, 1)
	rec := httptest.NewRecorder()
	APIV1UsersSearch(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestAPIV1UsersSearchUnauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/search?q=al", nil)
	rec := httptest.NewRecorder()
	APIV1UsersSearch(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAPIV1UsersSearchEmptyQuery(t *testing.T) {
	for _, path := range []string{
		"/api/v1/users/search",
		"/api/v1/users/search?q=",
		"/api/v1/users/search?q=%20",
		"/api/v1/users/search?q=a",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req = utils.SetAPIUserID(req, 1)
		rec := httptest.NewRecorder()
		APIV1UsersSearch(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d, want %d", path, rec.Code, http.StatusOK)
		}
		var body []map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s json: %v body=%q", path, err, rec.Body.String())
		}
		if len(body) != 0 {
			t.Fatalf("%s got %d hits, want empty array", path, len(body))
		}
	}
}
