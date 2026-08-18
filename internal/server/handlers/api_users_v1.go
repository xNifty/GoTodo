package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
)

const userSearchLimit = 10

type apiUserSearchHitJSON struct {
	UserName string `json:"user_name"`
}

// APIV1UsersSearch handles GET /api/v1/users/search?q=
func APIV1UsersSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.APIJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed.")
		return
	}
	userID, ok := utils.GetAPIUserID(r)
	if !ok {
		utils.APIJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if utf8.RuneCountInString(q) < 2 {
		_ = json.NewEncoder(w).Encode([]apiUserSearchHitJSON{})
		return
	}

	names, err := storage.SearchUsersByUsernamePrefix(q, userID, userSearchLimit)
	if err != nil {
		utils.APIJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to search users.")
		return
	}
	out := make([]apiUserSearchHitJSON, 0, len(names))
	for _, name := range names {
		out = append(out, apiUserSearchHitJSON{UserName: name})
	}
	_ = json.NewEncoder(w).Encode(out)
}
