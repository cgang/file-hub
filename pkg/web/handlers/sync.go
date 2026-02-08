package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
	"github.com/cgang/file-hub/pkg/sync"
	"github.com/cgang/file-hub/pkg/web/auth"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

const (
	DefaultLimit = 100
	MaxLimit     = 1000
)

type SyncHandler struct {
	svc *sync.Service
}

func NewSyncHandler(database *bun.DB) *SyncHandler {
	return &SyncHandler{
		svc: sync.NewService(database),
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type FileInfoResponse struct {
	Exists  bool              `json:"exists"`
	Info    *model.FileObject `json:"info,omitempty"`
	Message string            `json:"message,omitempty"`
}

type ListDirectoryResponse struct {
	Items   []*model.FileObject `json:"items"`
	Total   int64               `json:"total"`
	Offset  int                 `json:"offset"`
	Limit   int                 `json:"limit"`
	HasMore bool                `json:"has_more"`
	Message string              `json:"message,omitempty"`
}

type VersionResponse struct {
	Version   string    `json:"version"`
	Vector    string    `json:"vector"`
	Timestamp time.Time `json:"timestamp"`
}

type ChangesResponse struct {
	Version string             `json:"version"`
	Changes []*model.ChangeLog `json:"changes"`
	Changed int                `json:"changed"`
	Message string             `json:"message,omitempty"`
}

type UploadResponse struct {
	Etag    string `json:"etag"`
	Version string `json:"version"`
	Size    int64  `json:"size"`
	Message string `json:"message,omitempty"`
}

type BeginUploadResponse struct {
	UploadID       string `json:"upload_id"`
	TotalChunks    int    `json:"total_chunks"`
	ChunkSize      int64  `json:"chunk_size"`
	UploadedChunks []int  `json:"uploaded_chunks"`
	Message        string `json:"message,omitempty"`
}

type FinalizeUploadResponse struct {
	Etag    string `json:"etag"`
	Size    int64  `json:"size"`
	Message string `json:"message,omitempty"`
}

type SyncStatusResponse struct {
	Status  string            `json:"status"`
	Info    *model.FileObject `json:"info,omitempty"`
	Message string            `json:"message,omitempty"`
}

func (h *SyncHandler) GetFileInfo(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	path := c.Query("path")

	if repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo and path parameters are required"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	file, err := h.svc.GetFileInfo(c.Request.Context(), repo, path, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, FileInfoResponse{
			Exists:  false,
			Message: "File not found",
		})
		return
	}

	c.JSON(http.StatusOK, FileInfoResponse{
		Exists: true,
		Info:   file,
	})
}

func (h *SyncHandler) ListDirectory(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	path := c.Query("path")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "100")

	if repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo and path parameters are required"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > MaxLimit {
		limit = DefaultLimit
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	items, total, err := h.svc.ListDirectory(c.Request.Context(), repo, path, offset, limit, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list directory"})
		return
	}

	hasMore := int64(offset+limit) < total

	c.JSON(http.StatusOK, ListDirectoryResponse{
		Items:   items,
		Total:   total,
		Offset:  offset,
		Limit:   limit,
		HasMore: hasMore,
	})
}

func (h *SyncHandler) CreateDirectory(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	path := c.Query("path")

	if repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo and path parameters are required"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	if err := h.svc.CreateDirectory(c.Request.Context(), repo, path, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to create directory: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Directory created successfully"})
}

func (h *SyncHandler) Delete(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	path := c.Query("path")
	recursiveStr := c.DefaultQuery("recursive", "false")

	if repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo and path parameters are required"})
		return
	}

	recursive := recursiveStr == "true"

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), repo, path, recursive, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to delete: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted successfully"})
}

func (h *SyncHandler) Move(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	sourcePath := c.Query("source")
	destPath := c.Query("destination")

	if repoName == "" || sourcePath == "" || destPath == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo, source, and destination parameters are required"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	if err := h.svc.Move(c.Request.Context(), repo, sourcePath, destPath, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to move: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Moved successfully"})
}

func (h *SyncHandler) Copy(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	sourcePath := c.Query("source")
	destPath := c.Query("destination")

	if repoName == "" || sourcePath == "" || destPath == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo, source, and destination parameters are required"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	if err := h.svc.Copy(c.Request.Context(), repo, sourcePath, destPath, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to copy: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Copied successfully"})
}

func (h *SyncHandler) UploadFile(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	path := c.Query("path")

	if repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo and path parameters are required"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to read file data"})
		return
	}

	etag, version, size, err := h.svc.UploadFile(c.Request.Context(), repo, path, data, c.GetHeader("Content-Type"), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to upload file: %s", err)})
		return
	}

	c.JSON(http.StatusOK, UploadResponse{
		Etag:    etag,
		Version: version,
		Size:    size,
	})
}

func (h *SyncHandler) DownloadFile(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	path := c.Query("path")
	ifNoneMatch := c.GetHeader("If-None-Match")

	if repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo and path parameters are required"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	file, reader, err := h.svc.DownloadFile(c.Request.Context(), repo, path, ifNoneMatch, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to download file"})
		return
	}

	if reader == nil {
		c.Status(http.StatusNotModified)
		return
	}
	defer reader.Close()

	c.Header("Content-Type", file.ContentType())
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))
	if file.Checksum != nil {
		c.Header("ETag", *file.Checksum)
	}
	c.Header("Last-Modified", file.ModTime.Format(http.TimeFormat))

	c.DataFromReader(http.StatusOK, file.Size, file.ContentType(), reader, nil)
}

func (h *SyncHandler) GetCurrentVersion(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	if repoName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo parameter is required"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	version, err := h.svc.GetCurrentVersion(c.Request.Context(), repo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get version"})
		return
	}

	c.JSON(http.StatusOK, VersionResponse{
		Version:   version.CurrentVersion,
		Vector:    version.VersionVector,
		Timestamp: version.UpdatedAt,
	})
}

func (h *SyncHandler) ListChanges(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	sinceVersion := c.DefaultQuery("since", "")
	maxChangesStr := c.DefaultQuery("limit", "100")

	if repoName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo parameter is required"})
		return
	}

	maxChanges, err := strconv.Atoi(maxChangesStr)
	if err != nil || maxChanges <= 0 {
		maxChanges = 100
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	changes, err := h.svc.ListChanges(c.Request.Context(), repo.ID, sinceVersion, maxChanges)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get changes"})
		return
	}

	currentVersion, _ := h.svc.GetCurrentVersion(c.Request.Context(), repo.ID)

	c.JSON(http.StatusOK, ChangesResponse{
		Version: currentVersion.CurrentVersion,
		Changes: changes,
		Changed: len(changes),
	})
}

func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	path := c.Query("path")
	clientETag := c.Query("client_etag")
	clientVersionStr := c.Query("client_version")

	if repoName == "" || path == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo and path parameters are required"})
		return
	}

	clientVersion, _ := strconv.ParseInt(clientVersionStr, 10, 64)

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	status, file, err := h.svc.GetSyncStatus(c.Request.Context(), repo, path, clientETag, clientVersion, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get sync status"})
		return
	}

	c.JSON(http.StatusOK, SyncStatusResponse{
		Status: status,
		Info:   file,
	})
}

func (h *SyncHandler) BeginUpload(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	repoName := c.Query("repo")
	path := c.Query("path")
	totalSizeStr := c.Query("total_size")

	if repoName == "" || path == "" || totalSizeStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "repo, path, and total_size parameters are required"})
		return
	}

	totalSize, err := strconv.ParseInt(totalSizeStr, 10, 64)
	if err != nil || totalSize <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid total_size parameter"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	uploadID, uploadedChunks, err := h.svc.BeginUpload(c.Request.Context(), repo, path, totalSize, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to begin upload: %s", err)})
		return
	}

	c.JSON(http.StatusOK, BeginUploadResponse{
		UploadID:       uploadID,
		TotalChunks:    (int)(totalSize+1024*1024-1) / (1024 * 1024),
		ChunkSize:      1024 * 1024,
		UploadedChunks: uploadedChunks,
	})
}

func (h *SyncHandler) UploadChunk(c *gin.Context) {
	_, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	uploadID := c.Query("upload_id")
	chunkIndexStr := c.Query("chunk_index")

	if uploadID == "" || chunkIndexStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "upload_id and chunk_index parameters are required"})
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid chunk_index parameter"})
		return
	}

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to read chunk data"})
		return
	}

	if err := h.svc.UploadChunk(c.Request.Context(), uploadID, chunkIndex, data); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to upload chunk: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Chunk uploaded successfully"})
}

func (h *SyncHandler) FinalizeUpload(c *gin.Context) {
	user, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	uploadID := c.Query("upload_id")
	repoName := c.Query("repo")

	if uploadID == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "upload_id and repo parameters are required"})
		return
	}

	repo, err := db.GetRepositoryByNameAndOwner(c.Request.Context(), repoName, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Repository not found"})
		return
	}

	etag, size, err := h.svc.FinalizeUpload(c.Request.Context(), uploadID, repo, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to finalize upload: %s", err)})
		return
	}

	c.JSON(http.StatusOK, FinalizeUploadResponse{
		Etag: etag,
		Size: size,
	})
}

func (h *SyncHandler) CancelUpload(c *gin.Context) {
	_, ok := auth.GetAuthenticatedUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	uploadID := c.Query("upload_id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "upload_id parameter is required"})
		return
	}

	if err := h.svc.CancelUpload(c.Request.Context(), uploadID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to cancel upload: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Upload cancelled successfully"})
}

func RegisterSyncRoutes(router *gin.Engine, database *bun.DB) {
	handler := NewSyncHandler(database)

	api := router.Group("/api/sync")
	{
		api.GET("/info", handler.GetFileInfo)
		api.GET("/list", handler.ListDirectory)
		api.POST("/mkdir", handler.CreateDirectory)
		api.DELETE("/delete", handler.Delete)
		api.POST("/move", handler.Move)
		api.POST("/copy", handler.Copy)
		api.POST("/upload", handler.UploadFile)
		api.GET("/download", handler.DownloadFile)
		api.GET("/version", handler.GetCurrentVersion)
		api.GET("/changes", handler.ListChanges)
		api.GET("/status", handler.GetSyncStatus)
		api.POST("/upload/begin", handler.BeginUpload)
		api.POST("/upload/chunk", handler.UploadChunk)
		api.POST("/upload/finalize", handler.FinalizeUpload)
		api.DELETE("/upload/cancel", handler.CancelUpload)
	}
}
