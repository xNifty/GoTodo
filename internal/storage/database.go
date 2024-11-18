package storage

import (
        "database/sql"
        _ "github.com/mattn/go-sqlite3" 
        "fmt"
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

        _, err := db.Exec("CREATE TABLE tasks (id INTEGER PRIMARY KEY, name TEXT, description TEXT, completed INTEGER)")
        if err != nil {
                return
        } else {
                fmt.Println("Database created successfully")
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
