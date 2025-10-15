# Contributing to Document SMB Relay Service

Thank you for your interest in contributing to the Document SMB Relay Service! We welcome contributions from the community and appreciate your help in making this project better.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. We expect all contributors to:

- Be respectful and considerate in communication
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Respect differing viewpoints and experiences

## Getting Started

Before you begin contributing, please:

1. **Read the Documentation**: Familiarize yourself with the [README.md](README.md), [QUICKSTART.md](QUICKSTART.md), and other documentation
2. **Check Existing Issues**: Look through [open issues](https://github.com/bancey/document-smbrelay-service/issues) to see if your concern is already being addressed
3. **Join Discussions**: Participate in issue discussions to understand the project's direction

## Development Setup

### Prerequisites

- **Go 1.21 or higher** (Go 1.22+ recommended)
- **smbclient binary** (from Samba package)
- **Git** for version control
- **Make** (optional, for using Makefile commands)

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/document-smbrelay-service.git
   cd document-smbrelay-service
   ```

3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/bancey/document-smbrelay-service.git
   ```

4. **Install dependencies**:
   ```bash
   go mod download
   ```

5. **Build the application**:
   ```bash
   make build
   # or
   go build -o bin/server ./cmd/server
   ```

6. **Run tests** to ensure everything works:
   ```bash
   make test
   # or
   go test ./...
   ```

### Running the Application Locally

```bash
# Set required environment variables
export SMB_SERVER_NAME=testserver
export SMB_SERVER_IP=127.0.0.1
export SMB_SHARE_NAME=testshare
export SMB_USERNAME=testuser
export SMB_PASSWORD=testpass
export LOG_LEVEL=DEBUG

# Run the application
./bin/server
# or
make run
```

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue using the **Bug Report** template. Include:

- Clear description of the issue
- Steps to reproduce
- Expected vs. actual behavior
- Environment details (Go version, OS, etc.)
- Relevant logs or error messages

### Suggesting Features

For feature requests, use the **Feature Request** template. Include:

- Clear description of the proposed feature
- Use case and motivation
- Potential implementation approach (if applicable)
- Any alternatives you've considered

### Contributing Code

1. **Create a branch** for your work:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes** following our [coding standards](#coding-standards)

3. **Test your changes** thoroughly:
   ```bash
   make test
   go fmt ./...
   go vet ./...
   ```

4. **Commit your changes** with clear, descriptive messages:
   ```bash
   git commit -m "feat: add new feature description"
   # or
   git commit -m "fix: resolve specific bug"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request** against the `main` branch

## Coding Standards

### Go Code Style

- **Format code** with `go fmt` before committing:
  ```bash
  go fmt ./...
  # or
  make fmt
  ```

- **Run static analysis** with `go vet`:
  ```bash
  go vet ./...
  # or
  make vet
  ```

- **Follow Go best practices**:
  - Use meaningful variable and function names
  - Keep functions focused and concise
  - Add comments for exported functions and types
  - Handle errors appropriately
  - Use Go idioms and standard patterns

### Code Organization

- Place business logic in the `internal/` directory
- Keep handlers thin; move complex logic to appropriate packages
- Maintain separation of concerns (config, handlers, SMB operations, etc.)
- Follow the existing project structure

### Error Handling

- Always check and handle errors
- Provide meaningful error messages
- Use structured logging for debugging
- Don't panic in production code

## Testing Guidelines

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
make test-race

# Run tests in verbose mode
make test-verbose
```

### Writing Tests

- **Write unit tests** for new functionality
- **Follow existing test patterns** in the codebase
- **Use table-driven tests** where appropriate
- **Mock external dependencies** (SMB connections, file systems)
- **Test edge cases and error conditions**
- **Aim for high test coverage** for critical paths

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Pull Request Process

### Before Submitting

1. ✅ **Ensure all tests pass**: `make test`
2. ✅ **Format your code**: `make fmt`
3. ✅ **Run static analysis**: `make vet`
4. ✅ **Update documentation** if needed
5. ✅ **Add/update tests** for your changes
6. ✅ **Keep changes focused** and atomic

### PR Description

Use the Pull Request template to provide:

- Clear description of changes
- Related issue numbers (e.g., "Fixes #123")
- Type of change (bug fix, feature, docs, etc.)
- Testing performed
- Breaking changes (if any)
- Screenshots (for UI changes)

### Review Process

1. **Automated checks** will run on your PR
2. **Maintainers will review** your code
3. **Address feedback** promptly and professionally
4. **Update your PR** based on review comments
5. Once approved, a **maintainer will merge** your PR

### Commit Message Guidelines

We follow conventional commit format:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions or modifications
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `chore:` Maintenance tasks
- `ci:` CI/CD changes

Examples:
```
feat: add support for multiple file uploads
fix: resolve connection timeout on large files
docs: update API documentation for /upload endpoint
test: add unit tests for SMB connection handling
```

## Issue Guidelines

### Creating Issues

- **Search existing issues** first to avoid duplicates
- **Use issue templates** when available
- **Provide clear, detailed descriptions**
- **Include relevant context** (logs, environment, etc.)
- **Use appropriate labels** (if you have permission)

### Working on Issues

- **Comment on the issue** before starting work to avoid duplication
- **Ask questions** if requirements are unclear
- **Update the issue** with progress or blockers
- **Link your PR** to the issue

## Documentation

### What to Document

- **New features**: Add usage examples
- **API changes**: Update endpoint documentation
- **Configuration**: Document new environment variables
- **Breaking changes**: Clearly highlight in CHANGELOG

### Documentation Style

- Use clear, concise language
- Include code examples where helpful
- Keep formatting consistent with existing docs
- Test code examples to ensure they work

### Files to Update

- `README.md` - Main project documentation
- `QUICKSTART.md` - Quick start guide
- `README-GO.md` - Go implementation details
- Code comments for exported functions

## Community

### Getting Help

- **Issues**: Open an issue for bugs or questions
- **Discussions**: Use GitHub Discussions for general questions
- **Documentation**: Check existing documentation first

### Stay Updated

- **Watch the repository** for notifications
- **Review the roadmap** in README.md
- **Check recent issues and PRs** to see what's being worked on

## Recognition

We value all contributions, whether it's:

- 🐛 Bug reports and fixes
- ✨ New features
- 📚 Documentation improvements
- 🧪 Test additions
- 💡 Feature suggestions
- 🔍 Code reviews
- 📣 Spreading the word about the project

Thank you for contributing to Document SMB Relay Service! Your efforts help make this project better for everyone.

## License

By contributing to this project, you agree that your contributions will be licensed under the BSD-2-Clause License (same as the project).
