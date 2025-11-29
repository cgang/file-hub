package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/uptrace/bun"
)

// UserQuotaModel represents storage quota for a user
type UserQuotaModel struct {
	bun.BaseModel    `bun:"table:user_quota"`
	*model.UserQuota `bun:",inherit"`
}

func wrapQuota(mq *model.UserQuota) *UserQuotaModel {
	return &UserQuotaModel{UserQuota: mq}
}

// GetUserQuota retrieves the storage quota for a user
func GetUserQuota(ctx context.Context, userID int) (*UserQuotaModel, error) {
	quota := &UserQuotaModel{}
	err := db.NewSelect().
		Model(quota).
		Where("user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("quota record not found for user %d", userID)
		}
		return nil, fmt.Errorf("failed to get user quota: %w", err)
	}

	return quota, nil
}

// UpdateUserQuota updates the storage quota for a user
func UpdateUserQuota(ctx context.Context, userID int, totalQuotaBytes int64) error {
	quota := &model.UserQuota{UserID: userID, TotalQuotaBytes: totalQuotaBytes, UpdatedAt: time.Now()}
	result, err := db.NewUpdate().
		Model(wrapQuota(quota)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update user quota: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quota record not found for user %d", userID)
	}

	return nil
}

// GetUserQuotaUsage returns the used bytes for a user
func GetUserQuotaUsage(ctx context.Context, userID int) (int64, error) {
	var usedBytes int64
	err := db.NewSelect().
		Model((*UserQuotaModel)(nil)).
		Column("used_bytes").
		Where("user_id = ?", userID).
		Scan(ctx, &usedBytes)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("quota record not found for user %d", userID)
		}
		return 0, fmt.Errorf("failed to get user quota usage: %w", err)
	}

	return usedBytes, nil
}

// CheckUserQuota checks if a user has enough space for a file of given size
func CheckUserQuota(ctx context.Context, userID int, fileSize int64) (bool, error) {
	quota, err := GetUserQuota(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user quota: %w", err)
	}

	// Check if adding the new file would exceed the quota
	return (quota.UsedBytes + fileSize) <= quota.TotalQuotaBytes, nil
}
