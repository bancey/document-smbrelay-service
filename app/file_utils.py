import os
import tempfile
from pathlib import Path
from typing import Any

# Reuse CHUNK_SIZE from the original file to maintain behaviour
CHUNK_SIZE = 1024 * 64


async def save_upload_to_temp(upload_file: Any) -> str:
    """Save an UploadFile to a temporary file and return its path.

    Lazy-imports external dependencies so the module can be imported in
    environments where optional packages (aiofiles/fastapi) are not
    installed (useful for lightweight unit tests that patch functions).
    """
    # Local imports avoid raising ImportError at module import time
    import aiofiles
    try:
        from fastapi import UploadFile  # type: ignore
    except Exception:
        UploadFile = None  # type: ignore

    suffix = Path(getattr(upload_file, "filename", "")).suffix
    fd, path = tempfile.mkstemp(suffix=suffix)
    os.close(fd)
    async with aiofiles.open(path, "wb") as dest:
        while True:
            chunk = await upload_file.read(CHUNK_SIZE)
            if not chunk:
                break
            await dest.write(chunk)
    # upload_file may be an AsyncMock that implements close
    if hasattr(upload_file, "close"):
        await upload_file.close()
    return path
