# Testing DFS Connectivity with Docker Compose

This guide explains how to test the Document SMB Relay Service with a simulated DFS (Distributed File System) environment using Docker Compose.

## Overview

The `docker-compose.dfs.yml` file creates a simulated Windows DFS environment with:

- **DFS Namespace Server**: Acts as the entry point (like a Windows DFS namespace server)
- **File Server 1**: First DFS target with `documents` and `projects` shares
- **File Server 2**: Second DFS target with `documents` (redundant) and `archive` shares
- **SMB Relay Service**: Configured to connect to the DFS namespace

**Important Note**: This is a simulation of DFS behavior using standard Samba servers. True Windows DFS requires Active Directory, domain controllers, and the DFS Namespace service. However, this setup demonstrates the core concepts and allows testing the service's SMB connectivity patterns.

## Architecture

```
┌─────────────────────┐
│   SMB Relay Service │
│   (port 8080)       │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  DFS Namespace      │◄── Entry point (dfs-namespace)
│  dfs-namespace:445  │
└─────────────────────┘
           │
           ├──────────────┐
           ▼              ▼
┌──────────────┐  ┌──────────────┐
│ File Server 1│  │ File Server 2│
│ fileserver1  │  │ fileserver2  │
│ - documents  │  │ - documents  │
│ - projects   │  │ - archive    │
└──────────────┘  └──────────────┘
```

## Quick Start

### 1. Start the DFS Environment

```bash
# Start all services
docker-compose -f docker-compose.dfs.yml up -d

# Check all services are running
docker-compose -f docker-compose.dfs.yml ps

# Watch logs
docker-compose -f docker-compose.dfs.yml logs -f
```

### 2. Wait for Services to Initialize

```bash
# Wait for SMB servers to be ready (takes about 10-15 seconds)
sleep 15

# Check relay service health
curl http://localhost:8080/health | jq
```

### 3. Test File Upload to DFS Namespace

```bash
# Create a test file
echo "Testing DFS connectivity" > dfs-test.txt

# Upload to DFS namespace
curl -X POST http://localhost:8080/upload \
  -F file=@dfs-test.txt \
  -F remote_path=test/dfs-test.txt

# Verify upload on DFS namespace server
docker exec dfs-namespace ls -la /dfs-root/test/
```

### 4. Test Direct Access to File Servers

```bash
# Test accessing fileserver1 directly
# (requires uncommenting smb-relay-direct in docker-compose.dfs.yml)

# Access fileserver1 documents share
docker exec fileserver1 ls -la /share/documents/

# Access fileserver2 archive share
docker exec fileserver2 ls -la /share/archive/
```

### 5. Stop the Environment

```bash
docker-compose -f docker-compose.dfs.yml down

# Remove volumes as well
docker-compose -f docker-compose.dfs.yml down -v
```

## Testing Scenarios

### Scenario 1: DFS Namespace Connectivity

Test that the service can connect to and write files through the DFS namespace server:

```bash
# Upload file to DFS root
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=namespace-test.txt

# Verify on namespace server
docker exec dfs-namespace cat /dfs-root/namespace-test.txt
```

### Scenario 2: Multiple File Servers

Inspect the different file servers to understand the DFS topology:

```bash
# List shares on fileserver1
docker exec fileserver1 smbclient -L localhost -U testuser%testpass

# List shares on fileserver2
docker exec fileserver2 smbclient -L localhost -U testuser%testpass

# List shares on DFS namespace
docker exec dfs-namespace smbclient -L localhost -U testuser%testpass
```

### Scenario 3: Direct vs DFS Access

Compare accessing file servers directly vs through DFS namespace:

```bash
# Via DFS namespace (port 8080)
curl -X POST http://localhost:8080/upload \
  -F file=@test1.txt \
  -F remote_path=via-dfs.txt

# Enable smb-relay-direct in docker-compose.dfs.yml and restart
# Then via direct connection (port 8081)
curl -X POST http://localhost:8081/upload \
  -F file=@test2.txt \
  -F remote_path=via-direct.txt
```

### Scenario 4: Different Shares on Different Servers

```bash
# Upload to documents share (available on both servers)
curl -X POST http://localhost:8080/upload \
  -F file=@doc.txt \
  -F remote_path=documents/doc.txt

# Check both servers
docker exec fileserver1 ls -la /share/documents/
docker exec fileserver2 ls -la /share/documents/
```

### Scenario 5: Testing with Kerberos (Simulated)

Modify the relay service environment to use Kerberos authentication:

```bash
# Edit docker-compose.dfs.yml and change:
# SMB_AUTH_PROTOCOL=negotiate
# to:
# SMB_AUTH_PROTOCOL=kerberos

# Restart the service
docker-compose -f docker-compose.dfs.yml up -d smb-relay-dfs

# Test upload (will attempt Kerberos)
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=kerberos-test.txt
```

## Inspecting the Environment

### View SMB Server Configurations

```bash
# Check Samba configuration on DFS namespace
docker exec dfs-namespace testparm -s

# Check Samba configuration on fileserver1
docker exec fileserver1 testparm -s

# Check Samba status
docker exec dfs-namespace smbstatus
```

### View Network Connectivity

```bash
# Check which containers can reach each other
docker exec smb-relay-dfs ping -c 3 dfs-namespace
docker exec smb-relay-dfs ping -c 3 fileserver1
docker exec smb-relay-dfs ping -c 3 fileserver2

# Check DNS resolution
docker exec smb-relay-dfs nslookup dfs-namespace
docker exec smb-relay-dfs nslookup fileserver1.testdomain.local
```

### Access SMB Shares from Host

If you have `smbclient` installed on your host:

```bash
# Access DFS namespace
smbclient //localhost/dfs-root -U testuser%testpass

# List files
ls

# Download a file
get test.txt

# Exit
exit
```

## Understanding the Simulation

### How This Simulates DFS

1. **Namespace Server**: The `dfs-namespace` container acts as the entry point, similar to a Windows DFS namespace server.

2. **File Servers**: The `fileserver1` and `fileserver2` containers represent actual file servers that would be DFS targets.

3. **Share Distribution**: Different shares are distributed across file servers, simulating DFS link targets.

### Differences from Real Windows DFS

| Feature | Real Windows DFS | This Simulation |
|---------|------------------|-----------------|
| Referral System | Automatic DFS referrals | Manual share access |
| Active Directory | Required | Not used |
| Domain Controllers | Multiple DCs | Single namespace server |
| Failover | Automatic | Manual configuration |
| Load Balancing | Built-in | Not implemented |
| Cache | DFS client cache | Standard SMB caching |

### What This Tests

✅ **SMB Connectivity**: Verifies the service can connect to multiple SMB servers
✅ **Share Access**: Tests accessing different shares on different servers
✅ **Authentication**: Tests NTLM, negotiate, and Kerberos protocols
✅ **Network Topology**: Simulates a distributed file environment
✅ **Multiple Targets**: Shows how to configure multiple file server destinations

❌ **Real DFS Referrals**: Does not test actual Windows DFS referral mechanisms
❌ **Domain Authentication**: Does not use real Active Directory
❌ **Automatic Failover**: Does not test DFS automatic failover

## Real Windows DFS Testing

For testing with **real Windows DFS**, you would need:

1. **Windows Server** with DFS Namespace role installed
2. **Active Directory** domain
3. **Multiple file servers** joined to the domain
4. **DFS namespace** configured with links to shares on file servers
5. **DNS** properly configured for domain resolution

Then configure the service:

```bash
SMB_SERVER_NAME=dfs.yourdomain.com
SMB_SERVER_IP=dfs.yourdomain.com
SMB_SHARE_NAME=dfs-namespace
SMB_AUTH_PROTOCOL=kerberos
SMB_DOMAIN=YOURDOMAIN
```

See [DFS_KERBEROS.md](./DFS_KERBEROS.md) for detailed Windows DFS setup instructions.

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker-compose -f docker-compose.dfs.yml logs

# Check specific service
docker-compose -f docker-compose.dfs.yml logs dfs-namespace
docker-compose -f docker-compose.dfs.yml logs smb-relay-dfs
```

### Connection Refused

```bash
# Wait longer for services to initialize
sleep 20

# Check if ports are accessible
docker exec smb-relay-dfs nc -zv dfs-namespace 445
docker exec smb-relay-dfs nc -zv fileserver1 445
```

### Upload Fails

```bash
# Check relay service health
curl http://localhost:8080/health | jq

# Check SMB server status
docker exec dfs-namespace smbstatus

# View detailed logs
docker-compose -f docker-compose.dfs.yml logs -f smb-relay-dfs
```

### Permission Denied

```bash
# Verify credentials
docker exec dfs-namespace pdbedit -L

# Check share permissions
docker exec dfs-namespace testparm -s | grep -A 5 "dfs-root"
```

## Advanced Configuration

### Adding More File Servers

Edit `docker-compose.dfs.yml` and add additional fileserver services:

```yaml
fileserver3:
  image: dperson/samba:latest
  container_name: fileserver3
  environment:
    - USER=testuser;testpass
    - SHARE=backup;/share/backup;yes;no;no;testuser;testuser;;Backup Share
    - WORKGROUP=TESTDOMAIN
  volumes:
    - fileserver3-data:/share
  networks:
    - dfs-network
```

### Using Different Authentication

Change `SMB_AUTH_PROTOCOL` in the relay service:

- `ntlm`: Standard NTLM authentication (default)
- `negotiate`: Protocol negotiation (NTLM or Kerberos)
- `kerberos`: Kerberos authentication (simulated without real AD)

### Custom Network Configuration

Modify the network settings for specific IP addresses:

```yaml
networks:
  dfs-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.25.0.0/16
          gateway: 172.25.0.1
```

## Performance Testing

### Concurrent Uploads

```bash
# Upload 10 files concurrently
seq 1 10 | parallel -j 10 \
  "echo 'Test {}' > /tmp/test{}.txt && \
   curl -X POST http://localhost:8080/upload \
   -F file=@/tmp/test{}.txt \
   -F remote_path=concurrent/test{}.txt"
```

### Large File Upload

```bash
# Create 50MB file
dd if=/dev/zero of=large.bin bs=1M count=50

# Upload and time it
time curl -X POST http://localhost:8080/upload \
  -F file=@large.bin \
  -F remote_path=large.bin

# Verify size
docker exec dfs-namespace ls -lh /dfs-root/large.bin

# Cleanup
rm large.bin
```

## Cleanup

```bash
# Stop all services
docker-compose -f docker-compose.dfs.yml down

# Remove all data
docker-compose -f docker-compose.dfs.yml down -v

# Remove images (optional)
docker rmi dperson/samba
```

## Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Samba Documentation](https://www.samba.org/samba/docs/)
- [Windows DFS Documentation](https://docs.microsoft.com/en-us/windows-server/storage/dfs-namespaces/)
- [Main Documentation](./README.md)
- [DFS and Kerberos Guide](./DFS_KERBEROS.md)
- [Docker Testing Guide](./DOCKER_TESTING.md)
