# Docker Testing Guide

This guide explains how to test the Document SMB Relay Service locally using Docker Compose with a containerized SMB server.

**📁 Related Guides:**
- **[DFS_TESTING.md](./DFS_TESTING.md)** - Testing with simulated DFS environment (multiple file servers)
- **[DFS_KERBEROS.md](./DFS_KERBEROS.md)** - Production Windows DFS and Kerberos setup

## Quick Start

### Basic Testing Setup

Start both the SMB server and relay service:

```bash
docker-compose up -d
```

This will start:
- **Samba server** on port 445 (SMB) and 139 (NetBIOS)
- **SMB Relay Service** on port 8080

### Test the Setup

1. **Check service health:**
   ```bash
   curl http://localhost:8080/health | jq
   ```

2. **Upload a test file:**
   ```bash
   echo "Hello from Docker!" > test.txt
   curl -X POST http://localhost:8080/upload \
     -F file=@test.txt \
     -F remote_path=uploads/test.txt
   ```

3. **Verify the file was uploaded:**
   ```bash
   # Access the SMB share directly to verify
   docker exec smb-server ls -la /share/uploads/
   ```

### Stop the Services

```bash
docker-compose down
```

To also remove the volumes (delete uploaded files):
```bash
docker-compose down -v
```

## DFS Testing

For testing with a simulated DFS environment (DFS namespace + multiple file servers), see **[DFS_TESTING.md](./DFS_TESTING.md)**.

Quick DFS test:
```bash
# Start DFS environment
docker-compose -f docker-compose.dfs.yml up -d

# Test upload
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=test.txt

# Stop
docker-compose -f docker-compose.dfs.yml down
```

## Development Setup with Hot-Reload

For active development, use the development configuration:

```bash
docker-compose -f docker-compose.dev.yml up -d
```

This setup:
- Mounts your local `app/` directory for hot-reloading
- Sets `LOG_LEVEL=DEBUG` for detailed logging
- Enables uvicorn's `--reload` flag
- Creates multiple shares for testing

### Watch Logs

```bash
# Watch all logs
docker-compose -f docker-compose.dev.yml logs -f

# Watch only relay service logs
docker-compose -f docker-compose.dev.yml logs -f smb-relay

# Watch only Samba server logs
docker-compose -f docker-compose.dev.yml logs -f samba
```

## Configuration Options

### Environment Variables

You can override any environment variable in the `docker-compose.yml` file or via command line:

```bash
# Use different auth protocol
docker-compose up -d \
  -e SMB_AUTH_PROTOCOL=negotiate

# Use different share
docker-compose up -d \
  -e SMB_SHARE_NAME=documents
```

### Available Shares (dev configuration)

The development setup creates two shares:
- `testshare`: General testing share
- `documents`: Documents-specific share

Switch between them by changing the `SMB_SHARE_NAME` environment variable.

## Testing Different Authentication Methods

### NTLM (Default)

Already configured in `docker-compose.yml`:
```yaml
environment:
  - SMB_AUTH_PROTOCOL=ntlm
  - SMB_USERNAME=testuser
  - SMB_PASSWORD=testpass
```

### Negotiate

Edit `docker-compose.yml` or create an override:
```yaml
environment:
  - SMB_AUTH_PROTOCOL=negotiate
  - SMB_USERNAME=testuser
  - SMB_PASSWORD=testpass
```

### Testing Without Credentials

To test error handling when credentials are missing:

```bash
# Create a docker-compose.override.yml
cat > docker-compose.override.yml << 'EOF'
version: '3.8'
services:
  smb-relay:
    environment:
      - SMB_USERNAME=
      - SMB_PASSWORD=
EOF

docker-compose up -d
```

## Advanced Testing Scenarios

### Test with Multiple File Uploads

```bash
# Upload multiple files
for i in {1..5}; do
  echo "Test file $i" > test$i.txt
  curl -X POST http://localhost:8080/upload \
    -F file=@test$i.txt \
    -F remote_path=batch/file$i.txt
done

# Verify all files
docker exec smb-server ls -la /share/batch/
```

### Test Overwrite Protection

```bash
# First upload
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=test.txt

# Second upload (should fail)
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=test.txt

# Third upload with overwrite (should succeed)
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=test.txt \
  -F overwrite=true
```

### Test Nested Directories

```bash
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=dept/finance/reports/q1/test.txt
```

### Test Large Files

```bash
# Create a 10MB test file
dd if=/dev/zero of=largefile.bin bs=1M count=10

# Upload it
curl -X POST http://localhost:8080/upload \
  -F file=@largefile.bin \
  -F remote_path=large/test.bin

# Clean up
rm largefile.bin
```

## Inspecting the SMB Server

### Access the SMB Share from Container

```bash
# List files
docker exec smb-server ls -la /share/

# View file contents
docker exec smb-server cat /share/uploads/test.txt

# Check disk usage
docker exec smb-server du -sh /share/
```

### Access the SMB Share from Host (Linux/macOS)

If you have SMB client tools installed:

```bash
# Install smbclient (if not already installed)
# Ubuntu/Debian: sudo apt-get install smbclient
# macOS: brew install samba

# List shares
smbclient -L localhost -U testuser

# Access the share
smbclient //localhost/testshare -U testuser
# Enter password: testpass
# Then use: ls, cd, get, put commands
```

## Troubleshooting

### Service Won't Start

**Check if ports are already in use:**
```bash
# Check port 8080 (relay service)
lsof -i :8080

# Check port 445 (SMB)
sudo lsof -i :445
```

**View service logs:**
```bash
docker-compose logs smb-relay
docker-compose logs samba
```

### Connection Refused Errors

**Wait for services to be ready:**
```bash
# Wait for health check
docker-compose ps

# Check if Samba is accepting connections
docker exec smb-server netstat -tuln | grep 445
```

### Permission Issues

**Check Samba server status:**
```bash
docker exec smb-server testparm -s
docker exec smb-server smbstatus
```

### Reset Everything

```bash
# Stop and remove everything
docker-compose down -v

# Remove images (optional)
docker rmi smb-relay-service

# Start fresh
docker-compose build --no-cache
docker-compose up -d
```

## Running Tests

### Unit Tests

Run unit tests without starting the services:

```bash
# Install dependencies locally
pip install -r requirements.txt -r requirements-test.txt

# Run tests
./run_tests.sh unit
```

### Integration Tests

The repository includes integration tests that use Docker:

```bash
./run_tests.sh integration
```

These tests will automatically start a temporary SMB server.

## Performance Testing

### Concurrent Uploads

```bash
# Install parallel if needed: sudo apt-get install parallel

# Upload 10 files concurrently
seq 1 10 | parallel -j 10 "echo 'Test {}' > /tmp/test{}.txt && \
  curl -X POST http://localhost:8080/upload \
  -F file=@/tmp/test{}.txt \
  -F remote_path=concurrent/test{}.txt"
```

### Measure Upload Speed

```bash
# Create a 100MB file
dd if=/dev/zero of=testfile.bin bs=1M count=100

# Time the upload
time curl -X POST http://localhost:8080/upload \
  -F file=@testfile.bin \
  -F remote_path=performance/test.bin

# Clean up
rm testfile.bin
```

## Cleanup

### Remove All Containers and Volumes

```bash
docker-compose down -v
```

### Remove Images

```bash
docker rmi smb-relay-service dperson/samba
```

### Clean Docker System

```bash
# Remove all unused containers, networks, and images
docker system prune -a --volumes
```

## CI/CD Integration

The integration tests in this repository use a similar Docker-based setup. To run them locally:

```bash
# Run all tests (unit + integration)
./run_tests.sh

# Run only integration tests
./run_tests.sh integration
```

## Additional Resources

- [Samba Documentation](https://www.samba.org/samba/docs/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Main README](./README.md)
- [DFS and Kerberos Guide](./DFS_KERBEROS.md)
