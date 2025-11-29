package main

import (
	"GoTodo/internal/server"
	"GoTodo/internal/storage"
)

func main() {
	storage.CreateDatabase()
	storage.CreateUsersTable()
	storage.CreateRolesTable()
	storage.CreateInvitesTable()
	storage.MigrateInvitesTable() // Ensure inviteused column exists

	// The following is just for modifying columns during testing
	/**
	storage.AddColumns()
	storage.RemoveColumns()
	storage.MigrateTasksTable()
	*/

	server.StartServer()
}
