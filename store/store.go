package store

import (
	"errors"
	"sync"
	"time"
)

type URLData struct {
	OriginalURL string
	Clicks      int
	CreatedAt   time.Time
}

type Store struct {
	mu sync.RWMutex
	db map[string]*URLData
}

func New() *Store {
	return &Store{db: make(map[string]*URLData)}
}

func (s *Store) Save(slug string, original string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.db[slug]; exists {
		return errors.New("slug already exists")
	}

	s.db[slug] = &URLData{
		OriginalURL: original,
		CreatedAt:   time.Now(),
	}
	return nil
}

func (s *Store) Get(slug string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.db[slug]
	if !exists {
		return "", errors.New("slug not found")
	}
	return data.OriginalURL, nil
}

func (s *Store) IncrementClicks(slug string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if data, exists := s.db[slug]; exists {
		data.Clicks++
	}
}

func (s *Store) GetStats(slug string) (int, time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.db[slug]
	if !exists {
		return 0, time.Time{}, errors.New("slug not found")
	}
	return data.Clicks, data.CreatedAt, nil
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.db)
}
