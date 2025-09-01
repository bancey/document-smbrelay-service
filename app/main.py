from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse
from smb.SMBConnection import SMBConnection
import asyncio
import os
import tempfile

app = FastAPI(title="Document SMB Relay Service")


async def save_upload_to_temp(upload_file: UploadFile) -> str:
    suffix = os.path.splitext(upload_file.filename)[1]
    fd, path = tempfile.mkstemp(suffix=suffix)
    os.close(fd)
    with open(path, "wb") as dest:
        while True:
            chunk = await upload_file.read(1024 * 64)
            if not chunk:
                break
            dest.write(chunk)
    await upload_file.close()
    return path


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
    basename = os.path.basename(remote_path)
    remote_dir = os.path.dirname(remote_path)

    conn = SMBConnection(username, password, "fastapi-smb-relay", server_name, domain=domain, use_ntlm_v2=use_ntlm_v2)
    connected = conn.connect(server_ip, port)
    if not connected:
        raise ConnectionError("Could not connect to SMB server")

    # Create directories one by one for nested paths
    if remote_dir:
        parts = [p for p in remote_dir.split("/") if p]
        cur = ""
        for part in parts:
            cur = f"{cur}/{part}" if cur else part
            
            # First check if directory already exists
            directory_exists = False
            try:
                conn.listPath(share_name, cur)
                directory_exists = True
            except Exception:
                directory_exists = False
            
            # If directory doesn't exist, try to create it
            if not directory_exists:
                try:
                    conn.createDirectory(share_name, cur)
                except Exception as create_error:
                    # Directory creation failed - try alternative method
                    # Some SMB implementations are flaky with createDirectory
                    # Try creating a dummy file to force directory creation
                    try:
                        dummy_path = f"{cur}/.dummy_file_for_dir_creation"
                        import io
                        dummy_content = io.BytesIO(b"temp")
                        conn.storeFile(share_name, dummy_path, dummy_content)
                        # Try to delete the dummy file
                        try:
                            conn.deleteFiles(share_name, dummy_path)
                        except Exception:
                            pass  # Ignore deletion errors
                    except Exception:
                        # Both methods failed - directory creation might not be supported
                        # Continue anyway as the upload might still work
                        pass

    # If not allowed to overwrite, check if file exists first
    if not overwrite:
        try:
            conn.getAttributes(share_name, remote_path)
        except Exception:
            # If getAttributes fails, assume file doesn't exist and continue
            pass
        else:
            # getAttributes succeeded -> file exists
            conn.close()
            raise FileExistsError(f"Remote file already exists: {remote_path}")

    with open(local_path, "rb") as file_obj:
        try:
            conn.storeFile(share_name, remote_path, file_obj)
        except Exception as store_error:
            # If storeFile fails and we have nested directories, it might be because
            # the directory creation failed silently. Let's provide a better error message.
            if remote_dir and ("unable to open" in str(store_error).lower() or 
                             "no such file" in str(store_error).lower() or
                             "path not found" in str(store_error).lower()):
                raise ConnectionError(f"Failed to store {remote_path} on {share_name}: "
                                    f"Directory path may not exist. Original error: {store_error}")
            else:
                raise ConnectionError(f"Failed to store {remote_path} on {share_name}: {store_error}")

    conn.close()


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
    use_ntlm_v2 = os.environ.get("SMB_USE_NTLM_V2", "true").lower() in ("1", "true", "yes")

    missing = [k for k, v in (
        ("SMB_SERVER_NAME", server_name),
        ("SMB_SERVER_IP", server_ip),
        ("SMB_SHARE_NAME", share_name),
        ("SMB_USERNAME", username),
        ("SMB_PASSWORD", password),
    ) if not v]
    if missing:
        raise HTTPException(status_code=500, detail=f"Missing SMB configuration environment variables: {', '.join(missing)}")

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
