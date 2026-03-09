package admin

import (
	"database/sql"
	"errors"
	"time"
)

// AdminService сервис для админ-панели
type AdminService struct {
	db *sql.DB
}

// NewAdminService создаёт новый admin сервис
func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{
		db: db,
	}
}

// DashboardStats статистика для дашборда
type DashboardStats struct {
	TotalUsers       int64   `json:"total_users"`
	ActiveUsers      int64   `json:"active_users"`
	BannedUsers      int64   `json:"banned_users"`
	TotalQuestions   int64   `json:"total_questions"`
	TotalGames       int64   `json:"total_games"`
	AvgPlayTime      float64 `json:"avg_play_time"`
	NewUsersToday    int64   `json:"new_users_today"`
	NewUsersWeek     int64   `json:"new_users_week"`
	TopPlayers       []PlayerStats `json:"top_players"`
	RecentActivity   []ActivityLog `json:"recent_activity"`
}

// PlayerStats статистика игрока
type PlayerStats struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Level         int    `json:"level"`
	TotalEXP      int    `json:"total_exp"`
	CorrectAnswers int   `json:"correct_answers"`
	Rating        int    `json:"rating"`
	LastLogin     *time.Time `json:"last_login"`
}

// ActivityLog лог активности
type ActivityLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	UserName  string    `json:"user_name"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

// UserInfo информация о пользователе
type UserInfo struct {
	ID            int64      `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	Role          string     `json:"role"`
	IsActive      bool       `json:"is_active"`
	IsBanned      bool       `json:"is_banned"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLogin     *time.Time `json:"last_login"`
	Stats         UserStats  `json:"stats"`
}

// UserStats статистика пользователя
type UserStats struct {
	TotalEXP       int `json:"total_exp"`
	Level          int `json:"level"`
	CorrectAnswers int `json:"correct_answers"`
	WrongAnswers   int `json:"wrong_answers"`
	GoKnowledge    int `json:"go_knowledge"`
	TotalPlayTime  int `json:"total_play_time"`
	DaysPlayed     int `json:"days_played"`
}

// UpdateUserRequest запрос на обновление пользователя
type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
	IsBanned bool   `json:"is_banned"`
}

// BanUserRequest запрос на бан пользователя
type BanUserRequest struct {
	Reason string `json:"reason"`
}

// GetDashboardStats получает статистику для дашборда
func (s *AdminService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Общее количество пользователей
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	// Активные пользователи
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = 1").Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, err
	}

	// Забаненные пользователи
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_banned = 1").Scan(&stats.BannedUsers)
	if err != nil {
		return nil, err
	}

	// Новые пользователи сегодня
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE DATE(created_at) = DATE('now')
	`).Scan(&stats.NewUsersToday)
	if err != nil {
		return nil, err
	}

	// Новые пользователи за неделю
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE created_at >= DATE('now', '-7 days')
	`).Scan(&stats.NewUsersWeek)
	if err != nil {
		return nil, err
	}

	// Топ игроков
	rows, err := s.db.Query(`
		SELECT u.id, u.name, u.email, p.level, p.total_exp, p.correct_answers,
			   (p.go_knowledge * 10 + p.focus * 5 + p.willpower * 3 + p.level * 100) as rating,
			   u.last_login
		FROM users u
		JOIN player_stats p ON u.id = p.user_id
		ORDER BY rating DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var player PlayerStats
		err := rows.Scan(&player.ID, &player.Name, &player.Email, &player.Level,
			&player.TotalEXP, &player.CorrectAnswers, &player.Rating, &player.LastLogin)
		if err != nil {
			continue
		}
		stats.TopPlayers = append(stats.TopPlayers, player)
	}

	return stats, nil
}

// GetUsers получает список пользователей
func (s *AdminService) GetUsers(limit, offset int, search string) ([]UserInfo, int64, error) {
	// Общее количество
	var total int64
	searchPattern := "%" + search + "%"
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE name LIKE ? OR email LIKE ?
	`, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Получаем пользователей
	rows, err := s.db.Query(`
		SELECT id, email, name, role, is_active, is_banned, email_verified, 
			   created_at, last_login
		FROM users 
		WHERE name LIKE ? OR email LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]UserInfo, 0)
	for rows.Next() {
		var user UserInfo
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Role,
			&user.IsActive, &user.IsBanned, &user.EmailVerified,
			&user.CreatedAt, &user.LastLogin)
		if err != nil {
			continue
		}

		// Загружаем статистику
		err = s.db.QueryRow(`
			SELECT total_exp, level, correct_answers, wrong_answers, 
				   go_knowledge, total_play_time, days_played
			FROM player_stats WHERE user_id = ?
		`, user.ID).Scan(&user.Stats.TotalEXP, &user.Stats.Level,
			&user.Stats.CorrectAnswers, &user.Stats.WrongAnswers,
			&user.Stats.GoKnowledge, &user.Stats.TotalPlayTime,
			&user.Stats.DaysPlayed)

		if err != nil {
			// Статистика не обязательна
			user.Stats = UserStats{}
		}

		users = append(users, user)
	}

	return users, total, nil
}

// GetUserByID получает пользователя по ID
func (s *AdminService) GetUserByID(id int64) (*UserInfo, error) {
	var user UserInfo
	err := s.db.QueryRow(`
		SELECT id, email, name, role, is_active, is_banned, email_verified, 
			   created_at, last_login
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Email, &user.Name, &user.Role,
		&user.IsActive, &user.IsBanned, &user.EmailVerified,
		&user.CreatedAt, &user.LastLogin)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Загружаем статистику
	err = s.db.QueryRow(`
		SELECT total_exp, level, correct_answers, wrong_answers, 
			   go_knowledge, total_play_time, days_played
		FROM player_stats WHERE user_id = ?
	`, id).Scan(&user.Stats.TotalEXP, &user.Stats.Level,
		&user.Stats.CorrectAnswers, &user.Stats.WrongAnswers,
		&user.Stats.GoKnowledge, &user.Stats.TotalPlayTime,
		&user.Stats.DaysPlayed)

	if err != nil {
		user.Stats = UserStats{}
	}

	return &user, nil
}

// UpdateUser обновляет пользователя
func (s *AdminService) UpdateUser(id int64, req UpdateUserRequest) error {
	_, err := s.db.Exec(`
		UPDATE users 
		SET name = ?, email = ?, role = ?, is_active = ?, is_banned = ?, updated_at = ?
		WHERE id = ?
	`, req.Name, req.Email, req.Role, req.IsActive, req.IsBanned, time.Now(), id)
	return err
}

// BanUser банит пользователя
func (s *AdminService) BanUser(id int64, reason string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Баним пользователя
	_, err = tx.Exec(`
		UPDATE users SET is_banned = 1, is_active = 0, updated_at = ? WHERE id = ?
	`, time.Now(), id)
	if err != nil {
		return err
	}

	// Логируем бан
	_, err = tx.Exec(`
		INSERT INTO admin_logs (user_id, action, details, created_at)
		VALUES (?, 'ban', ?, ?)
	`, id, reason, time.Now())
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UnbanUser разбанивает пользователя
func (s *AdminService) UnbanUser(id int64) error {
	_, err := s.db.Exec(`
		UPDATE users SET is_banned = 0, updated_at = ? WHERE id = ?
	`, time.Now(), id)
	return err
}

// DeleteUser удаляет пользователя
func (s *AdminService) DeleteUser(id int64) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// GetRecentActivity получает последнюю активность
func (s *AdminService) GetRecentActivity(limit int) ([]ActivityLog, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, action, details, created_at
		FROM admin_logs
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activity := make([]ActivityLog, 0)
	for rows.Next() {
		var log ActivityLog
		err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.Details, &log.CreatedAt)
		if err != nil {
			continue
		}

		// Получаем имя пользователя
		err = s.db.QueryRow("SELECT name FROM users WHERE id = ?", log.UserID).Scan(&log.UserName)
		if err != nil {
			log.UserName = "Unknown"
		}

		activity = append(activity, log)
	}

	return activity, nil
}

// LogAction логирует действие админа
func (s *AdminService) LogAction(adminID int64, action, details string) error {
	_, err := s.db.Exec(`
		INSERT INTO admin_logs (user_id, action, details, created_at)
		VALUES (?, ?, ?, ?)
	`, adminID, action, details, time.Now())
	return err
}
