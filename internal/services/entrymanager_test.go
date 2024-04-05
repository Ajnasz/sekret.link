package services

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var timenow = time.Now()

func Test_EntryService_Create(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec("INSERT INTO entries").WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	ctx := context.Background()

	data := []byte("data")
	encryptedData := []byte("encrypted")
	entryModel := new(models.MockEntryModel)
	entryModel.
		On("CreateEntry", ctx, mock.Anything, mock.Anything, encryptedData, 1, mock.Anything).
		Return(&models.EntryMeta{
			UUID:           "uuid",
			RemainingReads: 1,
			DeleteKey:      "delete_key",
			Created:        timenow,
			Expire:         timenow.Add(time.Minute),
		}, nil)

	entryCrypto := new(MockEntryCrypto)
	entryCrypto.On("Encrypt", data).Return(encryptedData, nil)
	crypto := func(key []byte) Encrypter {
		return entryCrypto
	}

	keyManager := new(MockEntryKeyer)
	kek := key.NewKey()
	kek.Set([]byte("kek"))
	keyManager.On("CreateWithTx", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&EntryKey{}, kek, nil)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	meta, key, err := service.CreateEntry(ctx, data, 1, time.Minute)

	assert.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Equal(t, key, kek.Get())

	entryModel.AssertExpectations(t)
	keyManager.AssertExpectations(t)
	if meta.UUID == "" {
		t.Error("expected UUID to be set")
	}

	if meta.DeleteKey == "" {
		t.Error("expected delete key to be set")
	}

	if meta.Created.IsZero() {
		t.Error("expected created to be set")
	}

	if meta.Expire.IsZero() {
		t.Error("expected expire to be set")
	}

	if meta.RemainingReads != 1 {
		t.Errorf("expected remaining reads to be 1 got %d", meta.RemainingReads)
	}
}

func TestCreateError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec("INSERT INTO entries").WillReturnError(fmt.Errorf("error"))
	sqlMock.ExpectRollback()

	ctx := context.Background()

	data := []byte("data")
	encryptedData := []byte("encrypted")

	entryModel := new(models.MockEntryModel)
	entryModel.
		On("CreateEntry", ctx, mock.Anything, mock.Anything, encryptedData, 1, mock.Anything).
		Return(&models.EntryMeta{}, fmt.Errorf("error"))

	entryCrypto := new(MockEntryCrypto)
	entryCrypto.On("Encrypt", data).Return(encryptedData, nil)
	crypto := func(key []byte) Encrypter {
		return entryCrypto
	}

	keyManager := new(MockEntryKeyer)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	meta, key, err := service.CreateEntry(ctx, data, 1, time.Minute)

	assert.Error(t, err)
	assert.Nil(t, meta)
	assert.Nil(t, key)

	entryModel.AssertExpectations(t)
	keyManager.AssertExpectations(t)
}

func TestReadEntry(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	ctx := context.Background()

	entry := models.Entry{
		EntryMeta: models.EntryMeta{
			UUID:           "uuid",
			RemainingReads: 1,
			DeleteKey:      "delete_key",
			Created:        timenow,
			Accessed:       sql.NullTime{Time: timenow, Valid: true},
			Expire:         timenow.Add(time.Minute),
		},
		Data: []byte("encrypted"),
	}

	entryModel := new(models.MockEntryModel)
	entryModel.
		On("ReadEntry", ctx, mock.Anything, "uuid").
		Return(&entry, nil)
	entryModel.
		On("UpdateAccessed", ctx, mock.Anything, "uuid").
		Return(nil)

	key := []byte("key")
	entryCrypto := new(MockEntryCrypto)
	entryCrypto.On("Decrypt", []byte("encrypted")).Return([]byte("data"), nil)

	crypto := func(key []byte) Encrypter {
		return entryCrypto
	}

	keyManager := new(MockEntryKeyer)

	keyManager.On("GetDEKTx", ctx, mock.Anything, "uuid", key).Return([]byte("dek"), &EntryKey{}, nil)
	keyManager.On("UseTx", ctx, mock.Anything, entry.UUID).Return(nil)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	data, err := service.ReadEntry(ctx, "uuid", key)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, Entry{
		UUID:           entry.UUID,
		Data:           []byte("data"),
		RemainingReads: 0,
		DeleteKey:      entry.DeleteKey,
		Created:        entry.Created,
		Accessed:       entry.Accessed.Time,
		Expire:         entry.Expire,
	}, *data)

	entryModel.AssertExpectations(t)
}

func TestReadEntryError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	ctx := context.Background()

	entryModel := new(models.MockEntryModel)
	entryModel.
		On("ReadEntry", ctx, mock.Anything, "uuid").
		Return(&models.Entry{}, fmt.Errorf("error"))

	entryCrypto := new(MockEntryCrypto)
	crypto := func(key []byte) Encrypter {
		return entryCrypto
	}
	keyManager := new(MockEntryKeyer)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	data, err := service.ReadEntry(ctx, "uuid", []byte("key"))

	assert.Error(t, err)
	assert.Nil(t, data)

	entryModel.AssertExpectations(t)
	keyManager.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
}

func TestDeleteEntry(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	ctx := context.Background()

	entryModel := new(models.MockEntryModel)
	entryModel.
		On("DeleteEntry", ctx, mock.Anything, "uuid", "delete_key").
		Return(nil)

	entryCrypto := new(MockEntryCrypto)
	crypto := func(key []byte) Encrypter {
		return entryCrypto
	}
	keyManager := new(MockEntryKeyer)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	err = service.DeleteEntry(ctx, "uuid", "delete_key")

	assert.NoError(t, err)

	entryModel.AssertExpectations(t)
	keyManager.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
}

func TestDeleteEntryError(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	ctx := context.Background()

	entryModel := new(models.MockEntryModel)
	entryModel.
		On("DeleteEntry", ctx, mock.Anything, "uuid", "delete_key").
		Return(fmt.Errorf("error"))

	entryCrypto := new(MockEntryCrypto)
	crypto := func(key []byte) Encrypter {
		return entryCrypto
	}

	keyManager := new(MockEntryKeyer)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	err = service.DeleteEntry(ctx, "uuid", "delete_key")

	assert.Error(t, err)

	entryModel.AssertExpectations(t)
	keyManager.AssertExpectations(t)
	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
}

func TestDeleteEntryInvalidDeleteKey(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	sqlMock.ExpectBegin()
	// sqlMock.ExpectExec("DELETE FROM entries WHERE uuid=(.+) AND delete_key=(.+)").WillReturnResult(sqlmock.NewResult(2, 1))
	sqlMock.ExpectRollback()

	ctx := context.Background()

	entryModel := new(models.MockEntryModel)
	entryModel.
		On("DeleteEntry", ctx, mock.Anything, "uuid", "delete_key").
		Return(models.ErrEntryNotFound)

	entryCrypto := new(MockEntryCrypto)
	crypto := func(key []byte) Encrypter {
		return entryCrypto
	}

	keyManager := new(MockEntryKeyer)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	err = service.DeleteEntry(ctx, "uuid", "delete_key")

	assert.Error(t, err)
	assert.Equal(t, models.ErrEntryNotFound, err)

	entryModel.AssertExpectations(t)
	keyManager.AssertExpectations(t)

	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
}
