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

func (m *MockEntryKeyer) Create(ctx context.Context, entryUUID string, dek key.Key, expire time.Time, maxRead int) (*EntryKey, key.Key, error) {
	args := m.Called(ctx, entryUUID, dek, expire, maxRead)
	if args.Get(1) == nil {
		return args.Get(0).(*EntryKey), nil, args.Error(2)
	}
	return args.Get(0).(*EntryKey), args.Get(1).(key.Key), args.Error(2)
}

func (m *MockEntryKeyer) CreateWithTx(ctx context.Context, tx *sql.Tx, entryUUID string, dek key.Key, expire time.Time, maxRead int) (*EntryKey, key.Key, error) {
	args := m.Called(ctx, tx, entryUUID, dek, expire, maxRead)
	return args.Get(0).(*EntryKey), args.Get(1).(key.Key), args.Error(2)
}

func (m *MockEntryKeyer) GetDEK(ctx context.Context, entryUUID string, kek key.Key) (key.Key, *EntryKey, error) {
	args := m.Called(ctx, entryUUID, kek)
	return args.Get(0).(key.Key), args.Get(1).(*EntryKey), args.Error(2)
}

func (m *MockEntryKeyer) GetDEKTx(ctx context.Context, tx *sql.Tx, entryUUID string, kek key.Key) (key.Key, *EntryKey, error) {
	args := m.Called(ctx, tx, entryUUID, kek)
	return args.Get(0).(key.Key), args.Get(1).(*EntryKey), args.Error(2)
}

func (m *MockEntryKeyer) GenerateEncryptionKey(ctx context.Context, entryUUID string, existingKey key.Key, expire time.Time, maxRead int) (*EntryKey, key.Key, error) {
	args := m.Called(ctx, entryUUID, existingKey, expire, maxRead)
	return args.Get(0).(*EntryKey), args.Get(1).(key.Key), args.Error(2)
}

func (m *MockEntryKeyer) UseTx(ctx context.Context, tx *sql.Tx, entryUUID string) error {
	args := m.Called(ctx, tx, entryUUID)
	return args.Error(0)
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
