package database

import (
	"log"
)

// Migration представляет миграцию базы данных
type Migration struct {
	Version int
	Name    string
	Up      func() error
}

// migrations список миграций
var migrations = []Migration{
	{
		Version: 1,
		Name:    "initial_schema",
		Up: func() error {
			query := `
			CREATE TABLE IF NOT EXISTS players (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id TEXT UNIQUE NOT NULL,
				name TEXT DEFAULT 'Player',
				level INTEGER DEFAULT 1,
				experience INTEGER DEFAULT 0,
				go_knowledge INTEGER DEFAULT 0,
				focus INTEGER DEFAULT 100,
				willpower INTEGER DEFAULT 100,
				money INTEGER DEFAULT 0,
				dopamine INTEGER DEFAULT 100,
				play_time INTEGER DEFAULT 0,
				days_played INTEGER DEFAULT 1,
				current_day INTEGER DEFAULT 1,
				current_hour INTEGER DEFAULT 9,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
			
			CREATE TABLE IF NOT EXISTS skill_trees (
				user_id TEXT PRIMARY KEY,
				skill_points INTEGER DEFAULT 0,
				total_points INTEGER DEFAULT 0,
				FOREIGN KEY (user_id) REFERENCES players(user_id) ON DELETE CASCADE
			);
			
			CREATE TABLE IF NOT EXISTS skills (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				skill_id TEXT NOT NULL,
				level INTEGER DEFAULT 0,
				unlocked INTEGER DEFAULT 0,
				FOREIGN KEY (user_id) REFERENCES players(user_id) ON DELETE CASCADE,
				UNIQUE(user_id, skill_id)
			);
			
			CREATE TABLE IF NOT EXISTS daily_quests (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				quest_id TEXT NOT NULL,
				progress INTEGER DEFAULT 0,
				completed INTEGER DEFAULT 0,
				claimed INTEGER DEFAULT 0,
				day INTEGER DEFAULT 0,
				FOREIGN KEY (user_id) REFERENCES players(user_id) ON DELETE CASCADE,
				UNIQUE(user_id, quest_id)
			);
			
			CREATE TABLE IF NOT EXISTS achievements (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				achievement_id TEXT NOT NULL,
				unlocked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES players(user_id) ON DELETE CASCADE,
				UNIQUE(user_id, achievement_id)
			);
			
			CREATE INDEX IF NOT EXISTS idx_players_user_id ON players(user_id);
			CREATE INDEX IF NOT EXISTS idx_skills_user_id ON skills(user_id);
			CREATE INDEX IF NOT EXISTS idx_quests_user_id ON daily_quests(user_id);
			CREATE INDEX IF NOT EXISTS idx_achievements_user_id ON achievements(user_id);
			`
			_, err := DB.Exec(query)
			return err
		},
	},
	{
		Version: 2,
		Name:    "add_notification_settings",
		Up: func() error {
			query := `
			CREATE TABLE IF NOT EXISTS notification_settings (
				user_id TEXT PRIMARY KEY,
				enabled INTEGER DEFAULT 1,
				daily_reminder INTEGER DEFAULT 1,
				battle_reminder INTEGER DEFAULT 1,
				FOREIGN KEY (user_id) REFERENCES players(user_id) ON DELETE CASCADE
			);
			`
			_, err := DB.Exec(query)
			return err
		},
	},
	{
		Version: 3,
		Name:    "add_quiz_stats",
		Up: func() error {
			query := `
			ALTER TABLE players ADD COLUMN correct_answers INTEGER DEFAULT 0;
			ALTER TABLE players ADD COLUMN wrong_answers INTEGER DEFAULT 0;
			ALTER TABLE players ADD COLUMN asked_questions TEXT DEFAULT '[]';
			`
			_, err := DB.Exec(query)
			return err
		},
	},
}

// RunMigrations выполняет все миграции
func RunMigrations() error {
	// Создаём таблицу миграций
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		// Проверяем, применена ли миграция
		var exists int
		err := DB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.Version).Scan(&exists)
		if err != nil {
			return err
		}

		if exists > 0 {
			continue
		}

		log.Printf("🔄 Applying migration v%d: %s", m.Version, m.Name)

		// Выполняем миграцию
		if err := m.Up(); err != nil {
			return err
		}

		// Записываем в таблицу миграций
		_, err = DB.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", m.Version, m.Name)
		if err != nil {
			return err
		}

		log.Printf("✅ Migration v%d applied", m.Version)
	}

	return nil
}

// GetMigrationVersion возвращает текущую версию схемы
func GetMigrationVersion() (int, error) {
	var version int
	err := DB.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
