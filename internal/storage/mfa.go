package storage

import (
	"context"
	"fmt"
)

// MFAState is the persisted MFA configuration for a user.
type MFAState struct {
	Enabled         bool
	EncryptedSecret string
	LastStep        int64
}

// RecoveryCodeHash is an unused (or any) stored recovery-code bcrypt hash.
type RecoveryCodeHash struct {
	ID   int
	Hash string
}

// MigrateUsersAddMFA adds TOTP MFA columns to users.
func MigrateUsersAddMFA() error {
	pool, err := OpenDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer CloseDatabase(pool)

	_, err = pool.Exec(context.Background(),
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN DEFAULT FALSE")
	if err != nil {
		return fmt.Errorf("failed to add mfa_enabled: %v", err)
	}
	_, err = pool.Exec(context.Background(),
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_secret TEXT DEFAULT ''")
	if err != nil {
		return fmt.Errorf("failed to add mfa_secret: %v", err)
	}
	_, err = pool.Exec(context.Background(),
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_last_step BIGINT DEFAULT 0")
	if err != nil {
		return fmt.Errorf("failed to add mfa_last_step: %v", err)
	}
	return nil
}

// CreateMFARecoveryCodesTable ensures the recovery-codes table exists.
func CreateMFARecoveryCodesTable() error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS user_mfa_recovery_codes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			code_hash TEXT NOT NULL,
			used_at TIMESTAMPTZ
		)`)
	if err != nil {
		return fmt.Errorf("failed to create user_mfa_recovery_codes: %v", err)
	}
	_, err = pool.Exec(context.Background(),
		`CREATE INDEX IF NOT EXISTS idx_user_mfa_recovery_codes_user_id ON user_mfa_recovery_codes(user_id)`)
	if err != nil {
		return fmt.Errorf("failed to create user_mfa_recovery_codes index: %v", err)
	}
	return nil
}

// UserHasMFA reports whether MFA is enabled for the user.
func UserHasMFA(userID int) (bool, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return false, err
	}
	defer CloseDatabase(pool)

	var enabled bool
	err = pool.QueryRow(context.Background(),
		`SELECT COALESCE(mfa_enabled, false) FROM users WHERE id = $1`, userID).Scan(&enabled)
	return enabled, err
}

// GetMFAState loads MFA columns for a user.
func GetMFAState(userID int) (*MFAState, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	var s MFAState
	err = pool.QueryRow(context.Background(), `
		SELECT COALESCE(mfa_enabled, false), COALESCE(mfa_secret, ''), COALESCE(mfa_last_step, 0)
		FROM users WHERE id = $1`, userID).Scan(&s.Enabled, &s.EncryptedSecret, &s.LastStep)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// EnableUserMFA stores the encrypted TOTP secret, marks MFA on, and replaces recovery hashes.
func EnableUserMFA(userID int, encryptedSecret string, codeHashes []string) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE users SET mfa_enabled = TRUE, mfa_secret = $1, mfa_last_step = 0
		WHERE id = $2`, encryptedSecret, userID)
	if err != nil {
		return fmt.Errorf("enable mfa: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM user_mfa_recovery_codes WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear recovery codes: %w", err)
	}
	for _, hash := range codeHashes {
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_mfa_recovery_codes (user_id, code_hash) VALUES ($1, $2)`,
			userID, hash); err != nil {
			return fmt.Errorf("insert recovery code: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// DisableUserMFA clears MFA secret, last step, and recovery codes.
func DisableUserMFA(userID int) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE users SET mfa_enabled = FALSE, mfa_secret = '', mfa_last_step = 0
		WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("disable mfa: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM user_mfa_recovery_codes WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete recovery codes: %w", err)
	}
	return tx.Commit(ctx)
}

// ReplaceRecoveryCodeHashes deletes existing codes and inserts the given hashes.
func ReplaceRecoveryCodeHashes(userID int, codeHashes []string) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM user_mfa_recovery_codes WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear recovery codes: %w", err)
	}
	for _, hash := range codeHashes {
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_mfa_recovery_codes (user_id, code_hash) VALUES ($1, $2)`,
			userID, hash); err != nil {
			return fmt.Errorf("insert recovery code: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// UpdateMFALastStep records the last accepted TOTP timestep.
func UpdateMFALastStep(userID int, step int64) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)
	_, err = pool.Exec(context.Background(),
		`UPDATE users SET mfa_last_step = $1 WHERE id = $2`, step, userID)
	return err
}

// ListUnusedRecoveryHashes returns unused recovery-code hashes for a user.
func ListUnusedRecoveryHashes(userID int) ([]RecoveryCodeHash, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer CloseDatabase(pool)

	rows, err := pool.Query(context.Background(), `
		SELECT id, code_hash FROM user_mfa_recovery_codes
		WHERE user_id = $1 AND used_at IS NULL
		ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RecoveryCodeHash
	for rows.Next() {
		var h RecoveryCodeHash
		if err := rows.Scan(&h.ID, &h.Hash); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// MarkRecoveryCodeUsed sets used_at on a recovery code.
func MarkRecoveryCodeUsed(id int) error {
	pool, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer CloseDatabase(pool)
	_, err = pool.Exec(context.Background(),
		`UPDATE user_mfa_recovery_codes SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`, id)
	return err
}

// CountUnusedRecoveryCodes returns how many unused recovery codes remain.
func CountUnusedRecoveryCodes(userID int) (int, error) {
	pool, err := OpenDatabase()
	if err != nil {
		return 0, err
	}
	defer CloseDatabase(pool)

	var n int
	err = pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM user_mfa_recovery_codes
		WHERE user_id = $1 AND used_at IS NULL`, userID).Scan(&n)
	return n, err
}
