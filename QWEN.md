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

### Testing Approach
- Always generate corresponding unit tests for new functionality
- Follow table-driven test patterns for multiple test cases
- Use testify/assert for clear assertion messages
- Aim for >80% test coverage for new features
- Include both positive and negative test cases

### Error Handling Patterns
- Wrap errors with context using fmt.Errorf("message: %w", err)
- Return errors early to minimize nesting
- Use multi-value returns consistently (value, error)
- Log errors appropriately without exposing sensitive information

## AI-Assisted Development Workflow

### Code Review Focus Areas
When reviewing AI-generated code, pay special attention to:
- Proper resource cleanup (defer statements for file handles, database connections)
- Correct concurrency patterns (goroutines, channels, mutexes)
- Security considerations (input validation, authentication, authorization)
- Performance implications (memory usage, algorithmic complexity)
- Integration with existing codebase components

### Refactoring Guidance
- Preserve existing APIs and interfaces when possible
- Maintain backward compatibility during refactoring
- Update dependent code when changing exported functions/types
- Ensure database schema changes are properly migrated

### Best Practices for AI Collaboration
1. Understand the existing codebase before suggesting changes
2. Follow established patterns and conventions in the project
3. Ask clarifying questions when requirements are ambiguous
4. Provide multiple solution options when tradeoffs exist
5. Explain complex technical decisions clearly
6. Flag potential security or performance concerns

## Project-Specific AI Knowledge Base

### Architecture Overview
- Backend: Go with embedded Svelte frontend
- Database: PostgreSQL for metadata storage
- Storage: Native filesystem with WebDAV interface
- Sync: Binary diff algorithm for efficient transfers

### Key Components
- WebDAV server implementation
- File synchronization engine
- PostgreSQL metadata management
- Svelte web interface with Vite build system

### Important Implementation Details
- Frontend assets are embedded in the binary
- Configuration uses YAML files with environment variable override
- Authentication uses HTTP Basic Auth with database storage
- Makefile orchestrates both frontend and backend builds
- All frontend code must be kept under the `/web/` directory and based on Svelte

## AI Context Reminders

### Critical Success Factors
- Maintain simplicity for personal/family NAS deployment
- Ensure cross-platform compatibility
- Preserve data integrity during sync operations
- Keep dependencies minimal for easy installation

### Common Pitfalls to Avoid
- Over-engineering solutions beyond personal use case
- Introducing heavy external dependencies
- Compromising security for convenience
- Ignoring error conditions in file operations

### Optimization Priorities
1. Bandwidth efficiency through delta encoding
2. Storage efficiency with native filesystem
3. Memory efficiency for large file operations
4. CPU efficiency in sync algorithms

This document serves as context for AI assistants working on the File Hub project, focusing on AI-specific collaboration guidelines while avoiding duplication of project documentation found in README.md, docs/CONTRIBUTING.md, and other documentation files.