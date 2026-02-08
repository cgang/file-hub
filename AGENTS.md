# AI Agents Context - File Hub Project

## Table of Contents
- [Project Overview](#project-overview)
- [Project Structure and Architecture](#project-structure-and-architecture)
  - [Directory Layout](#directory-layout)
  - [Architecture Overview](#architecture-overview)
- [Technology Stack](#technology-stack)
  - [Go Backend](#go-backend)
  - [Svelte Frontend](#svelte-frontend)
  - [Database and ORM](#database-and-orm)
- [Coding Standards and Conventions](#coding-standards-and-conventions)
  - [General Development Conventions](#general-development-conventions)
  - [Go Patterns](#go-patterns)
  - [Svelte Patterns](#svelte-patterns)
- [Development Workflow](#development-workflow)
  - [Build Process](#build-process)
  - [Testing](#testing)
  - [Configuration Management](#configuration-management)
- [AI Agent Guidance](#ai-agent-guidance)
  - [Working with Go Code](#working-with-go-code)
  - [Working with Svelte Components](#working-with-svelte-components)
  - [Database Operations](#database-operations)
- [Project Goals and Strategic Principles](#project-goals-and-strategic-principles)

## Project Overview

File Hub is a personal file backup and synchronization service with WebDAV support, PostgreSQL metadata storage, and efficient binary diff synchronization. This document provides AI agents with comprehensive context about the project architecture, coding patterns, and development workflows.

## Project Structure and Architecture

### Directory Layout
```
├── cmd/                    # Main application entry point
│   └── main.go            # Application initialization and graceful shutdown
├── pkg/                   # Go source code organized by functionality
│   ├── config/           # Configuration loading and management
│   ├── db/              # Database operations using PostgreSQL with Bun ORM
│   ├── model/           # Data models (users, repositories, files, shares)
│   ├── stor/            # Storage abstraction (filesystem and AWS S3)
│   ├── users/           # User management and authentication
│   └── web/             # Web server, authentication, and WebDAV implementation
├── web/                  # Svelte frontend web interface
│   ├── src/
│   │   ├── components/  # Svelte components
│   │   ├── styles/      # CSS stylesheets
│   │   └── utils/       # Utility functions
│   ├── index.html       # HTML template
│   └── vite.config.js   # Build configuration
├── config/              # Configuration files
├── scripts/             # Database schema and initialization scripts
├── example/             # Example configurations and test files
├── docs/                # Documentation
└── local/               # Local development files (gitignored)
```

### Architecture Overview
The system follows a clean separation of concerns:
1. **Backend**: Go application with embedded Svelte frontend
2. **Database**: PostgreSQL for metadata storage using Bun ORM
3. **Storage**: Dual backend support (local filesystem and AWS S3)
4. **Protocol**: Full WebDAV implementation with PROPFIND, PUT, DELETE, MKCOL, COPY, MOVE
5. **Sync Protocol**: Protocol Buffer-based sync protocol for mobile-optimized synchronization (planned)
6. **Authentication**: HTTP Basic and Digest authentication with session management
7. **Frontend**: Modern Svelte-based web interface served from the same binary

## Technology Stack

### Go Backend
The backend is built with Go and follows these patterns:

#### Error Handling
- **Consistent error wrapping**: Use `fmt.Errorf("context: %w", err)` to preserve error context
- **Early returns**: Return errors early to minimize nesting and improve readability
- **Context-aware errors**: Propagate errors with context for better debugging
- **Critical errors**: Use `log.Fatalf()` for unrecoverable startup errors

#### Logging Patterns
- **Standard library**: Use Go's standard `log` package consistently
- **Contextual messages**: Include meaningful context in log messages
- **Error details**: Log full error details when handling failures

#### Package Organization
- **Separation of concerns**: Each package has a clear, focused responsibility
- **Service layer**: Business logic separated from data access (`users/`, `stor/`)
- **Interface-based design**: Storage abstraction through `Storage` interface
- **Clear naming**: Packages named after their functionality (db, web, stor, users)

#### Context Usage
- **Propagation**: Pass `context.Context` through function signatures for cancellation
- **Database operations**: Always use context with database queries
- **Graceful shutdown**: Implement proper context cancellation for clean shutdowns

#### Code Structure
- **Initialization**: Main entry point in `cmd/main.go` with graceful shutdown
- **Configuration**: Centralized config loading from multiple paths
- **Dependency injection**: Services initialized with required dependencies
- **Clean architecture**: Clear separation between models, services, and handlers

#### Testing Patterns
- **Test structure**: Use `testing` package with `testify/assert` for assertions
- **Comprehensive coverage**: Test core functionality including authentication
- **Test data**: Use YAML test data for configuration testing
- **File naming**: Test files follow `*_test.go` convention

### Svelte Frontend Architecture
The frontend is built with Svelte and follows these patterns:

#### Component Structure
- **Component-based**: Reusable Svelte components in `web/src/components/`
- **State management**: Local component state using Svelte's reactivity system
- **Event handling**: Component communication through Svelte's event dispatcher
- **Props passing**: Data passed through `export let` props
- **Lifecycle**: Use `onMount` for initialization and side effects

#### Key Components
1. **App.svelte** - Main application router (renders SetupPage, Login, or FileBrowser)
2. **FileBrowser.svelte** - Primary file management with directory listing and uploads
3. **NavigationBar.svelte** - Breadcrumb navigation with event dispatching
4. **FileCard.svelte** - Individual file/directory card for grid view
5. **UploadComponent.svelte** - File upload interface
6. **Login.svelte** - Authentication form with validation
7. **SetupPage.svelte** - Initial admin account setup

#### State Management
- **Local reactivity**: Use Svelte's built-in reactivity (`let variable = value`)
- **Event-driven**: Components communicate through dispatched events
- **No external libraries**: No Redux or Vuex; uses Svelte's native patterns
- **Data flow**: Parent to child through props, child to parent through events

#### Styling Approach
- **Custom CSS**: No CSS frameworks; uses pure CSS with variables
- **CSS variables**: Custom properties for theming in `main.css`
- **Responsive design**: Media queries for mobile/desktop adaptation
- **Component-scoped**: Styles scoped to components with global overrides

#### Build Configuration
- **Vite build**: Uses `@sveltejs/vite-plugin-svelte`
- **Development proxy**: API calls proxied during development
- **Production target**: Built to `web/dist/` directory
- **Base URL**: Served from `/ui/` path when embedded in backend

### Database and ORM Patterns
The database layer uses PostgreSQL with Bun ORM:

#### Core Tables
```sql
-- users: User accounts and authentication
-- repositories: File repositories owned by users
-- files: Metadata for files and directories in repositories
-- shares: Shared access to repository paths for specific users
-- user_quota: Storage quota management for users
```

#### ORM Usage
- **Bun ORM**: Uses `uptrace/bun` with PostgreSQL dialect
- **Model definitions**: Structs with Bun and JSON tags in `pkg/model/`
- **Database operations**: Context-aware queries in `pkg/db/`
- **Error handling**: Proper error wrapping for database failures

#### Schema Management
- **Initialization**: Schema defined in `scripts/database_schema.sql`
- **Setup script**: `scripts/init-database.sh` for easy database creation
- **Migrations**: Manual migration approach for schema changes

## Coding Standards and Conventions

### General Development Conventions
These conventions apply across all technologies and components in the project:

#### 1. Generated Files Should Not Be Committed
- Only source files that humans maintain should be in version control
- Generated files (like compiled code, auto-generated files) should be excluded from repositories
- Use `.gitignore` to exclude generated files and directories
- This prevents merge conflicts, reduces repository size, and ensures generated code stays in sync with source definitions

#### 2. Comprehensive Documentation
- Maintain clear documentation for all major components
- Include README files explaining purpose, files, features, and build processes
- Document the "why" behind design decisions, not just the "what"
- Good documentation improves maintainability and helps other developers understand the codebase

#### 3. Consistent Project Structure
- Organize code into logical, focused modules or directories
- Each component should have a clear, single purpose
- Follow consistent naming conventions and directory layouts
- This leads to better organization, easier maintenance, and clearer separation of concerns

#### 4. Robust Error Handling
- Plan for failure and edge cases during design phase
- Implement proper error propagation and handling strategies
- Consider real-world constraints and failure scenarios
- This creates robust systems that handle unexpected situations gracefully

#### 5. Communication Guidelines
- Avoid unnecessary acknowledgments like "You're absolutely right" or "I agree" in conversations
- Focus on providing valuable information and taking action
- Keep responses concise and to the point
- Only acknowledge when it adds value to the conversation

#### 6. Git Commit Messages
- Keep commit messages simple and short
- Focus on the key change made in the commit
- Use imperative mood (e.g., "Add feature" not "Added feature" or "Adds feature")
- Limit subject line to 50 characters when possible
- Add a blank line followed by a detailed description only when necessary
- Add `Co-authored-by:` lines in the commit message body to acknowledge all AI agents and assistants involved.

#### 7. Compatibility Considerations Pre-v1.0.0
- Ignore backward compatibility concerns before formal v1.0.0 release
- This project is under active development with no real users yet
- Breaking changes are acceptable during pre-release development
- Focus on improving functionality and architecture rather than maintaining compatibility
- Major API, database schema, and configuration changes can be made freely until v1.0.0

### Language-Specific Patterns

#### Go Patterns
These are specific to Go code:

##### Error Handling Examples
```go
func LoadConfig(name string) (*Config, error) {
    searchPaths, err := getConfigDirs()
    if err != nil {
        return nil, fmt.Errorf("failed to get config directories: %w", err)
    }
    // ... implementation
}
```

##### Context Usage Examples
```go
func GetUser(ctx context.Context, username string) (*User, error) {
    var user User
    err := db.NewSelect().Model(&user).Where("username = ?", username).Scan(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %s: %w", username, err)
    }
    return &user, nil
}
```

#### Svelte Patterns
These are specific to Svelte code:

##### Component Props and Events
```svelte
<!-- FileCard.svelte -->
<script>
  import { createEventDispatcher } from 'svelte';
  export let file; // Props
  const dispatch = createEventDispatcher();
  
  function handleClick() {
    dispatch('select', { file }); // Events
  }
</script>
```

##### State Management
```svelte
<script>
  let currentPath = '/'; // Reactive state
  let files = []; // Reactive array
  
  function updateFiles(newFiles) {
    files = newFiles; // Triggers reactivity
  }
</script>
```

## Development Workflow

### Build Process

#### Build System
- **Makefile**: Orchestrates frontend and backend builds
- **Frontend build**: `cd web && npm install && npm run build`
- **Backend build**: `go build -o bin/file-hub cmd/main.go`
- **Embedded assets**: Frontend assets embedded in Go binary

#### Development Workflow
1. **Frontend development**: Run `make web-dev` for hot-reload development
2. **Backend development**: Use `go run cmd/main.go` with local config
3. **Testing**: Run `go test ./...` for unit tests
4. **Linting**: Use `golangci-lint` via `make golint`

### Testing

#### Testing Patterns
- **Configuration testing**:
```go
func TestConfigWithS3(t *testing.T) {
    yamlData := `
web:
  port: 8080
database:
  uri: "postgresql://filehub:filehub@localhost:5432/filehub"
s3:
  endpoint: "https://s3.amazonaws.com"
  region: "us-east-1"
  access_key_id: "test-key"
  secret_access_key: "test-secret"
`
    var cfg Config
    err := yaml.Unmarshal([]byte(yamlData), &cfg)
    assert.NoError(t, err)
    assert.NotNil(t, cfg.S3)
    assert.Equal(t, "https://s3.amazonaws.com", cfg.S3.Endpoint)
    assert.Equal(t, "us-east-1", cfg.S3.Region)
    assert.Equal(t, "test-key", cfg.S3.AccessKeyID)
    assert.Equal(t, "test-secret", cfg.S3.SecretAccessKey)
}
```

### Configuration Management
- **YAML config**: Configuration files in YAML format
- **Multiple paths**: CONFIG_PATH environment variable for config search
- **Environment override**: Configuration can be overridden by environment variables
- **Example config**: `example/config.yaml` shows expected structure

### Database Operations
- **Model definition**:
```go
type User struct {
    bun.BaseModel `bun:"table:users"`
    ID       int    `bun:"id,pk,autoincrement"`
    Username string `bun:"username,unique,notnull"`
    Password string `bun:"password,notnull"`
}
```

## AI Agent Guidance

### Working with Go Code
Follow the established error handling and context usage patterns as shown in the examples above.

### Working with Svelte Components
Use the component structure and state management patterns demonstrated in the examples.

### Database Operations
Apply the ORM usage patterns and error handling strategies outlined in the database section.

### Sync Protocol Implementation
When implementing the Protocol Buffer-based sync protocol:
- Place sync protocol related code in a new `pkg/sync/` directory
- Define Protocol Buffer files in `api/proto/` directory and generate Go code
- Implement the service interface defined in the protocol specification
- Follow the same error handling and context patterns as the rest of the codebase
- Ensure authentication tokens work consistently with the existing WebDAV implementation

## Project Goals and Strategic Principles

### Architecture Goals
- **Simplicity**: Easy deployment for personal/family NAS systems
- **Data integrity**: Preserve data integrity during sync operations
- **Minimal dependencies**: Keep dependencies minimal for easy installation
- **Cross-platform**: Support Windows, macOS, Linux, Android, iOS

### Strategic Development Principles
- **Idiomatic Go**: Follow Effective Go conventions and best practices
- **Clean separation**: Clear boundaries between backend and frontend
- **Embedded assets**: Single binary with embedded frontend
- **Configurable storage**: Support for local filesystem and cloud storage

This document serves as context for AI agents working on the File Hub project, providing comprehensive guidance on architecture, patterns, and development workflows while avoiding duplication of project documentation found in README.md, docs/CONTRIBUTING.md, and other documentation files.

