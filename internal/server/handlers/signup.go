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

	context := map[string]interface{}{
		"LoggedIn":  false,
		"UserEmail": email,
		"Title":     "GoTodo - Sign Up",
	}

	utils.RenderTemplate(w, "signup.html", context)
}

func APISignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
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

	if email == "" || token == "" || password == "" || confirmPassword == "" {
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

	var existingUserID int
	err = pool.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", email).Scan(&existingUserID)
	if err == nil {
		w.Header().Set("HX-Retarget", "#signup-error")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.Header().Set("HX-Trigger", "clear-passwords")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Invalid email or invite token; please double check and try again")
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

	_, err = tx.Exec(ctx, "INSERT INTO users (email, password, role_id) VALUES ($1, $2, $3)", email, string(hashedPassword), defaultRoleID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}

	_, err = tx.Exec(ctx, "UPDATE invites SET inviteused = 1 WHERE id = $1", inviteID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	basePath := utils.GetBasePath()
	w.Header().Set("HX-Redirect", basePath+"/")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, " ")
}
