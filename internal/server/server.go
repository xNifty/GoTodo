package server

import (
	"GoTodo/internal/server/handlers"
	"GoTodo/internal/server/utils"
	"fmt"
	"net/http"
	"os"
)

// Literally just used to prevent favicon.ico from being requested
// TODO:: Add a favicon
func doNothing(w http.ResponseWriter, r *http.Request) {}

func StartServer() error {
	err := utils.InitializeTemplates()
	if err != nil {
		fmt.Println("Error initializing templates: ", err)
		return fmt.Errorf("failed to initialize templates: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)

	fs := http.FileServer(http.Dir("internal/server/public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/favicon.ico", doNothing)
	http.HandleFunc("/signup", handlers.SignupPageHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/api/signup", handlers.APISignup)
	http.HandleFunc("/api/login", handlers.APILogin)
	http.HandleFunc("/api/logout", handlers.APILogout)
	http.HandleFunc("/api/fetch-tasks", handlers.APIReturnTasks)
	http.HandleFunc("/partials/login", handlers.APIGetLoginPartial)
	http.HandleFunc("/api/add-task", handlers.APIAddTask)
	http.HandleFunc("/api/confirm", handlers.APIConfirmDelete)
	http.HandleFunc("/api/delete-task", handlers.APIDeleteTask)
	http.HandleFunc("/api/get-next-item", handlers.APIGetNextItem)
	http.HandleFunc("/api/update-status", handlers.APIUpdateTaskStatus)
	http.HandleFunc("/api/toggle-favorite", handlers.APIToggleFavorite)
	http.HandleFunc("/api/reorder-tasks", handlers.APIReorderTasks)
	http.HandleFunc("/about", handlers.AboutHandler)
	http.HandleFunc("/search", handlers.SearchHandler)

	// Profile routes
	http.HandleFunc("/profile", handlers.ProfilePage)
	http.HandleFunc("/api/update-timezone", handlers.APIUpdateTimezone)
	http.HandleFunc("/api/update-profile", handlers.APIUpdateProfile)

	// Invite routes
	http.HandleFunc("/createinvite", utils.RequirePermission("createinvites", handlers.CreateInvitePageHandler))
	http.HandleFunc("/api/create-invite", utils.RequirePermission("createinvites", handlers.APICreateInvite))
	http.HandleFunc("/api/invites", utils.RequirePermission("createinvites", handlers.APIGetInvites))
	http.HandleFunc("/api/confirm-invite-delete", utils.RequirePermission("createinvites", handlers.APIConfirmDeleteInvite))

	// Handle PUT and DELETE for invites with path parameters
	http.HandleFunc("/api/invite/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			utils.RequirePermission("createinvites", handlers.APIUpdateInvite)(w, r)
		case http.MethodDelete:
			utils.RequirePermission("createinvites", handlers.APIDeleteInvite)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Printf("Starting server on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}
