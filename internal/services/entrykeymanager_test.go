package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEntryKeyModel struct {
	mock.Mock
}

func (m *MockEntryKeyModel) Create(ctx context.Context, tx *sql.Tx, entryUUID string, encryptedKey []byte, hash []byte) (*models.EntryKey, error) {
	args := m.Called(ctx, tx, entryUUID, encryptedKey, hash)
	return args.Get(0).(*models.EntryKey), args.Error(1)
}

func (m *MockEntryKeyModel) Get(ctx context.Context, db *sql.DB, entryUUID string) ([]models.EntryKey, error) {
	args := m.Called(ctx, db, entryUUID)
	return args.Get(0).([]models.EntryKey), args.Error(1)
}

func (m *MockEntryKeyModel) Delete(ctx context.Context, tx *sql.Tx, uuid string) error {
	args := m.Called(ctx, tx, uuid)
	return args.Error(0)
}

func (m *MockEntryKeyModel) SetExpire(ctx context.Context, tx *sql.Tx, uuid string, expire time.Time) error {
	args := m.Called(ctx, tx, uuid, expire)
	return args.Error(0)
}

func (m *MockEntryKeyModel) SetMaxRead(ctx context.Context, tx *sql.Tx, uuid string, maxRead int) error {
	args := m.Called(ctx, tx, uuid, maxRead)
	return args.Error(0)
}

func (m *MockEntryKeyModel) Use(ctx context.Context, tx *sql.Tx, uuid string) error {
	args := m.Called(ctx, tx, uuid)
	return args.Error(0)
}

type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) Hash(data []byte) []byte {
	args := m.Called(data)
	return args.Get(0).([]byte)
}

type EncrypterMock struct {
	mock.Mock
}

func (e *EncrypterMock) Encrypt(data []byte) ([]byte, error) {
	args := e.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (e *EncrypterMock) Decrypt(data []byte) ([]byte, error) {
	args := e.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func TestEntryKeyManager_Create(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	entryUUID := "test-entry-uuid"
	dek := []byte("test-dek")
	encryptedKey := []byte("test-encrypted-key")
	hash := []byte("test-hash")
	expire := time.Now()
	maxRead := 10

	encrypter.On("Encrypt", dek).Return(encryptedKey, nil)
	hasher.On("Hash", dek).Return(hash)
	model.On("Create", ctx, mock.Anything, entryUUID, encryptedKey, hash).Return(&models.EntryKey{
		UUID:           "test-uuid",
		EntryUUID:      entryUUID,
		EncryptedKey:   encryptedKey,
		Created:        time.Now(),
		Expire:         sql.NullTime{Time: time.Now(), Valid: false},
		RemainingReads: sql.NullInt16{Int16: 0, Valid: false},
	}, nil)

	model.On("SetExpire", ctx, mock.Anything, "test-uuid", expire).Return(nil)
	model.On("SetMaxRead", ctx, mock.Anything, "test-uuid", maxRead).Return(nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	entryKey, key, err := manager.Create(ctx, entryUUID, dek, &expire, &maxRead)

	model.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	hasher.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.NoError(t, err)
	assert.Equal(t, "test-uuid", entryKey.UUID)
	assert.Equal(t, expire, entryKey.Expire.Time)
	assert.True(t, entryKey.Expire.Valid)
	assert.Equal(t, int16(maxRead), entryKey.RemainingReads.Int16)
	assert.NotEmpty(t, key.Get())
}

func TestEntryKeyManager_Create_NoExpire(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	dek := []byte("test-dek")
	entryUUID := "test-entry-uuid"
	encryptedKey := []byte("test-encrypted-key")
	hash := []byte("test-hash")

	hasher.On("Hash", dek).Return(hash)
	encrypter.On("Encrypt", dek).Return(encryptedKey, nil)
	model.On("Create", ctx, mock.Anything, entryUUID, encryptedKey, hash).Return(&models.EntryKey{
		UUID:           "test-uuid",
		EntryUUID:      entryUUID,
		EncryptedKey:   encryptedKey,
		KeyHash:        hash,
		Created:        time.Now(),
		Expire:         sql.NullTime{Time: time.Now(), Valid: false},
		RemainingReads: sql.NullInt16{Int16: 0, Valid: false},
	}, nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}
	manager := NewEntryKeyManager(db, model, hasher, crypto)
	entryKey, key, err := manager.Create(ctx, entryUUID, dek, nil, nil)

	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	model.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.NoError(t, err)
	assert.Equal(t, "test-uuid", entryKey.UUID)
	assert.False(t, entryKey.Expire.Valid)
	assert.NotEmpty(t, key.Get())
}

func TestEntryKeyManager_Create_NoMaxRead(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	entryUUID := "test-entry-uuid"
	dek := []byte("test-dek")
	encryptedKey := []byte("test-encrypted-key")
	hash := []byte("test-hash")

	hasher.On("Hash", dek).Return(hash)
	encrypter.On("Encrypt", dek).Return(encryptedKey, nil)
	model.On("Create", ctx, mock.Anything, entryUUID, encryptedKey, hash).Return(&models.EntryKey{
		UUID:           "test-uuid",
		EntryUUID:      entryUUID,
		EncryptedKey:   encryptedKey,
		KeyHash:        hash,
		Created:        time.Now(),
		Expire:         sql.NullTime{Time: time.Now(), Valid: false},
		RemainingReads: sql.NullInt16{Int16: 0, Valid: false},
	}, nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	entryKey, key, err := manager.Create(ctx, entryUUID, dek, nil, nil)

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.NoError(t, err)
	assert.Equal(t, "test-uuid", entryKey.UUID)
	assert.False(t, entryKey.RemainingReads.Valid)
	// key.Get should not return an empty string
	assert.NotEmpty(t, key.Get())
}
