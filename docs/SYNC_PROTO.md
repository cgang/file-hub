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