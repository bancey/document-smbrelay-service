# Contributing to Document SMB Relay Service

Welcome! We appreciate your interest in contributing to the Document SMB Relay Service. This guide will help you get started with contributing to this project.

## Overview

This is a minimal FastAPI service that accepts file uploads via HTTP and writes them directly to SMB shares without mounting them on the host filesystem. The service uses SMB protocols for connectivity and requires specific environment variables for configuration.

## Getting Started

### Prerequisites

- Python 3.10+ (Python 3.12+ recommended)
- Docker (for integration tests)
- Git

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/bancey/document-smbrelay-service.git
   cd document-smbrelay-service
   ```

2. **Install dependencies**
   ```bash
   pip install -r requirements.txt
   pip install -r requirements-test.txt
   ```

3. **Validate setup**
   ```bash
   python3 -m py_compile app/main.py
   python3 -c "import app.main; print('Import successful')"
   ```

## Development Workflow

### Before Making Changes

1. **Run existing tests** to ensure everything is working:
   ```bash
   ./run_tests.sh unit
   ```

2. **Start the application** to test functionality:
   ```bash
   SMB_SERVER_NAME=testserver \
   SMB_SERVER_IP=127.0.0.1 \
   SMB_SHARE_NAME=testshare \
   SMB_USERNAME=testuser \
   SMB_PASSWORD=testpass \
   uvicorn app.main:app --host 0.0.0.0 --port 8080
   ```

### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following these guidelines:
   - Keep changes minimal and focused
   - Follow existing code style and patterns
   - Update documentation if needed
   - Add tests for new functionality

3. **Validate your changes** frequently:
   ```bash
   # Syntax validation
   python3 -m py_compile app/main.py
   python3 -c "import app.main; print('Import successful')"
   
   # Run tests
   ./run_tests.sh unit
   
   # Test manually with the server running
   curl -s http://localhost:8080/docs | grep -q "Document SMB Relay Service"
   curl -s http://localhost:8080/openapi.json | python3 -m json.tool > /dev/null
   ```

## Testing

### Test Structure

- `tests/unit/` — Unit tests with mocked dependencies (fast)
- `tests/integration/` — End-to-end tests with real SMB server (requires Docker)

### Running Tests

```bash
# Run all tests
./run_tests.sh

# Run only unit tests (recommended for development)
./run_tests.sh unit

# Run only integration tests
./run_tests.sh integration

# Run specific test file
python -m pytest tests/unit/test_endpoints.py -v
```

### Writing Tests

- **Unit tests**: Mock external dependencies, focus on logic
- **Integration tests**: Only when real SMB behavior is required
- Use markers: `@pytest.mark.unit` or `@pytest.mark.integration`
- Keep tests isolated and clean up resources

Example unit test:
```python
@pytest.mark.unit
def test_your_feature(mock_dependency):
    # Arrange
    mock_dependency.return_value = expected_result
    
    # Act
    result = your_function()
    
    # Assert
    assert result == expected_result
```

## Code Style Guidelines

### Python Code
- Follow existing patterns in the codebase
- Use clear, descriptive variable names
- Keep functions focused and small
- Add docstrings for complex functions
- Handle errors appropriately

### Environment Variables
- All SMB configuration via environment variables
- Provide sensible defaults where possible
- Validate required variables with clear error messages

### Error Handling
- Return appropriate HTTP status codes
- Provide clear error messages
- Handle connection failures gracefully
- Clean up temporary files

## Documentation

### When to Update Documentation
- Adding new environment variables
- Changing API endpoints or behavior
- Adding new features or configuration options
- Fixing significant bugs

### Documentation Files
- `README.md` — Main usage documentation
- `tests/README.md` — Testing documentation
- This `CONTRIBUTING.md` — Development guidelines

## Submission Guidelines

### Pull Request Process

1. **Ensure tests pass**
   ```bash
   ./run_tests.sh unit
   ```

2. **Verify manual functionality**
   - Start server and test endpoints
   - Check error handling with missing environment variables
   - Test with unreachable SMB server (expected to fail gracefully)

3. **Create clear commit messages**
   ```
   feat: add new upload validation feature
   fix: handle connection timeout errors properly
   docs: update environment variable documentation
   test: add unit tests for file validation
   ```

4. **Fill out PR template** (use the provided template)

5. **Request review** from maintainers

### PR Requirements

- [ ] All unit tests pass
- [ ] Code follows existing style and patterns
- [ ] Documentation updated if needed
- [ ] Manual testing completed
- [ ] Clear description of changes
- [ ] No unrelated changes included

## CI/CD Pipeline

The project uses GitHub Actions for:
- **Automated testing** with unit and integration tests
- **Code quality** checking with SonarCloud
- **Docker image building** and publishing to GitHub Container Registry
- **Multi-architecture support** (linux/amd64, linux/arm64)

All PRs will automatically trigger:
1. Dependency installation
2. Code syntax validation
3. Complete test suite execution
4. Code quality analysis

## Common Issues and Solutions

### Import Errors
```bash
# Fix: Install dependencies
pip install -r requirements.txt
```

### Test Failures
```bash
# Fix: Install test dependencies
pip install -r requirements-test.txt
```

### Integration Test Issues
- Ensure Docker is running
- Check available ports (445, 137-139)
- Use unit tests for faster iteration

### Connection Refused Errors
- Expected when using test SMB values (127.0.0.1)
- Indicates service is working correctly
- Use real SMB server for actual testing

## Project Structure

```
.
├── app/
│   ├── main.py              # Main FastAPI application
│   ├── config/
│   │   └── smb_config.py    # SMB configuration management
│   ├── smb/
│   │   ├── connection.py    # SMB connection and health checks
│   │   └── operations.py    # SMB file operations
│   └── utils/
│       └── file_utils.py    # File utility functions
├── tests/
│   ├── unit/                # Unit tests
│   └── integration/         # Integration tests
├── requirements.txt         # Runtime dependencies
├── requirements-test.txt    # Test dependencies
├── Dockerfile              # Container configuration
├── run_tests.sh            # Test runner script
└── .github/
    └── workflows/           # CI/CD configuration
```

## Getting Help

- **Issues**: Open an issue using the provided templates
- **Questions**: Use the discussion feature or create an issue
- **Security**: Report security issues privately to the maintainers

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project.

Thank you for contributing! 🎉