package server

import (
	"GoTodo/internal/server/handlers"
	"GoTodo/internal/server/utils"
	"GoTodo/internal/storage"
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

	// Initialize Redis client for rate limiting (optional)
	if err := utils.InitRedis(); err != nil {
		fmt.Printf("Warning: Redis init failed: %v\n", err)
	}

	// Run DB migrations (create tables / add columns as needed)
	if err := storage.RunMigrations(); err != nil {
		fmt.Printf("Warning: migrations completed with errors: %v\n", err)
	}

	// Preload changelog from GitHub at startup to avoid runtime API calls
	if err := handlers.PreloadChangelog(); err != nil {
		fmt.Printf("Warning: Preloading changelog failed: %v\n", err)
	}

	fs := http.FileServer(http.Dir("internal/server/public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/favicon.ico", doNothing)
	http.HandleFunc("/signup", handlers.SignupPageHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	// Apply Redis-backed rate limiting to sensitive endpoints
	http.HandleFunc("/api/signup", utils.RateLimitMiddleware(5, 0.05, 900, utils.KeyByIP)(handlers.APISignup))
	http.HandleFunc("/api/login", utils.RateLimitMiddleware(10, 1.0, 60, utils.KeyByIP)(handlers.APILogin))
	http.HandleFunc("/api/logout", handlers.APILogout)
	http.HandleFunc("/api/fetch-tasks", handlers.APIReturnTasks)
	http.HandleFunc("/partials/login", handlers.APIGetLoginPartial)
	http.HandleFunc("/api/add-task", utils.RateLimitMiddleware(60, 1.0, 60, utils.KeyByUser)(handlers.APIAddTask))
	http.HandleFunc("/api/edit", handlers.APIEditTaskForm)
	http.HandleFunc("/api/edit-task", utils.RateLimitMiddleware(60, 1.0, 60, utils.KeyByUser)(handlers.APIEditTask))
	http.HandleFunc("/api/confirm", handlers.APIConfirmDelete)
	http.HandleFunc("/api/delete-task", utils.RateLimitMiddleware(60, 1.0, 60, utils.KeyByUser)(handlers.APIDeleteTask))
	http.HandleFunc("/api/get-next-item", handlers.APIGetNextItem)
	http.HandleFunc("/api/update-status", handlers.APIUpdateTaskStatus)
	http.HandleFunc("/api/toggle-favorite", handlers.APIToggleFavorite)
	http.HandleFunc("/api/reorder-tasks", handlers.APIReorderTasks)
	http.HandleFunc("/about", handlers.AboutHandler)
	http.HandleFunc("/changelog", handlers.ChangelogHandler)
	http.HandleFunc("/changelog/page", handlers.ChangelogPageHandler)
	http.HandleFunc("/search", handlers.SearchHandler)

	// Projects management (simple CRUD for user-owned projects)
	http.HandleFunc("/projects", utils.RequireAuth(handlers.ProjectsPageHandler))
	http.HandleFunc("/api/projects/create", utils.RequireAuth(handlers.APICreateProject))
	http.HandleFunc("/api/projects/delete", utils.RequireAuth(handlers.APIDeleteProject))
	http.HandleFunc("/api/projects/json", utils.RequireAuth(handlers.APIProjectsJSON))

	// Profile routes
	http.HandleFunc("/profile", handlers.ProfilePage)
	http.HandleFunc("/api/update-timezone", handlers.APIUpdateTimezone)
	http.HandleFunc("/api/update-profile", handlers.APIUpdateProfile)

	// Invite routes
	http.HandleFunc("/createinvite", utils.RequirePermission("createinvites", handlers.CreateInvitePageHandler))
	http.HandleFunc("/api/create-invite", utils.RequirePermission("createinvites", handlers.APICreateInvite))
	http.HandleFunc("/api/invites", utils.RequirePermission("createinvites", handlers.APIGetInvites))
	http.HandleFunc("/api/confirm-invite-delete", utils.RequirePermission("createinvites", handlers.APIConfirmDeleteInvite))

	// Ban/unban user actions (admin only)
	http.HandleFunc("/api/ban-user", utils.RequirePermission("createinvites", handlers.APIBanUser))
	http.HandleFunc("/api/unban-user", utils.RequirePermission("createinvites", handlers.APIUnbanUser))

	// Admin panel (admin permission required). Register both /admin and /admin/ to handle trailing slash.
	http.HandleFunc("/admin", utils.RequirePermission("admin", handlers.AdminPageHandler))
	http.HandleFunc("/admin/", utils.RequirePermission("admin", handlers.AdminPageHandler))
	http.HandleFunc("/api/admin/update-settings", utils.RequirePermission("admin", handlers.APIUpdateSiteSettings))

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
