import pytest
import os
import tempfile
from unittest.mock import AsyncMock, patch
from app.main import save_upload_to_temp


import pytest


@pytest.mark.unit
class TestSaveUploadToTemp:
    """Test cases for save_upload_to_temp function."""

    @pytest.mark.asyncio
    async def test_save_upload_to_temp_basic(self, mock_upload_file, sample_file_content):
        """Test basic functionality of save_upload_to_temp."""
        temp_path = await save_upload_to_temp(mock_upload_file)
        
        try:
            # Verify file was created
            assert os.path.exists(temp_path)
            
            # Verify content was written correctly
            with open(temp_path, "rb") as f:
                content = f.read()
            assert content == sample_file_content
            
            # Verify file has correct extension
            assert temp_path.endswith(".txt")
            
            # Verify upload file was closed
            mock_upload_file.close.assert_called_once()
            
        finally:
            # Cleanup
            if os.path.exists(temp_path):
                os.unlink(temp_path)

    @pytest.mark.asyncio
    async def test_save_upload_to_temp_no_extension(self, sample_file_content):
        """Test save_upload_to_temp with file that has no extension."""
        mock_file = AsyncMock()
        mock_file.filename = "test_file_no_extension"
        mock_file.read = AsyncMock(side_effect=[sample_file_content, b""])
        mock_file.close = AsyncMock()
        
        temp_path = await save_upload_to_temp(mock_file)
        
        try:
            # Verify file was created
            assert os.path.exists(temp_path)
            
            # Verify content
            with open(temp_path, "rb") as f:
                content = f.read()
            assert content == sample_file_content
            
        finally:
            if os.path.exists(temp_path):
                os.unlink(temp_path)

    @pytest.mark.asyncio
    async def test_save_upload_to_temp_large_file(self):
        """Test save_upload_to_temp with large file in chunks."""
        chunk_size = 1024 * 64  # Same as in the actual function
        large_content = b"A" * (chunk_size * 3 + 100)  # 3+ chunks
        
        mock_file = AsyncMock()
        mock_file.filename = "large_file.bin"
        
        # Split content into chunks to simulate real file reading
        chunks = []
        for i in range(0, len(large_content), chunk_size):
            chunks.append(large_content[i:i + chunk_size])
        chunks.append(b"")  # End of file marker
        
        mock_file.read = AsyncMock(side_effect=chunks)
        mock_file.close = AsyncMock()
        
        temp_path = await save_upload_to_temp(mock_file)
        
        try:
            # Verify file was created
            assert os.path.exists(temp_path)
            
            # Verify content
            with open(temp_path, "rb") as f:
                content = f.read()
            assert content == large_content
            
            # Verify read was called the right number of times
            assert mock_file.read.call_count == len(chunks)
            
        finally:
            if os.path.exists(temp_path):
                os.unlink(temp_path)

    @pytest.mark.asyncio
    async def test_save_upload_to_temp_empty_file(self):
        """Test save_upload_to_temp with empty file."""
        mock_file = AsyncMock()
        mock_file.filename = "empty.txt"
        mock_file.read = AsyncMock(side_effect=[b""])
        mock_file.close = AsyncMock()
        
        temp_path = await save_upload_to_temp(mock_file)
        
        try:
            # Verify file was created
            assert os.path.exists(temp_path)
            
            # Verify file is empty
            with open(temp_path, "rb") as f:
                content = f.read()
            assert content == b""
            
        finally:
            if os.path.exists(temp_path):
                os.unlink(temp_path)

    @pytest.mark.asyncio
    async def test_save_upload_to_temp_file_extension_preserved(self):
        """Test that file extension is preserved in temp file."""
        extensions = [".pdf", ".docx", ".jpg", ".zip", ".custom"]
        
        for ext in extensions:
            mock_file = AsyncMock()
            mock_file.filename = f"test{ext}"
            mock_file.read = AsyncMock(side_effect=[b"test content", b""])
            mock_file.close = AsyncMock()
            
            temp_path = await save_upload_to_temp(mock_file)
            
            try:
                assert temp_path.endswith(ext), f"Extension {ext} not preserved"
                assert os.path.exists(temp_path)
            finally:
                if os.path.exists(temp_path):
                    os.unlink(temp_path)