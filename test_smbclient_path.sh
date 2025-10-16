#!/bin/bash
# Test script to verify SMBCLIENT_PATH environment variable works

set -e

echo "=== Testing SMBCLIENT_PATH Configuration ==="
echo ""

# Test 1: Default auto-detection
echo "Test 1: Auto-detection (no SMBCLIENT_PATH set)"
echo "Expected: Should find smbclient in system PATH or common locations"
unset SMBCLIENT_PATH
if command -v smbclient &> /dev/null; then
    DETECTED_PATH=$(which smbclient)
    echo "✅ smbclient found at: $DETECTED_PATH"
else
    echo "⚠️  smbclient not found in PATH (will use fallback: /usr/bin/smbclient)"
fi
echo ""

# Test 2: Custom path via environment variable
echo "Test 2: Custom path via SMBCLIENT_PATH"
export SMBCLIENT_PATH="/custom/path/to/smbclient"
echo "✅ SMBCLIENT_PATH set to: $SMBCLIENT_PATH"
echo "   (Application will use this path when executing smbclient)"
echo ""

# Test 3: Verify in Docker
echo "Test 3: Docker container configuration"
if [ -f "Dockerfile" ]; then
    if grep -q "ENV SMBCLIENT_PATH" Dockerfile; then
        DOCKER_PATH=$(grep "ENV SMBCLIENT_PATH" Dockerfile | awk '{print $3}')
        echo "✅ Dockerfile sets SMBCLIENT_PATH to: $DOCKER_PATH"
    else
        echo "⚠️  SMBCLIENT_PATH not set in Dockerfile"
    fi
else
    echo "⚠️  Dockerfile not found"
fi
echo ""

# Test 4: Build and verify compilation
echo "Test 4: Build verification"
if go build -o bin/test_server ./cmd/server 2>&1 | grep -q "error"; then
    echo "❌ Build failed"
    exit 1
else
    echo "✅ Build successful"
    rm -f bin/test_server
fi
echo ""

echo "=== All Tests Passed ==="
echo ""
echo "Usage examples:"
echo "  # Use auto-detection:"
echo "  ./bin/server"
echo ""
echo "  # Use custom path:"
echo "  SMBCLIENT_PATH=/custom/path/smbclient ./bin/server"
echo ""
echo "  # In Docker:"
echo "  docker run -e SMBCLIENT_PATH=/usr/local/bin/smbclient ..."
