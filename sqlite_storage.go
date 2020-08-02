package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func (s *SQLiteStorage) Create(UUID string, entry []byte) error {
	ctx := context.Background()
	createStatement, err := s.db.PrepareContext(ctx, "INSERT INTO entries (uuid, data, created) VALUES  (?, ?, ?)")

	if err != nil {
		return err
	}

	_, err = createStatement.Exec(UUID, entry, time.Now())

	return err
}

func (s *SQLiteStorage) Get(UUID string) ([]byte, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT data FROM entries WHERE uuid=?", UUID)

	var data []byte
	err = row.Scan(&data)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "UPDATE entries SET accessed=? WHERE uuid=?", time.Now(), UUID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		log.Fatal(err)
	}

	return data, nil
}

func (s *SQLiteStorage) GetAndDelete(UUID string) ([]byte, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, "SELECT data FROM entries WHERE uuid=?", UUID)
	var data []byte

	err = row.Scan(&data)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "UPDATE entries SET data=?, accessed=? WHERE uuid=?", nil, time.Now(), UUID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	return data, nil
}

func NewSQLiteStorage(fileName string) *SQLiteStorage {
	db, err := sql.Open("sqlite3", fileName)

	if err != nil {
		log.Fatal(err)
	}

	createTable, err := db.Prepare("CREATE TABLE IF NOT EXISTS entries (uuid TEXT PRIMARY KEY, data BLOB, created TIMESTAMP, accessed TIMESTAMP)")

	if err != nil {
		log.Fatal(err)
	}
	_, err = createTable.Exec()

	if err != nil {
		log.Fatal(err)
	}

	return &SQLiteStorage{db}
}
