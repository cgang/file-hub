# Qwen AI Assistant Context - File Hub Project

## AI-Specific Development Guidelines

### Code Generation Principles
- Generate idiomatic Go code following Effective Go conventions
- Prioritize readability and maintainability over clever optimizations
- Use clear, descriptive variable and function names
- Follow the project's directory structure and organization patterns
- Ensure all generated code includes appropriate error handling

### Documentation Practices
- Focus documentation on explaining "why" rather than "what"
- Include inline comments for complex logic or non-obvious implementations
- Update godoc comments for all public functions and types
- Keep technical documentation concise but comprehensive

### Error Handling Patterns
- Wrap errors with context using fmt.Errorf("message: %w", err)
- Return errors early to minimize nesting
- Log errors appropriately without exposing sensitive information

### Refactoring Guidance
- Preserve existing APIs and interfaces when possible
- Maintain backward compatibility during refactoring
- Update dependent code when changing exported functions/types
- Ensure database schema changes are properly migrated

## Project-Specific AI Knowledge Base

### Architecture Overview
- Backend: Go with embedded Svelte frontend
- Database: PostgreSQL for metadata storage
- Storage: Native filesystem with WebDAV interface
- Sync: Binary diff algorithm for efficient transfers

### Key Components
- WebDAV server implementation
- PostgreSQL metadata management
- Svelte web interface with Vite build system
- File synchronization engine

### Important Implementation Details
- Frontend assets are embedded in the binary
- Configuration uses YAML files with environment variable override
- Authentication uses HTTP Basic Auth with database storage
- Makefile orchestrates both frontend and backend builds
- All frontend code must be kept under the `/web/` directory and based on Svelte

### Critical Success Factors
- Maintain simplicity for personal/family NAS deployment
- Preserve data integrity during sync operations
- Keep dependencies minimal for easy installation

This document serves as context for AI assistants working on the File Hub project, focusing on AI-specific collaboration guidelines while avoiding duplication of project documentation found in README.md, docs/CONTRIBUTING.md, and other documentation files.

