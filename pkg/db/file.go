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
	bun.BaseModel `bun:"table:files"`
	*model.FileObject
}

func wrapFile(mo *model.FileObject) *FileModel {
	return &FileModel{FileObject: mo}
}

func unwrapFiles(mos []*FileModel) []*model.FileObject {
	files := make([]*model.FileObject, len(mos))
	for i, mo := range mos {
		files[i] = mo.FileObject
	}
	return files
}

func newFile(id int) *FileModel {
	return &FileModel{FileObject: &model.FileObject{ID: id}}
}

// CreateFile creates a new file record in the database
func CreateFile(ctx context.Context, file *model.FileObject) error {
	// Set creation timestamp
	file.CreatedAt = time.Now()
	if file.UpdatedAt.IsZero() {
		file.UpdatedAt = file.CreatedAt
	}

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
		Where("id = ? AND deleted = ?", id, false).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file.FileObject, nil
}

// GetFile retrieves a file by repository ID and path
func GetFile(ctx context.Context, reposID int, path string) (*model.FileObject, error) {
	file := newFile(0)
	err := db.NewSelect().
		Model(file).
		Where("repo_id = ? AND path = ? AND deleted = ?", reposID, path, false).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return file.FileObject, nil
}

func GetChildFiles(ctx context.Context, parentID int) ([]*model.FileObject, error) {
	var files []*FileModel
	err := db.NewSelect().
		Model(&files).
		Where("parent_id = ? AND deleted = ?", parentID, false).
		Order("name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get child files: %w", err)
	}

	return unwrapFiles(files), nil
}

// GetFilesByUser retrieves all files for a specific user
func GetFilesByUser(ctx context.Context, userID int) ([]*FileModel, error) {
	var files []*FileModel
	err := db.NewSelect().
		Model(&files).
		Where("owner_id = ? AND deleted = ?", userID, false).
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
		Where("owner_id = ? AND (path = ? OR path LIKE ?) AND deleted = ?", userID, pathPrefix, pathPrefix+"%", false).
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

// UpsertFile creates a new file or updates an existing file in the database
// using PostgreSQL's UPSERT functionality based on repo_id and path
func UpsertFile(ctx context.Context, file *model.FileObject) error {
	// Ensure required fields are present
	if file.RepoID == 0 || file.Path == "" {
		return fmt.Errorf("repo_id and path are required for upsert")
	}

	// Set timestamps
	now := time.Now()
	if file.CreatedAt.IsZero() {
		file.CreatedAt = now
	}
	file.UpdatedAt = now

	// Use PostgreSQL 15+ MERGE command via bun's builder
	_, err := db.NewInsert().Model(wrapFile(file)).
		On("CONFLICT (repo_id, path) DO UPDATE").
		Set("mod_time = ?", file.ModTime).
		Set("size = ?", file.Size).
		Set("updated_at = ?", now).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to upsert file: %w", err)
	}

	return nil
}

// DeleteFileByPath marks a file as deleted by path and user
func DeleteFileByPath(ctx context.Context, repoID int, path string) error {
	result, err := db.NewDelete().
		Model((*FileModel)(nil)).
		Where("repo_id = ? AND path = ?", repoID, path).
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

// UpdateContentType updates content type of specified objects in database.
func UpdateContentType(ctx context.Context, objects []*model.FileObject) error {
	// TODO add batch implementation
	return nil
}
