package main

import (
	"sync"

	"github.com/google/uuid"
)

type EntryStorage interface {
	Create([]byte) string
	Get(string) []byte
}

type MemoryStorage struct {
	entries struct {
		sync.RWMutex
		m map[string][]byte
	}
}

func (m *MemoryStorage) Create(entry []byte) string {
	m.entries.RLock()
	defer m.entries.RUnlock()
	newUUID := uuid.New()

	uuidString := newUUID.String()

	m.entries.m[uuidString] = entry

	return uuidString
}

func (m *MemoryStorage) Get(UUID string) []byte {
	m.entries.RLock()
	defer m.entries.RUnlock()
	entry := m.entries.m[UUID]

	delete(m.entries.m, UUID)

	return entry
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		struct {
			sync.RWMutex
			m map[string][]byte
		}{m: make(map[string][]byte)},
	}
}
