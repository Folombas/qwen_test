package database

import "log"

// createAdminLogsTable создаёт таблицу логов администратора
func createAdminLogsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS admin_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		action TEXT NOT NULL,
		details TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_admin_logs_user ON admin_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_admin_logs_action ON admin_logs(action);
	CREATE INDEX IF NOT EXISTS idx_admin_logs_created ON admin_logs(created_at);
	`
	_, err := DB.Exec(query)
	return err
}

// createPlayerStatsTable создаёт таблицу статистики игроков
func createPlayerStatsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS player_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE NOT NULL,
		total_exp INTEGER DEFAULT 0,
		level INTEGER DEFAULT 1,
		correct_answers INTEGER DEFAULT 0,
		wrong_answers INTEGER DEFAULT 0,
		go_knowledge INTEGER DEFAULT 0,
		focus INTEGER DEFAULT 100,
		willpower INTEGER DEFAULT 100,
		total_play_time INTEGER DEFAULT 0,
		days_played INTEGER DEFAULT 1,
		current_streak INTEGER DEFAULT 0,
		max_streak INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_player_stats_user ON player_stats(user_id);
	CREATE INDEX IF NOT EXISTS idx_player_stats_level ON player_stats(level);
	`
	_, err := DB.Exec(query)
	return err
}

// RunAdminMigrations запускает миграции для админ-панели
func RunAdminMigrations() error {
	log.Println("🔄 Running admin migrations...")

	migrations := []struct {
		name string
		fn   func() error
	}{
		{"admin_logs table", createAdminLogsTable},
		{"player_stats table", createPlayerStatsTable},
	}

	for _, m := range migrations {
		log.Printf("  Creating %s...", m.name)
		if err := m.fn(); err != nil {
			log.Printf("  ❌ Failed to create %s: %v", m.name, err)
			return err
		}
		log.Printf("  ✅ Created %s", m.name)
	}

	log.Println("✅ Admin migrations completed")
	return nil
}
