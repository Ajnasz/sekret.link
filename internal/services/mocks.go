package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/stretchr/testify/mock"
)

type MockEntryKeyer struct {
	mock.Mock
}

func (m *MockEntryKeyer) Create(ctx context.Context, entryUUID string, dek []byte, expire *time.Time, maxRead *int) (*EntryKey, *key.Key, error) {
	args := m.Called(ctx, entryUUID, dek, expire, maxRead)
	return args.Get(0).(*EntryKey), args.Get(1).(*key.Key), args.Error(2)
}

func (m *MockEntryKeyer) CreateWithTx(ctx context.Context, tx *sql.Tx, entryUUID string, dek []byte, expire *time.Time, maxRead *int) (*EntryKey, *key.Key, error) {
	args := m.Called(ctx, tx, entryUUID, dek, expire, maxRead)
	return args.Get(0).(*EntryKey), args.Get(1).(*key.Key), args.Error(2)
}

func (m *MockEntryKeyer) GetDEK(ctx context.Context, entryUUID string, kek []byte) ([]byte, *EntryKey, error) {
	args := m.Called(ctx, entryUUID, kek)
	return args.Get(0).([]byte), args.Get(1).(*EntryKey), args.Error(2)
}

func (m *MockEntryKeyer) GetDEKTx(ctx context.Context, tx *sql.Tx, entryUUID string, kek []byte) ([]byte, *EntryKey, error) {
	args := m.Called(ctx, tx, entryUUID, kek)
	return args.Get(0).([]byte), args.Get(1).(*EntryKey), args.Error(2)
}

func (m *MockEntryKeyer) GenerateEncryptionKey(ctx context.Context, entryUUID string, existingKey []byte, expire *time.Time, maxRead *int) (*EntryKey, *key.Key, error) {
	args := m.Called(ctx, entryUUID, existingKey, expire, maxRead)
	return args.Get(0).(*EntryKey), args.Get(1).(*key.Key), args.Error(2)
}

type MockEntryCrypto struct {
	mock.Mock
}

func (m *MockEntryCrypto) Encrypt(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEntryCrypto) Decrypt(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}
