# File Hub - Test Plan

## Overview

This document outlines the testing strategy for the File Hub project.

### Testing Objectives
1. **Correctness**: Ensure all features work as specified
2. **Reliability**: Verify system handles errors gracefully
3. **Security**: Validate authentication, authorization, and data protection
4. **Performance**: Confirm acceptable response times and resource usage
5. **Compatibility**: Test across platforms and configurations

## Test Strategy

### Testing Pyramid
- **Unit Tests** (70%): Test individual functions and methods in isolation
- **Integration Tests** (20%): Test interactions between components
- **E2E Tests** (10%): Test complete user workflows

## Test Status

### Completed Tests ✅

| Package | Status | Coverage | Notes |
|---------|--------|----------|-------|
| `pkg/config` | ✅ Complete | 93.9% | Config loading, validation, path resolution |
| `pkg/model` | ✅ Complete | 100% | Model validation, serialization, quota calculations |
| `pkg/db` | ✅ Complete | N/A* | Database integration tests (requires PostgreSQL) |
| `pkg/stor` | ✅ Complete | 15.8% | Content types, path operations, storage interface |
| `pkg/users` | ✅ Complete | 7.4% | HA1 hash calculation, user requests |
| `pkg/sync` | ✅ Complete | 1.2% | Version generation, hash calculation, chunked uploads |
| `pkg/web/session` | ✅ Complete | 83.0% | Session management |

*Database tests skip gracefully when PostgreSQL is unavailable

### Partially Complete ⚠️

| Package | Status | Coverage | Remaining Work |
|---------|--------|----------|----------------|
| `pkg/web/auth` | ⚠️ Partial | 32.7% | More auth scenarios, token tests |
| `pkg/web/dav` | ⚠️ Partial | 12.2% | WebDAV method tests, error handling |

### Not Started ❌

| Package | Status | Priority |
|---------|--------|----------|
| `pkg/web/api` | ❌ Not Started | High |
| `pkg/web/handlers` | ❌ Not Started | Medium |

## Remaining Test Work

### High Priority

1. **Web API Handlers** (`pkg/web/api/`)
   - User CRUD API tests
   - Repository API tests
   - File operation API tests
   - Share API tests

2. **WebDAV Handlers** (`pkg/web/dav/`)
   - PROPFIND, PUT, DELETE, MKCOL tests
   - COPY, MOVE operations
   - Error response tests

3. **Authentication** (`pkg/web/auth/`)
   - Basic and Digest auth scenarios
   - Session security tests
   - Token-based auth

### Medium Priority

1. **Integration Tests**
   - Database transaction tests
   - Storage backend integration
   - WebDAV client integration

2. **E2E Tests**
   - User journey tests
   - File sync workflow
   - Sharing and collaboration

### Low Priority (Post-v1.0)

1. **Performance Tests**
   - Benchmark tests
   - Load tests
   - Stress tests

2. **Security Tests**
   - Authentication security
   - Authorization tests
   - API security

3. **Compatibility Tests**
   - Platform compatibility
   - Browser compatibility
   - WebDAV client compatibility

## Test Environment Setup

### Prerequisites
```bash
# Go testing
go get github.com/stretchr/testify

# PostgreSQL (for integration tests)
# Ubuntu/Debian:
sudo apt-get install postgresql postgresql-contrib
```

### Test Database
```bash
createdb filehub_test
psql -d filehub_test -f scripts/database_schema.sql
```

### Run Tests
```bash
# All tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test ./pkg/db/...
```

## Test Coverage Goals

| Package | Current | Target |
|---------|---------|--------|
| Core packages (config, model, db, stor, users) | Varies | 80%+ |
| Web packages (auth, dav, api, session) | Varies | 70%+ |
| **Overall** | ~45% | **80%+** |

## Test Organization

```
pkg/
├── config/
│   └── config_test.go       # ✅ Complete
├── model/
│   └── model_test.go        # ✅ Complete
├── db/
│   └── database_test.go     # ✅ Complete
├── stor/
│   └── storage_test.go      # ✅ Complete
├── sync/
│   └── service_test.go      # ✅ Complete
├── users/
│   └── users_test.go        # ✅ Complete
└── web/
    ├── auth/
    │   ├── auth_test.go     # Existing
    │   ├── basic_test.go    # Existing
    │   └── digest_test.go   # Existing
    ├── dav/
    │   └── dav_test.go      # Existing
    ├── session/
    │   └── session_test.go  # Existing
    └── api/
        └── api_test.go      # TODO
```

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | 2026-02-26 | Simplified - removed detailed test cases (now in code) |
| 1.0 | 2026-02-26 | Initial comprehensive test plan |
