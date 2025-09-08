"""SMB file operations for the SMB Relay Service."""

import os
from smb.SMBConnection import SMBConnection
from .connection import get_conn


def ensure_dirs(conn: SMBConnection, share_name: str, dir_path: str) -> None:
    """Ensure that directories exist on the SMB share, creating them if needed.
    
    Args:
        conn: Active SMB connection
        share_name: Name of the SMB share
        dir_path: Directory path to ensure exists
    """
    if not dir_path:
        return
    parts = [p for p in dir_path.split("/") if p]
    current_path = ""
    for part in parts:
        current_path = f"{current_path}/{part}" if current_path else part
        try:
            conn.listPath(share_name, current_path)
            continue
        except Exception:
            # If listPath fails, attempt to create and ignore errors
            try:
                conn.createDirectory(share_name, current_path)
            except Exception:
                # Some servers auto-create directories or deny listing; ignore
                pass


def remote_exists(conn: SMBConnection, share_name: str, path: str) -> bool:
    """Check if a file exists on the SMB share.
    
    Args:
        conn: Active SMB connection
        share_name: Name of the SMB share
        path: Path to the file to check
        
    Returns:
        bool: True if file exists, False otherwise
    """
    try:
        conn.getAttributes(share_name, path)
    except Exception:
        return False
    return True


def store(
    conn: SMBConnection, share_name: str, local: str, remote: str, remote_dir: str
) -> None:
    """Store a local file to the SMB share.
    
    Args:
        conn: Active SMB connection
        share_name: Name of the SMB share
        local: Local file path to upload
        remote: Remote file path on the share
        remote_dir: Remote directory path (for error messages)
        
    Raises:
        ConnectionError: If file storage fails
    """
    with open(local, "rb") as fh:
        try:
            conn.storeFile(share_name, remote, fh)
        except Exception as err:
            msg = str(err).lower()
            if remote_dir and (
                "unable to open" in msg
                or "no such file" in msg
                or "path not found" in msg
            ):
                raise ConnectionError(
                    f"Failed to store {remote} on {share_name}: Directory path may not exist. Original error: {err}"
                )
            raise ConnectionError(f"Failed to store {remote} on {share_name}: {err}")


def smb_upload_file(
    local_path: str,
    server_name: str,
    server_ip: str,
    share_name: str,
    remote_path: str,
    username: str,
    password: str,
    domain: str = "",
    port: int = 445,
    use_ntlm_v2: bool = True,
    overwrite: bool = False,
):
    """Upload a local file to an SMB share.
    
    Args:
        local_path: Path to the local file to upload
        server_name: NetBIOS name of the SMB server
        server_ip: IP address of the SMB server
        share_name: Name of the SMB share
        remote_path: Path on the SMB share where file will be stored
        username: SMB username
        password: SMB password
        domain: SMB domain/workgroup (optional)
        port: SMB port (default: 445)
        use_ntlm_v2: Whether to use NTLMv2 authentication (default: True)
        overwrite: Whether to overwrite existing files (default: False)
        
    Raises:
        FileExistsError: If file exists and overwrite is False
        ConnectionError: If SMB connection or file operations fail
    """
    remote_dir = os.path.dirname(remote_path)
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
        ensure_dirs(conn, share_name, remote_dir)

        if not overwrite and remote_exists(conn, share_name, remote_path):
            raise FileExistsError(f"Remote file already exists: {remote_path}")

        store(conn, share_name, local_path, remote_path, remote_dir)
    finally:
        try:
            conn.close()
        except Exception:
            pass