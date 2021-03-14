// +build !postgres,!sqlite,!redis

package main

import (
	"sync"
	"time"

	"github.com/Ajnasz/sekret.link/storage"
)

type memoryEntry struct {
	Data     []byte
	Created  time.Time
	Expire   time.Time
	Accessed time.Time
}

type memoryStorage struct {
	entries struct {
		sync.RWMutex
		m map[string]*memoryEntry
	}
}

func (m *memoryStorage) Close() error {
	return nil
}

func (m *memoryStorage) Create(UUID string, entry []byte, expire time.Duration) error {
	m.entries.Lock()
	defer m.entries.Unlock()

	now := time.Now()
	m.entries.m[UUID] = &memoryEntry{
		Data:    entry,
		Created: now,
		Expire:  now.Add(expire),
	}
	return nil
}

func (m *memoryStorage) GetMeta(UUID string) (*storage.EntryMeta, error) {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if entry, ok := m.entries.m[UUID]; ok {
		meta := &storage.EntryMeta{
			UUID:     UUID,
			Created:  entry.Created,
			Accessed: entry.Accessed,
			Expire:   entry.Expire,
		}

		if meta.IsExpired() {
			delete(m.entries.m, UUID)
			return nil, ErrEntryExpired
		}
		return meta, nil
	}

	return nil, ErrEntryNotFound
}

func (m *memoryStorage) Get(UUID string) (*storage.Entry, error) {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if entry, ok := m.entries.m[UUID]; ok {
		meta := storage.EntryMeta{
			UUID:     UUID,
			Created:  entry.Created,
			Accessed: entry.Accessed,
			Expire:   entry.Expire,
		}

		if meta.IsExpired() {
			delete(m.entries.m, UUID)
			return nil, ErrEntryExpired
		}
		return &storage.Entry{
			EntryMeta: meta,
			Data:      entry.Data,
		}, nil
	}

	return nil, ErrEntryNotFound
}

func (m *memoryStorage) GetAndDelete(UUID string) (*storage.Entry, error) {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if entry, ok := m.entries.m[UUID]; ok {
		delete(m.entries.m, UUID)
		meta := storage.EntryMeta{
			UUID:     UUID,
			Created:  entry.Created,
			Accessed: entry.Accessed,
			Expire:   entry.Expire,
		}

		if meta.IsExpired() {
			return nil, ErrEntryExpired
		}

		return &storage.Entry{
			EntryMeta: meta,
			Data:      entry.Data,
		}, nil
	}

	return nil, ErrEntryNotFound
}

func (m *memoryStorage) Delete(UUID string) error {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if _, ok := m.entries.m[UUID]; ok {
		delete(m.entries.m, UUID)
	}

	return nil
}
func (m *memoryStorage) DeleteExpired() error {
	now := time.Now()
	for UUID, entry := range m.entries.m {
		if entry.Expire.Before(now) {
			delete(m.entries.m, UUID)
		}
	}

	return nil
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{
		struct {
			sync.RWMutex
			m map[string]*memoryEntry
		}{m: make(map[string]*memoryEntry)},
	}
}

// func newStorage() storage.EntryStorage {
// 	return newMemoryStorage()
// }

type memoryCleanbleStorage struct {
	*memoryStorage
}

func (s memoryCleanbleStorage) Clean() {
	s.entries.RLock()
	defer s.entries.RUnlock()

	s.entries.m = make(map[string]*memoryEntry)
}
