package models

import (
	"context"
	"database/sql"
	"time"
)

type EntryKey struct {
	UUID           string
	EntryUUID      string
	EncryptedKey   []byte
	KeyHash        []byte
	Created        time.Time
	Expire         sql.NullTime
	RemainingReads sql.NullInt16
}

type EntryKeyModel struct{}

func (e *EntryKeyModel) Create(ctx context.Context, tx *sql.Tx, entryUUID string, encryptedKey []byte, hash []byte) (*EntryKey, error) {

	now := time.Now()
	res := tx.QueryRowContext(ctx, `
		INSERT INTO entry_key (uuid, entry_uuid, encrypted_key, key_hash, created)
		VALUES (gen_random_uuid(), $1, $2, $3, $4) RETURNING uuid, created;
	`, entryUUID, encryptedKey, hash, now)

	var uid string
	var created time.Time

	err := res.Scan(&uid, &created)

	if err != nil {
		return nil, err

	}

	return &EntryKey{
		UUID:         uid,
		EntryUUID:    entryUUID,
		EncryptedKey: encryptedKey,
		KeyHash:      hash,
		Created:      now,
	}, err
}

func (e *EntryKeyModel) Get(ctx context.Context, tx *sql.Tx, entryUUID string) ([]EntryKey, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT uuid, entry_uuid, encrypted_key, key_hash, created, expire, remaining_reads
		FROM entry_key
		WHERE entry_uuid = $1
		AND (expire IS NULL OR expire > NOW());
	`, entryUUID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var entryKeys []EntryKey

	for rows.Next() {
		var ek EntryKey
		err := rows.Scan(&ek.UUID, &ek.EntryUUID, &ek.EncryptedKey, &ek.KeyHash, &ek.Created, &ek.Expire, &ek.RemainingReads)
		if err != nil {
			return nil, err
		}

		entryKeys = append(entryKeys, ek)
	}

	return entryKeys, nil

}

// GetByUUID returns the entry key by its UUID
func (e *EntryKeyModel) Delete(ctx context.Context, tx *sql.Tx, uuid string) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM entry_key
		WHERE uuid = $1
	`, uuid)

	return err
}

func (e *EntryKeyModel) DeleteExpired(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM entry_key
		WHERE expire < NOW()
	`)

	return err
}

// SetExpire sets the expire time for the entry key
func (e *EntryKeyModel) SetExpire(ctx context.Context, tx *sql.Tx, uuid string, expire time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE entry_key
		SET expire = $1
		WHERE uuid = $2
	`, expire, uuid)

	return err
}

func (e *EntryKeyModel) SetMaxReads(ctx context.Context, tx *sql.Tx, uuid string, maxReads int) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE entry_key
		SET remaining_reads = $1
		WHERE uuid = $2
	`, maxReads, uuid)

	return err
}

func (e *EntryKeyModel) Use(ctx context.Context, tx *sql.Tx, uuid string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE entry_key
		SET remaining_reads = remaining_reads - 1
		WHERE uuid = $1 AND remaining_reads IS NOT NULL
	`, uuid)

	return err
}
