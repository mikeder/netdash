package storage

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		slog.Error("storage: failed to open database", "path", path, "err", err)
		os.Exit(1)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		message TEXT
	);`); err != nil {
		slog.Warn("storage: failed to create events table", "err", err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS device_uptime (
		ip TEXT,
		last_change DATETIME,
		status INTEGER
	);`); err != nil {
		slog.Warn("storage: failed to create device_uptime table", "err", err)
	}

	return db
}
