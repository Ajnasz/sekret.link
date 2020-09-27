// +build sqlite test

package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteStorage struct {
	db *sql.DB
}

func (s sqliteStorage) Close() error {
	return s.db.Close()
}

func (s sqliteStorage) Create(UUID string, entry []byte, expire time.Duration) error {
	ctx := context.Background()
	createStatement, err := s.db.PrepareContext(ctx, "INSERT INTO entries (uuid, data, created, expire) VALUES  (?, ?, ?, ?)")

	if err != nil {
		return err
	}

	now := time.Now()
	_, err = createStatement.Exec(UUID, entry, now, now.Add(expire))

	return err
}

func (s sqliteStorage) GetMeta(UUID string) (*EntryMeta, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT created, accessed, expire FROM entries WHERE uuid=?", UUID)

	var created time.Time
	var accessedNullTime sql.NullTime
	var expireNullTime sql.NullTime
	err = row.Scan(&created, &accessedNullTime, &expireNullTime)

	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, err
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

	meta := &EntryMeta{
		UUID:     UUID,
		Created:  created,
		Accessed: accessed,
		Expire:   expire,
	}

	if meta.IsExpired() {
		return nil, ErrEntryExpired
	}

	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (s sqliteStorage) Get(UUID string) (*Entry, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT data, created, accessed, expire FROM entries WHERE uuid=?", UUID)

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

	_, err = tx.ExecContext(ctx, "UPDATE entries SET accessed=? WHERE uuid=?", time.Now(), UUID)

	if err != nil {
		tx.Rollback()
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

	meta := EntryMeta{
		UUID:     UUID,
		Created:  created,
		Accessed: accessed,
		Expire:   expire,
	}

	if meta.IsExpired() {
		_, err = tx.ExecContext(ctx, "UPDATE entries SET data=?, accessed=? WHERE uuid=?", nil, time.Now(), UUID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		err := tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		return nil, ErrEntryExpired
	}

	err = tx.Commit()

	if err != nil {
		log.Fatal(err)
	}

	return &Entry{
		EntryMeta: meta,
		Data:      data,
	}, nil
}

func (s sqliteStorage) GetAndDelete(UUID string) (*Entry, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT data, created, accessed, expire FROM entries WHERE uuid=?", UUID)

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

	_, err = tx.ExecContext(ctx, "DELETE FROM entries WHERE uuid=?", UUID)

	if err != nil {
		tx.Rollback()
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

	meta := EntryMeta{
		UUID:     UUID,
		Created:  created,
		Accessed: accessed,
		Expire:   expire,
	}

	if meta.IsExpired() {
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

	return &Entry{
		EntryMeta: meta,
		Data:      data,
	}, nil
}

func (s sqliteStorage) Delete(UUID string) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM entries WHERE uuid=?", UUID)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
func (s sqliteStorage) DeleteExpired() error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM entries WHERE expire<?", time.Now())

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func newSQLiteStorage(fileName string) *sqliteStorage {
	db, err := sql.Open("sqlite3", fileName)

	if err != nil {
		log.Fatal(err)
	}

	createTable, err := db.Prepare("CREATE TABLE IF NOT EXISTS entries (uuid TEXT PRIMARY KEY, data BLOB, created TIMESTAMP, accessed TIMESTAMP, expire TIMESTAMP)")

	if err != nil {
		log.Fatal(err)
	}
	_, err = createTable.Exec()

	if err != nil {
		log.Fatal(err)
	}

	return &sqliteStorage{db}
}

type sqliteCleanableStorage struct {
	*sqliteStorage
}

func (s sqliteCleanableStorage) Clean() {
	_, err := s.db.Exec("DELETE FROM entries;")

	if err != nil {
		log.Fatal(err)
	}
}
