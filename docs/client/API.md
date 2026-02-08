# File Hub API Documentation

This document describes the complete API for interacting with File Hub, which clients can use to build file synchronization and management applications.

## Table of Contents

- [Overview](#overview)
- [Base URL](#base-url)
- [API Interfaces](#api-interfaces)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Overview

File Hub provides three main API interfaces:

1. **WebDAV API**: Standard WebDAV protocol for file operations (see [WEBDAV.md](WEBDAV.md))
2. **REST API**: HTTP endpoints for authentication and configuration
3. **Sync API**: Optimized sync protocol for mobile clients (see [SYNC.md](SYNC.md))

All APIs require authentication. See [AUTH.md](AUTH.md) for details.

## Base URL

The base URL depends on your File Hub server configuration:

```
http://your-server:8080
```

Or if configured with HTTPS:

```
https://your-server:8080
```

## API Interfaces

### WebDAV API

The WebDAV API is available at `/dav` and follows the WebDAV (RFC 4918) standard.
See [WEBDAV.md](WEBDAV.md) for complete details on supported methods and properties.

### Sync API

The Sync API is available at `/api/sync` and provides an optimized protocol for mobile clients.
See [SYNC.md](SYNC.md) for complete details on endpoints and sync workflow.

### REST API

The REST API is available at `/api` and provides endpoints for authentication and user operations.
Note: Initial server setup must be performed through the web console, not via API.

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content |
| 207 | Multi-Status (WebDAV) |
| 304 | Not Modified |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 500 | Internal Server Error |

### Error Response Format

For REST and Sync APIs, errors are returned as JSON:

```json
{
  "error": "Error message describing the problem"
}
```

For WebDAV APIs, errors are returned as XML:

```xml
<?xml version="1.0" encoding="utf-8" ?>
<D:error xmlns:D="DAV:">
  Error message
</D:error>
```

### Common Errors

- **401 Unauthorized**: Missing or invalid authentication
- **403 Forbidden**: Permission denied for the requested operation
- **404 Not Found**: Resource doesn't exist
- **500 Internal Server Error**: Server-side error

## Best Practices

### For Mobile Clients

- Use the **Sync API** for better performance and mobile optimizations
- Use **chunked uploads** for files larger than 10MB
- Implement **version-based sync** using `/api/sync/changes` and `/api/sync/version`
- Cache responses locally with **ETag** headers for conditional downloads
- Use **session-based authentication** for better UX on mobile

### For Desktop Clients

- Use the **WebDAV API** for broader compatibility with existing tools
- Implement standard **WebDAV clients** (e.g., Windows Explorer, macOS Finder)
- Support **both Basic and Digest authentication** for flexibility

### General Best Practices

1. **Pagination**: Always use pagination with `offset` and `limit` for directory listings
2. **Concurrency**: Limit concurrent requests to avoid overwhelming the server
3. **Retry Logic**: Implement exponential backoff for failed requests
4. **Caching**: Use ETag and Last-Modified headers for efficient caching
5. **Bandwidth**: Prefer differential sync using `/api/sync/changes` when possible
6. **Security**: Always use HTTPS in production environments
7. **Session Management**: Handle session timeouts gracefully
8. **Error Handling**: Implement proper error handling for all API calls
