package db

import (
	"context"
	"fmt"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/uptrace/bun"
)

type ChangeLogModel struct {
	bun.BaseModel `bun:"table:change_log"`
	*model.ChangeLog
}

type RepositoryVersionModel struct {
	bun.BaseModel `bun:"table:repository_versions"`
	*model.RepositoryVersion
}

type UploadSessionModel struct {
	bun.BaseModel `bun:"table:upload_sessions"`
	*model.UploadSession
}

type UploadChunkModel struct {
	bun.BaseModel `bun:"table:upload_chunks"`
	*model.UploadChunk
}

func wrapChangeLog(cl *model.ChangeLog) *ChangeLogModel {
	if cl == nil {
		return nil
	}
	return &ChangeLogModel{ChangeLog: cl}
}

func wrapRepositoryVersion(rv *model.RepositoryVersion) *RepositoryVersionModel {
	if rv == nil {
		return nil
	}
	return &RepositoryVersionModel{RepositoryVersion: rv}
}

func wrapUploadSession(us *model.UploadSession) *UploadSessionModel {
	if us == nil {
		return nil
	}
	return &UploadSessionModel{UploadSession: us}
}

func wrapUploadChunk(uc *model.UploadChunk) *UploadChunkModel {
	if uc == nil {
		return nil
	}
	return &UploadChunkModel{UploadChunk: uc}
}

func unwrapUploadChunks(ucs []*UploadChunkModel) []*model.UploadChunk {
	chunks := make([]*model.UploadChunk, len(ucs))
	for i, uc := range ucs {
		chunks[i] = uc.UploadChunk
	}
	return chunks
}

func RecordChange(ctx context.Context, change *model.ChangeLog) error {
	change.Timestamp = time.Now()
	_, err := db.NewInsert().Model(wrapChangeLog(change)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record change: %w", err)
	}
	return nil
}

func GetCurrentVersion(ctx context.Context, repoID int) (*model.RepositoryVersion, error) {
	var rv RepositoryVersionModel
	err := db.NewSelect().
		Model(&rv).
		Where("repo_id = ?", repoID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get repository version: %w", err)
	}
	return rv.RepositoryVersion, nil
}

func UpdateVersion(ctx context.Context, repoID int, newVersion string, versionVector string) error {
	now := time.Now()
	_, err := db.NewInsert().
		Model(wrapRepositoryVersion(&model.RepositoryVersion{
			RepoID:         repoID,
			CurrentVersion: newVersion,
			VersionVector:  versionVector,
			UpdatedAt:      now,
		})).
		On("CONFLICT (repo_id) DO UPDATE").
		Set("current_version = ?", newVersion).
		Set("version_vector = ?", versionVector).
		Set("updated_at = ?", now).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update repository version: %w", err)
	}
	return nil
}

func GetChangesSince(ctx context.Context, repoID int, sinceVersion string, limit int) ([]*model.ChangeLog, error) {
	var changes []*ChangeLogModel

	query := db.NewSelect().
		Model(&changes).
		Where("repo_id = ?", repoID)

	if sinceVersion != "" {
		query = query.Where("version > ?", sinceVersion)
	}

	query = query.
		Order("id ASC").
		Limit(limit)

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get changes since version %s: %w", sinceVersion, err)
	}

	result := make([]*model.ChangeLog, len(changes))
	for i, c := range changes {
		result[i] = c.ChangeLog
	}
	return result, nil
}

func CreateUploadSession(ctx context.Context, session *model.UploadSession) error {
	_, err := db.NewInsert().Model(wrapUploadSession(session)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload session: %w", err)
	}
	return nil
}

func GetUploadSession(ctx context.Context, uploadID string) (*model.UploadSession, error) {
	var us UploadSessionModel
	err := db.NewSelect().
		Model(&us).
		Where("upload_id = ?", uploadID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get upload session %s: %w", uploadID, err)
	}
	return us.UploadSession, nil
}

func UpdateUploadSessionStatus(ctx context.Context, uploadID string, status string) error {
	_, err := db.NewUpdate().
		Model((*UploadSessionModel)(nil)).
		Set("status = ?", status).
		Where("upload_id = ?", uploadID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update upload session status: %w", err)
	}
	return nil
}

func IncrementUploadedChunks(ctx context.Context, uploadID string) error {
	_, err := db.NewUpdate().
		Model((*UploadSessionModel)(nil)).
		Set("chunks_uploaded = chunks_uploaded + 1").
		Where("upload_id = ?", uploadID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to increment uploaded chunks: %w", err)
	}
	return nil
}

func CreateUploadChunk(ctx context.Context, chunk *model.UploadChunk) error {
	chunk.UploadedAt = time.Now()
	_, err := db.NewInsert().Model(wrapUploadChunk(chunk)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload chunk: %w", err)
	}
	return nil
}

func GetUploadedChunks(ctx context.Context, uploadID string) ([]*model.UploadChunk, error) {
	var chunks []*UploadChunkModel
	err := db.NewSelect().
		Model(&chunks).
		Where("upload_id = ?", uploadID).
		Order("chunk_index ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get uploaded chunks for %s: %w", uploadID, err)
	}
	return unwrapUploadChunks(chunks), nil
}

func GetUploadChunk(ctx context.Context, uploadID string, chunkIndex int) (*model.UploadChunk, error) {
	var uc UploadChunkModel
	err := db.NewSelect().
		Model(&uc).
		Where("upload_id = ? AND chunk_index = ?", uploadID, chunkIndex).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get upload chunk %d for %s: %w", chunkIndex, uploadID, err)
	}
	return uc.UploadChunk, nil
}

func CleanupExpiredUploadSessions(ctx context.Context) error {
	_, err := db.NewDelete().
		Model((*UploadSessionModel)(nil)).
		Where("expires_at < ?", time.Now()).
		Where("status != ?", "completed").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to cleanup expired upload sessions: %w", err)
	}
	return nil
}

func DeleteUploadSession(ctx context.Context, uploadID string) error {
	_, err := db.NewDelete().
		Model((*UploadSessionModel)(nil)).
		Where("upload_id = ?", uploadID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete upload session %s: %w", uploadID, err)
	}
	return nil
}
