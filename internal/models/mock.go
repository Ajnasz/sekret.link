package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockEntryModel struct {
	mock.Mock
}

func (m *MockEntryModel) CreateEntry(
	ctx context.Context,
	tx *sql.Tx,
	UUID string,
	data []byte,
	remainingReads int,
	expire time.Duration) (*EntryMeta, error) {
	args := m.Called(ctx, tx, UUID, data, remainingReads, expire)
	return args.Get(0).(*EntryMeta), args.Error(1)
}

func (m *MockEntryModel) ReadEntry(ctx context.Context, tx *sql.Tx, UUID string) (*Entry, error) {
	args := m.Called(ctx, tx, UUID)
	return args.Get(0).(*Entry), args.Error(1)
}

func (m *MockEntryModel) Use(ctx context.Context, tx *sql.Tx, UUID string) error {
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
