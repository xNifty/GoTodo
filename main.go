package main

import (
	"GoTodo/internal/server"
	"GoTodo/internal/storage"
)

func main() {
	storage.CreateDatabase()

	// The following is just for modifying columns during testing
	/**
	storage.AddColumns()
	storage.RemoveColumns()
	storage.CreateUsersTable()
	*/

	server.StartServer()
}
