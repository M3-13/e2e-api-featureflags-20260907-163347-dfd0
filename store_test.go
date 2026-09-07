package main

import (
	"errors"
	"strconv"
	"sync"
	"testing"
)

func TestStoreCreateAndGet(t *testing.T) {
	s := NewFlagStore()
	f := Flag{Key: "a", Enabled: true, Description: "d", RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := s.Get("a")
	if !ok {
		t.Fatal("expected flag to exist")
	}
	if got != f {
		t.Fatalf("got %+v, want %+v", got, f)
	}
}

func TestStoreCreateDuplicate(t *testing.T) {
	s := NewFlagStore()
	if err := s.Create(Flag{Key: "a"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Create(Flag{Key: "a"}); !errors.Is(err, ErrExists) {
		t.Fatalf("expected ErrExists, got %v", err)
	}
}

func TestStoreList(t *testing.T) {
	s := NewFlagStore()
	_ = s.Create(Flag{Key: "a"})
	_ = s.Create(Flag{Key: "b"})
	if list := s.List(); len(list) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(list))
	}
}

func TestStoreListEmpty(t *testing.T) {
	s := NewFlagStore()
	if list := s.List(); len(list) != 0 {
		t.Fatalf("expected empty list, got %d", len(list))
	}
}

func TestStoreGetMissing(t *testing.T) {
	s := NewFlagStore()
	if _, ok := s.Get("missing"); ok {
		t.Fatal("expected missing key to return ok=false")
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewFlagStore()
	_ = s.Create(Flag{Key: "a", Enabled: false})
	updated := Flag{Key: "a", Enabled: true, RolloutPercent: 100}
	got, ok := s.Update("a", updated)
	if !ok {
		t.Fatal("expected update to succeed")
	}
	if got != updated {
		t.Fatalf("got %+v, want %+v", got, updated)
	}
	if _, ok := s.Update("missing", updated); ok {
		t.Fatal("expected update of missing key to fail")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewFlagStore()
	_ = s.Create(Flag{Key: "a"})
	if !s.Delete("a") {
		t.Fatal("expected delete to succeed")
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("expected flag to be gone")
	}
	if s.Delete("a") {
		t.Fatal("expected delete of missing key to fail")
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewFlagStore()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := strconv.Itoa(i)
			if err := s.Create(Flag{Key: key}); err != nil {
				return
			}
			_, _ = s.Get(key)
			_ = s.List()
			_, _ = s.Update(key, Flag{Key: key, Enabled: true})
			_ = s.Delete(key)
		}(i)
	}
	wg.Wait()
}
