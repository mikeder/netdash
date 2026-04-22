package device

import (
	"database/sql"
	"log/slog"
	"sync"
	"time"
)

type Store struct {
	mu     sync.Mutex
	data   map[string]*Device
	alerts chan string
	db     *sql.DB
}

func NewStore() *Store {
	return &Store{data: make(map[string]*Device)}
}

func (s *Store) SetAlertChannel(ch chan string) {
	s.alerts = ch
}

func (s *Store) SetDB(db *sql.DB) {
	s.db = db
}

func (s *Store) Update(ip string, update func(d *Device)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, exists := s.data[ip]
	if !exists {
		d = &Device{IP: ip}
		s.data[ip] = d
		if s.alerts != nil {
			s.alerts <- "New device: " + ip
		}
	}

	prev := d.Online
	update(d)

	if prev != d.Online && s.db != nil {
		if _, err := s.db.Exec("INSERT INTO device_uptime(ip,last_change,status) VALUES(?,?,?)",
			ip, time.Now(), boolToInt(d.Online)); err != nil {
			slog.Warn("store: failed to record uptime change", "ip", ip, "err", err)
		}
	}
}

func boolToInt(b bool) int {
	if b { return 1 }
	return 0
}

func (s *Store) All() map[string]*Device {
	return s.data
}
