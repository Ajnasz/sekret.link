package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Ajnasz/sekret.link/entries"
	"github.com/Ajnasz/sekret.link/key"

	// Register postgresql driver
	_ "github.com/lib/pq"
)

// Storage stores secrets in postgresql
type Storage struct {
	db *sql.DB
}

// Close closes connection to the database
func (s Storage) Close() error {
	return s.db.Close()
}

// Write stores a new entry in database
func (s Storage) Write(ctx context.Context, UUID string, entry []byte, expire time.Duration, remainingReads int) (*entries.EntryMeta, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	if err = s.write(tx, UUID, entry, expire, remainingReads); err != nil {
		tx.Rollback()
		return nil, err
	}

	meta, err := s.readMeta(tx, UUID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return meta, entries.ErrEntryNotFound
		}
		return nil, err
	}

	if meta.IsExpired() {
		s.setAccessed(tx, UUID)
		tx.Commit()
		return nil, entries.ErrEntryExpired
	}

	tx.Commit()
	return meta, nil
}

func (s Storage) write(tx *sql.Tx, UUID string, entry []byte, expire time.Duration, remainingReads int) error {
	now := time.Now()
	k, err := key.NewGeneratedKey()
	if err != nil {
		return err
	}
	deleteKey := k.ToHex()

	_, err = tx.Exec(`INSERT INTO entries (uuid, data, created, expire, remaining_reads, delete_key) VALUES  ($1, $2, $3, $4, $5, $6) RETURNING uuid, delete_key;`, UUID, entry, now, now.Add(expire), remainingReads, deleteKey)

	return err
}

func (s Storage) setAccessed(tx *sql.Tx, UUID string) error {
	if _, err := tx.Exec("UPDATE entries SET accessed=$1 WHERE uuid=$2", time.Now(), UUID); err != nil {
		return err
	}

	return nil
}

func (s Storage) readMeta(tx *sql.Tx, UUID string) (*entries.EntryMeta, error) {
	row := tx.QueryRow(`
	SELECT
		created,
		accessed,
		expire,
		remaining_reads,
		delete_key
	FROM
		entries
	WHERE
		uuid=$1
		AND remaining_reads > 0
		`, UUID)

	var created time.Time
	var accessedNullTime sql.NullTime
	var expireNullTime sql.NullTime
	var remainingReadsNullInt32 sql.NullInt32
	var deleteKeyNullString sql.NullString
	err := row.Scan(&created, &accessedNullTime, &expireNullTime, &remainingReadsNullInt32, &deleteKeyNullString)

	if err != nil {
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

	meta := &entries.EntryMeta{
		UUID:      UUID,
		Created:   created,
		Accessed:  accessed,
		Expire:    expire,
		MaxReads:  maxReads,
		DeleteKey: deleteKey,
	}

	return meta, nil
}

func (s Storage) updateReadCount(tx *sql.Tx, UUID string) error {
	_, err := tx.Exec("UPDATE entries SET remaining_reads = remaining_reads - 1 WHERE uuid=$1;", UUID)
	return err
}

// read to get entry including the actual secret
// returns the data if the secret not expired yet
// updates read count
func (s Storage) read(tx *sql.Tx, UUID string) (*entries.Entry, error) {

	row := tx.QueryRow(`
		SELECT data, created, accessed, expire, remaining_reads, delete_key
		FROM entries
		WHERE uuid=$1
		AND remaining_reads > 0
		LIMIT 1`, UUID)

	var data []byte
	var created time.Time
	var accessedNullTime sql.NullTime
	var expireNullTime sql.NullTime
	var remainingReadsNullInt32 sql.NullInt32
	var deleteKeyNullString sql.NullString
	if err := row.Scan(&data, &created, &accessedNullTime, &expireNullTime, &remainingReadsNullInt32, &deleteKeyNullString); err != nil {
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

	meta := entries.EntryMeta{
		UUID:      UUID,
		Created:   created,
		Accessed:  accessed,
		Expire:    expire,
		MaxReads:  maxReads,
		DeleteKey: deleteKey,
	}

	return &entries.Entry{
		EntryMeta: meta,
		Data:      data,
	}, nil
}

// ReadConfirm to get entry including the actual secret then delete it
// returns the data if the secret not expired yet
// returns a bool channel which expects a true sent to so the read is confirmed
// if the message is false, the secret will be restored
// if the message is true, the data will be deleted
// updates read count
func (s Storage) ReadConfirm(ctx context.Context, UUID string) (*entries.Entry, chan bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	confirmChan := make(chan bool)
	entry, err := s.read(tx, UUID)

	if err != nil {
		tx.Rollback()
		close(confirmChan)
		if err == sql.ErrNoRows {
			return nil, nil, entries.ErrEntryNotFound
		}
		return nil, nil, err
	}

	if entry.IsExpired() {
		s.setAccessed(tx, UUID)
		tx.Commit()
		close(confirmChan)
		return nil, nil, entries.ErrEntryExpired
	}

	if err := s.updateReadCount(tx, UUID); err != nil {
		tx.Rollback()
		close(confirmChan)
		return nil, nil, err
	}

	go func() {
		select {
		case confirmed := <-confirmChan:
			if confirmed {
				if err := tx.Commit(); err != nil {
					fmt.Println(err)
				}
			} else {
				if err := tx.Rollback(); err != nil {
					fmt.Println(err)
				}
			}
		case <-ctx.Done():
			tx.Rollback()
		}
		close(confirmChan)
	}()

	return entry, confirmChan, nil
}

// Delete deletes the entry from the database
func (s Storage) Delete(ctx context.Context, UUID string) error {
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

// VerifyDelete returns true if the given deleteKey belongs to the given UUID
func (s Storage) VerifyDelete(ctx context.Context, UUID string, deleteKey string) (bool, error) {
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

// GetDeleteKey returns the delete key for the uuid
func (s Storage) GetDeleteKey(ctx context.Context, UUID string) (string, error) {
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

// DeleteExpired removes expired entries from the database
func (s Storage) DeleteExpired(ctx context.Context) error {
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
