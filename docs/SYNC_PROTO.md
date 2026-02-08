# File Hub Sync Protocol Specification

## Overview

The File Hub Sync Protocol is a custom synchronization protocol designed to complement the existing WebDAV implementation. It addresses specific limitations of WebDAV when used on mobile devices, particularly for uploading files to a home server. The protocol uses Protocol Buffers over HTTP/2 to provide efficient, low-overhead communication optimized for mobile environments.

## Motivation

While WebDAV provides excellent compatibility with existing file managers and tools, it has several limitations for mobile device synchronization:

1. **High overhead**: WebDAV's XML-based PROPFIND responses can be verbose, especially when requesting file metadata
2. **Authentication inefficiency**: HTTP Basic/Digest authentication requires sending credentials with each request
3. **Limited binary transfer optimization**: WebDAV's PUT/GET methods don't support efficient binary diff synchronization
4. **Lack of batch operations**: WebDAV requires separate requests for each operation, increasing round trips
5. **Poor mobile connectivity handling**: WebDAV doesn't handle intermittent connections well
6. **No delta sync**: WebDAV requires full directory listings even when only a few files changed
7. **No version tracking**: No mechanism to track changes between sync sessions

## Design Goals

1. **Optimized for mobile**: Minimize data usage and handle intermittent connections
2. **Efficient serialization**: Use Protocol Buffers for compact binary representation
3. **HTTP/2 transport**: Leverage HTTP/2 features like multiplexing and header compression
4. **Lightweight authentication**: Reduce authentication overhead compared to WebDAV
5. **Batch operations**: Support multiple operations in a single request
6. **Chunk-based transfers**: Handle unstable internet connections gracefully
7. **Version-based delta sync**: Only sync files that have changed since last sync
8. **Change tracking**: Track changes between sync sessions on both client and server
9. **Conflict resolution**: Built-in mechanisms for handling concurrent modifications

## Transport Layer

The protocol operates over HTTP/2 with Protocol Buffer-encoded messages. All requests and responses use the `application/grpc` content-type, though the protocol doesn't use gRPC directly.

### HTTP/2 Endpoints

```
POST /sync/upload
POST /sync/download
POST /sync/fileinfo
POST /sync/listdir
POST /sync/listchanges
POST /sync/getversion
POST /sync/mkdir
POST /sync/delete
POST /sync/move
POST /sync/copy
POST /sync/syncstatus
POST /sync/batch
```

Each endpoint accepts a Protocol Buffer request body and returns a Protocol Buffer response body.

## Authentication

The protocol uses a token-based authentication system optimized for mobile devices:

1. **Initial authentication**: Mobile clients authenticate once using Basic/Digest authentication to obtain a session token
2. **Token-based requests**: Subsequent requests use the session token in an `Authorization: Bearer <token>` header
3. **Token refresh**: Tokens expire after a configurable period, requiring re-authentication
4. **Token revocation**: Clients can explicitly revoke tokens when logging out

This approach minimizes the overhead of repeatedly sending credentials with each request, unlike WebDAV's Basic/Digest authentication.

## Version-Based Delta Synchronization

To minimize data transfer, the protocol supports version-based delta synchronization:

### Server-Side Version Tracking

1. **Version maintenance**: Server maintains recent versions of directory states and changes
2. **Expiration policy**: Old versions are periodically removed based on configurable retention policies
3. **Change logs**: Server keeps track of file operations (create, modify, delete, move) between versions

### Client-Side State Tracking

1. **Last synced version**: Client stores the version identifier from the last successful sync
2. **Unsynced changes**: Client tracks all local changes since the last sync (uploads, deletions, renames)
3. **Conflict detection**: Client compares local changes with server changes to detect conflicts

### Delta Operations

- **ListChanges**: Client provides its last synced version, server returns changes since that version
- **Fallback to full sync**: If the client's version is too old (expired on server), server returns full directory listing

## Chunk-Based File Transfer

To handle unstable internet connections, the protocol implements chunk-based file transfers:

### Upload with Chunks

For large files, the client can split the file into chunks and upload them individually:

1. **Initiate upload**: Client sends `BeginUploadRequest` to start a chunked upload session
2. **Upload chunks**: Client uploads individual chunks with `UploadChunkRequest`, specifying the chunk sequence number
3. **Verify chunks**: Server acknowledges each chunk and reports any missing chunks
4. **Finalize upload**: Client sends `FinalizeUploadRequest` to complete the file assembly

### Download with Chunks

For downloads, the protocol supports resumable downloads:

1. **Request download**: Client requests a file download with optional range parameters
2. **Stream chunks**: Server streams file content in chunks
3. **Resume capability**: If connection drops, client can resume from the last received byte

## Protocol Buffer Definitions

```protobuf
syntax = "proto3";

package filehub.sync;

option go_package = "github.com/cgang/file-hub/pkg/sync";

// SyncService defines the API for file synchronization
service SyncService {
  // Upload a file to the server (for small files)
  rpc UploadFile(UploadFileRequest) returns (UploadFileResponse);
  
  // Chunked upload operations (for large files or unstable connections)
  rpc BeginUpload(BeginUploadRequest) returns (BeginUploadResponse);
  rpc UploadChunk(UploadChunkRequest) returns (UploadChunkResponse);
  rpc FinalizeUpload(FinalizeUploadRequest) returns (FinalizeUploadResponse);
  rpc CancelUpload(CancelUploadRequest) returns (CancelUploadResponse);
  
  // Download a file from the server
  rpc DownloadFile(DownloadFileRequest) returns (stream DownloadFileResponse);
  
  // Get file metadata
  rpc GetFileInfo(GetFileInfoRequest) returns (GetFileInfoResponse);
  
  // List directory contents
  rpc ListDirectory(ListDirectoryRequest) returns (ListDirectoryResponse);
  
  // Get changes since a specific version
  rpc ListChanges(ListChangesRequest) returns (ListChangesResponse);
  
  // Get current version identifier
  rpc GetCurrentVersion(GetCurrentVersionRequest) returns (GetCurrentVersionResponse);
  
  // Create a directory
  rpc CreateDirectory(CreateDirectoryRequest) returns (CreateDirectoryResponse);
  
  // Delete a file or directory
  rpc Delete(DeleteRequest) returns (DeleteResponse);
  
  // Move/rename a file or directory
  rpc Move(MoveRequest) returns (MoveResponse);
  
  // Copy a file or directory
  rpc Copy(CopyRequest) returns (CopyResponse);
  
  // Get sync status for a path
  rpc GetSyncStatus(SyncStatusRequest) returns (SyncStatusResponse);
  
  // Batch operations for efficiency
  rpc BatchOperation(BatchOperationRequest) returns (BatchOperationResponse);
}

// Common types
message FileInfo {
  string path = 1;
  int64 size = 2;
  int64 mod_time = 3;  // Unix timestamp
  bool is_dir = 4;
  string mime_type = 5;
  string etag = 6;     // Content hash for change detection
  int64 version = 7;   // Version number for conflict resolution
}

message Repository {
  string name = 1;
  string owner = 2;
  string root_path = 3;
}

// Version tracking messages
message GetCurrentVersionRequest {
  string repo = 1;
  string path = 2;  // Directory path to get version for
}

message GetCurrentVersionResponse {
  bool success = 1;
  string version = 2;  // Unique identifier for current state
  int64 timestamp = 3; // Time when version was created
  string error_message = 4;
}

// Change tracking messages
message ListChangesRequest {
  string repo = 1;
  string path = 2;           // Directory path to sync
  string since_version = 3;  // Client's last synced version
  int32 max_changes = 4;     // Maximum number of changes to return (pagination)
  string continuation_token = 5; // Token for pagination of large change sets
}

message ListChangesResponse {
  bool success = 1;
  string current_version = 2; // Current server version
  bool version_expired = 3;   // True if since_version is too old and expired
  repeated FileInfo created = 4;      // Newly created files/directories
  repeated FileInfo modified = 5;     // Modified files/directories
  repeated string deleted = 6;        // Deleted file/directory paths
  repeated RenameOperation renamed = 7; // Renamed files/directories
  bool has_more = 8;                  // Whether more changes are available
  string continuation_token = 9;      // Token for getting next page of changes
  string error_message = 10;
}

message RenameOperation {
  string old_path = 1;
  string new_path = 2;
}

// Chunked upload messages
message BeginUploadRequest {
  string repo = 1;
  string path = 2;
  int64 total_size = 3;
  int64 mod_time = 4;
  string mime_type = 5;
  string upload_id = 6;  // Client-generated UUID for resumability
}

message BeginUploadResponse {
  bool success = 1;
  string upload_id = 2;  // Server-confirmed upload ID
  repeated int32 uploaded_chunks = 3;  // Already uploaded chunks (for resume)
  string error_message = 4;
}

message UploadChunkRequest {
  string upload_id = 1;  // Matches the ID from BeginUpload
  int32 chunk_index = 2; // Zero-based chunk index
  bytes data = 3;        // Chunk data
  int64 offset = 4;      // Byte offset in the original file (optional)
}

message UploadChunkResponse {
  bool success = 1;
  int32 chunk_index = 2;
  string error_message = 3;
}

message FinalizeUploadRequest {
  string upload_id = 1;
  string expected_etag = 2;  // Client-computed hash for verification
}

message FinalizeUploadResponse {
  bool success = 1;
  string etag = 2;  // Server-computed hash
  int64 version = 3;
  string error_message = 4;
}

message CancelUploadRequest {
  string upload_id = 1;
}

message CancelUploadResponse {
  bool success = 1;
  string error_message = 2;
}

// Regular upload (for small files)
message UploadFileRequest {
  string repo = 1;
  string path = 2;
  bytes content = 3;
  int64 size = 4;
  int64 mod_time = 5;
  string mime_type = 6;
  string etag = 7;  // Client-computed hash for verification
}

message UploadFileResponse {
  bool success = 1;
  string etag = 2;  // Server-computed hash
  int64 version = 3;
  string error_message = 4;
}

// DownloadFile with chunk support
message DownloadFileRequest {
  string repo = 1;
  string path = 2;
  string if_none_match = 3;  // For conditional download
  int64 offset = 4;          // Resume download from byte offset
  int64 length = 5;          // Number of bytes to download (optional)
}

message DownloadFileResponse {
  oneof response {
    FileInfo info = 1;
    bytes chunk = 2;
    DownloadComplete complete = 3;
  }
}

message DownloadComplete {
  string etag = 1;
  int64 version = 2;
}

// GetFileInfo
message GetFileInfoRequest {
  string repo = 1;
  string path = 2;
}

message GetFileInfoResponse {
  bool exists = 1;
  FileInfo info = 2;
  string error_message = 3;
}

// ListDirectory
message ListDirectoryRequest {
  string repo = 1;
  string path = 2;
  bool recursive = 3;
  int32 max_depth = 4;
  int32 offset = 5;
  int32 limit = 6;
}

message ListDirectoryResponse {
  repeated FileInfo items = 1;
  int32 total_count = 2;
  bool has_more = 3;
  string error_message = 4;
}

// CreateDirectory
message CreateDirectoryRequest {
  string repo = 1;
  string path = 2;
}

message CreateDirectoryResponse {
  bool success = 1;
  string error_message = 2;
}

// Delete
message DeleteRequest {
  string repo = 1;
  string path = 2;
  bool recursive = 3;  // For directories
}

message DeleteResponse {
  bool success = 1;
  string error_message = 2;
}

// Move
message MoveRequest {
  string repo = 1;
  string source_path = 2;
  string destination_path = 3;
  bool overwrite = 4;
}

message MoveResponse {
  bool success = 1;
  string error_message = 2;
}

// Copy
message CopyRequest {
  string repo = 1;
  string source_path = 2;
  string destination_path = 3;
  bool overwrite = 4;
}

message CopyResponse {
  bool success = 1;
  string error_message = 2;
}

// SyncStatus
message SyncStatusRequest {
  string repo = 1;
  string path = 2;
  string client_etag = 3;  // Client's version of the file
  int64 client_version = 4;  // Client's version number
}

message SyncStatusResponse {
  enum Status {
    UNKNOWN = 0;
    SYNCED = 1;
    MODIFIED = 2;
    DELETED = 3;
    CONFLICT = 4;
    NEW = 5;
  }
  
  Status status = 1;
  FileInfo server_info = 2;
  string error_message = 3;
}

// BatchOperation
message BatchOperationRequest {
  string repo = 1;
  repeated BatchItem items = 2;
}

message BatchItem {
  oneof operation {
    UploadFileRequest upload = 1;
    DownloadFileRequest download = 2;
    DeleteRequest delete = 3;
    MoveRequest move = 4;
    CopyRequest copy = 5;
    GetFileInfoRequest get_info = 6;
    CreateDirectoryRequest create_dir = 7;
    BeginUploadRequest begin_upload = 8;
    UploadChunkRequest upload_chunk = 9;
    FinalizeUploadRequest finalize_upload = 10;
    ListChangesRequest list_changes = 11;
    GetCurrentVersionRequest get_version = 12;
  }
}

message BatchOperationResponse {
  repeated BatchResult results = 1;
  string error_message = 2;
}

message BatchResult {
  int32 index = 1;  // Index in the original request
  bool success = 2;
  string result_data = 3;  // Serialized response for the specific operation
  string error_message = 4;
}
```

## Key Features

### 1. Version-Based Delta Synchronization

The protocol supports efficient version-based delta synchronization:

- **Server version tracking**: Server maintains recent versions of directory states and changes
- **Client state tracking**: Client stores its last synced version and tracks local changes
- **Change detection**: Server computes changes between client's last version and current state
- **Expiration handling**: If client's version is too old, server signals to perform full sync
- **Pagination**: Large change sets are returned in pages to prevent overwhelming clients

### 2. Unsynced Changes Tracking

Clients track all non-synced changes since the last state:

- **Local change log**: Client maintains a log of all local operations (uploads, deletions, renames)
- **Conflict resolution**: During sync, client resolves conflicts between local and server changes
- **Efficient sync**: Only unsynced changes are processed during sync operations

### 3. Chunk-Based Transfers

The protocol supports chunk-based file transfers to handle unstable internet connections:

- **Resumable uploads**: If a connection drops during upload, the client can resume from the last successfully uploaded chunk
- **Resumable downloads**: Similarly, downloads can be resumed from where they left off
- **Configurable chunk size**: Clients can adjust chunk sizes based on network conditions

### 4. Efficient Binary Transfer

The protocol supports both direct file uploads for small files and chunked uploads for large files or unstable connections.

### 5. Conditional Operations

Operations like download support conditional requests using ETags, reducing bandwidth when files haven't changed.

### 6. Batch Operations

Multiple operations can be combined in a single request, significantly reducing the number of round trips required for complex sync operations.

### 7. Conflict Detection

The protocol includes version numbers and ETags to detect conflicts when multiple clients modify the same file.

### 8. Connection Resilience

The chunk-based approach allows the protocol to gracefully handle connection interruptions, which is common on mobile networks.

### 9. Pagination Support

Directory listings and change responses support pagination to handle large directories efficiently on mobile devices with limited memory.

## Relationship with WebDAV

The Sync Protocol complements rather than replaces WebDAV:

- **WebDAV**: Best for general file access, integration with existing tools, and browsing
- **Sync Protocol**: Optimized for mobile apps, bulk operations, efficient synchronization, unreliable network conditions, and version-based delta sync

Both protocols operate on the same underlying file storage and share the same authentication system, ensuring consistent access controls.

## Mobile-Specific Optimizations

1. **Reduced bandwidth**: Protocol Buffers use significantly less bandwidth than XML
2. **Token authentication**: Reduces credential transmission overhead
3. **Batch operations**: Minimizes connection establishment overhead
4. **Conditional requests**: Avoids unnecessary data transfer
5. **Streamed transfers**: Allows for interruption and resumption
6. **Chunk-based transfers**: Handles unstable connections gracefully
7. **Progress tracking**: Clients can monitor upload/download progress
8. **Delta sync**: Only sync files that have changed since last sync
9. **Local change tracking**: Track unsynced changes to enable efficient sync operations

## Error Handling

All operations return structured error responses with specific error codes, allowing clients to implement appropriate retry and recovery strategies. The chunk-based approach allows for fine-grained error handling where individual chunks can be retried without restarting the entire file transfer. The version-based sync mechanism allows for efficient recovery from sync interruptions by resuming from the last known version, with fallback to full sync if the version has expired on the server.

This protocol design addresses the specific needs of mobile file synchronization while maintaining compatibility with the existing WebDAV infrastructure and adding robust support for unstable network conditions, efficient version-based delta synchronization, and comprehensive change tracking.