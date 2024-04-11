package models

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/internal/test/durable"
	"github.com/google/uuid"
)

func Test_EntryModel_CreateEntry(t *testing.T) {
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	uid := uuid.New().String()
	data := []byte("test data")
	remainingReads := 2
	expire := time.Hour * 24

	model := &EntryModel{}

	meta, err := model.CreateEntry(ctx, tx, uid, data, remainingReads, expire)
	if err != nil {
		t.Fatal(err)
	}

	if meta.RemainingReads != 2 {
		t.Errorf("expected %d got %d", remainingReads, meta.RemainingReads)
	}

	if meta.UUID != uid {
		t.Errorf("expected %s got %s", uid, meta.UUID)
	}

	if meta.DeleteKey == "" {
		t.Errorf("expected delete key to be set")
	}

	if meta.Created.IsZero() {
		t.Errorf("expected created to be set")
	}

	if meta.Expire.IsZero() {
		t.Errorf("expected expire to be set")
	}

	if meta.Accessed.Valid {
		t.Errorf("expected accessed not to be set")
	}

	if err := tx.Rollback(); err != nil {
		t.Fatal(errors.Join(err, errors.New("failed to rollback transaction")))
	}
}

func Test_EntryModel_Use(t *testing.T) {
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	uid := uuid.New().String()
	data := []byte("test data")
	remainingReads := 2
	expire := time.Hour * 24

	model := &EntryModel{}

	meta, err := model.CreateEntry(ctx, tx, uid, data, remainingReads, expire)
	if err != nil {
		t.Fatal(err)
	}

	if err := model.Use(ctx, tx, uid); err != nil {
		t.Fatal(errors.Join(err, errors.New("failed to access entry")))
	}

	entry, err := model.ReadEntry(ctx, tx, uid)
	if err != nil {
		t.Fatal(errors.Join(err, errors.New("failed to read entry")))
	}

	if entry.RemainingReads != 1 {
		t.Errorf("expected %d got %d", 0, entry.RemainingReads)
	}

	if !entry.Accessed.Valid {
		t.Errorf("expected accessed to be set")
	}

	if string(entry.Data) != string(data) {
		t.Errorf("expected %s got %s", string(data), string(entry.Data))
	}

	if err := model.DeleteEntry(ctx, tx, uid, "invalid delete key"); err != ErrInvalidKey {
		t.Fatal("expected error when deleting with invalid delete key")
	}

	if err := model.DeleteEntry(ctx, tx, uid, meta.DeleteKey); err != nil {
		t.Fatal(errors.Join(err, errors.New("failed to delete entry")))
	}

	if err := tx.Rollback(); err != nil {
		t.Fatal(errors.Join(err, errors.New("failed to rollback transaction")))
	}
}
