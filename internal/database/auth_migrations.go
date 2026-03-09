package database

import (
	"log"
)

// createUsersTable создаёт таблицу пользователей
func createUsersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		name TEXT NOT NULL,
		avatar TEXT DEFAULT '',
		role TEXT DEFAULT 'user',
		is_active INTEGER DEFAULT 1,
		is_banned INTEGER DEFAULT 0,
		email_verified INTEGER DEFAULT 0,
		last_login DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	`
	_, err := DB.Exec(query)
	return err
}

// createUserProfilesTable создаёт таблицу профилей
func createUserProfilesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS user_profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE NOT NULL,
		bio TEXT DEFAULT '',
		website TEXT DEFAULT '',
		location TEXT DEFAULT '',
		birth_date DATETIME,
		settings TEXT DEFAULT '{}',
		stats TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON user_profiles(user_id);
	`
	_, err := DB.Exec(query)
	return err
}

// createSessionsTable создаёт таблицу сессий
func createSessionsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		refresh_token TEXT NOT NULL,
		user_agent TEXT DEFAULT '',
		ip_address TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		last_active DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`
	_, err := DB.Exec(query)
	return err
}

// createEmailVerificationTokensTable создаёт таблицу токенов верификации
func createEmailVerificationTokensTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS email_verification_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		used INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_verification_tokens ON email_verification_tokens(token);
	`
	_, err := DB.Exec(query)
	return err
}

// createPasswordResetTokensTable создаёт таблицу токенов сброса пароля
func createPasswordResetTokensTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		used INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_reset_tokens ON password_reset_tokens(token);
	`
	_, err := DB.Exec(query)
	return err
}

// createRevokedTokensTable создаёт таблицу отозванных токенов
func createRevokedTokensTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS revoked_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL,
		revoked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		reason TEXT DEFAULT '',
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_revoked_tokens_user ON revoked_tokens(user_id);
	`
	_, err := DB.Exec(query)
	return err
}

// RunAuthMigrations запускает миграции для аутентификации
func RunAuthMigrations() error {
	log.Println("🔄 Running auth migrations...")

	migrations := []struct {
		name string
		fn   func() error
	}{
		{"users table", createUsersTable},
		{"user_profiles table", createUserProfilesTable},
		{"sessions table", createSessionsTable},
		{"email_verification_tokens table", createEmailVerificationTokensTable},
		{"password_reset_tokens table", createPasswordResetTokensTable},
		{"revoked_tokens table", createRevokedTokensTable},
	}

	for _, m := range migrations {
		log.Printf("  Creating %s...", m.name)
		if err := m.fn(); err != nil {
			log.Printf("  ❌ Failed to create %s: %v", m.name, err)
			return err
		}
		log.Printf("  ✅ Created %s", m.name)
	}

	log.Println("✅ Auth migrations completed")
	return nil
}
