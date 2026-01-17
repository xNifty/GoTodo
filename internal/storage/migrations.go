package storage

import (
	"fmt"
)

// RunMigrations attempts to create required tables and apply non-destructive migrations.
// It continues on errors, but returns an aggregated error if all operations fail.
func RunMigrations() error {
	errCount := 0

	if err := CreateUsersTable(); err != nil {
		fmt.Printf("migration: CreateUsersTable failed: %v\n", err)
		errCount++
	}
	if err := CreateRolesTable(); err != nil {
		fmt.Printf("migration: CreateRolesTable failed: %v\n", err)
		errCount++
	}
	if err := CreateInvitesTable(); err != nil {
		fmt.Printf("migration: CreateInvitesTable failed: %v\n", err)
		errCount++
	}
	if err := CreateTasksTable(); err != nil {
		fmt.Printf("migration: CreateTasksTable failed: %v\n", err)
		errCount++
	}

	// Non-breaking column migrations
	if err := MigrateUsersAddTimezone(); err != nil {
		fmt.Printf("migration: MigrateUsersAddTimezone failed: %v\n", err)
		errCount++
	}
	if err := MigrateUsersAddName(); err != nil {
		fmt.Printf("migration: MigrateUsersAddName failed: %v\n", err)
		errCount++
	}
	if err := MigrateUsersAddItemsPerPage(); err != nil {
		fmt.Printf("migration: MigrateUsersAddItemsPerPage failed: %v\n", err)
		errCount++
	}
	if err := MigrateUsersAddIsBanned(); err != nil {
		fmt.Printf("migration: MigrateUsersAddIsBanned failed: %v\n", err)
		errCount++
	}
	if err := MigrateTasksAddIsFavorite(); err != nil {
		fmt.Printf("migration: MigrateTasksAddIsFavorite failed: %v\n", err)
		errCount++
	}
	if err := MigrateTasksAddPosition(); err != nil {
		fmt.Printf("migration: MigrateTasksAddPosition failed: %v\n", err)
		errCount++
	}

	// Ensure site_settings table exists
	if err := CreateSiteSettingsTable(); err != nil {
		fmt.Printf("migration: CreateSiteSettingsTable failed: %v\n", err)
		errCount++
	}

	if errCount == 0 {
		return nil
	}
	return fmt.Errorf("migrations completed with %d errors (see logs)", errCount)
}
