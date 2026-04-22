package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(path string) *sql.DB {
	db, _ := sql.Open("sqlite3", path)

	db.Exec(`CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		message TEXT
	);`)

	db.Exec(`CREATE TABLE IF NOT EXISTS device_uptime (
		ip TEXT,
		last_change DATETIME,
		status INTEGER
	);`)

	return db
}
