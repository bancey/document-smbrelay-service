#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "  Document SMB Relay Service - Go Tests"
echo "========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go version:${NC} $(go version)"
echo ""

# Set environment variables for tests
export SMB_SERVER_NAME=testserver
export SMB_SERVER_IP=127.0.0.1
export SMB_SHARE_NAME=testshare
export SMB_USERNAME=testuser
export SMB_PASSWORD=testpass
export LOG_LEVEL=ERROR

# Run tests with different modes based on argument
MODE="${1:-all}"

case "$MODE" in
    "unit")
        echo -e "${YELLOW}Running unit tests only...${NC}"
        go test -v -race -short ./...
        ;;
    "coverage")
        echo -e "${YELLOW}Running tests with coverage...${NC}"
        go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
        echo ""
        echo -e "${GREEN}Coverage summary:${NC}"
        go tool cover -func=coverage.out | tail -n 1
        echo ""
        echo "To view detailed coverage report, run: go tool cover -html=coverage.out"
        ;;
    "verbose")
        echo -e "${YELLOW}Running tests in verbose mode...${NC}"
        go test -v -race ./...
        ;;
    "bench")
        echo -e "${YELLOW}Running benchmarks...${NC}"
        go test -bench=. -benchmem ./...
        ;;
    "all")
        echo -e "${YELLOW}Running all tests...${NC}"
        go test -v -race ./...
        ;;
    *)
        echo -e "${RED}Unknown mode: $MODE${NC}"
        echo "Usage: $0 [unit|coverage|verbose|bench|all]"
        exit 1
        ;;
esac

# Check test result
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}  ✓ All tests passed successfully!${NC}"
    echo -e "${GREEN}=========================================${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}=========================================${NC}"
    echo -e "${RED}  ✗ Tests failed${NC}"
    echo -e "${RED}=========================================${NC}"
    exit 1
fi
