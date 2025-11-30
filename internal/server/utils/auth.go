package utils

import (
	"GoTodo/internal/sessionstore"
	"fmt"
	"net/http"
)

func GetSessionUser(r *http.Request) (email string, roleID int, permissions []string, loggedIn bool) {
	session, err := sessionstore.Store.Get(r, "session")
	if err != nil {
		fmt.Printf("GetSessionUser error getting session: %v\n", err)
		return "", 0, nil, false
	}

	emailVal, ok := session.Values["email"]
	if !ok {
		return "", 0, nil, false
	}

	email, ok = emailVal.(string)
	if !ok {
		return "", 0, nil, false
	}

	roleIDVal, ok := session.Values["role_id"]
	if !ok {
		return email, 0, nil, true
	}

	roleID, ok = roleIDVal.(int)
	if !ok {
		return email, 0, nil, true
	}

	permissionsVal, ok := session.Values["permissions"]
	if !ok {
		return email, roleID, []string{}, true
	}

	permissions, ok = permissionsVal.([]string)
	if !ok {
		if permsInterface, ok := permissionsVal.([]interface{}); ok {
			permissions = make([]string, len(permsInterface))
			for i, v := range permsInterface {
				if str, ok := v.(string); ok {
					permissions[i] = str
				}
			}
		} else {
			permissions = []string{}
		}
	}

	return email, roleID, permissions, true
}

// GetSessionUserWithTimezone retrieves session user data including timezone
func GetSessionUserWithTimezone(r *http.Request) (email string, roleID int, permissions []string, timezone string, loggedIn bool) {
	session, err := sessionstore.Store.Get(r, "session")
	if err != nil {
		fmt.Printf("GetSessionUserWithTimezone error getting session: %v\n", err)
		return "", 0, nil, "America/New_York", false
	}

	emailVal, ok := session.Values["email"]
	if !ok {
		return "", 0, nil, "America/New_York", false
	}

	email, ok = emailVal.(string)
	if !ok {
		return "", 0, nil, "America/New_York", false
	}

	roleIDVal, ok := session.Values["role_id"]
	if !ok {
		return email, 0, nil, "America/New_York", true
	}

	roleID, ok = roleIDVal.(int)
	if !ok {
		return email, 0, nil, "America/New_York", true
	}

	permissionsVal, ok := session.Values["permissions"]
	if !ok {
		return email, roleID, []string{}, "America/New_York", true
	}

	permissions, ok = permissionsVal.([]string)
	if !ok {
		if permsInterface, ok := permissionsVal.([]interface{}); ok {
			permissions = make([]string, len(permsInterface))
			for i, v := range permsInterface {
				if str, ok := v.(string); ok {
					permissions[i] = str
				}
			}
		} else {
			permissions = []string{}
		}
	}

	timezoneVal, ok := session.Values["timezone"]
	if !ok {
		return email, roleID, permissions, "America/New_York", true
	}

	timezone, ok = timezoneVal.(string)
	if !ok {
		return email, roleID, permissions, "America/New_York", true
	}

	return email, roleID, permissions, timezone, true
}

// RequireAuth is a middleware that checks if a user is logged in
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, _, loggedIn := GetSessionUser(r)
		if !loggedIn {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// RequirePermission is a middleware that checks if a user has a specific permission
func RequirePermission(permission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _, permissions, loggedIn := GetSessionUser(r)
		if !loggedIn {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}

		hasPermission := false
		for _, p := range permissions {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			http.Error(w, "Forbidden: You don't have permission to access this resource", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
