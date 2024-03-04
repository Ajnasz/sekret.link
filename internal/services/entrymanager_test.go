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

var timenow = time.Now()

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
	entryModel := new(MockEntryModel)
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

	entryModel := new(MockEntryModel)
	entryModel.
		On("CreateEntry", ctx, mock.Anything, mock.Anything, encryptedData, 1, mock.Anything).
		Return(&models.EntryMeta{}, fmt.Errorf("error"))

	entryCrypto := new(MockEntryCrypto)
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

	entryModel := new(MockEntryModel)
	entryModel.
		On("ReadEntry", ctx, mock.Anything, "uuid").
		Return(&models.Entry{
			UUID:           "uuid",
			Data:           []byte("encrypted"),
			RemainingReads: 1,
			DeleteKey:      "delete_key",
			Created:        timenow,
			Accessed:       sql.NullTime{Time: timenow, Valid: true},
			Expire:         timenow.Add(time.Minute),
		}, nil)
	entryModel.
		On("UpdateAccessed", ctx, mock.Anything, "uuid").
		Return(nil)

	entryCrypto := new(MockEntryCrypto)

	entryCrypto.On("Decrypt", []byte("encrypted")).Return([]byte("data"), nil)

	service := NewEntryManager(db, entryModel, entryCrypto)
	data, err := service.ReadEntry(ctx, "uuid")

	assert.NoError(t, err)
	assert.NotNil(t, data)

	entryModel.AssertExpectations(t)
}
