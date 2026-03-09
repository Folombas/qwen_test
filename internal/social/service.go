// Package social предоставляет социальные функции (друзья, чат, дуэли)
package social

import (
	"database/sql"
	"errors"
	"time"
)

// SocialService сервис социальных функций
type SocialService struct {
	db *sql.DB
}

// NewSocialService создаёт новый сервис
func NewSocialService(db *sql.DB) *SocialService {
	return &SocialService{
		db: db,
	}
}

// FriendRequest запрос в друзья
type FriendRequest struct {
	ID         int64     `json:"id"`
	SenderID   int64     `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	ReceiverID int64     `json:"receiver_id"`
	ReceiverName string  `json:"receiver_name"`
	Status     string    `json:"status"` // pending, accepted, rejected
	CreatedAt  time.Time `json:"created_at"`
}

// Friend друг пользователя
type Friend struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Avatar       string    `json:"avatar"`
	Level        int       `json:"level"`
	Rating       int       `json:"rating"`
	IsOnline     bool      `json:"is_online"`
	LastSeen     time.Time `json:"last_seen"`
	FriendSince  time.Time `json:"friend_since"`
}

// Message сообщение в чате
type Message struct {
	ID        int64     `json:"id"`
	SenderID  int64     `json:"sender_id"`
	ReceiverID int64    `json:"receiver_id"`
	Content   string    `json:"content"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// Challenge вызов на дуэль
type Challenge struct {
	ID          int64     `json:"id"`
	SenderID    int64     `json:"sender_id"`
	SenderName  string    `json:"sender_name"`
	ReceiverID  int64     `json:"receiver_id"`
	ReceiverName string  `json:"receiver_name"`
	Status      string    `json:"status"` // pending, accepted, rejected, completed
	WinnerID    int64     `json:"winner_id"`
	SenderScore int       `json:"sender_score"`
	ReceiverScore int     `json:"receiver_score"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// Activity действие в ленте
type Activity struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	UserName    string    `json:"user_name"`
	Action      string    `json:"action"` // level_up, achievement, quiz_complete, challenge_win
	Description string    `json:"description"`
	Score       int       `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
}

// === Friends ===

// SendFriendRequest отправить запрос в друзья
func (s *SocialService) SendFriendRequest(senderID, receiverID int64) error {
	if senderID == receiverID {
		return errors.New("cannot add yourself as friend")
	}

	// Проверяем существующие запросы
	var exists int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM friend_requests 
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
	`, senderID, receiverID, receiverID, senderID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("friend request already exists")
	}

	_, err = s.db.Exec(`
		INSERT INTO friend_requests (sender_id, receiver_id, status, created_at)
		VALUES (?, ?, 'pending', ?)
	`, senderID, receiverID, time.Now())
	return err
}

// AcceptFriendRequest принять запрос в друзья
func (s *SocialService) AcceptFriendRequest(requestID, userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновляем статус запроса
	_, err = tx.Exec(`
		UPDATE friend_requests SET status = 'accepted' WHERE id = ? AND receiver_id = ?
	`, requestID, userID)
	if err != nil {
		return err
	}

	// Получаем ID друга
	var friendID int64
	err = tx.QueryRow(`SELECT sender_id FROM friend_requests WHERE id = ?`, requestID).Scan(&friendID)
	if err != nil {
		return err
	}

	// Создаём запись в friends
	_, err = tx.Exec(`
		INSERT INTO friends (user_id, friend_id, created_at) VALUES (?, ?, ?)
	`, userID, friendID, time.Now())
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO friends (user_id, friend_id, created_at) VALUES (?, ?, ?)
	`, friendID, userID, time.Now())
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RejectFriendRequest отклонить запрос
func (s *SocialService) RejectFriendRequest(requestID, userID int64) error {
	_, err := s.db.Exec(`
		UPDATE friend_requests SET status = 'rejected' WHERE id = ? AND receiver_id = ?
	`, requestID, userID)
	return err
}

// GetFriendRequests получить входящие запросы
func (s *SocialService) GetFriendRequests(userID int64) ([]FriendRequest, error) {
	rows, err := s.db.Query(`
		SELECT fr.id, fr.sender_id, u.name, fr.receiver_id, '', fr.status, fr.created_at
		FROM friend_requests fr
		JOIN users u ON fr.sender_id = u.id
		WHERE fr.receiver_id = ? AND fr.status = 'pending'
		ORDER BY fr.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]FriendRequest, 0)
	for rows.Next() {
		var req FriendRequest
		err := rows.Scan(&req.ID, &req.SenderID, &req.SenderName, &req.ReceiverID, 
			&req.ReceiverName, &req.Status, &req.CreatedAt)
		if err != nil {
			continue
		}
		requests = append(requests, req)
	}
	return requests, nil
}

// GetFriends получить список друзей
func (s *SocialService) GetFriends(userID int64) ([]Friend, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.name, u.avatar, 
			   (SELECT level FROM player_stats WHERE user_id = u.id) as level,
			   (SELECT go_knowledge * 10 + focus * 5 + willpower * 3 + level * 100 
			    FROM player_stats WHERE user_id = u.id) as rating,
			   u.last_login, f.created_at
		FROM friends f
		JOIN users u ON f.friend_id = u.id
		WHERE f.user_id = ?
		ORDER BY rating DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	friends := make([]Friend, 0)
	now := time.Now()
	for rows.Next() {
		var friend Friend
		var lastSeen sql.NullTime
		err := rows.Scan(&friend.ID, &friend.Name, &friend.Avatar, &friend.Level, 
			&friend.Rating, &lastSeen, &friend.FriendSince)
		if err != nil {
			continue
		}
		friend.LastSeen = lastSeen.Time
		// Считаем онлайн если был в последние 5 минут
		friend.IsOnline = now.Sub(friend.LastSeen) < 5*time.Minute
		friends = append(friends, friend)
	}
	return friends, nil
}

// RemoveFriend удалить друга
func (s *SocialService) RemoveFriend(userID, friendID int64) error {
	_, err := s.db.Exec(`DELETE FROM friends WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)`,
		userID, friendID, friendID, userID)
	return err
}

// === Messages ===

// SendMessage отправить сообщение
func (s *SocialService) SendMessage(senderID, receiverID int64, content string) (*Message, error) {
	result, err := s.db.Exec(`
		INSERT INTO messages (sender_id, receiver_id, content, created_at)
		VALUES (?, ?, ?, ?)
	`, senderID, receiverID, content, time.Now())
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Message{
		ID:         id,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  time.Now(),
	}, nil
}

// GetMessages получить переписку
func (s *SocialService) GetMessages(userID, friendID int64, limit int) ([]Message, error) {
	rows, err := s.db.Query(`
		SELECT id, sender_id, receiver_id, content, read, created_at
		FROM messages
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at DESC
		LIMIT ?
	`, userID, friendID, friendID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.Read, &msg.CreatedAt)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// MarkMessagesRead отметить сообщения как прочитанные
func (s *SocialService) MarkMessagesRead(userID, fromID int64) error {
	_, err := s.db.Exec(`
		UPDATE messages SET read = 1 WHERE receiver_id = ? AND sender_id = ? AND read = 0
	`, userID, fromID)
	return err
}

// GetUnreadCount получить количество непрочитанных
func (s *SocialService) GetUnreadCount(userID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM messages WHERE receiver_id = ? AND read = 0`, userID).Scan(&count)
	return count, err
}

// === Challenges ===

// SendChallenge отправить вызов на дуэль
func (s *SocialService) SendChallenge(senderID, receiverID int64, senderName, receiverName string) error {
	_, err := s.db.Exec(`
		INSERT INTO challenges (sender_id, sender_name, receiver_id, receiver_name, status, created_at)
		VALUES (?, ?, ?, ?, 'pending', ?)
	`, senderID, senderName, receiverID, receiverName, time.Now())
	return err
}

// AcceptChallenge принять вызов
func (s *SocialService) AcceptChallenge(challengeID, userID int64) error {
	now := time.Now()
	_, err := s.db.Exec(`
		UPDATE challenges SET status = 'accepted', started_at = ? WHERE id = ? AND receiver_id = ?
	`, now, challengeID, userID)
	return err
}

// CompleteChallenge завершить дуэль
func (s *SocialService) CompleteChallenge(challengeID, winnerID, senderScore, receiverScore int64) error {
	now := time.Now()
	_, err := s.db.Exec(`
		UPDATE challenges 
		SET status = 'completed', winner_id = ?, sender_score = ?, receiver_score = ?, completed_at = ?
		WHERE id = ?
	`, winnerID, senderScore, receiverScore, now, challengeID)
	return err
}

// GetChallenges получить активные вызовы
func (s *SocialService) GetChallenges(userID int64, status string) ([]Challenge, error) {
	rows, err := s.db.Query(`
		SELECT id, sender_id, sender_name, receiver_id, receiver_name, status, 
			   winner_id, sender_score, receiver_score, created_at, started_at, completed_at
		FROM challenges
		WHERE (receiver_id = ? OR sender_id = ?) AND status = ?
		ORDER BY created_at DESC
	`, userID, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	challenges := make([]Challenge, 0)
	for rows.Next() {
		var ch Challenge
		var startedAt, completedAt sql.NullTime
		err := rows.Scan(&ch.ID, &ch.SenderID, &ch.SenderName, &ch.ReceiverID, &ch.ReceiverName,
			&ch.Status, &ch.WinnerID, &ch.SenderScore, &ch.ReceiverScore, 
			&ch.CreatedAt, &startedAt, &completedAt)
		if err != nil {
			continue
		}
		if startedAt.Valid {
			ch.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			ch.CompletedAt = &completedAt.Time
		}
		challenges = append(challenges, ch)
	}
	return challenges, nil
}

// === Activity ===

// LogActivity записать действие в ленту
func (s *SocialService) LogActivity(userID int64, userName, action, description string, score int) error {
	_, err := s.db.Exec(`
		INSERT INTO activity (user_id, user_name, action, description, score, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, userName, action, description, score, time.Now())
	return err
}

// GetActivityFeed получить ленту активности (друзья + я)
func (s *SocialService) GetActivityFeed(userID int64, limit int) ([]Activity, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.user_id, a.user_name, a.action, a.description, a.score, a.created_at
		FROM activity a
		WHERE a.user_id = ? OR a.user_id IN (
			SELECT friend_id FROM friends WHERE user_id = ?
		)
		ORDER BY a.created_at DESC
		LIMIT ?
	`, userID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := make([]Activity, 0)
	for rows.Next() {
		var act Activity
		err := rows.Scan(&act.ID, &act.UserID, &act.UserName, &act.Action, &act.Description, &act.Score, &act.CreatedAt)
		if err != nil {
			continue
		}
		activities = append(activities, act)
	}
	return activities, nil
}
