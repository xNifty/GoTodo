package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"os"
)

const (
	RED   = "\033[31m"
	GREEN = "\033[32m"
	RESET = "\033[0m"
)

func OpenDatabase() (*pgxpool.Pool, error) {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return pool, err
}

func CloseDatabase(pool *pgxpool.Pool) {
	pool.Close()
}

// This will solely add new columns we need later down the line..it's dumb, but this is how I'm handling it for now
func AddColumns() (bool, error) {
	pool, err := OpenDatabase()
	defer CloseDatabase(pool)

	_, err = pool.Exec(context.Background(), "ALTER TABLE tasks ADD COLUMN IF NOT EXISTS time_stamp TIMESTAMP default NOW()")
	if err != nil {
		log.Printf("Error in AddColumns: %v\n", err)
		return false, err
	}

	return true, nil
}

func RemoveColumns() (bool, error) {
	pool, err := OpenDatabase()
	defer CloseDatabase(pool)

	_, err = pool.Exec(context.Background(), "ALTER TABLE tasks DROP COLUMN IF EXISTS time_stamp")
	if err != nil {
		log.Printf("Error in RemoveColumns: %v\n", err)
		return false, err
	}

	return true, nil
}

func CreateDatabase() {
	pool, err := OpenDatabase()
	defer CloseDatabase(pool)

	_, err = pool.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS tasks (id SERIAL PRIMARY KEY, title TEXT, description TEXT, completed BOOLEAN DEFAULT FALSE)")

	if err != nil {
		log.Fatalf("Unable to create table: %v\n", err)
	} else {
		fmt.Println("Database connection appears to be " + GREEN + "successful" + RESET + ".")
	}
}

func GetNextID() int {
	pool, err := OpenDatabase()
	defer CloseDatabase(pool)

	var nextID int
	err = pool.QueryRow(context.Background(), "SELECT COALESCE(MAX(id), 0) FROM tasks").Scan(&nextID)

	if err != nil {
		log.Printf("Error in GetNextID: %v\n", err)
		return 1
	}

	return nextID + 1
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
		pool, err := OpenDatabase()
		defer CloseDatabase(pool)

		_, err = pool.Exec(context.Background(), "DELETE FROM tasks")
		if err != nil {
			log.Printf("Error in DeleteAllTasks: %v\n", err)
		} else {
			fmt.Println("All tasks deleted successfully!")
		}
	} else {
		fmt.Println("Deletion cancelled.")
	}
}
