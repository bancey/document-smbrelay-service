"""File utility functions for the SMB Relay Service."""

import os
import tempfile
import aiofiles
from fastapi import UploadFile


async def save_upload_to_temp(upload_file: UploadFile) -> str:
    """Save an uploaded file to a temporary location.
    
    Args:
        upload_file: The uploaded file from FastAPI
        
    Returns:
        str: Path to the temporary file
    """
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