import os
from typing import Dict, Any
import logging
import importlib

logger = logging.getLogger("app.smb")


def get_conn(
    username: str,
    password: str,
    server_name: str,
    server_ip: str,
    domain: str = "",
    port: int = 445,
    use_ntlm_v2: bool = True,
):
    # Attempt to import pysmb's SMBConnection. If it's not available (e.g. in
    # environments where dependencies are not installed), fall back to trying
    # to read a patched SMBConnection object from app.main so unit tests can
    # monkeypatch app.main.SMBConnection.
    SMBConnection = None
    try:
        mod = importlib.import_module("smb.SMBConnection")
        SMBConnection = getattr(mod, "SMBConnection")
    except Exception:
        try:
            main_mod = importlib.import_module("app.main")
            SMBConnection = getattr(main_mod, "SMBConnection", None)
        except Exception:
            SMBConnection = None

    if SMBConnection is None:
        raise ImportError("SMBConnection implementation not available")

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


def ensure_dirs(conn: Any, share_name: str, dir_path: str) -> None:
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
            try:
                conn.createDirectory(share_name, current_path)
            except Exception:
                logger.debug("Could not create directory %s on share %s", current_path, share_name)
                pass


def remote_exists(conn: Any, share_name: str, path: str) -> bool:
    try:
        conn.getAttributes(share_name, path)
    except Exception:
        return False
    return True


def store(
    conn: Any, share_name: str, local: str, remote: str, remote_dir: str
) -> None:
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
            logger.debug("Error while closing SMB connection", exc_info=True)
            pass


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
            conn.listPath(share_name, "/")
            return {
                "status": "healthy",
                "smb_connection": "ok",
                "smb_share_accessible": True,
                "server": f"{server_name} ({server_ip}:{port})",
                "share": share_name,
            }
        finally:
            try:
                conn.close()
            except Exception:
                logger.debug("Error closing connection during health check", exc_info=True)
                pass
    except Exception as e:
        return {
            "status": "unhealthy",
            "smb_connection": "failed",
            "smb_share_accessible": False,
            "server": f"{server_name} ({server_ip}:{port})",
            "share": share_name,
            "error": str(e),
        }
