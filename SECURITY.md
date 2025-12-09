# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of the Document SMB Relay Service seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please Do Not

- **Do not** open a public GitHub issue for security vulnerabilities
- **Do not** disclose the vulnerability publicly until it has been addressed

### Reporting Process

**Please report security vulnerabilities by emailing the maintainer directly or using GitHub's private vulnerability reporting feature:**

1. **GitHub Private Reporting** (Preferred):
   - Go to the [Security tab](https://github.com/bancey/document-smbrelay-service/security)
   - Click "Report a vulnerability"
   - Fill in the details using the form

2. **Email**:
   - Contact the repository maintainer through GitHub
   - Include "SECURITY" in the subject line
   - Provide detailed information (see below)

### What to Include

When reporting a vulnerability, please include:

- **Description**: Clear description of the vulnerability
- **Impact**: Potential impact and severity assessment
- **Reproduction Steps**: Detailed steps to reproduce the issue
- **Affected Versions**: Which versions are affected
- **Proof of Concept**: Code or configuration demonstrating the issue (if applicable)
- **Suggested Fix**: If you have ideas for how to fix it (optional)
- **Your Contact Information**: How we can reach you for follow-up

### What to Expect

- **Acknowledgment**: We will acknowledge receipt within 48 hours
- **Initial Assessment**: We will provide an initial assessment within 5 business days
- **Updates**: We will keep you informed of our progress
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days
- **Credit**: We will credit you in the security advisory (unless you prefer to remain anonymous)

### Security Update Process

1. We will investigate and validate the vulnerability
2. We will develop and test a fix
3. We will prepare a security advisory
4. We will release a patched version
5. We will publish the security advisory with credit to the reporter (if desired)

## Security Best Practices

When deploying the Document SMB Relay Service, we recommend:

### Authentication & Access Control

- **Use strong credentials**: Ensure SMB credentials are complex and regularly rotated
- **Limit network access**: Use firewalls to restrict access to the service
- **Enable authentication**: Always use NTLM, Negotiate, or Kerberos authentication
- **Secure credential storage**: Use secrets management systems (not environment variables in production)

### Network Security

- **Use HTTPS**: Deploy behind a reverse proxy with TLS/HTTPS
- **Network segmentation**: Isolate the service in a secure network segment
- **SMB encryption**: Enable SMB encryption on your SMB server
- **Firewall rules**: Restrict SMB access to only necessary sources

### Application Security

- **Keep updated**: Regularly update to the latest version
- **Monitor logs**: Enable and monitor application logs (LOG_LEVEL=INFO or DEBUG)
- **File validation**: Validate uploaded files on the client side before upload
- **Path restrictions**: Use SMB_BASE_PATH to restrict file operations to specific directories
- **Rate limiting**: Implement rate limiting at the reverse proxy level

### Container Security

- **Use official images**: Only use official images from trusted registries
- **Scan for vulnerabilities**: Regularly scan container images
- **Run as non-root**: The service runs as non-root user in containers
- **Read-only filesystem**: Consider mounting the container filesystem as read-only
- **Resource limits**: Set appropriate CPU and memory limits

### Monitoring & Observability

- **Enable health checks**: Use the `/health` endpoint for monitoring
- **OpenTelemetry**: Enable OpenTelemetry for comprehensive observability
- **Alert on failures**: Set up alerts for authentication failures and errors
- **Audit logs**: Maintain audit logs of file uploads and access patterns

## Known Security Considerations

### SMB Protocol

- The service relies on SMB protocol security features
- Ensure your SMB server is properly configured and patched
- Consider using SMB3 with encryption for sensitive data

### File Uploads

- The service does not validate file contents or types
- Implement validation in your application layer
- Consider scanning uploaded files for malware

### Credential Exposure

- Credentials are required as environment variables
- Use secrets management in production (Kubernetes secrets, Azure Key Vault, etc.)
- Never commit credentials to version control

### DFS Considerations

- DFS referrals are automatically followed
- Ensure all DFS targets are trusted
- Monitor for unexpected referrals

## Security Features

The service includes several security features:

- **Authentication**: Supports NTLM, Negotiate, and Kerberos
- **Path validation**: Validates paths to prevent directory traversal
- **Error handling**: Careful error messages to avoid information disclosure
- **Health checks**: Validates SMB connectivity without exposing credentials
- **No filesystem mounting**: Direct SMB access without local mounting reduces attack surface
- **Minimal container**: Alpine-based container with minimal packages

## Compliance & Standards

- Follow OWASP guidelines for secure file upload handling
- Implement principle of least privilege for SMB access
- Maintain audit trails for compliance requirements
- Consider data retention policies for your use case

## Questions?

If you have questions about security that don't involve reporting a vulnerability, please:

- Open a [GitHub Discussion](https://github.com/bancey/document-smbrelay-service/discussions)
- Check the [documentation](README.md)
- Review existing [security advisories](https://github.com/bancey/document-smbrelay-service/security/advisories)

Thank you for helping keep the Document SMB Relay Service secure!
