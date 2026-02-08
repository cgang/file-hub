package sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
	"github.com/cgang/file-hub/pkg/stor"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	MaxSimpleUploadSize = 10 * 1024 * 1024 // 10MB
	ChunkSize           = 1024 * 1024      // 1MB chunks
	MaxConnectionTime   = 24 * time.Hour
)

type Service struct {
	db *bun.DB
}

func NewService(database *bun.DB) *Service {
	return &Service{
		db: database,
	}
}

func generateVersion() string {
	now := time.Now()
	return fmt.Sprintf("v%d-%d", now.Unix(), now.Nanosecond())
}

func calculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func calculateSHA256Reader(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *Service) GetCurrentVersion(ctx context.Context, repoID int) (*model.RepositoryVersion, error) {
	return db.GetCurrentVersion(ctx, repoID)
}

func (s *Service) ListChanges(ctx context.Context, repoID int, sinceVersion string, maxChanges int) ([]*model.ChangeLog, error) {
	if maxChanges <= 0 || maxChanges > 1000 {
		maxChanges = 100
	}
	return db.GetChangesSince(ctx, repoID, sinceVersion, maxChanges)
}

func (s *Service) GetFileInfo(ctx context.Context, repo *model.Repository, path string, userID int) (*model.FileObject, error) {
	resource := &model.Resource{
		Repo: repo,
		Path: path,
	}
	return stor.GetFileInfo(ctx, resource)
}

func (s *Service) ListDirectory(ctx context.Context, repo *model.Repository, path string, offset, limit int, userID int) ([]*model.FileObject, int64, error) {
	parent, err := db.GetFile(ctx, repo.ID, path)
	if err != nil {
		return nil, 0, err
	}

	files, err := db.GetChildFiles(ctx, parent.ID)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(files))
	if offset >= len(files) {
		return []*model.FileObject{}, total, nil
	}

	end := offset + limit
	if end > len(files) {
		end = len(files)
	}

	result := files[offset:end]
	return result, total, nil
}

func (s *Service) CreateDirectory(ctx context.Context, repo *model.Repository, path string, userID int) error {
	resource := &model.Resource{
		Repo: repo,
		Path: path,
	}

	if err := stor.CreateDir(ctx, resource); err != nil {
		return err
	}

	version := generateVersion()
	change := &model.ChangeLog{
		RepoID:    repo.ID,
		Operation: "create",
		Path:      path,
		UserID:    userID,
		Version:   version,
	}

	if err := db.RecordChange(ctx, change); err != nil {
		return fmt.Errorf("failed to record change: %w", err)
	}

	if err := db.UpdateVersion(ctx, repo.ID, version, "{}"); err != nil {
		return fmt.Errorf("failed to update repository version: %w", err)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, repo *model.Repository, path string, recursive bool, userID int) error {
	file, err := db.GetFile(ctx, repo.ID, path)
	if err != nil {
		return err
	}

	if file.IsDir && recursive {
		children, err := db.GetChildFiles(ctx, file.ID)
		if err != nil {
			return err
		}

		for _, child := range children {
			if deleteErr := s.Delete(ctx, repo, child.Path, true, userID); deleteErr != nil {
				return deleteErr
			}
		}
	}

	resource := &model.Resource{
		Repo: repo,
		Path: path,
	}

	if err := stor.DeleteFile(ctx, resource); err != nil {
		return err
	}

	version := generateVersion()
	change := &model.ChangeLog{
		RepoID:    repo.ID,
		Operation: "delete",
		Path:      path,
		UserID:    userID,
		Version:   version,
	}

	if err := db.RecordChange(ctx, change); err != nil {
		return fmt.Errorf("failed to record change: %w", err)
	}

	if err := db.UpdateVersion(ctx, repo.ID, version, "{}"); err != nil {
		return fmt.Errorf("failed to update repository version: %w", err)
	}

	return nil
}

func (s *Service) Move(ctx context.Context, repo *model.Repository, sourcePath, destPath string, userID int) error {
	srcResource := &model.Resource{
		Repo: repo,
		Path: sourcePath,
	}

	destResource := &model.Resource{
		Repo: repo,
		Path: destPath,
	}

	if err := stor.MoveFile(ctx, srcResource, destResource); err != nil {
		return err
	}

	version := generateVersion()
	change := &model.ChangeLog{
		RepoID:    repo.ID,
		Operation: "move",
		Path:      destPath,
		OldPath:   &sourcePath,
		UserID:    userID,
		Version:   version,
	}

	if err := db.RecordChange(ctx, change); err != nil {
		return fmt.Errorf("failed to record change: %w", err)
	}

	if err := db.UpdateVersion(ctx, repo.ID, version, "{}"); err != nil {
		return fmt.Errorf("failed to update repository version: %w", err)
	}

	return nil
}

func (s *Service) Copy(ctx context.Context, repo *model.Repository, sourcePath, destPath string, userID int) error {
	srcResource := &model.Resource{
		Repo: repo,
		Path: sourcePath,
	}

	destResource := &model.Resource{
		Repo: repo,
		Path: destPath,
	}

	if err := stor.CopyFile(ctx, srcResource, destResource); err != nil {
		return err
	}

	version := generateVersion()
	change := &model.ChangeLog{
		RepoID:    repo.ID,
		Operation: "copy",
		Path:      destPath,
		UserID:    userID,
		Version:   version,
	}

	if err := db.RecordChange(ctx, change); err != nil {
		return fmt.Errorf("failed to record change: %w", err)
	}

	if err := db.UpdateVersion(ctx, repo.ID, version, "{}"); err != nil {
		return fmt.Errorf("failed to update repository version: %w", err)
	}

	return nil
}

func (s *Service) UploadFile(ctx context.Context, repo *model.Repository, path string, data []byte, mimeType string, userID int) (string, string, int64, error) {
	if int64(len(data)) > MaxSimpleUploadSize {
		return "", "", 0, fmt.Errorf("file too large for simple upload, use chunked upload")
	}

	checksum := calculateSHA256(data)

	resource := &model.Resource{
		Repo: repo,
		Path: path,
	}

	if err := stor.PutFile(ctx, resource, nil); err != nil {
		return "", "", 0, err
	}

	version := generateVersion()
	change := &model.ChangeLog{
		RepoID:    repo.ID,
		Operation: "create",
		Path:      path,
		UserID:    userID,
		Version:   version,
	}

	if err := db.RecordChange(ctx, change); err != nil {
		return "", "", 0, fmt.Errorf("failed to record change: %w", err)
	}

	if err := db.UpdateVersion(ctx, repo.ID, version, "{}"); err != nil {
		return "", "", 0, fmt.Errorf("failed to update repository version: %w", err)
	}

	return checksum, version, int64(len(data)), nil
}

func (s *Service) DownloadFile(ctx context.Context, repo *model.Repository, path string, ifNoneMatch string, userID int) (*model.FileObject, io.ReadCloser, error) {
	resource := &model.Resource{
		Repo: repo,
		Path: path,
	}

	file, err := stor.GetFileInfo(ctx, resource)
	if err != nil {
		return nil, nil, err
	}

	if ifNoneMatch != "" && file.Checksum != nil && *file.Checksum == ifNoneMatch {
		return nil, nil, nil
	}

	reader, err := stor.OpenFile(ctx, resource)
	if err != nil {
		return nil, nil, err
	}

	return file, reader, nil
}

func (s *Service) BeginUpload(ctx context.Context, repo *model.Repository, path string, totalSize int64, userID int) (string, []int, error) {
	uploadID := uuid.New().String()
	totalChunks := int((totalSize + ChunkSize - 1) / ChunkSize)

	session := &model.UploadSession{
		UploadID:       uploadID,
		RepoID:         repo.ID,
		Path:           path,
		TotalSize:      totalSize,
		UserID:         userID,
		ChunksUploaded: 0,
		TotalChunks:    totalChunks,
		ExpiresAt:      time.Now().Add(MaxConnectionTime),
		Status:         "active",
	}

	if err := db.CreateUploadSession(ctx, session); err != nil {
		return "", nil, fmt.Errorf("failed to create upload session: %w", err)
	}

	return uploadID, []int{}, nil
}

func (s *Service) UploadChunk(ctx context.Context, uploadID string, chunkIndex int, data []byte) error {
	session, err := db.GetUploadSession(ctx, uploadID)
	if err != nil {
		return fmt.Errorf("upload session not found: %w", err)
	}

	if session.Status != "active" {
		return fmt.Errorf("upload session is not active")
	}

	if time.Now().After(session.ExpiresAt) {
		return fmt.Errorf("upload session has expired")
	}

	checksum := calculateSHA256(data)
	chunk := &model.UploadChunk{
		UploadID:   uploadID,
		ChunkIndex: chunkIndex,
		Offset:     int64(chunkIndex) * ChunkSize,
		Size:       int64(len(data)),
		Checksum:   &checksum,
	}

	if err := db.CreateUploadChunk(ctx, chunk); err != nil {
		return fmt.Errorf("failed to store chunk: %w", err)
	}

	if err := db.IncrementUploadedChunks(ctx, uploadID); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (s *Service) FinalizeUpload(ctx context.Context, uploadID string, repo *model.Repository, userID int) (string, int64, error) {
	session, err := db.GetUploadSession(ctx, uploadID)
	if err != nil {
		return "", 0, fmt.Errorf("upload session not found: %w", err)
	}

	chunks, err := db.GetUploadedChunks(ctx, uploadID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get uploaded chunks: %w", err)
	}

	if len(chunks) != session.TotalChunks {
		return "", 0, fmt.Errorf("not all chunks uploaded: %d/%d", len(chunks), session.TotalChunks)
	}

	if err := db.UpdateUploadSessionStatus(ctx, uploadID, "completed"); err != nil {
		return "", 0, fmt.Errorf("failed to update session status: %w", err)
	}

	fileSize := session.TotalSize
	checksum := "finalized-" + uuid.New().String()[:8]

	version := generateVersion()
	change := &model.ChangeLog{
		RepoID:    session.RepoID,
		Operation: "create",
		Path:      session.Path,
		UserID:    session.UserID,
		Version:   version,
	}

	if err := db.RecordChange(ctx, change); err != nil {
		return "", 0, fmt.Errorf("failed to record change: %w", err)
	}

	if err := db.UpdateVersion(ctx, session.RepoID, version, "{}"); err != nil {
		return "", 0, fmt.Errorf("failed to update repository version: %w", err)
	}

	return checksum, fileSize, nil
}

func (s *Service) CancelUpload(ctx context.Context, uploadID string) error {
	if err := db.UpdateUploadSessionStatus(ctx, uploadID, "cancelled"); err != nil {
		return fmt.Errorf("failed to cancel upload: %w", err)
	}

	if err := db.DeleteUploadSession(ctx, uploadID); err != nil {
		return fmt.Errorf("failed to delete upload session: %w", err)
	}

	return nil
}

func (s *Service) GetSyncStatus(ctx context.Context, repo *model.Repository, path string, clientETag string, clientVersion int64, userID int) (string, *model.FileObject, error) {
	file, err := s.GetFileInfo(ctx, repo, path, userID)
	if err != nil {
		if err.Error() == "file not found" {
			return "new", nil, nil
		}
		return "", nil, err
	}

	if clientETag == "" {
		if file.Checksum != nil {
			return "modified", file, nil
		}
		return "new", file, nil
	}

	if file.Checksum != nil && *file.Checksum == clientETag {
		return "synced", file, nil
	}

	return "modified", file, nil
}
