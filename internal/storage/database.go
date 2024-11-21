package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func OpenDatebase() *sql.DB {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return db
}

func CloseDatabase(db *sql.DB) {
	db.Close()
}

func CreateDatabase() {
	db := OpenDatebase()
	defer CloseDatabase(db)

	_, err := db.Exec("CREATE TABLE tasks (id INTEGER PRIMARY KEY, title TEXT, description TEXT, completed INTEGER)")
	if err != nil {
		return
	} else {
		fmt.Println("Database created successfully")
	}
}

func GetNextID() int {
	db := OpenDatebase()
	defer db.Close()

	var id int

	err := db.QueryRow("SELECT MAX(id) FROM tasks").Scan(&id)
	if err != nil {
		return 1
	} else {
		return id + 1
	}
}

func DeleteDatabase() {
	var confirm bool = false

	if confirm {
		err := os.Remove("./tasks.db")
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println("Database deleted successfully")
		}
	} else {
		fmt.Println("Database deletion cancelled")
	}
}

func DeleteAllTasks() {
	fmt.Print("\nAre you sure you want to delete all tasks? (y/n): ")
	var confirm string
	_, err := fmt.Scanln(&confirm)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}
	if confirm == "y" {
		db := OpenDatebase()
		defer db.Close()

		_, err := db.Exec("DELETE FROM tasks")
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println("All tasks deleted successfully!")
		}
	} else {
		fmt.Println("Deletion cancelled.")
	}
}
