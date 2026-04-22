package alerts

import (
	"database/sql"
	"log/slog"
)

type Manager struct {
	ch chan string
	db *sql.DB
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{
		ch: make(chan string, 100),
		db: db,
	}
}

func (a *Manager) Channel() chan string {
	return a.ch
}

func (a *Manager) Run() {
	for msg := range a.ch {
		slog.Info("alert", "message", msg)
		if _, err := a.db.Exec("INSERT INTO events(message) VALUES(?)", msg); err != nil {
			slog.Warn("alerts: failed to persist event", "err", err)
		}
	}
}
