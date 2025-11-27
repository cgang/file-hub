package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CreateFile creates a new file record in the database
func (d *DB) CreateFile(file *File) error {
	// Set creation timestamp
	file.CreatedAt = time.Now()
	file.UpdatedAt = file.CreatedAt

	_, err := d.NewInsert().Model(file).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetFileByID retrieves a file by ID
func (d *DB) GetFileByID(id int) (*File, error) {
	file := &File{ID: id}
	err := d.NewSelect().
		Model(file).
		Where("id = ? AND is_deleted = ?", id, false).
		Scan(context.Background())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

// GetFileByPath retrieves a file by its path
func (d *DB) GetFileByPath(path string, userID int) (*File, error) {
	file := &File{}
	err := d.NewSelect().
		Model(file).
		Where("path = ? AND user_id = ? AND is_deleted = ?", path, userID, false).
		Scan(context.Background())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

// GetFilesByUser retrieves all files for a specific user
func (d *DB) GetFilesByUser(userID int) ([]*File, error) {
	var files []*File
	err := d.NewSelect().
		Model(&files).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("updated_at DESC").
		Scan(context.Background())

	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	return files, nil
}

// GetFilesByUserAndPathPrefix retrieves files under a specific path for a user
func (d *DB) GetFilesByUserAndPathPrefix(userID int, pathPrefix string) ([]*File, error) {
	// Ensure pathPrefix ends with a slash to avoid matching partial directory names
	// For example, if pathPrefix is "/docs/", we don't want to match "/docs-old/"
	if pathPrefix != "/" && pathPrefix[len(pathPrefix)-1] != '/' {
		pathPrefix += "/"
	}

	var files []*File
	err := d.NewSelect().
		Model(&files).
		Where("user_id = ? AND (path = ? OR path LIKE ?) AND is_deleted = ?", userID, pathPrefix, pathPrefix+"%", false).
		Order("path").
		Scan(context.Background())

	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	return files, nil
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
func (d *DB) UpdateFile(id int, update *FileUpdate) error {
	file := &File{ID: id}

	// Get the existing file first
	err := d.NewSelect().Model(file).Where("id = ?", id).Scan(context.Background())
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

	result, err := d.NewUpdate().Model(file).Where("id = ?", id).Exec(context.Background())
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

// DeleteFile marks a file as deleted (soft delete)
func (d *DB) DeleteFile(id int) error {
	now := time.Now()
	file := &File{ID: id, IsDeleted: true, DeletedAt: &now, UpdatedAt: now}
	result, err := d.NewUpdate().Model(file).Where("id = ?", id).Exec(context.Background())

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
func (d *DB) DeleteFileByPath(path string, userID int) error {
	now := time.Now()
	file := &File{IsDeleted: true, DeletedAt: &now, UpdatedAt: now}
	result, err := d.NewUpdate().
		Model(file).
		Where("path = ? AND user_id = ?", path, userID).
		Exec(context.Background())

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