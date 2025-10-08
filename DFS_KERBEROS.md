# Windows DFS and Kerberos Support

This document provides detailed information about using the Document SMB Relay Service with Windows Distributed File System (DFS) and Kerberos authentication.

## Overview

The Document SMB Relay Service fully supports:
- **Windows DFS**: Automatic DFS referral handling and path resolution
- **Kerberos Authentication**: Native support for Active Directory authentication
- **Domain Integration**: Seamless integration with Windows domain environments

## Windows DFS Support

### What is DFS?

Windows Distributed File System (DFS) allows organizations to:
- Create a unified namespace across multiple file servers
- Provide automatic failover and load balancing
- Simplify file share access with logical paths

### How DFS Works with This Service

The underlying `smbprotocol` library automatically handles DFS:
1. **DFS Referral Resolution**: When accessing a DFS path, the service automatically receives and follows DFS referrals to the actual file server
2. **No Special Configuration**: Simply point to the DFS namespace - no additional setup required
3. **Transparent Operation**: All DFS operations are handled transparently by the library

### DFS Configuration Example

```bash
# Basic DFS setup
SMB_SERVER_NAME=dfs.corp.example.com
SMB_SERVER_IP=dfs.corp.example.com
SMB_SHARE_NAME=documents
SMB_USERNAME=myuser
SMB_PASSWORD=mypassword
SMB_DOMAIN=CORP
```

## Kerberos Authentication

### Why Use Kerberos?

Kerberos provides:
- **Enhanced Security**: No passwords transmitted over the network
- **Single Sign-On**: Use cached Kerberos tickets
- **Domain Integration**: Natural fit for Active Directory environments
- **Better DFS Experience**: Optimal authentication for domain-based DFS shares

### Kerberos Configuration

#### Option 1: Using Cached Tickets (Recommended for Containers)

```bash
# Obtain Kerberos ticket first (outside container)
kinit user@CORP.EXAMPLE.COM

# Then start the service (no credentials needed)
SMB_SERVER_NAME=dfs.corp.example.com
SMB_SERVER_IP=dfs.corp.example.com
SMB_SHARE_NAME=documents
SMB_AUTH_PROTOCOL=kerberos
```

#### Option 2: Using Explicit Credentials

```bash
SMB_SERVER_NAME=dfs.corp.example.com
SMB_SERVER_IP=dfs.corp.example.com
SMB_SHARE_NAME=documents
SMB_USERNAME=myuser
SMB_PASSWORD=mypassword
SMB_DOMAIN=CORP
SMB_AUTH_PROTOCOL=kerberos
```

### Kerberos Prerequisites

For Kerberos authentication to work, ensure:

1. **Kerberos Configuration** (`/etc/krb5.conf`):
   ```ini
   [libdefaults]
       default_realm = CORP.EXAMPLE.COM
       dns_lookup_realm = true
       dns_lookup_kdc = true
   
   [realms]
       CORP.EXAMPLE.COM = {
           kdc = dc1.corp.example.com
           kdc = dc2.corp.example.com
           admin_server = dc1.corp.example.com
       }
   
   [domain_realm]
       .corp.example.com = CORP.EXAMPLE.COM
       corp.example.com = CORP.EXAMPLE.COM
   ```

2. **DNS Resolution**: Ensure proper DNS resolution for:
   - Domain controllers
   - DFS namespace servers
   - Actual file servers (for DFS referrals)

3. **Time Synchronization**: Kerberos requires synchronized time (within 5 minutes by default)
   ```bash
   # Check time sync
   ntpdate -q pool.ntp.org
   ```

## Complete Docker Example with DFS and Kerberos

### Dockerfile for Kerberos Support

```dockerfile
FROM python:3.13-slim

# Install Kerberos client libraries
RUN apt-get update && apt-get install -y \
    krb5-user \
    libkrb5-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy application
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY app ./app

ENV PYTHONUNBUFFERED=1
EXPOSE 8080

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
```

### Docker Run with Kerberos Configuration

```bash
docker run -d \
  --name smb-relay \
  -p 8080:8080 \
  -v /etc/krb5.conf:/etc/krb5.conf:ro \
  -v /tmp/krb5cc:/tmp/krb5cc:ro \
  -e KRB5CCNAME=/tmp/krb5cc \
  -e SMB_SERVER_NAME=dfs.corp.example.com \
  -e SMB_SERVER_IP=dfs.corp.example.com \
  -e SMB_SHARE_NAME=documents \
  -e SMB_AUTH_PROTOCOL=kerberos \
  document-smb-relay:latest
```

### Kubernetes Example with Kerberos

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: krb5-config
data:
  krb5.conf: |
    [libdefaults]
        default_realm = CORP.EXAMPLE.COM
        dns_lookup_realm = true
        dns_lookup_kdc = true
---
apiVersion: v1
kind: Secret
metadata:
  name: smb-credentials
type: Opaque
stringData:
  username: myuser
  password: mypassword
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: smb-relay
spec:
  replicas: 1
  selector:
    matchLabels:
      app: smb-relay
  template:
    metadata:
      labels:
        app: smb-relay
    spec:
      containers:
      - name: smb-relay
        image: document-smb-relay:latest
        ports:
        - containerPort: 8080
        env:
        - name: SMB_SERVER_NAME
          value: "dfs.corp.example.com"
        - name: SMB_SERVER_IP
          value: "dfs.corp.example.com"
        - name: SMB_SHARE_NAME
          value: "documents"
        - name: SMB_USERNAME
          valueFrom:
            secretKeyRef:
              name: smb-credentials
              key: username
        - name: SMB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: smb-credentials
              key: password
        - name: SMB_DOMAIN
          value: "CORP"
        - name: SMB_AUTH_PROTOCOL
          value: "kerberos"
        volumeMounts:
        - name: krb5-config
          mountPath: /etc/krb5.conf
          subPath: krb5.conf
          readOnly: true
      volumes:
      - name: krb5-config
        configMap:
          name: krb5-config
```

## Troubleshooting

### DFS Issues

**Problem**: "Path not found" or "Access denied" errors
- **Solution**: Verify DFS namespace is accessible from the service host
- **Check**: DNS resolution of DFS namespace server
- **Test**: `smbclient //dfs.corp.example.com/documents -U username`

**Problem**: Connection to wrong server
- **Solution**: DFS is working correctly - the service automatically follows referrals
- **Note**: Logs will show the actual file server after referral resolution

### Kerberos Issues

**Problem**: "Kerberos credentials not available"
- **Solution**: Ensure valid Kerberos ticket or provide explicit credentials
- **Check**: `klist` to view cached tickets
- **Fix**: `kinit user@REALM` to obtain ticket

**Problem**: "Clock skew too great"
- **Solution**: Synchronize system time with domain controllers
- **Check**: `ntpdate -q domain-controller.example.com`
- **Fix**: Enable NTP synchronization

**Problem**: "Server not found in Kerberos database"
- **Solution**: Verify DNS reverse lookup for the SMB server
- **Check**: Ensure SPN (Service Principal Name) is registered for the server
- **Note**: Use IP address instead of hostname if DNS issues persist

### Authentication Issues

**Problem**: "Access denied" with Kerberos
- **Solution**: Verify user has permissions on the share
- **Check**: Try with NTLM to isolate Kerberos-specific issues
- **Debug**: Set `LOG_LEVEL=DEBUG` to see detailed authentication logs

**Problem**: Works with NTLM but not with Kerberos
- **Solution**: Check Kerberos configuration and ticket validity
- **Verify**: Ensure proper DNS resolution and time synchronization
- **Test**: Try accessing the share with `smbclient` using Kerberos

## Testing

### Testing DFS Connectivity

```bash
# Test health endpoint
curl http://localhost:8080/health | jq

# Expected: Should show connection to actual file server after DFS referral
```

### Testing Kerberos Authentication

```bash
# Obtain Kerberos ticket
kinit user@CORP.EXAMPLE.COM

# Start service with Kerberos
SMB_AUTH_PROTOCOL=kerberos \
SMB_SERVER_NAME=dfs.corp.example.com \
SMB_SERVER_IP=dfs.corp.example.com \
SMB_SHARE_NAME=documents \
uvicorn app.main:app

# Test upload
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=test/upload.txt
```

## Best Practices

1. **Use Kerberos for DFS**: Kerberos provides the best experience with domain-based DFS shares
2. **Set Proper DNS**: Ensure reliable DNS resolution for DFS and domain controllers
3. **Monitor Time Sync**: Implement NTP synchronization for Kerberos environments
4. **Use Service Accounts**: Consider using dedicated service accounts with minimal required permissions
5. **Enable Debug Logging**: Use `LOG_LEVEL=DEBUG` during initial setup to troubleshoot issues
6. **Test Connectivity**: Validate DFS and Kerberos setup before deploying the service

## Additional Resources

- [Windows DFS Documentation](https://docs.microsoft.com/en-us/windows-server/storage/dfs-namespaces/dfs-overview)
- [Kerberos Authentication](https://web.mit.edu/kerberos/krb5-latest/doc/)
- [smbprotocol Library](https://github.com/jborean93/smbprotocol)
- [Active Directory Integration](https://docs.microsoft.com/en-us/windows-server/identity/ad-ds/get-started/virtual-dc/active-directory-domain-services-overview)
