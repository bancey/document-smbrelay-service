"""SMB connection management for the SMB Relay Service."""

import logging
from smb.SMBConnection import SMBConnection

logger = logging.getLogger(__name__)


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
    logger.info(
        f"Attempting SMB connection to {server_name} ({server_ip}:{port}) "
        f"with user '{username}'{f' in domain {domain}' if domain else ''}, "
        f"NTLMv2={'enabled' if use_ntlm_v2 else 'disabled'}"
    )
    
    try:
        conn = SMBConnection(
            username,
            password,
            "fastapi-smb-relay",
            server_name,
            domain=domain,
            use_ntlm_v2=use_ntlm_v2,
        )
        
        logger.debug(f"SMBConnection object created, attempting connection to {server_ip}:{port}")
        
        if not conn.connect(server_ip, port):
            logger.error(f"SMB connection failed to {server_name} ({server_ip}:{port})")
            raise ConnectionError("Could not connect to SMB server")
        
        logger.info(f"SMB connection established successfully to {server_name} ({server_ip}:{port})")
        return conn
        
    except Exception as e:
        logger.error(f"SMB connection error to {server_name} ({server_ip}:{port}): {e}")
        raise


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
            logger.debug(f"Testing share access by listing root directory of '{share_name}'")
            conn.listPath(share_name, "/")
            
            logger.info(f"SMB health check successful - share '{share_name}' is accessible")
            return {
                "status": "healthy",
                "smb_connection": "ok",
                "smb_share_accessible": True,
                "server": f"{server_name} ({server_ip}:{port})",
                "share": share_name
            }
        finally:
            try:
                logger.debug("Closing SMB connection")
                conn.close()
            except Exception as close_error:
                logger.warning(f"Error closing SMB connection: {close_error}")
                
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