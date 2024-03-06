package services

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/test"
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
	entryModel := new(test.MockEntryModel)
	entryModel.
		On("CreateEntry", ctx, mock.Anything, mock.Anything, encryptedData, 1, mock.Anything).
		Return(&models.EntryMeta{
			UUID:           "uuid",
			RemainingReads: 1,
			DeleteKey:      "delete_key",
			Created:        timenow,
			Expire:         timenow.Add(time.Minute),
		}, nil)

	entryCrypto := new(test.MockEntryCrypto)
	entryCrypto.On("Encrypt", data).Return(encryptedData, nil)

	service := NewEntryManager(db, entryModel, entryCrypto)
	meta, err := service.CreateEntry(ctx, data, 1, time.Minute)

	assert.NoError(t, err)
	assert.NotNil(t, meta)

	entryModel.AssertExpectations(t)
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

	entryModel := new(test.MockEntryModel)
	entryModel.
		On("CreateEntry", ctx, mock.Anything, mock.Anything, encryptedData, 1, mock.Anything).
		Return(&models.EntryMeta{}, fmt.Errorf("error"))

	entryCrypto := new(test.MockEntryCrypto)
	entryCrypto.On("Encrypt", data).Return(encryptedData, nil)

	service := NewEntryManager(db, entryModel, entryCrypto)
	meta, err := service.CreateEntry(ctx, data, 1, time.Minute)

	assert.Error(t, err)
	assert.Nil(t, meta)

	entryModel.AssertExpectations(t)
}

func TestReadEntry(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery("SELECT uuid, data, remaining_reads, delete_key, created, accessed, expire FROM entries").
		WillReturnRows(sqlmock.NewRows([]string{"uuid", "data", "remaining_reads", "delete_key", "created", "accessed", "expire"}).AddRow("uuid", "data", 1, "delete_key", timenow, timenow, timenow.Add(time.Minute)))
	sqlMock.ExpectExec("UPDATE entries SET accessed = NOW\\(\\), remaining_reads = remaining_reads - 1 WHERE uuid =").
		WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	ctx := context.Background()

	entry := models.Entry{
		UUID:           "uuid",
		Data:           []byte("encrypted"),
		RemainingReads: 1,
		DeleteKey:      "delete_key",
		Created:        timenow,
		Accessed:       sql.NullTime{Time: timenow, Valid: true},
		Expire:         timenow.Add(time.Minute),
	}

	entryModel := new(test.MockEntryModel)
	entryModel.
		On("ReadEntry", ctx, mock.Anything, "uuid").
		Return(&entry, nil)
	entryModel.
		On("UpdateAccessed", ctx, mock.Anything, "uuid").
		Return(nil)

	entryCrypto := new(test.MockEntryCrypto)

	entryCrypto.On("Decrypt", []byte("encrypted")).Return([]byte("data"), nil)

	service := NewEntryManager(db, entryModel, entryCrypto)
	data, err := service.ReadEntry(ctx, "uuid")

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

	entryModel := new(test.MockEntryModel)
	entryModel.
		On("ReadEntry", ctx, mock.Anything, "uuid").
		Return(&models.Entry{}, fmt.Errorf("error"))

	entryCrypto := new(test.MockEntryCrypto)

	service := NewEntryManager(db, entryModel, entryCrypto)
	data, err := service.ReadEntry(ctx, "uuid")

	assert.Error(t, err)
	assert.Nil(t, data)

	entryModel.AssertExpectations(t)
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

	entryModel := new(test.MockEntryModel)
	entryModel.
		On("DeleteEntry", ctx, mock.Anything, "uuid", "delete_key").
		Return(nil)

	entryCrypto := new(test.MockEntryCrypto)

	service := NewEntryManager(db, entryModel, entryCrypto)
	err = service.DeleteEntry(ctx, "uuid", "delete_key")

	assert.NoError(t, err)

	entryModel.AssertExpectations(t)
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

	entryModel := new(test.MockEntryModel)
	entryModel.
		On("DeleteEntry", ctx, mock.Anything, "uuid", "delete_key").
		Return(fmt.Errorf("error"))

	entryCrypto := new(test.MockEntryCrypto)

	service := NewEntryManager(db, entryModel, entryCrypto)
	err = service.DeleteEntry(ctx, "uuid", "delete_key")

	assert.Error(t, err)

	entryModel.AssertExpectations(t)
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

	entryModel := new(test.MockEntryModel)
	entryModel.
		On("DeleteEntry", ctx, mock.Anything, "uuid", "delete_key").
		Return(models.ErrEntryNotFound)

	entryCrypto := new(test.MockEntryCrypto)

	service := NewEntryManager(db, entryModel, entryCrypto)
	err = service.DeleteEntry(ctx, "uuid", "delete_key")

	assert.Error(t, err)
	assert.Equal(t, models.ErrEntryNotFound, err)

	entryModel.AssertExpectations(t)

	if sqlMock.ExpectationsWereMet() != nil {
		t.Errorf("there were unfulfilled expectations: %s", sqlMock.ExpectationsWereMet())
	}
}
