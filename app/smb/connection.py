"""SMB connection management for the SMB Relay Service."""

from smb.SMBConnection import SMBConnection


def get_conn(
    username: str,
    password: str,
    server_name: str,
    server_ip: str,
    domain: str = "",
    port: int = 445,
    use_ntlm_v2: bool = True,
) -> SMBConnection:
    """Create and establish an SMB connection.
    
    Args:
        username: SMB username
        password: SMB password  
        server_name: NetBIOS name of the SMB server
        server_ip: IP address of the SMB server
        domain: SMB domain/workgroup (optional)
        port: SMB port (default: 445)
        use_ntlm_v2: Whether to use NTLMv2 authentication (default: True)
        
    Returns:
        SMBConnection: Connected SMB connection object
        
    Raises:
        ConnectionError: If connection to SMB server fails
    """
    conn = SMBConnection(
        username,
        password,
        "fastapi-smb-relay",
        server_name,
        domain=domain,
        use_ntlm_v2=use_ntlm_v2,
    )
    if not conn.connect(server_ip, port):
        raise ConnectionError("Could not connect to SMB server")
    return conn


def check_smb_health(
    server_name: str,
    server_ip: str,
    share_name: str,
    username: str,
    password: str,
    domain: str = "",
    port: int = 445,
    use_ntlm_v2: bool = True,
) -> dict:
    """Check SMB server connectivity and share accessibility.
    
    Args:
        server_name: NetBIOS name of the SMB server
        server_ip: IP address of the SMB server
        share_name: Name of the SMB share
        username: SMB username
        password: SMB password
        domain: SMB domain/workgroup (optional)
        port: SMB port (default: 445)
        use_ntlm_v2: Whether to use NTLMv2 authentication (default: True)
        
    Returns:
        dict: Health check result with status information
    """
    try:
        conn = get_conn(
            username,
            password,
            server_name,
            server_ip,
            domain,
            port,
            use_ntlm_v2,
        )
        try:
            # Test basic share access by listing root directory
            conn.listPath(share_name, "/")
            return {
                "status": "healthy",
                "smb_connection": "ok",
                "smb_share_accessible": True,
                "server": f"{server_name} ({server_ip}:{port})",
                "share": share_name
            }
        finally:
            try:
                conn.close()
            except Exception:
                pass
    except Exception as e:
        return {
            "status": "unhealthy", 
            "smb_connection": "failed",
            "smb_share_accessible": False,
            "server": f"{server_name} ({server_ip}:{port})",
            "share": share_name,
            "error": str(e)
        }