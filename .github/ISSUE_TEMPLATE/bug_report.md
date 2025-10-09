---
name: Bug Report
about: Create a report to help us improve the service
title: '[BUG] '
labels: ['bug']
assignees: ''

---

## Bug Description
A clear and concise description of what the bug is.

## Steps to Reproduce
Steps to reproduce the behavior:
1. Set environment variables: `SMB_SERVER_NAME=...`, `SMB_SERVER_IP=...`, etc.
2. Start the service with: `uvicorn app.main:app --host 0.0.0.0 --port 8080`
3. Make request: `curl -X POST http://localhost:8080/upload -F file=@example.pdf -F remote_path=test.pdf`
4. See error

## Expected Behavior
A clear and concise description of what you expected to happen.

## Actual Behavior
A clear and concise description of what actually happened.

## Error Messages
If applicable, include any error messages or logs:
```
Paste error messages here
```

## Environment Information
- **Operating System**: [e.g., Ubuntu 22.04, Windows 11, macOS 14]
- **Python Version**: [e.g., 3.12.3]
- **Service Version**: [e.g., latest, commit hash, or Docker tag]
- **Deployment Method**: [e.g., Docker, direct Python, Kubernetes]

## SMB Server Details
- **SMB Server Type**: [e.g., Windows Server, Samba, NAS device]
- **SMB Protocol Version**: [e.g., SMB 2.1, SMB 3.0]
- **Authentication Method**: [e.g., NTLM, Kerberos]

## Configuration
**Environment Variables** (remove sensitive values):
```
SMB_SERVER_NAME=example-server
SMB_SERVER_IP=192.168.1.100
SMB_SHARE_NAME=documents
SMB_USERNAME=username
SMB_PASSWORD=*** (hidden)
SMB_DOMAIN=domain
SMB_PORT=445
SMB_USE_NTLM_V2=true
LOG_LEVEL=INFO
```

## Request Details
If the issue is related to file uploads:
- **File Type**: [e.g., PDF, DOCX, image]
- **File Size**: [e.g., 2MB, 50KB]
- **Remote Path**: [e.g., `inbox/document.pdf`, `reports/2024/report.xlsx`]
- **Overwrite Setting**: [e.g., true, false]

## Additional Context
Add any other context about the problem here, such as:
- Network configuration details
- Firewall settings
- SMB share permissions
- Any recent changes to the environment

## Possible Solution
If you have an idea for how to fix the issue, please describe it here.

## Checklist
- [ ] I have searched existing issues to ensure this is not a duplicate
- [ ] I have included all relevant environment information
- [ ] I have removed sensitive information from configuration examples
- [ ] I have tested with the latest version of the service