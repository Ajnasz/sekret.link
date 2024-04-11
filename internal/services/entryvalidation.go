package services

import (
	"time"

	"github.com/Ajnasz/sekret.link/internal/models"
)

func validateEntry(entry *models.Entry) error {
	if entry == nil {
		return ErrEntryNotFound
	}

	if entry.Expire.Before(time.Now()) {
		return ErrEntryExpired
	}

	if entry.RemainingReads <= 0 {
		return ErrEntryExpired
	}

	return nil
}

func validateEntryKey(entryKey *models.EntryKey) error {
	if entryKey == nil {
		return ErrEntryKeyNotFound
	}

	if entryKey.Expire.Valid && entryKey.Expire.Time.Before(time.Now()) {
		return ErrEntryExpired
	}

	if entryKey.RemainingReads.Valid && entryKey.RemainingReads.Int16 <= 0 {
		return ErrEntryNoRemainingReads
	}

	return nil
}
