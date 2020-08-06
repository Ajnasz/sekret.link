package main

import (
	"fmt"
	"sync"
)

type MemoryStorage struct {
	entries struct {
		sync.RWMutex
		m map[string][]byte
	}
}

func (m *MemoryStorage) Create(UUID string, entry []byte) error {
	m.entries.RLock()
	defer m.entries.RUnlock()

	m.entries.m[UUID] = entry
	return nil
}

func (m *MemoryStorage) Get(UUID string) (*Entry, error) {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if entry, ok := m.entries.m[UUID]; ok {
		return &Entry{
			UUID: UUID,
			Data: entry,
		}, nil
	}

	return nil, fmt.Errorf("Entry not found")
}

func (m *MemoryStorage) GetAndDelete(UUID string) (*Entry, error) {
	m.entries.RLock()
	defer m.entries.RUnlock()

	if entry, ok := m.entries.m[UUID]; ok {
		delete(m.entries.m, UUID)

		return &Entry{
			UUID: UUID,
			Data: entry,
		}, nil
	}

	return nil, fmt.Errorf("Entry not found")
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		struct {
			sync.RWMutex
			m map[string][]byte
		}{m: make(map[string][]byte)},
	}
}
