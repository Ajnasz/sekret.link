package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
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
		_ = db.Close()
		return nil, nil, err
	}

	return db, tx, nil
}

func createTestEntryKey(ctx context.Context, tx *sql.Tx) (string, string, error) {
	uid := uuid.New().String()

	entryModel := &EntryModel{}

	_, err := entryModel.CreateEntry(ctx, tx, uid, "text/plain", []byte("test data"))

	if err != nil {
		return "", "", err
	}

	model := &EntryKeyModel{}

	expire := time.Now().Add(time.Hour)
	maxReads := 2
	entryKey, err := model.Create(ctx, tx, uid, []byte("test"), []byte("hash entrykey use tx"), &expire, &maxReads)

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
	_, err = entryModel.CreateEntry(ctx, tx, uid, "text/plain", []byte("test data"))
	if err != nil {
		t.Fatal(err)
	}

	model := &EntryKeyModel{}

	expire := time.Now().Add(time.Hour)
	remainingReads := 2
	entryKey, err := model.Create(ctx, tx, uid, []byte("test"), []byte("hashke"), &expire, &remainingReads)

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
	_, err = entryModel.CreateEntry(ctx, tx, uid, "text/plain", []byte("test data"))
	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
		t.Fatal(err)
	}

	model := &EntryKeyModel{}

	for i := range 10 {
		expire := time.Now().Add(time.Hour)
		maxReads := 2
		_, err = model.Create(ctx, tx, uid, []byte("test"), fmt.Appendf(nil, "hashke %d", i), &expire, &maxReads)

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
		t.Fatalf("expected 10 got %d", len(entryKeys))
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
	defer func() { if err := db.Close(); err != nil { t.Logf("db.Close: %v", err) } }()

	uid, entryKeyUUID, err := createTestEntryKey(ctx, tx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

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
	defer func() { if err := db.Close(); err != nil { t.Logf("db.Close: %v", err) } }()

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

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Errorf("rollback failed: %v", err)
		}
		t.Fatal(err)
	}

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

func Test_EntryKeyModel_Use_Exhausted(t *testing.T) {
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

	_, entryKeyUUID, err := createTestEntryKey(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}

	model := &EntryKeyModel{}
	if err := model.SetMaxReads(ctx, tx, entryKeyUUID, 1); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	// First Use: should succeed (1 → 0)
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := model.Use(ctx, tx, entryKeyUUID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("first Use should succeed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	// Second Use: remaining_reads=0, must return ErrEntryKeyExhausted
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = model.Use(ctx, tx, entryKeyUUID)
	_ = tx.Rollback()
	if !errors.Is(err, ErrEntryKeyExhausted) {
		t.Errorf("expected ErrEntryKeyExhausted, got %v", err)
	}
}

func Test_EntryKeyModel_Use_Concurrent(t *testing.T) {
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close failed: %v", err)
		}
	}()

	setupTx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	_, entryKeyUUID, err := createTestEntryKey(ctx, setupTx)
	if err != nil {
		_ = setupTx.Rollback()
		t.Fatal(err)
	}
	model := &EntryKeyModel{}
	if err := model.SetMaxReads(ctx, setupTx, entryKeyUUID, 1); err != nil {
		_ = setupTx.Rollback()
		t.Fatal(err)
	}
	if err := setupTx.Commit(); err != nil {
		t.Fatal(err)
	}

	// Both goroutines open a transaction, then both call Use() simultaneously.
	// Exactly one must succeed; the other must get ErrEntryKeyExhausted.
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []error
		ready   = make(chan struct{})
	)

	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx, err := db.Begin()
			if err != nil {
				mu.Lock()
				results = append(results, err)
				mu.Unlock()
				return
			}
			<-ready
			useErr := model.Use(ctx, tx, entryKeyUUID)
			if useErr != nil {
				_ = tx.Rollback()
			} else {
				_ = tx.Commit()
			}
			mu.Lock()
			results = append(results, useErr)
			mu.Unlock()
		}()
	}

	close(ready)
	wg.Wait()

	var successes, exhausted int
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrEntryKeyExhausted):
			exhausted++
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}

	if successes != 1 {
		t.Errorf("expected 1 success, got %d", successes)
	}
	if exhausted != 1 {
		t.Errorf("expected 1 ErrEntryKeyExhausted, got %d", exhausted)
	}
}
