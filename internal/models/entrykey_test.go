package models

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/internal/test/durable"
	"github.com/google/uuid"
)

func getTestDbTx(ctx context.Context) (*sql.DB, *sql.Tx, error) {
	db, err := durable.TestConnection(ctx)

	if err != nil {
		return nil, nil, err
	}

	tx, err := db.Begin()

	if err != nil {
		defer db.Close()
		return nil, nil, err
	}

	return db, tx, nil
}

func createTestEntryKey(ctx context.Context, tx *sql.Tx) (string, string, error) {
	uid := uuid.New().String()

	entryModel := &EntryModel{}

	_, err := entryModel.CreateEntry(ctx, tx, uid, []byte("test data"), 2, 3600)

	if err != nil {
		return "", "", err
	}

	model := &EntryKeyModel{}

	entryKey, err := model.Create(ctx, tx, uid, []byte("test"), []byte("hash entrykey use tx"))

	if err != nil {
		return "", "", err
	}

	return uid, entryKey.UUID, nil
}

func Test_EntryKeyModel_Create(t *testing.T) {
	ctx := context.Background()
	db, tx, err := getTestDbTx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	uid := uuid.New().String()

	entryModel := &EntryModel{}
	_, err = entryModel.CreateEntry(ctx, tx, uid, []byte("test data"), 2, 3600)
	if err != nil {
		t.Fatal(err)
	}

	model := &EntryKeyModel{}

	entryKey, err := model.Create(ctx, tx, uid, []byte("test"), []byte("hashke"))

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	if entryKey.UUID == "" {
		t.Error("expected uuid to be set")
	}

	if entryKey.Created.IsZero() {
		t.Error("expected created to be set")
	}

	if entryKey.EntryUUID != uid {
		t.Errorf("expected %s got %s", uid, entryKey.EntryUUID)
	}

	if entryKey.EncryptedKey == nil {
		t.Error("expected encrypted data to be set")
	}

	if entryKey.KeyHash == nil {
		t.Error("expected encrypted key to be set")
	}

}

func Test_EntryKeyModel_Get(t *testing.T) {
	ctx := context.Background()
	db, tx, err := getTestDbTx(ctx)

	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	uid := uuid.New().String()

	entryModel := &EntryModel{}
	_, err = entryModel.CreateEntry(ctx, tx, uid, []byte("test data"), 2, 3600)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	model := &EntryKeyModel{}

	for i := 0; i < 10; i++ {
		_, err = model.Create(ctx, tx, uid, []byte("test"), []byte(fmt.Sprintf("hashke %d", i)))

		if err != nil {
			if err := tx.Rollback(); err != nil {
				t.Error(err)
			}
			t.Fatal(err)
		}

	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	entryKeys, err := model.Get(ctx, tx, uid)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	if len(entryKeys) != 10 {
		t.Fatalf("expected 1 got %d", len(entryKeys))
	}

	if entryKeys[0].EntryUUID != uid {
		t.Errorf("expected %s got %s", uid, entryKeys[0].EntryUUID)
	}

	if entryKeys[0].EncryptedKey == nil {
		t.Error("expected encrypted data to be set")
	}

	if entryKeys[0].KeyHash == nil {
		t.Error("expected encrypted key to be set")
	}
}

func Test_EntryKeyModel_Get_Empty(t *testing.T) {
	ctx := context.Background()
	db, tx, err := getTestDbTx(ctx)

	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	model := &EntryKeyModel{}

	entryKeys, err := model.Get(ctx, tx, uuid.New().String())

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	if len(entryKeys) != 0 {
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
		t.Fatalf("expected 0 got %d", len(entryKeys))
	}
}

func Test_EntryKeyModel_Delete(t *testing.T) {
	ctx := context.Background()
	db, tx, err := getTestDbTx(ctx)

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	uid, entryKeyUUID, err := createTestEntryKey(ctx, tx)

	model := &EntryKeyModel{}

	err = model.Delete(ctx, tx, entryKeyUUID)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	tx, err = db.Begin()

	if err != nil {
		t.Fatal(err)
	}

	entryKeys, err := model.Get(ctx, tx, uid)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	if len(entryKeys) != 0 {
		t.Fatalf("expected 0 got %d", len(entryKeys))
	}
}

func Test_EntryKeyModel_Delete_Empty(t *testing.T) {
	ctx := context.Background()
	db, tx, err := getTestDbTx(ctx)

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	model := &EntryKeyModel{}
	err = model.Delete(ctx, tx, uuid.New().String())

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func Test_EntryKeyModel_SetExpire(t *testing.T) {
	ctx := context.Background()
	db, tx, err := getTestDbTx(ctx)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	model := &EntryKeyModel{}

	uid, entryKeyUUID, err := createTestEntryKey(ctx, tx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	err = model.SetExpire(ctx, tx, entryKeyUUID, time.Now().Add(time.Hour))

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	tx, err = db.Begin()

	if err != nil {
		t.Fatal(err)
	}

	entryKeys, err := model.Get(ctx, tx, uid)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("rollback failed: %v", err)
	}

	if len(entryKeys) != 1 {
		t.Fatalf("expected 1 got %d", len(entryKeys))
	}

	if entryKeys[0].Expire.Time.IsZero() {
		t.Error("expected expire to be set")
	}

	if entryKeys[0].Expire.Time.Before(time.Now()) {
		t.Error("expected expire to be in the future")
	}
}

func Test_EntryKeyModel_SetExpire_Empty(t *testing.T) {
	ctx := context.Background()
	db, tx, err := getTestDbTx(ctx)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	model := &EntryKeyModel{}

	err = model.SetExpire(ctx, tx, uuid.New().String(), time.Now().Add(time.Hour))

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}
}

func Test_EntryKeyModel_UseTx(t *testing.T) {
	ctx := context.Background()
	db, tx, err := getTestDbTx(ctx)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	uid, entryKeyUUID, err := createTestEntryKey(ctx, tx)

	model := &EntryKeyModel{}

	if err := model.SetMaxReads(ctx, tx, entryKeyUUID, 2); err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	err = model.Use(ctx, tx, entryKeyUUID)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	tx, err = db.Begin()

	if err != nil {
		t.Fatal(err)
	}

	entryKeys, err := model.Get(ctx, tx, uid)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	if len(entryKeys) != 1 {
		t.Fatalf("expected 1 got %d", len(entryKeys))
	}

	if entryKeys[0].RemainingReads.Int16 != 1 {
		t.Errorf("expected 1 got %d", entryKeys[0].RemainingReads.Int16)
	}
}
