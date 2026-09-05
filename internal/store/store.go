package store

import (
	"sync"
)

type Store struct {
	Data map[string]interface{}
	mux  sync.RWMutex
}

func Newstore(s Store) *Store {
	return &Store{
		Data: make(map[string]interface{}),
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
	val, ok := s.Data[key]
	if !ok {
		return "", false
	}
	str, isStr := val.(string)
	if !isStr {
		// Try to return string representation for other types?
		// Or return false because it's not a string? Redis returns WRONGTYPE. Let's return false for now.
		return "", false
	}
	return str, true
}

func (s *Store) GetRaw(key string) (interface{}, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	val, ok := s.Data[key]
	return val, ok
}

func (s *Store) SetRaw(key string, value interface{}) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Data[key] = value
}

func (s *Store) Delete(key string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	_, exists := s.Data[key]
	if exists {
		delete(s.Data, key)
	}
	return exists
}

func (s *Store) Exists(key string) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	_, exists := s.Data[key]
	return exists
}
