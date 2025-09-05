#!/bin/bash

# Test runner script for document SMB relay service

set -e

export PATH="$HOME/.local/bin:$PATH"

echo "🧪 Document SMB Relay Service Test Suite"
echo "========================================"

# Prefer `python`, fall back to `python3`
if command -v python >/dev/null 2>&1; then
    PYTHON=python
elif command -v python3 >/dev/null 2>&1; then
    PYTHON=python3
else
    echo "❌ Neither 'python' nor 'python3' is available on PATH."
    exit 1
fi

# Function to run tests with error handling
run_tests() {
    local test_type="$1"
    local test_args="$2"
    
    echo
    echo "🔧 Running $test_type tests..."
    echo "Command: $PYTHON -m pytest $test_args"
    
    if $PYTHON -m pytest $test_args; then
        echo "✅ $test_type tests passed!"
        return 0
    else
        echo "❌ $test_type tests failed!"
        return 1
    fi
}

# Check if specific test type was requested
case "${1:-all}" in
    "unit")
        run_tests "unit" "-m unit"
        ;;
    "integration")
        if [[ -n "$CI_SMB_PORT" ]]; then
            echo "🔧 Running integration tests in CI environment (port: $CI_SMB_PORT)"
            run_tests "integration" "-m integration"
        else
            echo "⚠️  Integration tests in local environment require CI setup"
            echo "For local testing, use GitHub Actions or set up SMB server manually"
            echo "To run in CI: set CI_SMB_PORT and CI_SMB_SERVER_IP environment variables"
            echo
            echo "Skipping integration tests in local environment"
            exit 0
        fi
        ;;
    "all")
        # Run unit tests first
        if run_tests "unit" "-m unit"; then
            echo
            if [[ -n "$CI_SMB_PORT" ]]; then
                echo "🔧 Running integration tests in CI environment"
                run_tests "integration" "-m integration"
            else
                echo "⚠️  Integration tests skipped - not in CI environment"
                echo "Integration tests require GitHub Actions services or manual SMB setup"
            fi
        fi
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [unit|integration|all|help]"
        echo
        echo "Options:"
        echo "  unit         Run only unit tests (fast, no external dependencies)"
        echo "  integration  Run only integration tests (requires CI environment)"
        echo "  all          Run all tests (default)"
        echo "  help         Show this help message"
        echo
        echo "Environment:"
        echo "  CI_SMB_PORT       SMB server port in CI environment"
        echo "  CI_SMB_SERVER_IP  SMB server IP in CI environment"
        echo
        echo "Examples:"
        echo "  $0           # Run all tests"
        echo "  $0 unit      # Run only unit tests"
        echo "  $0 integration  # Run integration tests (CI only)"
        echo
        echo "Note: Integration tests are designed for CI environments with"
        echo "      GitHub Actions services and will skip in local development."
        exit 0
        ;;
    *)
        echo "❌ Unknown option: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac

echo
echo "🎉 Test run completed!"