package db

import (
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
)

// GetUserQuota retrieves the storage quota for a user
func (d *DB) GetUserQuota(userID int) (*UserQuota, error) {
	quota := &UserQuota{}
	err := d.Model(quota).
		Where("user_id = ?user_id", userID).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, fmt.Errorf("quota record not found for user %d", userID)
		}
		return nil, fmt.Errorf("failed to get user quota: %w", err)
	}

	return quota, nil
}

// UpdateUserQuota updates the storage quota for a user
func (d *DB) UpdateUserQuota(userID int, totalQuotaBytes int64) error {
	quota := &UserQuota{UserID: userID, TotalQuotaBytes: totalQuotaBytes, UpdatedAt: time.Now()}
	result, err := d.Model(quota).
		Where("user_id = ?user_id", userID).
		Update()

	if err != nil {
		return fmt.Errorf("failed to update user quota: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("quota record not found for user %d", userID)
	}

	return nil
}

// GetUserQuotaUsage returns the used bytes for a user
func (d *DB) GetUserQuotaUsage(userID int) (int64, error) {
	var usedBytes int64
	err := d.Model((*UserQuota)(nil)).
		Column("used_bytes").
		Where("user_id = ?user_id", userID).
		Select(&usedBytes)

	if err != nil {
		if err == pg.ErrNoRows {
			return 0, fmt.Errorf("quota record not found for user %d", userID)
		}
		return 0, fmt.Errorf("failed to get user quota usage: %w", err)
	}

	return usedBytes, nil
}

// CheckUserQuota checks if a user has enough space for a file of given size
func (d *DB) CheckUserQuota(userID int, fileSize int64) (bool, error) {
	quota, err := d.GetUserQuota(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user quota: %w", err)
	}

	// Check if adding the new file would exceed the quota
	return (quota.UsedBytes + fileSize) <= quota.TotalQuotaBytes, nil
}