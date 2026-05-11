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
	mu   sync.RWMutex
	urls map[string]*URLData
}

func New() *Store {
	return &Store{
		urls: make(map[string]*URLData),
	}
}

// slug = the random token that it will created for url
func (s *Store) Save(slug string, original string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.urls[slug]; exists {
		return errors.New("slug already exists!")
	}

	s.urls[slug] = &URLData{
		OriginalURL: original,
		Clicks:      0,
		CreatedAt:   time.Now(),
	}

	return nil
}

func (s *Store) Get(slug string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.urls[slug]
	if !exists {
		return "", errors.New("slug not found!")
	}

	return data.OriginalURL, nil
}

func (s *Store) IncrementClicks(slug string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if data, exists := s.urls[slug]; exists {
		data.Clicks++
	}
}

func (s *Store) GetStats(slug string) (int, time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.urls[slug]
	if !exists {
		return 0, time.Time{}, errors.New("slug not found!")
	}

	return data.Clicks, data.CreatedAt, nil
}
