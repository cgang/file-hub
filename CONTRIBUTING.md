# Contributing to File Hub - Personal File Sync Service

## Getting Started
Before contributing, please read the [Go Proverbs](https://go-proverbs.github.io/) and familiarize yourself with Go community conventions. We follow Go's project layout standards and engineering philosophy for overall architecture.

## Project Structure
This project follows the [Go Standards Project Layout](https://github.com/golang-standards/project-layout) with these key directories:
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

## Contribution Guidelines
### Code Style
1. **Go Code**: Follow [Effective Go](https://golang.org/doc/effective_go) conventions
2. **Git Commits**: Use imperative mood (e.g., "Fix bug" not "Fixed bug")
3. **Documentation**: Keep godoc up-to-date

### Development Process
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit changes (`git commit -am 'Add feature'`)
4. Push to branch (`git push origin feature/my-feature`)
5. Create a pull request

## Technical Requirements
### Testing
- All new features must include unit tests
- Run tests with `make test` before submitting PRs
- Maintain 80%+ test coverage

### Dependency Management
- Use Go modules for dependencies
- Install golint for code quality checks (`make golint`)
- Document any system-level dependencies

## Review Process
1. Ensure PRs reference an issue
2. Maintainers will assign reviewers
3. Address feedback iteratively
4. Merge after approval and passing CI

## Community Standards
This project follows Go's community conduct standards. Please be respectful, constructive, and adhere to:
- [Go Code of Conduct](https://golang.org/conduct)
- [Contributor Covenant](https://www.contributor-covenant.org/version/2/0/code_of_conduct/)
