package main

import (
	"errors"
	"sync"
)

var ErrExists = errors.New("flag already exists")

type FlagStore struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

func NewFlagStore() *FlagStore {
	return &FlagStore{flags: make(map[string]Flag)}
}

var store = NewFlagStore()

func (s *FlagStore) Create(f Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[f.Key]; ok {
		return ErrExists
	}
	s.flags[f.Key] = f
	return nil
}

func (s *FlagStore) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		out = append(out, f)
	}
	return out
}

func (s *FlagStore) Get(key string) (Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[key]
	return f, ok
}

func (s *FlagStore) Update(key string, f Flag) (Flag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return Flag{}, false
	}
	s.flags[key] = f
	return f, true
}

func (s *FlagStore) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return false
	}
	delete(s.flags, key)
	return true
}
