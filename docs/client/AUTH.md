# File Hub Authentication Documentation

This document describes the authentication mechanisms supported by File Hub and how to implement them in client applications.

## Table of Contents

- [Overview](#overview)
- [Authentication Methods](#authentication-methods)
- [Session-Based Authentication](#session-based-authentication)
- [HTTP Basic Authentication](#http-basic-authentication)
- [HTTP Digest Authentication](#http-digest-authentication)
- [Authentication Flow](#authentication-flow)
- [Security Considerations](#security-considerations)

## Overview

File Hub supports three authentication methods:

1. **Session-Based Authentication**: Best for mobile and web clients
2. **HTTP Basic Authentication**: Simple but less secure
3. **HTTP Digest Authentication**: More secure, prevents password exposure

The server checks for authentication in this order:
1. Session cookie (if present)
2. Authorization header (Basic or Digest)

## Authentication Methods

### Session-Based Authentication (Recommended for Mobile)

Uses HTTP cookies to maintain authenticated sessions.

**Pros:**
- Secure - credentials sent only once
- Efficient - no need to include credentials in every request
- Standard - works well with HTTP clients

**Cons:**
- Requires cookie management
- Sessions expire after 24 hours

**Use Case:** Mobile apps, web applications

### HTTP Basic Authentication

Simple username/password authentication using Base64 encoding.

**Pros:**
- Simple to implement
- Widely supported

**Cons:**
- Passwords sent with every request (can be mitigated with HTTPS)
- Vulnerable if not using HTTPS

**Use Case:** Desktop clients, testing

### HTTP Digest Authentication

Challenge-response authentication that doesn't send passwords.

**Pros:**
- Passwords never transmitted
- More secure than Basic auth
- Works over plain HTTP (though HTTPS still recommended)

**Cons:**
- More complex to implement
- Requires nonce management
- Less widely supported

**Use Case:** Security-sensitive applications

## Session-Based Authentication

### Creating a Session

Login to create a session and receive a cookie.

**Request:**
```http
POST /api/login HTTP/1.1
Host: server:8080
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

**Response (200 OK):**
```http
HTTP/1.1 200 OK
Set-Cookie: filehub_session=session_id_here; Path=/; HttpOnly; SameSite=Lax
Content-Type: application/json

{
  "success": true
}
```

### Using a Session

Include the session cookie in subsequent requests.

**Request:**
```http
GET /api/sync/info?repo=myrepo&path=/file.txt HTTP/1.1
Host: server:8080
Cookie: filehub_session=session_id_here
```

### Destroying a Session

Logout to end the session.

**Request:**
```http
POST /api/logout HTTP/1.1
Host: server:8080
Cookie: filehub_session=session_id_here
```

**Response (200 OK):**
```http
HTTP/1.1 200 OK
Set-Cookie: filehub_session=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax
```

### Session Cookie Attributes

- **Name**: `filehub_session`
- **Path**: `/` (valid for all paths)
- **Secure**: `false` (localhost) or `true` (HTTPS)
- **HttpOnly**: `true` (not accessible via JavaScript)
- **SameSite**: `Lax` (CSRF protection)
- **Expiration**: 24 hours from creation

## HTTP Basic Authentication

### Authentication Header Format

```
Authorization: Basic {base64(username:password)}
```

### Example

**Request:**
```http
GET /dav/myrepo/file.txt HTTP/1.1
Host: server:8080
Authorization: Basic YWxhZGRpbjpvcGVuc2VzYW1l
```

Where `YWxhZGRpbjpvcGVuc2VzYW1l` is `base64("username:password")`.

### Server Response

On successful authentication:
- Request proceeds normally

On failed authentication:
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Basic realm="file-hub"
Content-Type: text/plain

Unauthorized
```

## HTTP Digest Authentication

### Authentication Header Format

```
Authorization: Digest username="{username}", realm="{realm}", nonce="{nonce}", uri="{uri}", qop="auth", nc="{nc}", cnonce="{cnonce}", response="{response}", opaque="{opaque}", algorithm="MD5"
```

### Fields Explained

| Field | Description |
|-------|-------------|
| `username` | User's username |
| `realm` | Authentication realm (from server) |
| `nonce` | Random string from server (prevents replay) |
| `uri` | Request URI being accessed |
| `qop` | Quality of protection (always "auth") |
| `nc` | Nonce count (hex, increments each request) |
| `cnonce` | Client-generated random string |
| `response` | Calculated response hash |
| `opaque` | Server string (echoed back) |
| `algorithm` | Hash algorithm (always "MD5") |

### Calculating the Response

1. **Calculate HA1:**
   ```
   HA1 = MD5(username:realm:password)
   ```

2. **Calculate HA2:**
   ```
   HA2 = MD5(method:uri)
   ```

3. **Calculate Response:**
   ```
   response = MD5(HA1:nonce:nc:cnonce:qop:HA2)
   ```

### Example Flow

#### Step 1: Server Challenge (401 Unauthorized)

```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Digest realm="file-hub", nonce="a1b2c3d4e5f6g7h8", opaque="x9y8z7w6v5u4t3s2", algorithm="MD5", qop="auth"
```

#### Step 2: Client Response

```http
GET /dav/myrepo/file.txt HTTP/1.1
Host: server:8080
Authorization: Digest username="alice", realm="file-hub", nonce="a1b2c3d4e5f6g7h8", uri="/dav/myrepo/file.txt", qop="auth", nc="00000001", cnonce="0a4b113c", response="6629fae49393a05397450978507c4ef1", opaque="x9y8z7w6v5u4t3s2", algorithm="MD5"
```


### Incrementing NC (Nonce Count)

The `nc` field is a hexadecimal number that increments with each request using the same nonce.

- First request: `00000001`
- Second request: `00000002`
- Third request: `00000003`

## Authentication Flow

### Initial Request (No Auth)

```http
GET /dav/myrepo/file.txt HTTP/1.1
Host: server:8080
```

**Response (401 Unauthorized):**
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Digest realm="file-hub", nonce="random_nonce", opaque="random_opaque", algorithm="MD5", qop="auth"
```

### Authenticated Request

```http
GET /dav/myrepo/file.txt HTTP/1.1
Host: server:8080
Authorization: Digest username="alice", realm="file-hub", nonce="random_nonce", uri="/dav/myrepo/file.txt", qop="auth", nc="00000001", cnonce="client_nonce", response="calculated_hash", opaque="random_opaque", algorithm="MD5"
```

**Response (200 OK):**
```http
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 1024

<file content>
```

## Security Considerations

### HTTPS vs HTTP

- **Always use HTTPS** in production
- HTTP sends credentials in clear text (except for Digestauth)
- HTTPS encrypts all communication

### Password Storage on Server

- Server stores `HA1 = MD5(username:realm:password)` (for Digest auth)
- Plain-text passwords are never stored
- MD5 is used for compatibility with HTTP Digest auth specification

### Nonce Management

- Nonces prevent replay attacks
- Server should invalidate nonces after use or time-based expiration
- Current implementation accepts any nonce (production should enforce expiration)

### Session Security

- Sessions expire after 24 hours
- Cookies are `HttpOnly` (not accessible to JavaScript)
- Use HTTPS to protect session cookies in transit

### Credential Storage on Client

- Never store passwords in plain text
- Use Android Keystore or iOS Keychain for secure storage
- Consider encrypting credentials using a device-specific key


## Authentication Best Practices

1. **Prefer Sessions** for mobile apps - simpler and more efficient
2. **Use HTTPS** - encrypts all communication
3. **Handle 401 Responses** - gracefully prompt user for credentials
4. **Cache Credentials Securely** - use platform keychain/keystore
5. **Implement Timeout** - re-authenticate when session expires
6. **Retry Logic** - handle transient authentication failures
7. **Logout** - provide user ability to end session securely

## Testing Authentication

### Testing with curl

#### Session-Based

```bash
# Login
curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'

# Use session
curl -b cookies.txt http://localhost:8080/api/sync/info?repo=myrepo&path=/

# Logout
curl -b cookies.txt -X POST http://localhost:8080/api/logout
```

#### Basic Auth

```bash
curl -u username:password http://localhost:8080/dav/myrepo/file.txt
```

#### Digest Auth (requires --digest flag)

```bash
curl --digest -u username:password http://localhost:8080/dav/myrepo/file.txt
```

## Next Steps

- Read [API.md](API.md) for complete API reference
- Read [SYNC.md](SYNC.md) for sync implementation
