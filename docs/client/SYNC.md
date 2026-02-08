# File Hub Sync Protocol Documentation

This document describes the sync protocol for efficient file synchronization optimized for mobile clients.

## Table of Contents

- [Overview](#overview)
- [Protocol Design](#protocol-design)
- [Version Control](#version-control)
- [Sync Workflow](#sync-workflow)
- [Change Log](#change-log)
- [Chunked Upload](#chunked-upload)
- [Conflict Resolution](#conflict-resolution)
- [Performance Optimization](#performance-optimization)

## Overview

The File Hub sync protocol is designed for efficient synchronization between clients and the server, with special consideration for mobile devices:

- **Version-based sync**: Uses version numbers and version vectors
- **Change tracking**: Logs all file operations
- **Chunked uploads**: Supports resumable uploads for large files
- **Conditional downloads**: ETag-based caching
- **Bandwidth efficient**: Transfers only changed data

## Protocol Design

### Core Concepts

1. **Repository Version**: Each repository has a version number that increments on every change
2. **Change Log**: Tracks all operations (create, modify, delete, move, copy) with version info
3. **Version Vector**: Tracks changes from multiple users for conflict detection
4. **ETag**: SHA-256 checksum for file content verification

### Simplified Data Models for Client

#### RepositoryVersion

Simplified version state for client synchronization:

```json
{
  "version": "5",
  "vector": "{\"user1\":3,\"user2\":2}",
  "timestamp": "2026-02-08T18:00:00Z"
}
```

#### ChangeLog

Simplified change log for client synchronization:

```json
{
  "operation": "modify",
  "path": "/documents/report.txt",
  "old_path": null,
  "version": "5",
  "timestamp": "2026-02-08T18:00:00Z"
}
```

## Version Control

### Version Numbers

Each repository maintains a monotonically increasing version number that increments on every change. This provides:

- **Change detection**: Compare version numbers to detect changes
- **Sync point**: Client tracks last synced version
- **Conflict detection**: Compare version vectors

### Version Vectors

Version vectors track changes from multiple users to detect conflicts:

Format: JSON object mapping user IDs to their last change version

```json
{
  "user1": 3,
  "user2": 2
}
```

This means:
- User1 has made changes up to version 3
- User2 has made changes up to version 2

**Conflict Detection:**
- No conflict: Client's version is subset of server's version vector
- Conflict: Client and server have divergent changes

### Getting Current Version

**Endpoint:** `GET /api/sync/version?repo={repo_name}`

**Response:**
```json
{
  "version": "5",
  "vector": "{\"user1\":3,\"user2\":2}",
  "timestamp": "2026-02-08T18:00:00Z"
}
```

## Sync Workflow

### Initial Sync (First Time)

1. **Get current repository version**
2. **Get all changes** (empty `since` parameter)
3. **Download/create files** based on change log
4. **Store synced version** locally

```http
GET /api/sync/changes?repo=myrepo&since=&limit=1000 HTTP/1.1
Host: server:8080
Cookie: filehub_session=session_id
```

**Response:**
```json
{
  "version": "10",
  "changes": [
    {
      "repo_id": 1,
      "operation": "create",
      "path": "/photo.jpg",
      "user_id": 1,
      "version": "1",
      "timestamp": "2026-02-08T10:00:00Z"
    },
    {
      "repo_id": 1,
      "operation": "create",
      "path": "/document.pdf",
      "user_id": 1,
      "version": "2",
      "timestamp": "2026-02-08T11:00:00Z"
    }
  ],
  "changed": 2
}
```

### Incremental Sync

1. **Get current repository version**
2. **Get changes since last sync** (using stored version)
3. **Process changes**:
   - `create`: Download new file
   - `modify`: Download updated file
   - `delete`: Remove local file
   - `move`: Move/rename local file
   - `copy`: Copy local file
4. **Update stored version** locally

```http
GET /api/sync/changes?repo=myrepo&since=5&limit=100 HTTP/1.1
Host: server:8080
Cookie: filehub_session=session_id
```

**Response:**
```json
{
  "version": "8",
  "changes": [
    {
      "repo_id": 1,
      "operation": "modify",
      "path": "/document.pdf",
      "user_id": 1,
      "version": "6",
      "timestamp": "2026-02-08T12:00:00Z"
    },
    {
      "repo_id": 1,
      "operation": "create",
      "path": "/newfile.txt",
      "user_id": 1,
      "version": "7",
      "timestamp": "2026-02-08T13:00:00Z"
    },
    {
      "repo_id": 1,
      "operation": "delete",
      "path": "/oldfile.txt",
      "user_id": 1,
      "version": "8",
      "timestamp": "2026-02-08T14:00:00Z"
    }
  ],
  "changed": 3
}
```

### Bidirectional Sync

For clients that also upload changes:

1. **Download server changes** (as above)
2. **Check local changes** against server version
3. **Upload local changes** if no conflict
4. **Handle conflicts** if detected (see Conflict Resolution)
5. **Update local version** after successful sync

## Change Log

### Operations

| Operation | Description | Additional Fields |
|-----------|-------------|-------------------|
| `create` | New file/directory created | `path` |
| `modify` | File content modified | `path` |
| `delete` | File/directory deleted | `path` |
| `move` | File/directory moved/renamed | `path`, `old_path` |
| `copy` | File/directory copied | `path`, `old_path` |

### Processing Changes

#### Create Operation
```json
{
  "operation": "create",
  "path": "/newfile.txt",
  "version": "5"
}
```
**Action:** Download and create the file

#### Modify Operation
```json
{
  "operation": "modify",
  "path": "/existing.txt",
  "version": "6"
}
```
**Action:**
1. Check local file version
2. Download updated file if server version is newer

#### Delete Operation
```json
{
  "operation": "delete",
  "path": "/removed.txt",
  "version": "7"
}
```
**Action:** Delete local file or move to trash

#### Move Operation
```json
{
  "operation": "move",
  "path": "/newlocation/file.txt",
  "old_path": "/oldlocation/file.txt",
  "version": "8"
}
```
**Action:** Move/rename local file from `old_path` to `path`

#### Copy Operation
```json
{
  "operation": "copy",
  "path": "/copy.txt",
  "old_path": "/original.txt",
  "version": "9"
}
```
**Action:** Copy local file from `old_path` to `path`

## Chunked Upload

For large files (>10MB), use chunked uploads for better reliability and resume capability:

### Upload Flow

#### 1. Begin Upload

**Request:**
```http
POST /api/sync/upload/begin?repo=myrepo&path=/largefile.zip&total_size=15728640 HTTP/1.1
Host: server:8080
Cookie: filehub_session=session_id
```

**Response:**
```json
{
  "upload_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef",
  "total_chunks": 15,
  "chunk_size": 1048576,
  "uploaded_chunks": []
}
```

#### 2. Upload Chunks

Upload each chunk sequentially (Chunk 0, 1, 2, ...):

**Request:**
```http
POST /api/sync/upload/chunk?upload_id=a1b2c3d4...&chunk_index=0 HTTP/1.1
Host: server:8080
Cookie: filehub_session=session_id
Content-Length: 1048576

<1MB of chunk data>
```

**Response:**
```json
{
  "success": true,
  "message": "Chunk uploaded successfully"
}
```

#### 3. Finalize Upload

After uploading all chunks, assemble the file:

**Request:**
```http
POST /api/sync/upload/finalize?upload_id=a1b2c3d4...&repo=myrepo HTTP/1.1
Host: server:8080
Cookie: filehub_session=session_id
```

**Response:**
```json
{
  "etag": "sha256_hash_of_complete_file",
  "size": 15728640
}
```

#### 4. Resume Interrupted Upload

If upload fails or is interrupted, resume by:

1. **Query existing upload** (call `/api/sync/upload/begin` again with same parameters)
2. **Get list of uploaded chunks** from response
3. **Resume upload** from missing chunk indices

```json
{
  "upload_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef",
  "total_chunks": 15,
  "chunk_size": 1048576,
  "uploaded_chunks": [0, 1, 2, 3, 4]  // Chunks 0-4 already uploaded
}
```

Resume from chunk index 5.

#### 5. Cancel Upload

**Request:**
```http
DELETE /api/sync/upload/cancel?upload_id=a1b2c3d4... HTTP/1.1
Host: server:8080
Cookie: filehub_session=session_id
```

**Response:**
```json
{
  "success": true,
  "message": "Upload cancelled successfully"
}
```

### Chunk Size

- **Default chunk size**: 1 MiB (1,048,576 bytes)
- **Chunk size is fixed** by server configuration
- Total chunks calculated as: `ceil(total_size / chunk_size)`

### Upload Session Expiration

- Upload sessions expire after **24 hours**
- Clean up old sessions to free server resources
- Resume uploads before expiration

## Sync Status

Check sync status for individual files before uploading to detect conflicts.

**Endpoint:** `GET /api/sync/status?repo={repo}&path={path}&client_etag={etag}&client_version={version}`

**Parameters:**
- `repo`: Repository name
- `path`: File path
- `client_etag` (optional): Local file's ETag (SHA-256)
- `client_version` (optional): Client's stored version number

**Response:**
```json
{
  "status": "outdated",
  "info": {
    "id": 42,
    "name": "file.txt",
    "path": "/file.txt",
    "size": 2048,
    "checksum": "new_sha256_hash",
    "mod_time": "2026-02-08T18:10:00Z",
    "version": "8"
  }
}
```

### Status Values

| Status | Description | Action |
|--------|-------------|--------|
| `synced` | File is up to date | No action needed |
| `outdated` | Server has newer version | Download server version |
| `conflict` | Both client and server have changes | Resolve conflict (see below) |
| `not_found` | File doesn't exist on server | Upload client version |

## Conflict Resolution

### When Conflicts Occur

Conflicts happen when:
- Same file modified on both client and server
- File deleted on one side, modified on other
- Network partition causes divergent changes

### Version Vector Comparison

Compare client and server version vectors:

**No Conflict** (client can upload):
```
Server: {"user1": 5, "user2": 3}
Client: {"user1": 3}
```
- Client's changes are behind server's
- Safe to upload after pulling server changes

**Conflict** (needs resolution):
```
Server: {"user1": 5, "user2": 3}
Client: {"user2": 4}
```
- Divergent changes from user2
- Requires manual resolution

### Conflict Resolution Strategies

1. **Last Write Wins**: Use the file with higher timestamp
2. **User Selection**: Prompt user to choose version
3. **Both Versions**: Keep both files with different names
4. **Merge**: Attempt automatic merge (text files only)


## Performance Optimization

### Conditional Downloads

Use ETag and `If-None-Match` header to avoid downloading unchanged files:

```http
GET /api/sync/download?repo=myrepo&path=/file.txt HTTP/1.1
Host: server:8080
If-None-Match: abc123def456
Cookie: filehub_session=session_id
```

**Response (200 OK):** File data (ETag doesn't match)
**Response (304 Not Modified:** No data (ETag matches)

### Pagination

Use pagination for directory listings to avoid loading all items at once:

```http
GET /api/sync/list?repo=myrepo&path=/&offset=0&limit=100 HTTP/1.1
```

**Response:**
```json
{
  "items": [...],
  "total": 500,
  "offset": 0,
  "limit": 100,
  "has_more": true
}
```

Load subsequent pages with `offset=100`, `offset=200`, etc.

### Batch Operations

Group multiple operations to reduce HTTP calls:

- Batch multiple file uploads in single network request (future feature)
- Batch multiple change log queries
- Parallel independent downloads

### Background Sync

- Sync periodically in background
- Use WorkManager (Android) or background tasks
- Respect battery and network constraints
- Only sync when on WiFi if configured

