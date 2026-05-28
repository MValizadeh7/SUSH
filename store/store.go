package store

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/lib/pq"
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
		db.Close()
		return nil, err
	}

	cache := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := cache.Ping(context.Background()).Err(); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db, cache: cache}, nil
}

func (s *Store) Close() {
	s.cache.Close()
	s.db.Close()
}

func Duplicate(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
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
	if _, err := s.db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE slug = $1", slug); err != nil {
		log.Printf("click bump failed for %s: %v", slug, err)
	}
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
	if err := s.db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count); err != nil {
		log.Printf("count query failed: %v", err)
	}
	return count
}
