#!/bin/bash

# Go application build and test script

set -e

echo "🔧 Document SMB Relay Service (Go) - Build & Test"
echo "=================================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

echo "✅ Go version: $(go version)"

# Function to run build
build_app() {
    echo ""
    echo "🔨 Building Go application..."
    
    # Download dependencies
    echo "📦 Downloading dependencies..."
    go mod download
    go mod tidy
    
    # Build the application
    echo "🔨 Compiling..."
    go build -o bin/server ./cmd/server
    
    if [ $? -eq 0 ]; then
        echo "✅ Build successful! Binary created at: bin/server"
        return 0
    else
        echo "❌ Build failed!"
        return 1
    fi
}

# Function to run tests
run_tests() {
    echo ""
    echo "🧪 Running tests..."
    
    if go test -v ./...; then
        echo "✅ All tests passed!"
        return 0
    else
        echo "❌ Tests failed!"
        return 1
    fi
}

# Function to format code
format_code() {
    echo ""
    echo "🎨 Formatting code..."
    go fmt ./...
    echo "✅ Code formatted!"
}

# Function to vet code
vet_code() {
    echo ""
    echo "🔍 Running go vet..."
    if go vet ./...; then
        echo "✅ Code passed vet checks!"
        return 0
    else
        echo "❌ Vet checks failed!"
        return 1
    fi
}

# Main script
case "${1:-build}" in
    "build")
        build_app
        ;;
    "test")
        run_tests
        ;;
    "fmt")
        format_code
        ;;
    "vet")
        vet_code
        ;;
    "all")
        format_code
        vet_code
        build_app
        run_tests
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [build|test|fmt|vet|all|help]"
        echo ""
        echo "Options:"
        echo "  build    Build the Go application (default)"
        echo "  test     Run tests"
        echo "  fmt      Format code"
        echo "  vet      Run go vet"
        echo "  all      Run all checks and build"
        echo "  help     Show this help message"
        exit 0
        ;;
    *)
        echo "❌ Unknown option: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac

echo ""
echo "🎉 Done!"
