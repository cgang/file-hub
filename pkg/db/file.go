package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/uptrace/bun"
)

// FileModel represents a file object for database operations
type FileModel struct {
	bun.BaseModel     `bun:"table:files"`
	*model.FileObject `bun:",inherit"`
}

func wrapFile(mo *model.FileObject) *FileModel {
	return &FileModel{FileObject: mo}
}
func newFile(id int) *FileModel {
	return &FileModel{FileObject: &model.FileObject{ID: id}}
}

// CreateFile creates a new file record in the database
func CreateFile(ctx context.Context, file *model.FileObject) error {
	// Set creation timestamp
	file.CreatedAt = time.Now()
	file.UpdatedAt = file.CreatedAt

	_, err := db.NewInsert().Model(wrapFile(file)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetFileByID retrieves a file by ID
func GetFileByID(ctx context.Context, id int) (*model.FileObject, error) {
	file := newFile(id)
	err := db.NewSelect().
		Model(file).
		Where("id = ? AND is_deleted = ?", id, false).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file.FileObject, nil
}

// GetFileByPath retrieves a file by its path
func GetFileByPath(ctx context.Context, path string, userID int) (*model.FileObject, error) {
	file := newFile(0)
	err := db.NewSelect().
		Model(file).
		Where("path = ? AND user_id = ? AND is_deleted = ?", path, userID, false).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file.FileObject, nil
}

// GetFilesByUser retrieves all files for a specific user
func GetFilesByUser(ctx context.Context, userID int) ([]*FileModel, error) {
	var files []*FileModel
	err := db.NewSelect().
		Model(&files).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("updated_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	return files, nil
}

// GetFilesByUserAndPathPrefix retrieves files under a specific path for a user
func GetFilesByUserAndPathPrefix(ctx context.Context, userID int, pathPrefix string) ([]*model.FileObject, error) {
	// Ensure pathPrefix ends with a slash to avoid matching partial directory names
	// For example, if pathPrefix is "/docs/", we don't want to match "/docs-old/"
	if pathPrefix != "/" && pathPrefix[len(pathPrefix)-1] != '/' {
		pathPrefix += "/"
	}

	var files []*FileModel
	err := db.NewSelect().
		Model(&files).
		Where("user_id = ? AND (path = ? OR path LIKE ?) AND is_deleted = ?", userID, pathPrefix, pathPrefix+"%", false).
		Order("path").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	var results []*model.FileObject
	for _, f := range files {
		results = append(results, f.FileObject)
	}

	return results, nil
}

// FileUpdate contains fields that can be updated for a file
type FileUpdate struct {
	MimeType  *string    `json:"mime_type,omitempty"`
	Size      *int64     `json:"size,omitempty"`
	Checksum  *string    `json:"checksum,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	IsDeleted *bool      `json:"is_deleted,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UpdateFile updates a file in the database
func UpdateFile(ctx context.Context, id int, update *FileUpdate) error {
	file := newFile(id)

	// Get the existing file first
	err := db.NewSelect().Model(file).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("file not found")
		}
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Update fields if they are provided
	if update.MimeType != nil {
		file.MimeType = update.MimeType
	}
	if update.Size != nil {
		file.Size = *update.Size
	}
	if update.Checksum != nil {
		file.Checksum = update.Checksum
	}
	if update.IsDeleted != nil {
		file.IsDeleted = *update.IsDeleted
		if update.DeletedAt != nil {
			file.DeletedAt = update.DeletedAt
		} else if *update.IsDeleted {
			now := time.Now()
			file.DeletedAt = &now
		}
	}

	// Always update updated_at
	file.UpdatedAt = time.Now()

	result, err := db.NewUpdate().Model(file).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// DeleteFile deletes a file with the given ID
func DeleteFile(ctx context.Context, id int) error {
	result, err := db.NewDelete().Model((*FileModel)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// DeleteFileByPath marks a file as deleted by path and user
func DeleteFileByPath(ctx context.Context, path string, userID int) error {
	result, err := db.NewDelete().
		Model((*FileModel)(nil)).
		Where("path = ? AND user_id = ?", path, userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}
