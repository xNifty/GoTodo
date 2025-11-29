package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	RED   = "\033[31m"
	GREEN = "\033[32m"
	RESET = "\033[0m"
)

func OpenDatabase() (*pgxpool.Pool, error) {
	// Try to load .env, but don't crash if it's not there.
	// In Cloud Run, there won't be a .env file at all.
	_ = godotenv.Load()

	required := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
	}

	config := make(map[string]string)
	//missing := []string{}

	for _, key := range required {
		val := os.Getenv(key)
		if val == "" {
			//missing = append(missing, key)
			log.Fatalf("missing env variables: %v", key)
		} else {
			config[key] = val
		}
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config["DB_USER"],
		config["DB_PASSWORD"],
		config["DB_HOST"],
		config["DB_PORT"],
		config["DB_NAME"],
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	return pool, nil
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

func CreateTable(tableName string, columns []string) error {
	pool, err := OpenDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer CloseDatabase(pool)

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columns, ", "))

	_, err = pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", tableName, err)
	}

	fmt.Printf("Table %s created successfully\n", tableName)
	return nil
}

// CreateUsersTable creates the users table with predefined columns
func CreateUsersTable() error {
	columns := []string{
		"id SERIAL PRIMARY KEY",
		"email VARCHAR(255) UNIQUE NOT NULL",
		"password VARCHAR(255) NOT NULL",
		"role_id INTEGER NOT NULL",
		"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}

	return CreateTable("users", columns)
}

func CreateInvitesTable() error {
	columns := []string{
		"id SERIAL PRIMARY KEY",
		"email VARCHAR(255) UNIQUE NOT NULL",
		"token VARCHAR(255) UNIQUE NOT NULL",
		"inviteused INTEGER DEFAULT 0",
	}
	return CreateTable("invites", columns)
}

// MigrateInvitesTable adds the inviteused column if it doesn't exist
func MigrateInvitesTable() error {
	pool, err := OpenDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer CloseDatabase(pool)

	// Add inviteused column if it doesn't exist
	_, err = pool.Exec(context.Background(), "ALTER TABLE invites ADD COLUMN IF NOT EXISTS inviteused INTEGER DEFAULT 0")
	if err != nil {
		return fmt.Errorf("failed to add inviteused column to invites table: %v", err)
	}
	return nil
}

func CreateRolesTable() error {
	columns := []string{
		"id SERIAL PRIMARY KEY",
		"name VARCHAR(50) UNIQUE NOT NULL",
		"permissions TEXT[] NOT NULL",
	}
	return CreateTable("roles", columns)
}

func CreateTasksTable() error {
	columns := []string{
		"id SERIAL PRIMARY KEY",
		"title TEXT NOT NULL",
		"description TEXT",
		"completed BOOLEAN DEFAULT FALSE",
		"time_stamp TIMESTAMP DEFAULT NOW()",
		"user_id INTEGER",
	}
	return CreateTable("tasks", columns)
}

// MigrateTasksTable adds a user_id column and a foreign key constraint to the tasks table
func MigrateTasksTable() error {
	pool, err := OpenDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer CloseDatabase(pool)

	// Add user_id column
	_, err = pool.Exec(context.Background(), "ALTER TABLE tasks ADD COLUMN IF NOT EXISTS user_id INTEGER")
	if err != nil {
		return fmt.Errorf("failed to add user_id column to tasks table: %v", err)
	}

	// Add foreign key constraint
	_, err = pool.Exec(context.Background(), "ALTER TABLE tasks ADD CONSTRAINT fk_tasks_users FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE")
	if err != nil {
		return fmt.Errorf("failed to add foreign key constraint to tasks table: %v", err)
	}
	return nil
}

type User struct {
	ID       int
	Email    string
	Password string
}

func GetUserByEmail(email string) (*User, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var user User
	err = pool.QueryRow(context.Background(), "SELECT id, email, password FROM users WHERE email=$1", email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetPermissionsByRoleID(roleID int) ([]string, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var permissions []string
	err = pool.QueryRow(context.Background(), "SELECT permissions FROM roles WHERE id=$1", roleID).Scan(&permissions)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func GetDefaultRoleID() (int, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return 0, err
	}
	defer CloseDatabase(pool)

	var roleID int
	err = pool.QueryRow(context.Background(), "SELECT id FROM roles WHERE name = 'user'").Scan(&roleID)
	if err != nil {
		return 1, nil
	}
	return roleID, nil
}
