"""SMB connection management for the SMB Relay Service."""

import logging
import smbclient

logger = logging.getLogger(__name__)


def get_conn(
    username: str,
    password: str,
    server_name: str,
    server_ip: str,
    domain: str = "",
    port: int = 445,
    use_ntlm_v2: bool = True,
) -> dict:
    """Create and establish an SMB connection using smbprotocol.
    
    Args:
        username: SMB username
        password: SMB password  
        server_name: NetBIOS name of the SMB server
        server_ip: IP address of the SMB server
        domain: SMB domain/workgroup (optional)
        port: SMB port (default: 445)
        use_ntlm_v2: Whether to use NTLMv2 authentication (default: True)
        
    Returns:
        dict: Connection info with server details
        
    Raises:
        ConnectionError: If connection to SMB server fails
    """
    # Use IP address if provided, otherwise use server_name
    server = server_ip or server_name
    
    logger.info(
        f"Attempting SMB connection to {server_name} ({server_ip}:{port}) "
        f"with user '{username}'{f' in domain {domain}' if domain else ''}, "
        f"NTLMv2={'enabled' if use_ntlm_v2 else 'disabled'}"
    )
    
    try:
        # Construct username with domain if provided  
        auth_username = f"{domain}\\{username}" if domain else username
        
        # Register session with smbclient
        logger.debug(f"Registering SMB session for {server}:{port}")
        session = smbclient.register_session(
            server=server,
            username=auth_username,
            password=password,
            port=port,
            auth_protocol='ntlm' if use_ntlm_v2 else 'negotiate',
        )
        
        logger.info(f"SMB session established successfully to {server_name} ({server_ip}:{port})")
        
        return {
            'server': server,
            'port': port,
            'session': session,
            'server_name': server_name,
            'server_ip': server_ip
        }
        
    except Exception as e:
        logger.error(f"SMB connection error to {server_name} ({server_ip}:{port}): {e}")
        raise ConnectionError(f"Could not connect to SMB server: {e}")


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
    logger.info(f"Starting SMB health check for share '{share_name}' on {server_name} ({server_ip}:{port})")
    
    try:
        conn_info = get_conn(
            username,
            password,
            server_name,
            server_ip,
            domain,
            port,
            use_ntlm_v2,
        )
        
        # Test basic share access by listing root directory
        server = conn_info['server']
        unc_path = f"//{server}/{share_name}/"
        
        logger.debug(f"Testing share access by listing root directory of '{share_name}' at {unc_path}")
        smbclient.listdir(unc_path)
        
        logger.info(f"SMB health check successful - share '{share_name}' is accessible")
        return {
            "status": "healthy",
            "smb_connection": "ok",
            "smb_share_accessible": True,
            "server": f"{server_name} ({server_ip}:{port})",
            "share": share_name
        }
                
    except Exception as e:
        logger.error(f"SMB health check failed for {server_name} ({server_ip}:{port}), share '{share_name}': {e}")
        return {
            "status": "unhealthy", 
            "smb_connection": "failed",
            "smb_share_accessible": False,
            "server": f"{server_name} ({server_ip}:{port})",
            "share": share_name,
            "error": str(e)
        }