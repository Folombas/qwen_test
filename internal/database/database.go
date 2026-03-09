// Package database предоставляет функции для работы с базой данных SQLite.
package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB инициализирует базу данных
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	// Включаем foreign keys
	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	// Применяем миграции
	return RunMigrations()
}

// CloseDB закрывает соединение с БД
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// CreateBackup создаёт резервную копию базы данных
func CreateBackup(dbPath string) (string, error) {
	if DB == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// Создаём директорию для бэкапов
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Имя файла бэкапа
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("qwen_backup_%s.db", timestamp))

	// Копируем файл БД
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to read database file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	log.Printf("✅ Бэкап создан: %s", backupPath)
	return backupPath, nil
}

// CleanupOldBackups удаляет старые бэкапы, оставляя только последние n
func CleanupOldBackups(keep int) error {
	backupDir := "backups"
	files, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(files) <= keep {
		return nil
	}

	// Сортируем по времени модификации
	type fileInfo struct {
		name    string
		modTime time.Time
	}

	var infos []fileInfo
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}
		infos = append(infos, fileInfo{name: f.Name(), modTime: info.ModTime()})
	}

	// Сортировка по убыванию времени
	for i := 0; i < len(infos)-1; i++ {
		for j := i + 1; j < len(infos); j++ {
			if infos[i].modTime.Before(infos[j].modTime) {
				infos[i], infos[j] = infos[j], infos[i]
			}
		}
	}

	// Удаляем старые
	for i := keep; i < len(infos); i++ {
		path := filepath.Join(backupDir, infos[i].name)
		if err := os.Remove(path); err != nil {
			log.Printf("⚠️  WARNING: Не удалось удалить старый бэкап %s: %v", path, err)
		} else {
			log.Printf("🗑️  Удалён старый бэкап: %s", path)
		}
	}

	return nil
}

// SaveJSON сохраняет данные в JSON файл
func SaveJSON(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

// LoadJSON загружает данные из JSON файла
func LoadJSON(filename string, dest interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
