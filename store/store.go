package store

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type URLData struct {
	OriginalURL string
	Clicks      int
	CreatedAt   time.Time
}

type Store struct {
	db *sql.DB
}

func New(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Save(slug string, original string) error {
	_, err := s.db.Exec(
		"INSERT INTO urls (slug, original_url, clicks, created_at) VALUES ($1, $2, 0, $3)",
		slug, original, time.Now(),
	)
	return err
}

func (s *Store) Get(slug string) (string, error) {
	var original string
	err := s.db.QueryRow(
		"SELECT original_url FROM urls WHERE slug = $1", slug,
	).Scan(&original)
	if err != nil {
		return "", err
	}
	return original, nil
}

func (s *Store) IncrementClicks(slug string) {
	s.db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE slug = $1", slug)
}

func (s *Store) GetStats(slug string) (int, time.Time, error) {
	var clicks int
	var createdAt time.Time
	err := s.db.QueryRow(
		"SELECT clicks, created_at FROM urls WHERE slug = $1", slug,
	).Scan(&clicks, &createdAt)
	if err != nil {
		return 0, time.Time{}, err
	}
	return clicks, createdAt, nil
}

func (s *Store) Count() int {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)
	return count
}
