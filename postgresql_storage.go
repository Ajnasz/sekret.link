package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type postgresqlStorage struct {
	db *sql.DB
}

func (s postgresqlStorage) Close() error {
	return s.db.Close()
}

func (s postgresqlStorage) Create(UUID string, entry []byte, expire time.Duration) error {
	ctx := context.Background()
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `INSERT INTO entries (uuid, data, created, expire) VALUES  ($1, $2, $3, $4) RETURNING uuid;`, UUID, entry, now, now.Add(expire))
	return err
}

func (s postgresqlStorage) GetMeta(UUID string) (*EntryMeta, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT created, accessed, expire FROM entries WHERE uuid=$1", UUID)

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

func (s postgresqlStorage) Get(UUID string) (*Entry, error) {
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

	meta := EntryMeta{
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

	return &Entry{
		EntryMeta: meta,
		Data:      data,
	}, nil
}

func (s postgresqlStorage) GetAndDelete(UUID string) (*Entry, error) {
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

	_, err = tx.ExecContext(ctx, "DELETE FROM entries WHERE uuid=$1", UUID)

	if err != nil {
		tx.Rollback()
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

	meta := EntryMeta{
		UUID:     UUID,
		Created:  created,
		Accessed: accessed,
		Expire:   expire,
	}

	if meta.IsExpired() {
		return nil, ErrEntryExpired
	}

	return &Entry{
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

func (s postgresqlStorage) DeleteExpired() error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM entries WHERE expire < NOW()")

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func newPostgresqlStorage(psqlconn string) *postgresqlStorage {
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}

	createTable, err := db.Prepare("CREATE TABLE IF NOT EXISTS entries (uuid uuid PRIMARY KEY, data BYTEA, created TIMESTAMPTZ, accessed TIMESTAMPTZ, expire TIMESTAMPTZ)")

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}
	_, err = createTable.Exec()

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}

	return &postgresqlStorage{db}
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
