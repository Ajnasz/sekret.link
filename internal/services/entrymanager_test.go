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
	crypto := func(key key.Key) Encrypter {
		return entryCrypto
	}

	keyManager := new(MockEntryKeyer)
	kek := key.NewKey()
	kek.Set([]byte("kek"))
	keyManager.On("CreateWithTx", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&EntryKey{}, *kek, nil)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	meta, key, err := service.CreateEntry(ctx, data, 1, time.Minute)

	assert.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Equal(t, key.Get(), kek.Get())

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
	crypto := func(key key.Key) Encrypter {
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
	t.Run("read entry with valid data", func(t *testing.T) {

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
			On("Use", ctx, mock.Anything, "uuid").
			Return(nil)

		k, err := key.NewGeneratedKey()
		if err != nil {
			t.Fatal(err)
		}

		entryCrypto := new(MockEntryCrypto)
		entryCrypto.On("Decrypt", []byte("encrypted")).Return([]byte("data"), nil)

		crypto := func(key key.Key) Encrypter {
			return entryCrypto
		}

		keyManager := new(MockEntryKeyer)

		dek, err := key.NewGeneratedKey()
		if err != nil {
			t.Fatal(err)
		}

		keyManager.On("GetDEKTx", ctx, mock.Anything, "uuid", *k).Return(*dek, &EntryKey{
			UUID: "entrykey uuid",
		}, nil)
		keyManager.On("UseTx", ctx, mock.Anything, "entrykey uuid").Return(nil)

		service := NewEntryManager(db, entryModel, crypto, keyManager)
		data, err := service.ReadEntry(ctx, "uuid", *k)

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
	})

	t.Run("should return notfound error when entry not found", func(t *testing.T) {
		db, sqlMock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		sqlMock.ExpectBegin()
		sqlMock.ExpectRollback()

		ctx := context.Background()

		var emptyEntry *models.Entry
		entryModel := new(models.MockEntryModel)
		entryModel.
			On("ReadEntry", ctx, mock.Anything, "uuid").
			Return(emptyEntry, models.ErrEntryNotFound)

		entryCrypto := new(MockEntryCrypto)
		crypto := func(key key.Key) Encrypter {
			return entryCrypto
		}
		keyManager := new(MockEntryKeyer)

		service := NewEntryManager(db, entryModel, crypto, keyManager)

		k, err := key.NewGeneratedKey()
		assert.NoError(t, err)
		data, err := service.ReadEntry(ctx, "uuid", *k)

		assert.Error(t, err)
		assert.Nil(t, data)

		entryModel.AssertExpectations(t)
		keyManager.AssertExpectations(t)
	})

	t.Run("it should try to decrypt with legacy method when key not found", func(t *testing.T) {
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
		entryModel.On("Use", ctx, mock.Anything, "uuid").Return(nil)

		entryCrypto := new(MockEntryCrypto)
		entryCrypto.On("Decrypt", []byte("encrypted")).Return([]byte("decrypted"), nil)

		crypto := func(key key.Key) Encrypter {
			return entryCrypto
		}

		var emptyEntryKey *EntryKey
		var emptyDEK key.Key

		k, err := key.NewGeneratedKey()
		assert.NoError(t, err)
		keyManager := new(MockEntryKeyer)
		keyManager.On("GetDEKTx", ctx, mock.Anything, "uuid", *k).Return(emptyDEK, emptyEntryKey, ErrEntryKeyNotFound)

		service := NewEntryManager(db, entryModel, crypto, keyManager)
		data, err := service.ReadEntry(ctx, "uuid", *k)

		assert.Nil(t, err)
		assert.Equal(t, "decrypted", string(data.Data))

		entryModel.AssertExpectations(t)
		keyManager.AssertExpectations(t)
	})
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
		Return(&models.Entry{}, models.ErrEntryNotFound)

	entryCrypto := new(MockEntryCrypto)
	crypto := func(key key.Key) Encrypter {
		return entryCrypto
	}
	keyManager := new(MockEntryKeyer)

	service := NewEntryManager(db, entryModel, crypto, keyManager)
	data, err := service.ReadEntry(ctx, "uuid", []byte("key"))

	assert.Error(t, ErrEntryNotFound)
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
	crypto := func(key key.Key) Encrypter {
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
	crypto := func(key key.Key) Encrypter {
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
	crypto := func(key key.Key) Encrypter {
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

func Test_EntryManager_DeleteExpired(t *testing.T) {
	t.Run("call delete method on the model", func(t *testing.T) {
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
			On("DeleteExpired", ctx, mock.Anything).
			Return(nil)

		entryCrypto := new(MockEntryCrypto)
		crypto := func(key key.Key) Encrypter {
			return entryCrypto
		}

		keyManager := new(MockEntryKeyer)

		service := NewEntryManager(db, entryModel, crypto, keyManager)
		err = service.DeleteExpired(ctx)

		assert.NoError(t, err)

		entryModel.AssertExpectations(t)
		keyManager.AssertExpectations(t)
		if sqlMock.ExpectationsWereMet() != nil {
			t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
		}
	})

	t.Run("call delete method on the model with error", func(t *testing.T) {
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
			On("DeleteExpired", ctx, mock.Anything).
			Return(fmt.Errorf("error"))

		entryCrypto := new(MockEntryCrypto)
		crypto := func(key key.Key) Encrypter {
			return entryCrypto
		}

		keyManager := new(MockEntryKeyer)

		service := NewEntryManager(db, entryModel, crypto, keyManager)
		err = service.DeleteExpired(ctx)

		assert.Error(t, err)

		entryModel.AssertExpectations(t)
		keyManager.AssertExpectations(t)
		if sqlMock.ExpectationsWereMet() != nil {
			t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
		}
	})
}

func Test_EntryManager_GenerateEntryKey(t *testing.T) {
	t.Run("call generate entry key method on the key manager", func(t *testing.T) {
		entryUUID := "entry-uuid"
		dek, err := key.NewGeneratedKey()
		if err != nil {
			t.Fatal(err)
		}
		kek, err := key.NewGeneratedKey()
		if err != nil {
			t.Fatal(err)
		}

		keyManager := new(MockEntryKeyer)
		keyManager.On("GenerateEncryptionKey", mock.Anything, entryUUID, *dek, mock.Anything, mock.Anything).
			Return(&EntryKey{
				EntryUUID:      entryUUID,
				RemainingReads: 1,
				Expire:         time.Now().Add(time.Minute),
			}, *kek, nil)

		service := NewEntryManager(nil, nil, nil, keyManager)

		entryKey, err := service.GenerateEntryKey(context.Background(), entryUUID, *dek)

		assert.NoError(t, err)
		assert.Equal(t, entryUUID, entryKey.EntryUUID)
		assert.Equal(t, *kek, entryKey.KEK)
	})

	t.Run("call generate entry key method on the key manager with error", func(t *testing.T) {
		entryUUID := "entry-uuid"
		dek, err := key.NewGeneratedKey()
		if err != nil {
			t.Fatal(err)
		}

		var emptyEntryKey *EntryKey
		var emptyKey key.Key

		keyManager := new(MockEntryKeyer)
		keyManager.On("GenerateEncryptionKey", mock.Anything, entryUUID, *dek, mock.Anything, mock.Anything).
			Return(emptyEntryKey, emptyKey, fmt.Errorf("error"))

		service := NewEntryManager(nil, nil, nil, keyManager)

		entryKey, err := service.GenerateEntryKey(context.Background(), entryUUID, *dek)

		assert.Error(t, err)
		assert.Nil(t, entryKey)
	})
}
