package storage

import (
	"context"
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
	// Ensure projects table exists
	if err := CreateProjectsTable(); err != nil {
		fmt.Printf("migration: CreateProjectsTable failed: %v\n", err)
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
	// Add project_id column to tasks (nullable)
	if err := MigrateTasksAddProjectID(); err != nil {
		fmt.Printf("migration: MigrateTasksAddProjectID failed: %v\n", err)
		errCount++
	}
	// Add date_modified column to tasks
	if err := MigrateTasksAddDateModified(); err != nil {
		fmt.Printf("migration: MigrateTasksAddDateModified failed: %v\n", err)
		errCount++
	}
	// Add due_date column to tasks
	if err := MigrateTasksAddDueDate(); err != nil {
		fmt.Printf("migration: MigrateTasksAddDueDate failed: %v\n", err)
		errCount++
	}

	// Ensure site_settings table exists
	if err := CreateSiteSettingsTable(); err != nil {
		fmt.Printf("migration: CreateSiteSettingsTable failed: %v\n", err)
		errCount++
	}
	if err := MigrateSiteSettingsAddRegistrationOptions(); err != nil {
		fmt.Printf("migration: MigrateSiteSettingsAddRegistrationOptions failed: %v\n", err)
		errCount++
	}

	// Ensure password_reset table exists
	if err := CreatePasswordResetTable(); err != nil {
		fmt.Printf("migration: CreatePasswordResetTable failed: %v\n", err)
		errCount++
	}

	// Ensure 'admin' permission exists on the admin role
	if err := MigrateEnsureAdminPermission(); err != nil {
		fmt.Printf("migration: MigrateEnsureAdminPermission failed: %v\n", err)
		errCount++
	}

	if errCount == 0 {
		return nil
	}
	return fmt.Errorf("migrations completed with %d errors (see logs)", errCount)
}

// MigrateEnsureAdminPermission ensures the roles table contains an 'admin' role
// and that its permissions array includes the "admin" permission.
func MigrateEnsureAdminPermission() error {
	pool, err := OpenDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer CloseDatabase(pool)

	// Required permissions for admin role
	required := []string{"add", "edit", "delete", "viewall", "createinvites", "admin"}

	// Try to read existing permissions for role 'admin'
	var perms []string
	err = pool.QueryRow(context.Background(), "SELECT permissions FROM roles WHERE name = 'admin'").Scan(&perms)
	if err != nil {
		// Role doesn't exist (or other error) — attempt to create it with full permission set
		_, insErr := pool.Exec(context.Background(), "INSERT INTO roles (name, permissions) VALUES ($1, $2)", "admin", required)
		if insErr != nil {
			return fmt.Errorf("failed to create admin role: %v (scan error: %v)", insErr, err)
		}
		return nil
	}

	// Compute missing permissions and append them individually to avoid duplicates
	have := map[string]bool{}
	for _, p := range perms {
		have[p] = true
	}
	missing := make([]string, 0)
	for _, r := range required {
		if !have[r] {
			missing = append(missing, r)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	for _, m := range missing {
		// Append only if not present (WHERE clause protects against duplicates)
		_, err = pool.Exec(context.Background(), "UPDATE roles SET permissions = array_append(permissions, $1) WHERE name = 'admin' AND NOT (permissions @> $2)", m, []string{m})
		if err != nil {
			return fmt.Errorf("failed to append permission %s: %v", m, err)
		}
	}
	return nil
}
