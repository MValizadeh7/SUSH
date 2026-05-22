package store

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type URLData struct {
	OriginalURL string
	Clicks      int
	CreatedAt   time.Time
}

type Store struct {
	db    *sql.DB
	cache *redis.Client
}

func New(connStr, redisAddr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	cache := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := cache.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Store{db: db, cache: cache}, nil
}

func (s *Store) Save(slug string, original string) error {
	_, err := s.db.Exec(
		"INSERT INTO urls (slug, original_url, clicks, created_at) VALUES ($1, $2, 0, $3)",
		slug, original, time.Now(),
	)
	return err
}

func (s *Store) Get(slug string) (string, error) {
	ctx := context.Background()

	if url, err := s.cache.Get(ctx, slug).Result(); err == nil {
		return url, nil
	}

	var url string
	err := s.db.QueryRow("SELECT original_url FROM urls WHERE slug = $1", slug).Scan(&url)
	if err != nil {
		return "", err
	}

	s.cache.Set(ctx, slug, url, time.Hour)

	return url, nil
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
