package alerts

import (
	"database/sql"
	"log"
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
		log.Println(msg)
		a.db.Exec("INSERT INTO events(message) VALUES(?)", msg)
	}
}
