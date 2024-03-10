package test

import (
	"context"
	"database/sql"
	"time"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/stretchr/testify/mock"
)

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

type MockEntryModel struct {
	mock.Mock
}

func (m *MockEntryModel) CreateEntry(
	ctx context.Context,
	tx *sql.Tx,
	UUID string,
	data []byte,
	remainingReads int,
	expire time.Duration) (*models.EntryMeta, error) {
	args := m.Called(ctx, tx, UUID, data, remainingReads, expire)
	return args.Get(0).(*models.EntryMeta), args.Error(1)
}

func (m *MockEntryModel) ReadEntry(ctx context.Context, tx *sql.Tx, UUID string) (*models.Entry, error) {
	args := m.Called(ctx, tx, UUID)
	return args.Get(0).(*models.Entry), args.Error(1)
}

func (m *MockEntryModel) UpdateAccessed(ctx context.Context, tx *sql.Tx, UUID string) error {
	args := m.Called(ctx, tx, UUID)
	return args.Error(0)
}

func (m *MockEntryModel) DeleteEntry(ctx context.Context, tx *sql.Tx, UUID string, deleteKey string) error {
	args := m.Called(ctx, tx, UUID, deleteKey)
	return args.Error(0)
}

func (m *MockEntryModel) DeleteExpired(ctx context.Context, tx *sql.Tx) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}
