package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	// import postgresql driver
	_ "github.com/lib/pq"

	"github.com/Ajnasz/sekret.link/internal/key"
)

var ErrEntryNotFound = errors.New("entry not found")
var ErrInvalidKey = errors.New("invalid key")
var ErrCreateEntry = errors.New("failed to create entry")

type EntryMeta struct {
	UUID           string
	RemainingReads int
	DeleteKey      string
	Created        time.Time
	Accessed       sql.NullTime
	Expire         time.Time
	ContentType    string
}

// uuid uuid PRIMARY KEY,
// data BYTEA,
// remaining_reads SMALLINT DEFAULT 1,
// delete_key CHAR(256) NOT NULL,
// created TIMESTAMPTZ,
// accessed TIMESTAMPTZ,
// expire TIMESTAMPTZ
type Entry struct {
	EntryMeta
	Data []byte
}

type EntryModel struct {
}

func (e *EntryModel) getDeleteKey() (string, error) {
	k, err := key.NewGeneratedKey()
	if err != nil {
		return "", err
	}
	return k.String(), nil
}

// CreateEntry creates a new entry into the database
func (e *EntryModel) CreateEntry(ctx context.Context, tx *sql.Tx, uuid string, contenType string, data []byte, remainingReads int, expire time.Duration) (*EntryMeta, error) {
	deleteKey, err := e.getDeleteKey()
	if err != nil {
		return nil, errors.Join(err, ErrCreateEntry)
	}

	now := time.Now()
	res, err := tx.ExecContext(ctx, `INSERT INTO entries (uuid, data, created, expire, remaining_reads, delete_key, content_type) VALUES  ($1, $2, $3, $4, $5, $6, $7) RETURNING uuid, delete_key;`, uuid, data, now, now.Add(expire), remainingReads, deleteKey, contenType)

	if err != nil {
		return nil, errors.Join(err, ErrCreateEntry)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, errors.Join(err, ErrCreateEntry)
	}

	if rows != 1 {
		return nil, ErrCreateEntry
	}

	return &EntryMeta{
		UUID:           uuid,
		RemainingReads: remainingReads,
		DeleteKey:      deleteKey,
		Created:        now,
		Expire:         now.Add(expire),
	}, err
}

func (e *EntryModel) Use(ctx context.Context, tx *sql.Tx, uuid string) error {
	_, err := tx.ExecContext(ctx, "UPDATE entries SET accessed = NOW(), remaining_reads = remaining_reads - 1 WHERE uuid = $1 AND remaining_reads > 0", uuid)
	return err
}

// ReadEntry reads a entry from the database
// and updates the read count
func (e *EntryModel) ReadEntry(ctx context.Context, tx *sql.Tx, uuid string) (*Entry, error) {
	row := tx.QueryRow("SELECT uuid, data, remaining_reads, delete_key, created, accessed, expire, content_type FROM entries WHERE uuid=$1 AND remaining_reads > 0 LIMIT 1", uuid)
	var s Entry
	err := row.Scan(&s.UUID, &s.Data, &s.RemainingReads, &s.DeleteKey, &s.Created, &s.Accessed, &s.Expire, &s.ContentType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	return &s, nil
}

func (e *EntryModel) ReadEntryMeta(ctx context.Context, tx *sql.Tx, uuid string) (*EntryMeta, error) {
	row := tx.QueryRow("SELECT created, accessed, expire, remaining_reads, delete_key, content_type FROM entries WHERE uuid=$1 AND remaining_reads > 0 LIMIT 1", uuid)
	var s EntryMeta
	err := row.Scan(&s.Created, &s.Accessed, &s.Expire, &s.RemainingReads, &s.DeleteKey, &s.ContentType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrEntryNotFound
		}
		return nil, err
	}

	s.UUID = uuid

	return &s, nil
}

// DeleteEntry deletes a entry from the database
// if the delete key matches
// returns an error if the delete key does not match
func (e *EntryModel) DeleteEntry(ctx context.Context, tx *sql.Tx, uuid string, deleteKey string) error {
	row := tx.QueryRowContext(ctx, "SELECT delete_key FROM entries WHERE uuid=$1", uuid)

	var storedDeleteKey string
	err := row.Scan(&storedDeleteKey)

	if err != nil {
		return err
	}

	// TODO check how come the storeDeleteKey has a new line
	if strings.TrimSpace(storedDeleteKey) != deleteKey {
		return ErrInvalidKey
	}

	ret, err := tx.ExecContext(ctx, "DELETE FROM entries WHERE uuid=$1 AND delete_key=$2", uuid, deleteKey)

	if err != nil {
		return err
	}

	rows, err := ret.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrEntryNotFound
	}

	return nil
}

func (e *EntryModel) DeleteExpired(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM entries WHERE expire < NOW()")

	return err
}
