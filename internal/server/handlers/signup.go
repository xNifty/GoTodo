package handlers

import (
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func SignupPageHandler(w http.ResponseWriter, r *http.Request) {
	email, _, _, loggedIn := utils.GetSessionUser(r)
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	enableRegistration := true
	if settings, err := storage.GetSiteSettings(); err == nil && settings != nil {
		enableRegistration = settings.EnableRegistration
	}
	if !enableRegistration {
		utils.SetFlash(w, r, "This site is currently not accepting new signups")
		basePath := utils.GetBasePath()
		if basePath == "/" {
			basePath = ""
		}
		http.Redirect(w, r, basePath+"/", http.StatusSeeOther)
		return
	}

	context := map[string]interface{}{
		"LoggedIn":  false,
		"UserEmail": email,
		"Title":     "GoTodo - Sign Up",
		"Token":     "",
	}

	// If a token query parameter is provided (e.g., via /register?token=XYZ), preserve it
	if t := r.URL.Query().Get("token"); t != "" {
		context["Token"] = t
	}

	utils.RenderTemplate(w, r, "signup.html", context)
}

// RegisterHandler redirects to signup page while populating the token from the URL
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	email, _, _, loggedIn := utils.GetSessionUser(r)
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	enableRegistration := true
	if settings, err := storage.GetSiteSettings(); err == nil && settings != nil {
		enableRegistration = settings.EnableRegistration
	}
	if !enableRegistration {
		utils.SetFlash(w, r, "This site is currently not accepting new signups")
		basePath := utils.GetBasePath()
		if basePath == "/" {
			basePath = ""
		}
		http.Redirect(w, r, basePath+"/", http.StatusSeeOther)
		return
	}

	token := r.URL.Query().Get("token")

	context := map[string]interface{}{
		"LoggedIn":    false,
		"UserEmail":   email,
		"Title":       "GoTodo - Sign Up",
		"Token":       token,
		"TokenLocked": true,
	}

	utils.RenderTemplate(w, r, "signup.html", context)
}

func APISignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	enableRegistration := true
	inviteOnly := true
	if settings, err := storage.GetSiteSettings(); err == nil && settings != nil {
		enableRegistration = settings.EnableRegistration
		inviteOnly = settings.InviteOnly
	}
	if !enableRegistration {
		w.Header().Set("HX-Retarget", "#signup-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "This site is currently not accepting new signups")
		return
	}

	_, _, _, loggedIn := utils.GetSessionUser(r)
	if loggedIn {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "You are already logged in")
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error parsing form data")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	token := strings.TrimSpace(r.FormValue("token"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	timezone := strings.TrimSpace(r.FormValue("timezone"))

	if email == "" || password == "" || confirmPassword == "" || timezone == "" || (inviteOnly && token == "") {
		w.Header().Set("HX-Retarget", "#signup-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "All fields are required")
		return
	}

	if password != confirmPassword {
		w.Header().Set("HX-Retarget", "#signup-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.Header().Set("HX-Trigger", "clear-passwords")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Passwords do not match")
		return
	}

	pool, err := storage.OpenDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	defer storage.CloseDatabase(pool)

	var inviteID int
	var inviteUsed int
	if inviteOnly {
		fmt.Println("signup attempt:", email, token)
		err = pool.QueryRow(context.Background(), "SELECT id, inviteused FROM invites WHERE email = $1 AND token = $2", email, token).Scan(&inviteID, &inviteUsed)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || err.Error() == "no rows in result set" {
				w.Header().Set("HX-Retarget", "#signup-error")
				w.Header().Set("HX-Reswap", "innerHTML")
				w.Header().Set("HX-Trigger", "clear-passwords")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "Invalid email or invite token; please double check and try again")
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal server error")
			return
		}

		if inviteUsed == 1 {
			w.Header().Set("HX-Retarget", "#signup-error")
			w.Header().Set("HX-Reswap", "innerHTML")
			w.Header().Set("HX-Trigger", "clear-passwords")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Invalid email or invite token; please double check and try again")
			return
		}
	}

	var existingUserID int
	err = pool.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", email).Scan(&existingUserID)
	if err == nil {
		w.Header().Set("HX-Retarget", "#signup-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.Header().Set("HX-Trigger", "clear-passwords")
		w.WriteHeader(http.StatusOK)
		if inviteOnly {
			fmt.Fprint(w, "Invalid email or invite token; please double check and try again")
		} else {
			fmt.Fprint(w, "An account with this email already exists")
		}
		return
	} else if err.Error() != "no rows in result set" && !errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}

	defaultRoleID, err := storage.GetDefaultRoleID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "INSERT INTO users (email, password, role_id, timezone) VALUES ($1, $2, $3, $4)", email, string(hashedPassword), defaultRoleID, timezone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}

	if inviteOnly {
		_, err = tx.Exec(ctx, "UPDATE invites SET inviteused = 1 WHERE id = $1", inviteID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal server error")
			return
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	basePath := utils.GetBasePath()
	// Redirect to home with a flag so the home page can show a splash banner
	w.Header().Set("HX-Redirect", basePath+"/?account_created=true")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, " ")
}
