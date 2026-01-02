package main

import (
	"GoTodo/internal/server"
	"GoTodo/internal/storage"
	"fmt"
)

func main() {
	fmt.Println("Application main function started.")
	storage.CreateDatabase()
	storage.CreateUsersTable()
	storage.CreateRolesTable()
	storage.CreateInvitesTable()
	storage.CreateTasksTable()
	storage.MigrateInvitesTable()     // Ensure inviteused column exists
	storage.MigrateUsersAddTimezone() // Ensure timezone column exists
	storage.MigrateUsersAddName()
	storage.MigrateUsersAddIsBanned()
	storage.MigrateUsersAddItemsPerPage()
	storage.MigrateTasksAddIsFavorite()
	storage.MigrateTasksAddPosition()

	// The following is just for modifying columns during testing
	/**
	storage.AddColumns()
	storage.RemoveColumns()
	storage.MigrateTasksTable()
	*/

	err := server.StartServer()
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
