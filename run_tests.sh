#!/bin/bash

# Test runner script for document SMB relay service

set -e

export PATH="$HOME/.local/bin:$PATH"

echo "🧪 Document SMB Relay Service Test Suite"
echo "========================================"

# Function to run tests with error handling
run_tests() {
    local test_type="$1"
    local test_args="$2"
    
    echo
    echo "🔧 Running $test_type tests..."
    echo "Command: python -m pytest $test_args"
    
    if python -m pytest $test_args; then
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
        echo "⚠️  Integration tests require Docker to be running"
        echo "Docker will be used to start an SMB server for testing"
        read -p "Continue? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            run_tests "integration" "-m integration"
        else
            echo "Integration tests skipped"
        fi
        ;;
    "all")
        # Run unit tests first
        if run_tests "unit" "-m unit"; then
            echo
            echo "⚠️  Integration tests require Docker to be running"
            echo "Docker will be used to start an SMB server for testing"
            read -p "Run integration tests? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                run_tests "integration" "-m integration"
            else
                echo "Integration tests skipped"
            fi
        fi
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [unit|integration|all|help]"
        echo
        echo "Options:"
        echo "  unit         Run only unit tests (fast, no external dependencies)"
        echo "  integration  Run only integration tests (requires Docker)"
        echo "  all          Run all tests (default)"
        echo "  help         Show this help message"
        echo
        echo "Examples:"
        echo "  $0           # Run all tests"
        echo "  $0 unit      # Run only unit tests"
        echo "  $0 integration  # Run only integration tests"
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