# Sync Protocol

This directory contains the Protocol Buffer definitions for the File Hub Sync Protocol, which is designed to complement the existing WebDAV implementation. The protocol is optimized for mobile devices to upload files to a home server using protobuf over HTTP/2.

## Status: ✅ COMPLETE

The sync protocol implementation is complete with both HTTP REST API and gRPC support.

## Features

### Implemented
- ✅ Protocol Buffer definitions (`sync.proto`) - 17 RPC methods
- ✅ Generated Go code (`sync.pb.go`, `sync_grpc.pb.go`)
- ✅ HTTP REST API (16 endpoints under `/api/sync/*`)
- ✅ gRPC service with authentication interceptors
- ✅ Chunked upload with resume capability
- ✅ Version-based change tracking
- ✅ Database schema and operations
- ✅ Comprehensive unit tests

### File Operations
- `GetFileInfo` - Get file/directory metadata
- `ListDirectory` - List directory contents with pagination
- `CreateDirectory` - Create new directories
- `Delete` - Delete files and directories (recursive support)
- `Move` - Move/rename files and directories
- `Copy` - Copy files and directories

### Upload Operations
- `UploadFile` - Simple upload for small files (<10MB)
- `BeginUpload` - Start chunked upload session
- `UploadChunk` - Upload individual chunks (1MB each)
- `FinalizeUpload` - Assemble chunks and complete upload
- `CancelUpload` - Cancel upload and cleanup

### Download Operations
- `DownloadFile` - Stream file download with conditional support

### Sync Operations
- `GetCurrentVersion` - Get repository version identifier
- `ListChanges` - Get changes since a version (delta sync)
- `GetSyncStatus` - Compare client vs server state
- `BatchOperation` - Batch multiple operations (stub)

## Files

- `sync.proto` - Protocol Buffer definitions
- `sync.pb.go` - Generated protobuf code
- `sync_grpc.pb.go` - Generated gRPC service code
- `service.go` - Core business logic
- `grpc_service.go` - gRPC service implementation
- `grpc_auth.go` - gRPC authentication interceptors
- `service_test.go` - Unit tests

## Building

The protobuf Go code is automatically generated during the build process:

```bash
make build
```

Or manually generate:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/sync/sync.proto
```

## Configuration

Enable gRPC server by adding to `config.yaml`:

```yaml
web:
  port: 8080
  grpc_port: 9090  # gRPC server port (0 to disable)
```

## API Endpoints

### HTTP REST API
All endpoints are under `/api/sync/` and require authentication:

```
GET    /api/sync/info         - Get file info
GET    /api/sync/list         - List directory
POST   /api/sync/mkdir        - Create directory
DELETE /api/sync/delete       - Delete file/directory
POST   /api/sync/move         - Move/rename
POST   /api/sync/copy         - Copy
POST   /api/sync/upload       - Simple upload
GET    /api/sync/download     - Download file
GET    /api/sync/version      - Get current version
GET    /api/sync/changes      - Get changes since version
GET    /api/sync/status       - Get sync status
POST   /api/sync/upload/begin - Begin chunked upload
POST   /api/sync/upload/chunk - Upload chunk
POST   /api/sync/upload/finalize - Finalize upload
DELETE /api/sync/upload/cancel - Cancel upload
```

### gRPC Service
The gRPC service runs on the configured `grpc_port` and provides all the same functionality
with better performance for mobile clients and support for streaming downloads.

## Authentication

Both HTTP and gRPC APIs support:
- Session cookie authentication (`filehub_session`)
- Basic authentication (Authorization header)
- Token authentication (`x-session-token` header)

## Chunked Upload Flow

1. Call `BeginUpload` to create upload session
2. Upload chunks using `UploadChunk` (can resume from interruption)
3. Call `FinalizeUpload` to assemble and store final file
4. Call `CancelUpload` to abort (optional, cleanup is automatic)

## Change Tracking

The sync protocol uses version-based change tracking:
- Each repository maintains a version string (format: `v{timestamp}-{nanoseconds}`)
- All operations increment the version
- `ListChanges` returns changes since a specific version
- Enables efficient delta synchronization