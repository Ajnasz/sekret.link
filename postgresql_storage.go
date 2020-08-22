package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type PostgresqlStorage struct {
	db *sql.DB
}

func (s *PostgresqlStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresqlStorage) Create(UUID string, entry []byte) error {
	ctx := context.Background()
	_, err := s.db.ExecContext(ctx, `INSERT INTO entries (uuid, data, created) VALUES  ($1, $2, $3) RETURNING uuid;`, UUID, entry, time.Now())
	return err
}

func (s *PostgresqlStorage) GetMeta(UUID string) (*EntryMeta, error) {
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
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "UPDATE entries SET accessed=$1 WHERE uuid=$2", time.Now(), UUID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		log.Fatal(err)
	}

	var accessed time.Time
	var expire time.Time

	if accessedNullTime.Valid {
		accessed = accessedNullTime.Time
	}
	if expireNullTime.Valid {
		expire = expireNullTime.Time
	}

	return &EntryMeta{
		UUID:     UUID,
		Created:  created,
		Accessed: accessed,
		Expire:   expire,
	}, nil
}

func (s *PostgresqlStorage) Get(UUID string) (*Entry, error) {
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
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "UPDATE entries SET accessed=$1 WHERE uuid=$2", time.Now(), UUID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		log.Fatal(err)
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

	return &Entry{
		EntryMeta: meta,
		Data:      data,
	}, nil
}

func (s *PostgresqlStorage) GetAndDelete(UUID string) (*Entry, error) {
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
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "UPDATE entries SET data=$1, accessed=$2 WHERE uuid=$3", nil, time.Now(), UUID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
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

	return &Entry{
		EntryMeta: meta,
		Data:      data,
	}, nil
}

func NewPostgresqlStorage(psqlconn string) *PostgresqlStorage {
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}

	createTable, err := db.Prepare("CREATE TABLE IF NOT EXISTS entries (uuid uuid PRIMARY KEY, data BYTEA, created TIMESTAMP, accessed TIMESTAMP, expire TIMESTAMP)")

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}
	_, err = createTable.Exec()

	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}

	return &PostgresqlStorage{db}
}
