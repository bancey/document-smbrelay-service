from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse
from smb.SMBConnection import SMBConnection
import aiofiles
import asyncio
import os
import tempfile

app = FastAPI(title="Document SMB Relay Service")


async def save_upload_to_temp(upload_file: UploadFile) -> str:
    suffix = os.path.splitext(upload_file.filename)[1]
    fd, path = tempfile.mkstemp(suffix=suffix)
    os.close(fd)
    async with aiofiles.open(path, "wb") as dest:
        while True:
            chunk = await upload_file.read(1024 * 64)
            if not chunk:
                break
            await dest.write(chunk)
    await upload_file.close()
    return path


def get_conn(
    username: str,
    password: str,
    server_name: str,
    server_ip: str,
    domain: str = "",
    port: int = 445,
    use_ntlm_v2: bool = True,
) -> SMBConnection:
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


def ensure_dirs(conn: SMBConnection, share_name: str, dir_path: str) -> None:
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
    try:
        conn.getAttributes(share_name, path)
    except Exception:
        return False
    return True


def store(
    conn: SMBConnection, share_name: str, local: str, remote: str, remote_dir: str
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
    """Check SMB server connectivity and share accessibility."""
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


@app.get("/health")
async def health():
    """Health check endpoint that verifies application responsiveness and SMB connectivity."""
    # Load SMB configuration from environment variables
    server_name = os.environ.get("SMB_SERVER_NAME")
    server_ip = os.environ.get("SMB_SERVER_IP")
    share_name = os.environ.get("SMB_SHARE_NAME")
    username = os.environ.get("SMB_USERNAME")
    password = os.environ.get("SMB_PASSWORD")
    domain = os.environ.get("SMB_DOMAIN", "")
    port = int(os.environ.get("SMB_PORT", "445"))
    use_ntlm_v2 = os.environ.get("SMB_USE_NTLM_V2", "true").lower() in (
        "1",
        "true", 
        "yes",
    )

    # Check for missing required environment variables
    missing = [
        k
        for k, v in (
            ("SMB_SERVER_NAME", server_name),
            ("SMB_SERVER_IP", server_ip),
            ("SMB_SHARE_NAME", share_name),
            ("SMB_USERNAME", username),
            ("SMB_PASSWORD", password),
        )
        if not v
    ]
    
    if missing:
        return JSONResponse(
            status_code=503,
            content={
                "status": "unhealthy",
                "app_status": "ok",
                "smb_connection": "not_configured",
                "smb_share_accessible": False,
                "error": f"Missing SMB configuration environment variables: {', '.join(missing)}"
            }
        )

    # Test SMB connectivity
    loop = asyncio.get_event_loop()
    smb_health = await loop.run_in_executor(
        None,
        check_smb_health,
        server_name,
        server_ip,
        share_name,
        username,
        password,
        domain,
        port,
        use_ntlm_v2,
    )

    # Add app status to health response
    smb_health["app_status"] = "ok"
    
    if smb_health["status"] == "healthy":
        return JSONResponse(content=smb_health)
    else:
        return JSONResponse(status_code=503, content=smb_health)


@app.post("/upload")
async def upload(
    file: UploadFile = File(...),
    remote_path: str = Form(...),
    overwrite: bool = Form(False),
):
    # Load SMB configuration from environment variables
    server_name = os.environ.get("SMB_SERVER_NAME")
    server_ip = os.environ.get("SMB_SERVER_IP")
    share_name = os.environ.get("SMB_SHARE_NAME")
    username = os.environ.get("SMB_USERNAME")
    password = os.environ.get("SMB_PASSWORD")
    domain = os.environ.get("SMB_DOMAIN", "")
    port = int(os.environ.get("SMB_PORT", "445"))
    use_ntlm_v2 = os.environ.get("SMB_USE_NTLM_V2", "true").lower() in (
        "1",
        "true",
        "yes",
    )

    missing = [
        k
        for k, v in (
            ("SMB_SERVER_NAME", server_name),
            ("SMB_SERVER_IP", server_ip),
            ("SMB_SHARE_NAME", share_name),
            ("SMB_USERNAME", username),
            ("SMB_PASSWORD", password),
        )
        if not v
    ]
    if missing:
        raise HTTPException(
            status_code=500,
            detail=f"Missing SMB configuration environment variables: {', '.join(missing)}",
        )

    # normalize remote path (pysmb expects no leading slash)
    if not remote_path or remote_path.startswith("/"):
        remote_path = remote_path.lstrip("/")

    tmp_path = await save_upload_to_temp(file)
    try:
        loop = asyncio.get_event_loop()
        await loop.run_in_executor(
            None,
            smb_upload_file,
            tmp_path,
            server_name,
            server_ip,
            share_name,
            remote_path,
            username,
            password,
            domain,
            port,
            use_ntlm_v2,
            overwrite,
        )
    except FileExistsError as e:
        raise HTTPException(status_code=409, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    finally:
        try:
            os.remove(tmp_path)
        except Exception:
            pass

    return JSONResponse({"status": "ok", "remote_path": remote_path})
