package main

import (
	"sync"
	"time"
)

type memoryEntry struct {
	Data     []byte
	Created  time.Time
	Expire   time.Time
	Accessed time.Time
}

type MemoryStorage struct {
	entries struct {
		sync.RWMutex
		m map[string]*memoryEntry
	}
}

func (m *MemoryStorage) Close() error {
	return nil
}

func (m *MemoryStorage) Create(UUID string, entry []byte, expire time.Duration) error {
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

func (m *MemoryStorage) GetMeta(UUID string) (*EntryMeta, error) {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if entry, ok := m.entries.m[UUID]; ok {
		meta := &EntryMeta{
			UUID:     UUID,
			Created:  entry.Created,
			Accessed: entry.Accessed,
			Expire:   entry.Expire,
		}

		if meta.IsExpired() {
			delete(m.entries.m, UUID)
			return nil, entryExpiredError
		}
		return meta, nil
	}

	return nil, entryNotFound
}

func (m *MemoryStorage) Get(UUID string) (*Entry, error) {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if entry, ok := m.entries.m[UUID]; ok {
		meta := EntryMeta{
			UUID:     UUID,
			Created:  entry.Created,
			Accessed: entry.Accessed,
			Expire:   entry.Expire,
		}

		if meta.IsExpired() {
			delete(m.entries.m, UUID)
			return nil, entryExpiredError
		}
		return &Entry{
			EntryMeta: meta,
			Data:      entry.Data,
		}, nil
	}

	return nil, entryNotFound
}

func (m *MemoryStorage) GetAndDelete(UUID string) (*Entry, error) {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if entry, ok := m.entries.m[UUID]; ok {
		delete(m.entries.m, UUID)
		meta := EntryMeta{
			UUID:     UUID,
			Created:  entry.Created,
			Accessed: entry.Accessed,
			Expire:   entry.Expire,
		}

		if meta.IsExpired() {
			return nil, entryExpiredError
		}

		return &Entry{
			EntryMeta: meta,
			Data:      entry.Data,
		}, nil
	}

	return nil, entryNotFound
}

func (m *MemoryStorage) Delete(UUID string) error {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if _, ok := m.entries.m[UUID]; ok {
		delete(m.entries.m, UUID)
	}

	return nil
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		struct {
			sync.RWMutex
			m map[string]*memoryEntry
		}{m: make(map[string]*memoryEntry)},
	}
}
