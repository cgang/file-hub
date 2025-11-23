# File Hub - Personal File Sync Service

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/cgang/file-hub)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

> ⚠️ **WORK IN PROGRESS** - This project is currently under active development and not yet ready for production use.

A personal file backup and synchronization service with WebDAV support, PostgreSQL metadata storage, and efficient binary diff synchronization.

## 📌 Project Goals
- Real-time file synchronization across devices
- Web interface for file management
- Cross-platform compatibility (Windows, macOS, Linux, Android, iOS)
- Simple deployment for personal/family NAS systems
- Client code maintained in a separate repository

## 🔍 Core Features
### 📁 Storage Architecture
- Native filesystem storage with PostgreSQL metadata
- Simple directory structure configuration
- Basic quota management

### 🔄 Sync Mechanism
- Binary diff algorithm for bandwidth optimization
- Conflict resolution strategy
- WebDAV protocol implementation

### 🔐 Security
- End-to-end encryption for data in transit and at rest
- Database-stored authentication credentials

### ⚡ Performance
- Delta encoding transfers
- Caching for frequent files
- Parallel sync operations

##  Quick Start
```bash
# Clone repository
git clone https://github.com/cgang/file-hub.git
cd file-hub

# Setup database
createdb filehub
psql -d filehub -f scripts/database_schema.sql

# Build and run the integrated service (backend + frontend in single binary)
make build
./bin/file-hub
```

## 🌐 Integrated Web UI
The project features a modern web UI built with Svelte that is embedded directly in the binary. The single binary serves both the WebDAV API and the web interface:

- **Web UI**: Accessible at `http://localhost:8080` (or your configured host/port)
- **WebDAV API**: Available at `http://localhost:8080/webdav`

For development:
```bash
# Install web UI dependencies
make web-install

# Run web UI in development mode (separate from backend)
make web-dev

# Build frontend assets into the binary
make build
```

## 💡 Contributing
Interested in contributing? See our [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## 🤖 AI Assistant
This project has received assistance from Qwen AI Assistant during development. The AI has helped with code generation, documentation, refactoring, and bug fixes as part of the development workflow.
