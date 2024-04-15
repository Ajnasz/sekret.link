package services

import (
	"context"
	"database/sql"
	"fmt"
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

func (m *MockEntryKeyModel) Get(ctx context.Context, tx *sql.Tx, entryUUID string) ([]models.EntryKey, error) {
	args := m.Called(ctx, tx, entryUUID)
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

func (m *MockEntryKeyModel) SetMaxReads(ctx context.Context, tx *sql.Tx, uuid string, maxRead int) error {
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
	model.On("SetMaxReads", ctx, mock.Anything, "test-uuid", maxRead).Return(nil)

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
	assert.Equal(t, expire, entryKey.Expire)
	assert.Equal(t, maxRead, entryKey.RemainingReads)
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
	fmt.Println("__------------------------", entryKey.Expire)
	// assert.False(nil, entryKey.Expire)
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
	// key.Get should not return an empty string
	assert.NotEmpty(t, key.Get())
}

func TestEntryKeyManager_GetDEK(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	entryUUID := "test-entry-uuid"
	encryptedKey := []byte("test-encrypted-key")
	dek := []byte("test-dek")
	hash := []byte("test-hash")

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()
	hasher.On("Hash", dek).Return(hash)
	encrypter.On("Decrypt", encryptedKey).Return(dek, nil)
	model.On("Get", ctx, mock.Anything, entryUUID).Return([]models.EntryKey{
		{
			UUID:         "test-uuid",
			EntryUUID:    entryUUID,
			EncryptedKey: encryptedKey,
			KeyHash:      hash,
			Created:      time.Now(),
		},
	}, nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	foundDEK, entryKey, err := manager.GetDEK(ctx, entryUUID, dek)

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.NoError(t, err)
	assert.Equal(t, dek, foundDEK)
	assert.Equal(t, "test-uuid", entryKey.UUID)
}

// TestEntryKeyManager_GetDEK_NotFound tests the case when the entry key is not
// found so the function should return an ErrEntryKeyNotFound error
func TestEntryKeyManager_GetDEK_NotFound(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	entryUUID := "test-entry-uuid"
	dek := []byte("test-dek")

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()
	model.On("Get", ctx, mock.Anything, entryUUID).Return([]models.EntryKey{}, nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	foundDEK, entryKey, err := manager.GetDEK(ctx, entryUUID, dek)

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.Error(t, ErrEntryKeyNotFound, err)
	assert.Nil(t, foundDEK)
	assert.Nil(t, entryKey)
}

func TestEntryKeyManager_GetDEK_DecryptError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	entryUUID := "test-entry-uuid"
	encryptedKey := []byte("test-encrypted-keyke")
	dek := []byte("test-dekk")
	hash := []byte("test-hashh")

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	encrypter.On("Decrypt", encryptedKey).Return([]byte{}, assert.AnError)

	model.On("Get", ctx, mock.Anything, entryUUID).Return([]models.EntryKey{
		{
			UUID:         "test-uuid",
			EntryUUID:    entryUUID,
			EncryptedKey: encryptedKey,
			KeyHash:      hash,
			Created:      time.Now(),
		},
	}, nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	foundDEK, entryKey, err := manager.GetDEK(ctx, entryUUID, dek)

	assert.Error(t, err)
	assert.Nil(t, foundDEK)
	assert.Nil(t, entryKey)
	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
}

func TestEntryManager_InvalidDEK(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	entryUUID := "test-entry-uuid"
	encryptedKey := []byte("test-encrypted-key")
	dek := []byte("test-dek")
	badDEK := []byte("bad-dek")
	hash := []byte("test-hash")
	badHash := []byte("bad-hash")

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()
	encrypter.On("Decrypt", encryptedKey).Return(badDEK, nil)
	hasher.On("Hash", badDEK).Return(badHash)

	model.On("Get", ctx, mock.Anything, entryUUID).Return([]models.EntryKey{
		{
			UUID:         "test-uuid",
			EntryUUID:    entryUUID,
			EncryptedKey: encryptedKey,
			KeyHash:      hash,
			Created:      time.Now(),
		},
	}, nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	foundDEK, entryKey, err := manager.GetDEK(ctx, entryUUID, dek)

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.Error(t, err)
	assert.Nil(t, foundDEK)
	assert.Nil(t, entryKey)
}

func TestEntryKeyManager_GenerateEncryptionKey(t *testing.T) {
	// reads an existing key from the db, creates a new key, and returns the new key

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}

	entryUUID := "test-entry-uuid"
	encryptedKey := []byte("test-encrypted-key")
	newEncryptedKey := []byte("new-test-encrypted-key")
	dek := []byte("test-dek")
	hash := []byte("test-hash")
	expire := time.Now()
	maxRead := 10

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	model.On("Get", ctx, mock.Anything, entryUUID).Return([]models.EntryKey{
		{
			UUID:         "test-uuid",
			EntryUUID:    entryUUID,
			EncryptedKey: encryptedKey,
			KeyHash:      hash,
			Created:      time.Now(),
		},
	}, nil)
	encrypter.On("Decrypt", encryptedKey).Return(dek, nil)
	hasher.On("Hash", dek).Return(hash)

	encrypter.On("Encrypt", mock.Anything).Return(newEncryptedKey, nil)
	model.On("Create", ctx, mock.Anything, entryUUID, newEncryptedKey, hash).Return(&models.EntryKey{
		UUID:           "new-test-uuid",
		EntryUUID:      entryUUID,
		EncryptedKey:   newEncryptedKey,
		KeyHash:        hash,
		Created:        time.Now(),
		Expire:         sql.NullTime{Time: time.Now(), Valid: false},
		RemainingReads: sql.NullInt16{Int16: 0, Valid: false},
	}, nil)
	model.On("SetExpire", ctx, mock.Anything, "new-test-uuid", expire).Return(nil)
	model.On("SetMaxReads", ctx, mock.Anything, "new-test-uuid", maxRead).Return(nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)

	entryKey, key, err := manager.GenerateEncryptionKey(ctx, entryUUID, encryptedKey, &expire, &maxRead)

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.NoError(t, err)
	assert.Equal(t, "new-test-uuid", entryKey.UUID)
	assert.Equal(t, expire, entryKey.Expire)
	assert.Equal(t, maxRead, entryKey.RemainingReads)
	assert.NotEmpty(t, key.Get())
}

// TestEntryKeyManager_GenerateEncryptionKey_DecryptError tests if the UseTx method correctly calls the model's Use method
func TestEntryKeyManager_UseTx(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	model.On("Use", ctx, mock.Anything, "test-uuid").Return(nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	tx.Commit()

	err = manager.UseTx(ctx, tx, "test-uuid")

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.NoError(t, err)
}

func Test_EntryKeyManager_GetDEKTx_NoRemainingReads(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	entryUUID := "test-entry-uuid"
	encryptedKey := []byte("test-encrypted-key")
	dek := []byte("test-dek")
	hash := []byte("test-hash")

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()
	hasher.On("Hash", dek).Return(hash)
	encrypter.On("Decrypt", encryptedKey).Return(dek, nil)
	model.On("Get", ctx, mock.Anything, entryUUID).Return([]models.EntryKey{
		{
			UUID:           "test-uuid",
			EntryUUID:      entryUUID,
			EncryptedKey:   encryptedKey,
			KeyHash:        hash,
			Created:        time.Now(),
			RemainingReads: sql.NullInt16{Int16: 0, Valid: true},
		},
	}, nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	foundDEK, entryKey, err := manager.GetDEK(ctx, entryUUID, dek)

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.Error(t, ErrEntryNoRemainingReads, err)
	assert.Nil(t, foundDEK)
	assert.Nil(t, entryKey)
}

func Test_EntryKeyManager_GetDEKTx_Expired(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	entryUUID := "test-entry-uuid"
	encryptedKey := []byte("test-encrypted-key")
	dek := []byte("test-dek")
	hash := []byte("test-hash")

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()
	hasher.On("Hash", dek).Return(hash)
	encrypter.On("Decrypt", encryptedKey).Return(dek, nil)
	model.On("Get", ctx, mock.Anything, entryUUID).Return([]models.EntryKey{
		{
			UUID:           "test-uuid",
			EntryUUID:      entryUUID,
			EncryptedKey:   encryptedKey,
			KeyHash:        hash,
			Created:        time.Now(),
			Expire:         sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
			RemainingReads: sql.NullInt16{Int16: 1, Valid: true},
		},
	}, nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	foundDEK, entryKey, err := manager.GetDEK(ctx, entryUUID, dek)

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.Error(t, ErrEntryExpired, err)
	assert.Nil(t, foundDEK)
	assert.Nil(t, entryKey)
}

func Test_EntryKeyManager_Delete(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	ctx := context.Background()
	model := &MockEntryKeyModel{}
	hasher := &MockHasher{}
	encrypter := &EncrypterMock{}
	uuid := "test-uuid"

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()
	model.On("Delete", ctx, mock.Anything, uuid).Return(nil)

	crypto := func(key []byte) Encrypter {
		return encrypter
	}

	manager := NewEntryKeyManager(db, model, hasher, crypto)
	err = manager.Delete(ctx, uuid)

	model.AssertExpectations(t)
	hasher.AssertExpectations(t)
	encrypter.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
	assert.NoError(t, err)
}
