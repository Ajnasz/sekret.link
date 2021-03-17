package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/Ajnasz/sekret.link/storage"
	_ "github.com/lib/pq"
)

type postgresqlStorage struct {
	db *sql.DB
}

func (s postgresqlStorage) Close() error {
	return s.db.Close()
}

func (s postgresqlStorage) Create(UUID string, entry []byte, expire time.Duration, remainingReads int) error {
	ctx := context.Background()
	now := time.Now()
	_, deleteKey, err := createKey()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO entries (uuid, data, created, expire, remaining_reads, delete_key) VALUES  ($1, $2, $3, $4, $5, $6) RETURNING uuid, delete_key;`, UUID, entry, now, now.Add(expire), remainingReads, deleteKey)
	return err
}

func (s postgresqlStorage) GetMeta(UUID string) (*storage.EntryMeta, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT created, accessed, expire, remaining_reads, delete_key FROM entries WHERE uuid=$1", UUID)

	var created time.Time
	var accessedNullTime sql.NullTime
	var expireNullTime sql.NullTime
	var remainingReadsNullInt32 sql.NullInt32
	var deleteKeyNullString sql.NullString
	err = row.Scan(&created, &accessedNullTime, &expireNullTime, &remainingReadsNullInt32, &deleteKeyNullString)

	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	var accessed time.Time
	var expire time.Time
	var maxReads int32
	var deleteKey string

	if accessedNullTime.Valid {
		accessed = accessedNullTime.Time
	}
	if expireNullTime.Valid {
		expire = expireNullTime.Time
	}
	if remainingReadsNullInt32.Valid {
		maxReads = remainingReadsNullInt32.Int32
	}
	if deleteKeyNullString.Valid {
		deleteKey = strings.TrimSpace(deleteKeyNullString.String)
	}

	meta := &storage.EntryMeta{
		UUID:      UUID,
		Created:   created,
		Accessed:  accessed,
		Expire:    expire,
		MaxReads:  maxReads,
		DeleteKey: deleteKey,
	}

	if meta.IsExpired() {
		_, err = tx.ExecContext(ctx, "UPDATE entries SET data=$1, accessed=$2 WHERE uuid=$3", nil, time.Now(), UUID)

		if err != nil {
			tx.Rollback()
			return nil, err
		}
		err := tx.Commit()
		if err != nil {
			return nil, err
		}

		return nil, ErrEntryExpired
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (s postgresqlStorage) Get(UUID string) (*storage.Entry, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT data, created, accessed, expire FROM entries WHERE uuid=$1", UUID)

	var data []byte
	var created time.Time
	var accessedNullTime sql.NullTime
	var expireNullTime sql.NullTime
	err = row.Scan(&data, &created, &accessedNullTime, &expireNullTime)

	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}

		return nil, err
	}

	var accessed time.Time
	var expire time.Time

	if accessedNullTime.Valid {
		accessed = accessedNullTime.Time
	}
	if expireNullTime.Valid {
		expire = expireNullTime.Time
	}

	meta := storage.EntryMeta{
		UUID:     UUID,
		Created:  created,
		Accessed: accessed,
		Expire:   expire,
	}

	if meta.IsExpired() {
		_, err = tx.ExecContext(ctx, "UPDATE entries SET data=$1, accessed=$2 WHERE uuid=$3", nil, time.Now(), UUID)

		if err != nil {
			tx.Rollback()
			return nil, err
		}
		err := tx.Commit()
		if err != nil {
			return nil, err
		}

		return nil, ErrEntryExpired
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return &storage.Entry{
		EntryMeta: meta,
		Data:      data,
	}, nil
}

func (s postgresqlStorage) GetAndDelete(UUID string) (*storage.Entry, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT data, created, accessed, expire FROM entries WHERE uuid=$1", UUID)

	var data []byte
	var created time.Time
	var accessedNullTime sql.NullTime
	var expireNullTime sql.NullTime
	err = row.Scan(&data, &created, &accessedNullTime, &expireNullTime)

	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	queries := []string{
		"UPDATE entries SET remaining_reads = remaining_reads - 1 WHERE uuid=$1;",
		"DELETE FROM entries WHERE uuid=$1 AND remaining_reads < 1;",
	}
	for _, query := range queries {
		_, err = tx.ExecContext(ctx, query, UUID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var accessed time.Time
	var expire time.Time

	if accessedNullTime.Valid {
		accessed = accessedNullTime.Time
	}
	if expireNullTime.Valid {
		expire = expireNullTime.Time
	}

	meta := storage.EntryMeta{
		UUID:     UUID,
		Created:  created,
		Accessed: accessed,
		Expire:   expire,
	}

	if meta.IsExpired() {
		return nil, ErrEntryExpired
	}

	return &storage.Entry{
		EntryMeta: meta,
		Data:      data,
	}, nil
}

func (s postgresqlStorage) Delete(UUID string) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM entries WHERE uuid=$1", UUID)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s postgresqlStorage) VerifyDelete(UUID string, deleteKey string) (bool, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	row := tx.QueryRowContext(ctx, "SELECT 1 FROM entries WHERE uuid=$1 AND delete_key=$2;", UUID, deleteKey)

	if row.Err() != nil {
		tx.Rollback()
		return false, row.Err()
	}

	var found bool

	row.Scan(&found)

	if err = tx.Commit(); err != nil {
		return false, err
	}
	return found, nil
}

func (s postgresqlStorage) GetDeleteKey(UUID string) (string, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}

	row := tx.QueryRowContext(ctx, "SELECT delete_key FROM entries WHERE uuid=$1;", UUID)

	if err != nil {
		tx.Rollback()
		return "", err
	}
	var deleteKey string

	row.Scan(&deleteKey)

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return deleteKey, nil
}

func (s postgresqlStorage) DeleteExpired() error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM entries WHERE expire < NOW() OR remaining_reads < 1;")

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

type postgresCleanableStorage struct {
	*postgresqlStorage
}

func (s postgresCleanableStorage) Clean() {
	_, err := s.db.Exec("TRUNCATE entries;")

	if err != nil {
		log.Fatal(err)
	}
}
