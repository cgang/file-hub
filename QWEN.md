# Qwen AI Assistant Context - File Hub Project

## Project Overview
**Name:** File Hub - Personal File Sync Service
**Description:** A personal file backup and synchronization service with WebDAV support, PostgreSQL metadata storage, and efficient binary diff synchronization.
**Purpose:** Real-time file synchronization across devices with web interface for file management, designed for personal/family NAS systems.

## 📌 Project Goals
- Real-time file synchronization across devices
- Web interface for file management
- Cross-platform compatibility (Windows, macOS, Linux, Android, iOS)
- Simple deployment for personal/family NAS systems
- Client code maintained in a separate repository

## 🔍 Core Features
- **Storage Architecture:** Native filesystem storage with PostgreSQL metadata
- **Sync Mechanism:** Binary diff algorithm for bandwidth optimization with conflict resolution
- **Security:** End-to-end encryption for data in transit and at rest
- **Performance:** Delta encoding transfers, caching, and parallel sync operations

## 🏗️ Project Structure
```
├── cmd/                # Main application entry points
├── internal/             # Private application/business logic
├── pkg/                  # Library code
├── config/               # Configuration files
├── web/                  # WebDAV interface and templates
├── scripts/              # Development/deployment scripts
├── test/                 # Test files
└── docs/                 # Documentation
```

## 🛠️ Development Guidelines

### Code Style & Standards
- Follow idiomatic Go conventions (gofmt, godoc)
- Use Go modules for dependency management
- Prefer clear, readable code over clever optimizations
- Follow project-specific coding standards from CONTRIBUTING.md
- Follow [Effective Go](https://golang.org/doc/effective_go) conventions

### Contribution Process
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit changes in imperative mood (e.g., "Fix bug" not "Fixed bug")
4. Push to branch (`git push origin feature/my-feature`)
5. Create a pull request

### Technical Requirements
- All new features must include unit tests
- Run tests with `make test` before submitting PRs
- Maintain 80%+ test coverage
- Use Go modules for dependencies
- Install golint for code quality checks (`make golint`)

### Error Handling
- Emphasize proper error wrapping and handling
- Suggest appropriate logging with structured logging
- Identify edge cases in network operations
- Review deferred cleanup patterns

### Testing & Quality
- Focus on core functionality: file synchronization and storage
- Use standard Go libraries where possible
- Optimize for readability and maintainability
- Prioritize Linux-specific optimizations and features

### Code Review Focus Areas
- Proper error handling (multi-value returns, wrapping)
- Goroutine leak risks and channel misuse
- Concurrency patterns improvements
- Proper file handling with os/io packages
- API documentation with Godoc
- Synchronization algorithms clearly explained
- Performance considerations documented
- Configuration documentation kept current

## 🚀 Quick Start Commands
```bash
# Setup database
createdb filehub
make migrate

# Run service
make run
```

## 🔧 Common Commands
- `make test` - Run tests
- `make migrate` - Run database migrations
- `make run` - Start the service

## 🤝 Community Standards
This project follows Go's community conduct standards, including:
- Go Code of Conduct
- Contributor Covenant

## 📚 Additional Resources
- [CONTRIBUTING.md](CONTRIBUTING.md) - Detailed contribution guidelines
- [Go Proverbs](https://go-proverbs.github.io/) - Go community conventions to follow