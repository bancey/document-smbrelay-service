from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse
import asyncio
import logging
import os

from .config.smb_config import load_smb_config_from_env
from .smb.connection import check_smb_health
from .smb.operations import smb_upload_file
from .utils.file_utils import save_upload_to_temp

# Configure logging
def configure_logging():
    """Configure application logging based on LOG_LEVEL environment variable."""
    log_level = os.getenv("LOG_LEVEL", "INFO").upper()
    
    # Validate log level
    valid_levels = ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]
    if log_level not in valid_levels:
        log_level = "INFO"
    
    # Configure root logger
    logging.basicConfig(
        level=getattr(logging, log_level),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S"
    )
    
    # Set log level for application modules
    logging.getLogger("app").setLevel(getattr(logging, log_level))
    
    logger = logging.getLogger(__name__)
    logger.info(f"Logging configured with level: {log_level}")

# Initialize logging
configure_logging()

app = FastAPI(title="Document SMB Relay Service")

@app.get("/health")
async def health():
    """
    Health check endpoint that verifies application responsiveness and SMB connectivity.

    Returns:
        JSONResponse: A JSON object with the following structure:
            {
                "status": "healthy" or "unhealthy",
                "app_status": "ok",
                "smb_connection": "ok" or "failed" or "not_configured",
                "smb_share_accessible": True or False,
                "server": "<server_name> (<server_ip>:<port>)",
                "share": "<share_name>",
                "error": "<error_message>" (only present if unhealthy or misconfigured)
            }

    HTTP Status Codes:
        200: Returned when the application and SMB server are healthy and accessible.
        503: Returned when the application is unhealthy, SMB configuration is missing, or SMB server/share is inaccessible.
    """
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
                "error": f"Missing SMB configuration environment variables: {', '.join(missing)}"
            }
        )

    # Test SMB connectivity
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
        try:
            os.remove(tmp_path)
        except Exception:
            pass

    return JSONResponse({"status": "ok", "remote_path": remote_path})
