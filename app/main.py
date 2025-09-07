from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse
import asyncio
import os
from contextlib import suppress

from .config import load_smb_config_from_env
from .file_utils import save_upload_to_temp
from .smb import smb_upload_file, check_smb_health

app = FastAPI(title="Document SMB Relay Service")

# Placeholder to allow tests to patch `app.main.SMBConnection` when pysmb
# is not installed in the test environment.
SMBConnection = None


# Re-export functions for backward compatibility with tests that import from app.main
# (tests patch app.main.smb_upload_file and app.main.check_smb_health)
__all__ = [
    "app",
    "save_upload_to_temp",
    "load_smb_config_from_env",
    "smb_upload_file",
    "check_smb_health",
]


@app.get("/health")
async def health():
    # Load SMB configuration from environment variables
    config, missing = load_smb_config_from_env()

    if missing:
        return JSONResponse(
            status_code=503,
            content={
                "status": "unhealthy",
                "app_status": "ok",
                "smb_connection": "not_configured",
                "smb_share_accessible": False,
                "error": f"Missing SMB configuration environment variables: {', '.join(missing)}",
            },
        )

    # Test SMB connectivity in a thread to avoid blocking the event loop
    smb_health = await asyncio.to_thread(
        check_smb_health,
        config["server_name"],
        config["server_ip"],
        config["share_name"],
        config["username"],
        config["password"],
        config["domain"],
        config["port"],
        config["use_ntlm_v2"],
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
    config, missing = load_smb_config_from_env()
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
        await asyncio.to_thread(
            smb_upload_file,
            tmp_path,
            config["server_name"],
            config["server_ip"],
            config["share_name"],
            remote_path,
            config["username"],
            config["password"],
            config["domain"],
            config["port"],
            config["use_ntlm_v2"],
            overwrite,
        )
    except FileExistsError as e:
        raise HTTPException(status_code=409, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    finally:
        with suppress(Exception):
            os.remove(tmp_path)

    return JSONResponse({"status": "ok", "remote_path": remote_path})
