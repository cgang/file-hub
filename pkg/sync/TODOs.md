# Sync Protocol Implementation Plan

**Status**: ✅ **COMPLETED** - MVP implementation with Gin integration

## Current State

✅ **Completed**:
- ✅ Protocol Buffer definitions (`pkg/sync/sync.proto`) - All 17 RPC methods defined
- ✅ Generated Go code (`pkg/sync/sync.pb.go`, `pkg/sync/sync_grpc.pb.go`)
- ✅ Database Schema (`scripts/sync_schema.sql`) - Complete with change_log, repository_versions, upload_sessions, upload_chunks
- ✅ Data Models (`pkg/model/sync.go`) - All sync entities
- ✅ Database Layer (`pkg/db/sync.go`) - 12 database operation functions
- ✅ Business Logic (`pkg/sync/service.go`) - Core sync service with file operations
- ✅ HTTP API (`pkg/web/handlers/sync.go`) - 16 REST endpoints integrated with Gin
- ✅ Integration (`pkg/web/server.go`) - Routes registered and connected
- ✅ Documentation (`pkg/sync/README.md`, `docs/SYNC_IMPLEMENTATION.md`)

🔄 **Optional Enhancements**:
- Recursive directory listing (flat is default - see docs/SYNC_IMPLEMENTATION.md)
- Unit tests (in progress)
- Integration tests
- Rate limiting
- Prometheus metrics

---

## Implementation Roadmap

### ✅ Phase 1: Foundation (COMPLETED)

#### 1.1 Database Schema Extensions
**File**: `scripts/sync_schema.sql`

✅ **COMPLETED**:
- ✅ Created `change_log` table with indexes
- ✅ Created `repository_versions` table with JSONB support
- ✅ Created `upload_sessions` table for chunked uploads
- ✅ Created `upload_chunks` table for chunk tracking
- ✅ Added all necessary indexes for performance

**Model Files**: `pkg/model/sync.go`

✅ **COMPLETED**:
- ✅ Created `ChangeLog` struct
- ✅ Created `RepositoryVersion` struct
- ✅ Created `UploadSession` struct
- ✅ Created `UploadChunk` struct
- ✅ Added Bun ORM tags for all models

#### 1.2 Database Operations Layer
**File**: `pkg/db/sync.go`

✅ **COMPLETED**:
- ✅ `RecordChange` - Record file operation
- ✅ `GetChangesSince` - Retrieve changes by version
- ✅ `GetCurrentVersion` - Get repository version
- ✅ `UpdateVersion` - Update repository version
- ✅ `CreateUploadSession` - Create upload session
- ✅ `GetUploadSession` - Retrieve upload session
- ✅ `UpdateUploadSessionStatus` - Update session status
- ✅ `IncrementUploadedChunks` - Track upload progress
- ✅ `CreateUploadChunk` - Store chunk data
- ✅ `GetUploadedChunks` - Get all chunks
- ✅ `GetUploadChunk` - Get specific chunk
- ✅ `CleanupExpiredUploadSessions` - Clean up old sessions
- ✅ `DeleteUploadSession` - Delete session

---

### ✅ Phase 2: Core Service Structure (COMPLETED)

#### 2.1 Service Interface and Initialization
**File**: `pkg/sync/service.go`

✅ **COMPLETED**:
- ✅ Created `SyncService` struct with database dependency
- ✅ Implemented `NewSyncService` constructor
- ✅ Integrated with Bun DB connection

#### 2.2 Version Management Implementation
**File**: `pkg/sync/service.go`

✅ **COMPLETED**:
- ✅ `generateVersion()` - Timestamp-based version: `v{unix_timestamp}-{nanoseconds}`
- ✅ Integrated into all file operations
- ✅ Database storage in repository_versions table

#### 2.3 File Metadata Conversion
**File**: `pkg/sync/service.go`

✅ **COMPLETED**:
- ✅ Direct model reuse from `pkg/model/file.go`
- ✅ File I/O via `pkg/stor` abstraction layer
- ✅ Metadata conversion through existing patterns

---

### ✅ Phase 3: Basic File Operations (COMPLETED)

#### 3.1 GetFileInfo
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `GetFileInfo()` function
- ✅ HTTP: `GET /api/sync/info` endpoint
- ✅ Authentication: Session-based
- ✅ Integration: Uses `pkg/stor.GetFileInfo`

#### 3.2 ListDirectory
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `ListDirectory()` function
- ✅ HTTP: `GET /api/sync/list` endpoint
- ✅ Pagination: offset/limit support
- ✅ **Decision**: Flat listing by default (see docs/SYNC_IMPLEMENTATION.md)
- ✅ Integration: Uses `db.GetChildFiles`

#### 3.3 CreateDirectory
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `CreateDirectory()` function
- ✅ HTTP: `POST /api/sync/mkdir` endpoint
- ✅ Change tracking: Records in change_log
- ✅ Version update: Increments repository version

#### 3.4 Delete
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `Delete()` function
- ✅ HTTP: `DELETE /api/sync/delete` endpoint
- ✅ Recursive deletion support
- ✅ Change tracking: Records in change_log

#### 3.5 Move and Copy
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `Move()` and `Copy()` functions
- ✅ HTTP: `POST /api/sync/move`, `POST /api/sync/copy` endpoints
- ✅ Change tracking: Records old_path for moves
- ✅ Version update: Increments repository version

---

### ✅ Phase 4: Upload and Download Operations (COMPLETED)

#### 4.1 Simple Upload (UploadFile)
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `UploadFile()` function
- ✅ HTTP: `POST /api/sync/upload` endpoint
- ✅ Size limit: 10MB for simple upload
- ✅ SHA-256 hashing for integrity
- ✅ Change tracking: Records create operation

#### 4.2 Simple Download (DownloadFile)
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `DownloadFile()` function
- ✅ HTTP: `GET /api/sync/download` endpoint
- ✅ Conditional support: If-None-Match header
- ✅ Stream support: Via storage layer

#### 4.3 Chunked Upload System
**Files**: `pkg/sync/service.go`, `pkg/db/sync.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `BeginUpload()`, `UploadChunk()`, `FinalizeUpload()`, `CancelUpload()`
- ✅ HTTP: `POST /api/sync/upload/begin`, `/chunk`, `/finalize`, `DELETE /cancel`
- ✅ Database: upload_sessions and upload_chunks tables
- ✅ Resume capability: Returns already uploaded chunks
- ✅ Session expiration: 24 hours
- ✅ Chunk size: 1MB

---

### ✅ Phase 5: Change Tracking and Sync Management (COMPLETED)

#### 5.1 GetVersion and ListChanges
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `GetCurrentVersion()` and `ListChanges()` functions
- ✅ HTTP: `GET /api/sync/version`, `GET /api/sync/changes` endpoints
- ✅ Change query: By version timestamp
- ✅ Pagination: limit parameter with max 1000

#### 5.2 GetSyncStatus
**File**: `pkg/sync/service.go`, `pkg/web/handlers/sync.go`

✅ **COMPLETED**:
- ✅ Service: `GetSyncStatus()` function
- ✅ HTTP: `GET /api/sync/status` endpoint
- ✅ Status detection: synced, modified, new, deleted
- ✅ Comparison: client_etag vs server_etag

#### 5.3 BatchOperation
**Status**: DEFERRED - Optimization for later

---

### ✅ Phase 6: Authentication and Integration (COMPLETED)

#### 6.1 gRPC Authentication Middleware
**Decision**: Using Gin HTTP middleware instead of gRPC

✅ **COMPLETED**:
- ✅ Reused existing `pkg/web/auth` session authentication
- ✅ All endpoints protected by `auth.GetAuthenticatedUser`
- ✅ Session cookie-based authentication
- ✅ User ID validation per request

#### 6.2 Register HTTP Server
**File**: `pkg/web/server.go`

✅ **COMPLETED**:
- ✅ Registered sync routes via `handlers.RegisterSyncRoutes`
- ✅ Integrated with existing Gin router
- ✅ Database connection passed to handlers
- ✅ Route group: `/api/sync`

#### 6.3 Server Options
**File**: `pkg/web/server.go`

✅ **COMPLETED**:
- ✅ Reuses existing server configuration
- ✅ Debug mode support
- ✅ Metrics endpoint support
- ✅ Pprof support when enabled

---

### 🔄 Phase 7: Testing (IN PROGRESS)

#### 7.1 Unit Tests (This Phase)
**Files**: `pkg/sync/service_test.go`, `pkg/db/sync_test.go`

**Task**: Write comprehensive unit tests
- [ ] Test version generation
- [ ] Test model-to-proto conversions
- [ ] Test database operations (RecordChange, GetChangesSince, etc.)
- [ ] Test file operations (GetFileInfo, ListDirectory, CreateDirectory, Delete, Move, Copy)
- [ ] Test upload operations (UploadFile, chunked upload)
- [ ] Test download operations (DownloadFile)
- [ ] Test change tracking logic
- [ ] Test sync status determination
- [ ] Test upload session management
- [ ] Test chunk operations

**Pattern**: Use `testing` package with `github.com/stretchr/testify/assert`

#### 7.2 Integration Tests
**Status**: DEFERRED - Add after unit tests complete

**Task**: Write integration tests with real database
- [ ] Set up test database
- [ ] Create test user and repository
- [ ] Test full sync workflow
- [ ] Verify change log entries
- [ ] Cleanup

#### 7.3 End-to-End Tests
**Status**: DEFERRED - Add after integration tests complete

**Task**: Write end-to-end tests with HTTP client
- [ ] Start test HTTP server
- [ ] Test all API endpoints
- [ ] Verify authentication
- [ ] Test chunked upload flow
- [ ] Cleanup

---

### 🔄 Phase 8: Optimization and Advanced Features (DEFERRED)

#### 8.1 Hash-Based Diff Calculation
**Status**: DEFERRED - Not needed for current requirements

**Rationale**: Simple file transfer + SHA-256 verification is sufficient

#### 8.2 Content-Defined Chunking (CDC)
**Status**: DEFERRED - Not needed for current requirements

**Rationale**: Fixed-size chunks (1MB) are simpler and adequate

#### 8.3 Recursive Directory Listing
**Decision**: Implement as opt-in parameter

**Status**: DEFERRED - Implement if needed

**Rationale**: Flat listing by default is more efficient. See docs/SYNC_IMPLEMENTATION.md

---

## Quick Start

### 1. Apply Database Schema
```bash
psql -d filehub -f scripts/sync_schema.sql
```

### 2. Build and Run
```bash
go build -o bin/file-hub cmd/main.go
./bin/file-hub
```

### 3. Test the API
```bash
# Get file info
curl "http://localhost:8080/api/sync/info?repo=myrepo&path=/documents/file.txt" \
  -H "Cookie: filehub_session=<session_id>"

# List directory (flat listing)
curl "http://localhost:8080/api/sync/list?repo=myrepo&path=/documents&offset=0&limit=50"

# Get changes for incremental sync
curl "http://localhost:8080/api/sync/changes?repo=myrepo&since=v1234567890-123456&limit=100"
```

---

## Documentation

- **docs/SYNC_IMPLEMENTATION.md** - Complete implementation details and usage examples
- **README.md** - Project overview
- **API documentation** - See docs/SYNC_IMPLEMENTATION.md for endpoint details

---

**File**: `pkg/sync/conversion.go`

**Task**: Convert between database models and protobuf types
- [ ] `ModelToProto(file *model.FileObject) *FileInfo` - Convert DB model to protobuf
- [ ] `ProtoToModel(repoID int, info *FileInfo) *model.FileObject` - Convert protobuf to DB model
- [ ] Test with: `TestModelToProtoConversion`, `TestProtoToModelConversion`

**Pattern**: Follow existing conversion patterns in `pkg/stor/storage.go` for `FileMeta` to `FileObject` conversion

---

### Phase 3: Basic File Operations (Simple RPCs)

#### 3.1 GetFileInfo
**File**: `pkg/sync/service.go`

**Task**: Implement GetFileInfo RPC method
- [ ] Extract user ID from context
- [ ] Get repository by name from database
- [ ] Get file metadata using `pkg/db.GetFile(ctx, repoID, path)`
- [ ] Convert to protobuf FileInfo
- [ ] Return response with success=true or appropriate error

**Dependencies**:
- File path: `pkg/model/resource.go` (for Resource struct)
- Database function: `pkg/db/file.go::GetFile`
- Authentication context: Need to implement in Phase 5

#### 3.2 ListDirectory
**File**: `pkg/sync/service.go`

**Task**: Implement ListDirectory RPC method
- [ ] Validate request parameters
- [ ] Get parent directory ID from database
- [ ] Use `pkg/db.GetChildFiles(ctx, parentID)`
- [ ] Implement pagination if offset/limit provided
- [ ] Convert results to protobuf FileInfo array
- [ ] Return response with items, total_count, has_more

**Pattern**: Reuse `pkg/stor/ListDir` logic

#### 3.3 CreateDirectory
**File**: `pkg/sync/service.go`

**Task**: Implement CreateDirectory RPC method
- [ ] Check if path already exists
- [ ] Create directory entry using `pkg/db.CreateFile(ctx, fileObject)`
- [ ] Create actual directory in storage using `pkg/stor.CreateDir`
- [ ] Record change in change_log
- [ ] Update repository version
- [ ] Return success response

**Dependencies**:
- Database: `pkg/db/file.go::CreateFile`
- Storage: `pkg/stor/storage.go::CreateDir`

#### 3.4 Delete
**File**: `pkg/sync/service.go`

**Task**: Implement Delete RPC method
- [ ] Check if path exists
- [ ] If directory and recursive=true, delete all children
- [ ] Mark as deleted in database using `pkg/db.DeleteFileByPath`
- [ ] Delete from storage using `pkg/stor.DeleteFile`
- [ ] Record change in change_log
- [ ] Update repository version
- [ ] Return success response

**Dependencies**:
- Database: `pkg/db/file.go::DeleteFileByPath`
- Storage: `pkg/stor/storage.go::DeleteFile`

#### 3.5 Move and Copy
**File**: `pkg/sync/service.go`

**Task**: Implement Move and Copy RPC methods

For Move:
- [ ] Validate source and destination
- [ ] Use `pkg/stor.MoveFile(srcResource, destResource)`
- [ ] Record change_log with operation='move', old_path, path
- [ ] Update repository version
- [ ] Return success response

For Copy:
- [ ] Validate source and destination
- [ ] Use `pkg/stor.CopyFile(srcResource, destResource)`
- [ ] Record change_log with operation='copy'
- [ ] Update repository version
- [ ] Return success response

**Dependencies**:
- Storage: `pkg/stor/storage.go::MoveFile`, `pkg/stor/storage.go::CopyFile`
- Model: `pkg/model/resource.go` (for Resource struct)

---

### Phase 4: Upload and Download Operations

#### 4.1 Simple Upload (UploadFile)
**File**: `pkg/sync/service.go`

**Task**: Implement UploadFile RPC method (for small files)
- [ ] Validate file size (limit to 10MB for simple upload)
- [ ] Calculate SHA-256 hash of content
- [ ] Create Resource object from request
- [ ] Use `pkg/stor.PutFile(ctx, resource, bytes.NewReader(content))`
- [ ] Update database with checksum and size
- [ ] Record change_log with operation='create'
- [ ] Update repository version
- [ ] Return response with server-computed etag and version

**Dependencies**:
- Storage: `pkg/stor/storage.go::PutFile`
- Database: `pkg/db/file.go::UpsertFile`

#### 4.2 Simple Download (DownloadFile - streaming)
**File**: `pkg/sync/service.go`

**Task**: Implement DownloadFile with streaming support
- [ ] Check if-none-match header for conditional download
- [ ] Get file info from database
- [ ] Open file using `pkg/stor.OpenFile(ctx, resource)`
- [ ] Stream file in chunks (e.g., 1MB chunks)
- [ ] Send FileInfo first
- [ ] Send chunks
- [ ] Send DownloadComplete at the end

**Pattern**: gRPC streaming - send multiple messages using stream.Send()

#### 4.3 Chunked Upload System
**Files**: `pkg/sync/upload.go`, `pkg/sync/models.go`

**Task**: Implement chunked upload for large files

**Database Schema** - Add to Phase 1.1:
```sql
CREATE TABLE upload_sessions (
    id SERIAL PRIMARY KEY,
    upload_id VARCHAR(64) UNIQUE NOT NULL,  -- Client-generated UUID
    repo_id INTEGER NOT NULL REFERENCES repositories(id),
    path TEXT NOT NULL,
    total_size BIGINT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 day'
);

CREATE TABLE upload_chunks (
    id SERIAL PRIMARY KEY,
    upload_id VARCHAR(64) NOT NULL REFERENCES upload_sessions(upload_id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    offset BIGINT NOT NULL,
    size BIGINT NOT NULL,
    checksum VARCHAR(64),
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(upload_id, chunk_index)
);
```

**Implementation**:
- [ ] `BeginUpload` - Create upload session, return existing chunks for resume
- [ ] `UploadChunk` - Store chunk data, mark as uploaded
- [ ] `FinalizeUpload` - Validate all chunks, assemble file, create database entry
- [ ] `CancelUpload` - Clean up upload session and chunks

**Implementation Details**:

```go
// BeginUpload
func (s *SyncService) BeginUpload(ctx context.Context, req *BeginUploadRequest) (*BeginUploadResponse, error) {
    // 1. Validate upload_id is a valid UUID
    // 2. Check if session already exists (resume case)
    // 3. If exists, return list of already uploaded chunks
    // 4. If new, create upload_session entry
    // 5. Return upload_id and empty uploaded_chunks array
}

// UploadChunk
func (s *SyncService) UploadChunk(ctx context.Context, req *UploadChunkRequest) (*UploadChunkResponse, error) {
    // 1. Validate session exists and not expired
    // 2. Store chunk data in temporary storage (filesystem or memory)
    // 3. Calculate checksum of chunk
    // 4. Insert/update upload_chunks record
    // 5. Return success
}

// FinalizeUpload
func (s *SyncService) FinalizeUpload(ctx context.Context, req *FinalizeUploadRequest) (*FinalizeUploadResponse, error) {
    // 1. Verify all chunks uploaded
    // 2. Assemble chunks into final file
    // 3. Calculate final SHA-256
    // 4. Compare with expected_etag
    // 5. Store file using pkg/stor.PutFile
    // 6. Create database entry
    // 7. Clean up upload session
    // 8. Record change_log
    // 9. Update repository version
    // 10. Return success with etag and version
}

// CancelUpload
func (s *SyncService) CancelUpload(ctx context.Context, req *CancelUploadRequest) (*CancelUploadResponse, error) {
    // 1. Delete upload session and chunks
    // 2. Clean up temporary chunk storage
    // 3. Return success
}
```

---

### Phase 5: Change Tracking and Sync Management

#### 5.1 GetVersion and ListChanges
**File**: `pkg/sync/service.go`

**Task**: Implement version and change tracking RPCs

GetVersion:
- [ ] Get repository version from database
- [ ] Return version identifier and timestamp
- [ ] If no version yet, return empty version

ListChanges:
- [ ] Get current version for comparison
- [ ] Query change_log: WHERE repo_id = ? AND timestamp > version_timestamp
- [ ] Categorize changes into created, modified, deleted, renamed
- [ ] Implement pagination if max_changes specified
- [ ] Return changes with current_version
- [ ] If since_version is too old, set version_expired=true

#### 5.2 GetSyncStatus
**File**: `pkg/sync/service.go`

**Task**: Implement sync status comparison
- [ ] Get file info from server
- [ ] Compare client_etag and client_version with server
- [ ] Determine status:
    - SYNCED: etag and version match
    - MODIFIED: etag different (server modified)
    - NEW: file doesn't exist on server
    - DELETED: file exists on client but not server
    - CONFLICT: client claims modification but server also modified (different version)
- [ ] Return status with server info

#### 5.3 BatchOperation (Optional - Defer to later)
**File**: `pkg/sync/service.go`

**Task**: Implement batch operations for efficiency
- [ ] Iterate through batch items
- [ ] Execute each operation sequentially
- [ ] Collect results
- [ ] Return array of BatchResult with success/failure for each
- [ ] If one operation fails, continue with others

**Note**: Implement last as optimization after all individual operations work.

---

### Phase 6: Authentication and Integration

#### 6.1 gRPC Authentication Middleware
**File**: `pkg/sync/auth.go`

**Task**: Create gRPC authentication interceptor
```go
// AuthInterceptor validates session cookies or Basic/Digest auth for gRPC
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // Extract user from context (similar to pkg/web/auth/auth.go::GetSessionUser)
    // Support both session cookie and Authorization header
    // For gRPC, need to parse metadata from context
    // Set user in context using context.WithValue(ctx, "user", user)
    // Call handler(ctx, req)
    // Return nil, status.Error(codes.Unauthenticated, ...) if auth fails
}
```

**Pattern**: gRPC server interceptors use metadata from context, not HTTP headers directly.

#### 6.2 Register gRPC Server
**File**: `cmd/main.go`

**Task**: Register sync service with gRPC server
- [ ] Create gRPC server with auth interceptor
- [ ] Register sync service
- [ ] Create listener (reuse existing port or add separate port for gRPC)
- [ ] Start gRPC server

**Alternative**: Use grpc-gateway to expose gRPC via REST API (defer to later).

#### 6.3 gRPC Server Options
**File**: `pkg/sync/server.go`

**Task**: Configure gRPC server properly
- [ ] Configure max message size (for large file uploads) - e.g., 100MB
- [ ] Configure keepalive parameters
- [ ] Configure connection timeout

---

### Phase 7: Testing

#### 7.1 Unit Tests
**Files**: `pkg/sync/service_test.go`, `pkg/sync/version_test.go`

**Task**: Write comprehensive unit tests
- [ ] Test version generation and incrementing
- [ ] Test model-to-proto conversions
- [ ] Test each RPC method with mocked database
- [ ] Test change tracking logic
- [ ] Test chunked upload/resume scenarios
- [ ] Test sync status determination

**Pattern**: Use `testing` package with `github.com/stretchr/testify/assert` for assertions (matching existing test patterns in `pkg/web/auth/auth_test.go`)

#### 7.2 Integration Tests
**File**: `pkg/sync/integration_test.go`

**Task**: Write integration tests with real database
- [ ] Set up test database
- [ ] Create test user and repository
- [ ] Initialize sync service
- [ ] Test full sync workflow:
    - Upload file
    - Download file
    - List directory
    - Get changes
    - Delete file
- [ ] Verify change log entries
- [ ] Verify version updates
- [ ] Cleanup

#### 7.3 End-to-End Tests
**File**: `pkg/sync/e2e_test.go`

**Task**: Write end-to-end tests with gRPC client
- [ ] Start test gRPC server
- [ ] Create gRPC client connection
- [ ] Create sync client
- [ ] Test complete sync cycle:
    - BeginUpload -> UploadChunk -> FinalizeUpload
    - Check version
    - Download file
    - Verify checksum
- [ ] Cleanup

---

### Phase 8: Optimization and Advanced Features (Defer)

#### 8.1 Hash-Based Diff Calculation
**Recommendation**: **DEFER - Not Needed for MVP**

The protobuf sync protocol with version-based change detection is sufficient. Binary diff algorithms add complexity without significant benefit for most use cases.

**If implementing later**, consider:
- Use `github.com/pierrec/lz4` for compression
- Use `github.com/cespare/xxhash` for fast hashing
- Don't implement rolling hash - keep it simple

#### 8.2 Content-Defined Chunking (CDC)
**Recommendation**: **DEFER - Not Needed for MVP**

Fixed-size chunks (1MB default) are simpler and work well enough. CDC (variable-size chunks) is only beneficial for deduplication at the block level.

#### 8.3 gRPC-Gateway for REST API
**Recommendation**: **DEFER - Optional**

The primary client will use gRPC directly. REST API can be added later if needed for browser-based clients.

---

## External Dependencies

### Required Packages
```go
require (
    google.golang.org/grpc v1.60.0  // gRPC framework
    google.golang.org/protobuf v1.32.0  // Protocol buffers
    github.com/google/uuid v1.5.0  // UUID generation for upload sessions
)
```

### Recommended Packages (for later)
```go
require (
    github.com/pierrec/lz4/v4 v4.1.17  // Compression (optional)
    github.com/cespare/xxhash/v2 v2.2.0  // Fast hashing (optional)
)
```

**Avoid**:
- Rolling hash implementations (Rabin fingerprint) - overkill for this use case
- Binary diff libraries (bsdiff, xdelta) - not needed with full file sync

---

## Implementation Priority

### Must Have (MVP):
1. Phase 1: Database schema extensions
2. Phase 2: Core service structure
3. Phase 3: Basic file operations (GetFileInfo, ListDirectory, CreateDirectory, Delete, Move, Copy)
4. Phase 4: Simple Upload/Download (no chunking yet)
5. Phase 5: GetVersion, ListChanges, GetSyncStatus
6. Phase 6: Authentication and gRPC server registration
7. Phase 7: Basic unit tests

### Should Have (Week 2-3):
1. Phase 4.3: Chunked upload/resume system
2. Phase 7.2-7.3: Integration and E2E tests

### Nice to Have (Later):
1. Phase 5.3: BatchOperation
2. Phase 8.1: Diff algorithms (consider if bandwidth is critical)
3. Phase 8.2: Content-defined chunking (consider if deduplication is needed)
4. Phase 8.3: gRPC-Gateway for REST API

---

## Key Integration Points

### Existing Code Patterns to Follow

1. **Database Operations**: Follow `pkg/db/file.go` pattern
   - Use context in all database calls
   - Wrap errors with `fmt.Errorf("context: %w", err)`
   - Create separate model files in `pkg/model/`

2. **Storage Operations**: Reuse `pkg/stor/storage.go` functions
   - `PutFile`, `OpenFile`, `DeleteFile`, `MoveFile`, `CopyFile`
   - Create `*model.Resource` objects from requests

3. **Authentication**: Reuse `pkg/web/auth/auth.go` patterns
   - Session-based authentication for web clients
   - Support Basic/Digest auth for API clients
   - Extract gRPC metadata for authentication

4. **Testing**: Follow existing test patterns
   - Use `testing` package with `testify/assert`
   - Place tests in `*_test.go` files
   - Use table-driven tests for multiple scenarios

---

## Success Criteria

- [ ] All RPC methods in `sync.proto` are implemented
- [ ] File operations (create, read, update, delete, move, copy) work correctly
- [ ] Change tracking captures all file operations
- [ ] Version management allows incremental sync
- [ ] Chunked upload supports interruptions and resume
- [ ] Authentication integrates with existing system
- [ ] Unit tests cover main code paths (>70% coverage)
- [ ] Integration tests verify end-to-end workflows
- [ ] gRPC server is registered and accessible
- [ ] Documentation updated with API usage examples

---

## Notes for AI agents

1. **Incremental Implementation**: Do not try to implement all phases at once. Start with Phase 1, then Phase 2, etc.

2. **Reuse Existing Code**: heavily leverage existing packages (pkg/db, pkg/stor, pkg/model, pkg/web/auth) rather than re-implementing.

3. **Simple Over Complex**: The original TODOs.md described complex algorithms (rolling hash, delta compression). For MVP, simple SHA-256 hashing and full file transfer are sufficient. Complex algorithms can be added later if needed.

4. **Test As You Go**: Write tests immediately after implementing each function. The test files show exactly what the function should do.

5. **Database-First**: Always create/update database schema before implementing business logic. The database layer is the foundation.

6. **Follow Existing Patterns**: The codebase has consistent patterns for error handling, context usage, and resource management. Follow these patterns strictly.

7. **gRPC-Specific**: Remember that gRPC works differently from HTTP. Use streaming for large files, metadata for headers, and interceptors for middleware.

8. **Version Management**: Keep versioning simple. Timestamp-based versions work well for sync. Complex version vectors are overkill for single-server sync.

9. **Change Tracking**: The change_log is the heart of sync. Every file operation must be logged. This is how clients know what changed.

10. **Authentication Integration**: The gRPC server must validate users using the same system as the web interface. Reuse session management and HA1 hash validation.
