package sync

import (
	"context"
	"io"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
	"github.com/uptrace/bun"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCService implements the gRPC SyncService server
type GRPCService struct {
	UnimplementedSyncServiceServer
	service *Service
}

// NewGRPCService creates a new gRPC service wrapper
func NewGRPCService(database *bun.DB) *GRPCService {
	return &GRPCService{
		service: NewService(database),
	}
}

// getRepositoryFromContext extracts repository from context using user ID
func (g *GRPCService) getRepositoryFromContext(ctx context.Context, repoName string) (*model.Repository, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	repo, err := db.GetRepositoryByNameAndOwner(ctx, repoName, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "repository not found: %v", err)
	}

	return repo, nil
}

// UploadFile implements the UploadFile RPC
func (g *GRPCService) UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &UploadFileResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	etag, _, _, err := g.service.UploadFile(ctx, repo, req.Path, req.Content, req.MimeType, 0)
	if err != nil {
		return &UploadFileResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	// Parse version string to int64 if needed
	versionNum := int64(0) // Default to 0, or parse from version string if format allows

	return &UploadFileResponse{
		Success: true,
		Etag:    etag,
		Version: versionNum,
	}, nil
}

// BeginUpload implements the BeginUpload RPC
func (g *GRPCService) BeginUpload(ctx context.Context, req *BeginUploadRequest) (*BeginUploadResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &BeginUploadResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	uploadID, uploadedChunks, err := g.service.BeginUpload(ctx, repo, req.Path, req.TotalSize, 0)
	if err != nil {
		return &BeginUploadResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &BeginUploadResponse{
		Success:        true,
		UploadId:       uploadID,
		UploadedChunks: int32SliceToInt32(uploadedChunks),
	}, nil
}

// UploadChunk implements the UploadChunk RPC
func (g *GRPCService) UploadChunk(ctx context.Context, req *UploadChunkRequest) (*UploadChunkResponse, error) {
	err := g.service.UploadChunk(ctx, req.UploadId, int(req.ChunkIndex), req.Data)
	if err != nil {
		return &UploadChunkResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &UploadChunkResponse{
		Success:    true,
		ChunkIndex: req.ChunkIndex,
	}, nil
}

// FinalizeUpload implements the FinalizeUpload RPC
func (g *GRPCService) FinalizeUpload(ctx context.Context, req *FinalizeUploadRequest) (*FinalizeUploadResponse, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return &FinalizeUploadResponse{Success: false, ErrorMessage: "user not authenticated"}, nil
	}

	repoName, ok := ctx.Value("repoName").(string)
	if !ok {
		return &FinalizeUploadResponse{Success: false, ErrorMessage: "repository name not found in context"}, nil
	}

	repo, err := g.getRepositoryFromContext(ctx, repoName)
	if err != nil {
		return &FinalizeUploadResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	etag, _, err := g.service.FinalizeUpload(ctx, req.UploadId, repo, userID)
	if err != nil {
		return &FinalizeUploadResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &FinalizeUploadResponse{
		Success: true,
		Etag:    etag,
	}, nil
}

// CancelUpload implements the CancelUpload RPC
func (g *GRPCService) CancelUpload(ctx context.Context, req *CancelUploadRequest) (*CancelUploadResponse, error) {
	err := g.service.CancelUpload(ctx, req.UploadId)
	if err != nil {
		return &CancelUploadResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &CancelUploadResponse{
		Success: true,
	}, nil
}

// DownloadFile implements the DownloadFile streaming RPC
func (g *GRPCService) DownloadFile(req *DownloadFileRequest, stream grpc.ServerStreamingServer[DownloadFileResponse]) error {
	ctx := stream.Context()

	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return status.Errorf(codes.NotFound, "repository not found: %v", err)
	}

	file, reader, err := g.service.DownloadFile(ctx, repo, req.Path, req.IfNoneMatch, 0)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to download file: %v", err)
	}

	if reader == nil {
		// File not modified (304 Not Modified)
		return stream.Send(&DownloadFileResponse{
			Response: &DownloadFileResponse_Info{
				Info: &FileInfo{},
			},
		})
	}
	defer reader.Close()

	// Convert file to protobuf FileInfo
	fileInfo := fileToProto(file)
	if err := stream.Send(&DownloadFileResponse{
		Response: &DownloadFileResponse_Info{
			Info: fileInfo,
		},
	}); err != nil {
		return err
	}

	// Stream file content in chunks
	buffer := make([]byte, ChunkSize)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			if sendErr := stream.Send(&DownloadFileResponse{
				Response: &DownloadFileResponse_Chunk{
					Chunk: buffer[:n],
				},
			}); sendErr != nil {
				return sendErr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to read file: %v", err)
		}
	}

	// Send completion message
	etag := ""
	if file.Checksum != nil {
		etag = *file.Checksum
	}
	return stream.Send(&DownloadFileResponse{
		Response: &DownloadFileResponse_Complete{
			Complete: &DownloadComplete{
				Etag:    etag,
				Version: 0, // Version from file if available
			},
		},
	})
}

// GetFileInfo implements the GetFileInfo RPC
func (g *GRPCService) GetFileInfo(ctx context.Context, req *GetFileInfoRequest) (*GetFileInfoResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &GetFileInfoResponse{Exists: false, ErrorMessage: err.Error()}, nil
	}

	file, err := g.service.GetFileInfo(ctx, repo, req.Path, 0)
	if err != nil {
		return &GetFileInfoResponse{Exists: false, ErrorMessage: "file not found"}, nil
	}

	return &GetFileInfoResponse{
		Exists: true,
		Info:   fileToProto(file),
	}, nil
}

// ListDirectory implements the ListDirectory RPC
func (g *GRPCService) ListDirectory(ctx context.Context, req *ListDirectoryRequest) (*ListDirectoryResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &ListDirectoryResponse{ErrorMessage: err.Error()}, nil
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}

	items, total, err := g.service.ListDirectory(ctx, repo, req.Path, offset, limit, 0)
	if err != nil {
		return &ListDirectoryResponse{ErrorMessage: err.Error()}, nil
	}

	protoItems := make([]*FileInfo, len(items))
	for i, item := range items {
		protoItems[i] = fileToProto(item)
	}

	hasMore := int64(offset+limit) < total

	return &ListDirectoryResponse{
		Items:      protoItems,
		TotalCount: int32(total),
		HasMore:    hasMore,
	}, nil
}

// ListChanges implements the ListChanges RPC
func (g *GRPCService) ListChanges(ctx context.Context, req *ListChangesRequest) (*ListChangesResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &ListChangesResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	maxChanges := int(req.MaxChanges)
	if maxChanges <= 0 || maxChanges > 1000 {
		maxChanges = 100
	}

	changes, err := g.service.ListChanges(ctx, repo.ID, req.SinceVersion, maxChanges)
	if err != nil {
		return &ListChangesResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	// Categorize changes
	created := make([]*FileInfo, 0)
	modified := make([]*FileInfo, 0)
	deleted := make([]string, 0)
	renamed := make([]*RenameOperation, 0)

	for _, change := range changes {
		switch change.Operation {
		case "create":
			// Get file info for created items
			file, err := g.service.GetFileInfo(ctx, repo, change.Path, 0)
			if err == nil {
				created = append(created, fileToProto(file))
			}
		case "modify":
			file, err := g.service.GetFileInfo(ctx, repo, change.Path, 0)
			if err == nil {
				modified = append(modified, fileToProto(file))
			}
		case "delete":
			deleted = append(deleted, change.Path)
		case "move":
			if change.OldPath != nil {
				renamed = append(renamed, &RenameOperation{
					OldPath: *change.OldPath,
					NewPath: change.Path,
				})
			}
		}
	}

	currentVersion, _ := g.service.GetCurrentVersion(ctx, repo.ID)

	return &ListChangesResponse{
		Success:       true,
		CurrentVersion: currentVersion.CurrentVersion,
		Created:       created,
		Modified:      modified,
		Deleted:       deleted,
		Renamed:       renamed,
		HasMore:       len(changes) >= maxChanges,
	}, nil
}

// GetCurrentVersion implements the GetCurrentVersion RPC
func (g *GRPCService) GetCurrentVersion(ctx context.Context, req *GetCurrentVersionRequest) (*GetCurrentVersionResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &GetCurrentVersionResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	version, err := g.service.GetCurrentVersion(ctx, repo.ID)
	if err != nil {
		return &GetCurrentVersionResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &GetCurrentVersionResponse{
		Success:   true,
		Version:   version.CurrentVersion,
		Timestamp: version.UpdatedAt.Unix(),
	}, nil
}

// CreateDirectory implements the CreateDirectory RPC
func (g *GRPCService) CreateDirectory(ctx context.Context, req *CreateDirectoryRequest) (*CreateDirectoryResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &CreateDirectoryResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	if err := g.service.CreateDirectory(ctx, repo, req.Path, 0); err != nil {
		return &CreateDirectoryResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &CreateDirectoryResponse{
		Success: true,
	}, nil
}

// Delete implements the Delete RPC
func (g *GRPCService) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &DeleteResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	if err := g.service.Delete(ctx, repo, req.Path, req.Recursive, 0); err != nil {
		return &DeleteResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &DeleteResponse{
		Success: true,
	}, nil
}

// Move implements the Move RPC
func (g *GRPCService) Move(ctx context.Context, req *MoveRequest) (*MoveResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &MoveResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	if err := g.service.Move(ctx, repo, req.SourcePath, req.DestinationPath, 0); err != nil {
		return &MoveResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &MoveResponse{
		Success: true,
	}, nil
}

// Copy implements the Copy RPC
func (g *GRPCService) Copy(ctx context.Context, req *CopyRequest) (*CopyResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &CopyResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	if err := g.service.Copy(ctx, repo, req.SourcePath, req.DestinationPath, 0); err != nil {
		return &CopyResponse{Success: false, ErrorMessage: err.Error()}, nil
	}

	return &CopyResponse{
		Success: true,
	}, nil
}

// GetSyncStatus implements the GetSyncStatus RPC
func (g *GRPCService) GetSyncStatus(ctx context.Context, req *SyncStatusRequest) (*SyncStatusResponse, error) {
	repo, err := g.getRepositoryFromContext(ctx, req.Repo)
	if err != nil {
		return &SyncStatusResponse{ErrorMessage: err.Error()}, nil
	}

	status, file, err := g.service.GetSyncStatus(ctx, repo, req.Path, req.ClientEtag, req.ClientVersion, 0)
	if err != nil {
		return &SyncStatusResponse{ErrorMessage: err.Error()}, nil
	}

	var protoStatus SyncStatusResponse_Status
	switch status {
	case "synced":
		protoStatus = SyncStatusResponse_SYNCED
	case "modified":
		protoStatus = SyncStatusResponse_MODIFIED
	case "new":
		protoStatus = SyncStatusResponse_NEW
	case "deleted":
		protoStatus = SyncStatusResponse_DELETED
	case "conflict":
		protoStatus = SyncStatusResponse_CONFLICT
	default:
		protoStatus = SyncStatusResponse_UNKNOWN
	}

	var serverInfo *FileInfo
	if file != nil {
		serverInfo = fileToProto(file)
	}

	return &SyncStatusResponse{
		Status:     protoStatus,
		ServerInfo: serverInfo,
	}, nil
}

// BatchOperation implements the BatchOperation RPC
func (g *GRPCService) BatchOperation(ctx context.Context, req *BatchOperationRequest) (*BatchOperationResponse, error) {
	// TODO: Implement batch operations
	return &BatchOperationResponse{
		ErrorMessage: "batch operations not yet implemented",
	}, nil
}

// Helper function to convert model.FileObject to protobuf FileInfo
func fileToProto(file *model.FileObject) *FileInfo {
	if file == nil {
		return nil
	}

	etag := ""
	if file.Checksum != nil {
		etag = *file.Checksum
	}

	mimeType := ""
	if file.MimeType != nil {
		mimeType = *file.MimeType
	}

	return &FileInfo{
		Path:     file.Path,
		Size:     file.Size,
		ModTime:  file.ModTime.Unix(),
		IsDir:    file.IsDir,
		MimeType: mimeType,
		Etag:     etag,
		Version:  0, // Add version field if available
	}
}

func int32SliceToInt32(slice []int) []int32 {
	result := make([]int32, len(slice))
	for i, v := range slice {
		result[i] = int32(v)
	}
	return result
}
