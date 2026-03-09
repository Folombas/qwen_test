package database

import "log"

// createFriendsTable создаёт таблицу друзей
func createFriendsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS friends (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		friend_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (friend_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, friend_id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_friends_user ON friends(user_id);
	CREATE INDEX IF NOT EXISTS idx_friends_friend ON friends(friend_id);
	`
	_, err := DB.Exec(query)
	return err
}

// createFriendRequestsTable создаёт таблицу запросов в друзья
func createFriendRequestsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS friend_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER NOT NULL,
		receiver_id INTEGER NOT NULL,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(sender_id, receiver_id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_requests_sender ON friend_requests(sender_id);
	CREATE INDEX IF NOT EXISTS idx_requests_receiver ON friend_requests(receiver_id);
	CREATE INDEX IF NOT EXISTS idx_requests_status ON friend_requests(status);
	`
	_, err := DB.Exec(query)
	return err
}

// createMessagesTable создаёт таблицу сообщений
func createMessagesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER NOT NULL,
		receiver_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		read INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
	CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages(receiver_id);
	CREATE INDEX IF NOT EXISTS idx_messages_read ON messages(read);
	`
	_, err := DB.Exec(query)
	return err
}

// createChallengesTable создаёт таблицу дуэлей
func createChallengesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS challenges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER NOT NULL,
		sender_name TEXT NOT NULL,
		receiver_id INTEGER NOT NULL,
		receiver_name TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		winner_id INTEGER,
		sender_score INTEGER DEFAULT 0,
		receiver_score INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		started_at DATETIME,
		completed_at DATETIME,
		FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_challenges_sender ON challenges(sender_id);
	CREATE INDEX IF NOT EXISTS idx_challenges_receiver ON challenges(receiver_id);
	CREATE INDEX IF NOT EXISTS idx_challenges_status ON challenges(status);
	`
	_, err := DB.Exec(query)
	return err
}

// createActivityTable создаёт таблицу активности
func createActivityTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS activity (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		user_name TEXT NOT NULL,
		action TEXT NOT NULL,
		description TEXT DEFAULT '',
		score INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_activity_user ON activity(user_id);
	CREATE INDEX IF NOT EXISTS idx_activity_action ON activity(action);
	CREATE INDEX IF NOT EXISTS idx_activity_created ON activity(created_at);
	`
	_, err := DB.Exec(query)
	return err
}

// RunSocialMigrations запускает миграции социальных функций
func RunSocialMigrations() error {
	log.Println("🔄 Running social migrations...")

	migrations := []struct {
		name string
		fn   func() error
	}{
		{"friends table", createFriendsTable},
		{"friend_requests table", createFriendRequestsTable},
		{"messages table", createMessagesTable},
		{"challenges table", createChallengesTable},
		{"activity table", createActivityTable},
	}

	for _, m := range migrations {
		log.Printf("  Creating %s...", m.name)
		if err := m.fn(); err != nil {
			log.Printf("  ❌ Failed to create %s: %v", m.name, err)
			return err
		}
		log.Printf("  ✅ Created %s", m.name)
	}

	log.Println("✅ Social migrations completed")
	return nil
}
