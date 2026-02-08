# File Hub WebDAV API

The WebDAV API is available at `/dav` and follows the WebDAV (RFC 4918) standard.

## Base Path

```
/dav/{repository_name}/{path}
```

## Supported WebDAV Methods

### PROPFIND - List directory contents or get file properties

Lists files and directories or retrieves metadata for a specific resource.

**Request:**
```
PROPFIND /dav/{repo}/{path}
```

**Headers:**
- `Authorization`: Basic or Digest auth (required)
- `Depth`: `0` (single resource) or `1` (resource + children)
- `Content-Type`: `text/xml`

**Request Body (optional):**
```xml
<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:">
  <D:prop>
    <D:displayname/>
    <D:getcontenttype/>
    <D:getcontentlength/>
    <D:getlastmodified/>
    <D:resourcetype/>
    <D:getetag/>
  </D:prop>
</D:propfind>
```

**Response (207 Multi-Status):**
```xml
<?xml version="1.0" encoding="utf-8" ?>
<D:multistatus xmlns:D="DAV:">
  <D:response>
    <D:href>/dav/{repo}/{path}</D:href>
    <D:propstat>
      <D:prop>
        <D:displayname>filename.txt</D:displayname>
        <D:getcontenttype>text/plain</D:getcontenttype>
        <D:getcontentlength>1024</D:getcontentlength>
        <D:getlastmodified>Mon, 08 Feb 2026 18:00:00 GMT</D:getlastmodified>
        <D:resourcetype/>
        <D:getetag>1a2b3c4d-400</D:getetag>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>
```

**Depth Header Values:**
- `0`: Returns metadata only for the specified resource
- `1`: Returns metadata for the resource and its direct children (one level)

### PUT - Upload a file

Uploads a file to the server.

**Request:**
```
PUT /dav/{repo}/{path}
```

**Headers:**
- `Authorization`: Basic or Digest auth (required)
- `Content-Type`: MIME type of the file
- `Content-Length`: Size of the file in bytes

**Response:** 201 Created (on success)

### GET - Download a file

Downloads a file from the server.

**Request:**
```
GET /dav/{repo}/{path}
```

**Headers:**
- `Authorization`: Basic or Digest auth (required)

**Response (200 OK):**
- Body: File content
- Headers:
  - `Content-Type`: MIME type
  - `Content-Length`: Size in bytes
  - `Last-Modified`: Last modified timestamp
  - `ETag`: File identifier

### DELETE - Delete a file or directory

Deletes a file or directory.

**Request:**
```
DELETE /dav/{repo}/{path}
```

**Headers:**
- `Authorization`: Basic or Digest auth (required)

**Response:** 204 No Content (on success)

### MKCOL - Create a directory

Creates a new directory.

**Request:**
```
MKCOL /dav/{repo}/{path}
```

**Headers:**
- `Authorization`: Basic or Digest auth (required)

**Response:** 201 Created (on success)

### COPY - Copy a file or directory

Copies a file or directory to a new location.

**Request:**
```
COPY /dav/{repo}/{source_path}
```

**Headers:**
- `Authorization`: Basic or Digest auth (required)
- `Destination`: `/dav/{repo}/{destination_path}`

**Response:** 201 Created (on success)

### MOVE - Move a file or directory

Moves a file or directory to a new location.

**Request:**
```
MOVE /dav/{repo}/{source_path}
```

**Headers:**
- `Authorization`: Basic or Digest auth (required)
- `Destination`: `/dav/{repo}/{destination_path}`

**Response:** 201 Created (on success)

### OPTIONS - List supported methods

Returns the HTTP methods supported by the WebDAV endpoint.

**Request:**
```
OPTIONS /dav/{repo}/{path}
```

**Headers:**
- `Authorization`: Not required for OPTIONS

**Response (200 OK):**
- Headers:
  - `Allow`: `OPTIONS,GET,POST,PUT,DELETE,COPY,MOVE,PROPFIND,MKCOL,LOCK,UNLOCK`
  - `DAV`: `1`

## WebDAV Headers

The server includes these headers in all WebDAV responses:

- `DAV: 1` - Indicates WebDAV Level 1 support
- `MS-Author-Via: DAV` - Indicates Microsoft client compatibility

## WebDAV Properties

The server supports the following WebDAV properties:

| Property | Description | Type |
|----------|-------------|------|
| `displayname` | Display name of the resource | String |
| `resourcetype` | Type of resource (empty or collection) | XML |
| `getcontenttype` | MIME type | String |
| `getcontentlength` | Size in bytes | Integer |
| `getlastmodified` | Last modified timestamp | String (RFC 1123) |
| `creationdate` | Creation timestamp | String (RFC 3339) |
| `getetag` | Entity tag for caching | String |

For directories, `resourcetype` is set to `<D:collection/>`, and `getcontenttype` is `httpd/unix-directory`.