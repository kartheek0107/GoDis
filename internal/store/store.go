package store

import (
	"sync"
)

type Store struct {
	Data map[string]string
	mux  sync.RWMutex
}

func Newstore(s Store) *Store {
	return &Store{
		Data: make(map[string]string),
	}
}

func (s *Store) Set(key string, value string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Data[key] = value
	return nil
}

func (s *Store) Get(key string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	str, ok := s.Data[key]
	return str, ok
}
