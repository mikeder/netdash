package device

import (
	"database/sql"
	"log/slog"
	"sync"
	"time"
)

type Store struct {
	mu          sync.Mutex
	data        map[string]*Device
	alerts      chan string
	db          *sql.DB
	subscribers []chan struct{}
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

func (s *Store) Subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.subscribers = append(s.subscribers, ch)
	s.mu.Unlock()
	return ch
}

func (s *Store) Unsubscribe(ch chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sub := range s.subscribers {
		if sub == ch {
			s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
			return
		}
	}
}

func (s *Store) Update(ip string, update func(d *Device)) {
	s.mu.Lock()

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

	subs := make([]chan struct{}, len(s.subscribers))
	copy(subs, s.subscribers)

	s.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func boolToInt(b bool) int {
	if b { return 1 }
	return 0
}

func (s *Store) All() map[string]*Device {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]*Device, len(s.data))
	for k, v := range s.data {
		d := *v
		out[k] = &d
	}
	return out
}
