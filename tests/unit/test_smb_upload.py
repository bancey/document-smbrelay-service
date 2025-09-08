import pytest
import os
from unittest.mock import Mock, patch, mock_open
from app.smb.operations import smb_upload_file


@pytest.mark.unit
class TestSMBUploadFile:
    """Test cases for smb_upload_file function."""

    def test_smb_upload_file_basic_success(self, temp_file):
        """Test successful SMB file upload."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.listPath.side_effect = Exception("Directory doesn't exist")
        mock_conn.createDirectory.return_value = None
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        mock_conn.storeFile.return_value = None
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="test/file.txt",
                username="testuser",
                password="testpass"
            )

        # Verify connection was attempted
        mock_conn.connect.assert_called_once_with("127.0.0.1", 445)
        
        # Verify directory creation was attempted for "test"
        mock_conn.createDirectory.assert_called_once_with("testshare", "test")
        
        # Verify file existence check
        mock_conn.getAttributes.assert_called_once_with("testshare", "test/file.txt")
        
        # Verify file upload
        mock_conn.storeFile.assert_called_once()
        args = mock_conn.storeFile.call_args[0]
        assert args[0] == "testshare"
        assert args[1] == "test/file.txt"
        
        # Verify connection closed
        mock_conn.close.assert_called_once()

    def test_smb_upload_file_connection_failure(self, temp_file):
        """Test SMB connection failure."""
        mock_conn = Mock()
        mock_conn.connect.return_value = False

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            with pytest.raises(ConnectionError, match="Could not connect to SMB server"):
                smb_upload_file(
                    local_path=temp_file,
                    server_name="testserver",
                    server_ip="127.0.0.1",
                    share_name="testshare",
                    remote_path="file.txt",
                    username="testuser",
                    password="testpass"
                )

    def test_smb_upload_file_exists_no_overwrite(self, temp_file):
        """Test upload when file exists and overwrite is False."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.getAttributes.return_value = Mock()  # File exists
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            with pytest.raises(FileExistsError, match="Remote file already exists: file.txt"):
                smb_upload_file(
                    local_path=temp_file,
                    server_name="testserver",
                    server_ip="127.0.0.1",
                    share_name="testshare",
                    remote_path="file.txt",
                    username="testuser",
                    password="testpass",
                    overwrite=False
                )

        # Verify connection was closed
        mock_conn.close.assert_called_once()

    def test_smb_upload_file_exists_with_overwrite(self, temp_file):
        """Test upload when file exists and overwrite is True."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.getAttributes.return_value = Mock()  # File exists
        mock_conn.storeFile.return_value = None
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="file.txt",
                username="testuser",
                password="testpass",
                overwrite=True
            )

        # Verify file was uploaded despite existing
        mock_conn.storeFile.assert_called_once()

    def test_smb_upload_file_nested_directory_creation(self, temp_file):
        """Test creation of nested directories."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.listPath.side_effect = Exception("Directory doesn't exist")
        mock_conn.createDirectory.return_value = None
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        mock_conn.storeFile.return_value = None
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="level1/level2/level3/file.txt",
                username="testuser",
                password="testpass"
            )

        # Verify all directory levels were attempted to be created
        expected_calls = [
            (("testshare", "level1"),),
            (("testshare", "level1/level2"),),
            (("testshare", "level1/level2/level3"),)
        ]
        
        assert mock_conn.createDirectory.call_count == 3
        actual_calls = mock_conn.createDirectory.call_args_list
        for i, expected_call in enumerate(expected_calls):
            assert actual_calls[i][0] == expected_call[0]

    def test_smb_upload_file_directory_already_exists(self, temp_file):
        """Test when directories already exist."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.listPath.return_value = []  # Directory exists
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        mock_conn.storeFile.return_value = None
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="existing_dir/file.txt",
                username="testuser",
                password="testpass"
            )

        # Directory creation should not be attempted
        mock_conn.createDirectory.assert_not_called()

    def test_smb_upload_file_no_directory(self, temp_file):
        """Test upload to root directory (no subdirectories)."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        mock_conn.storeFile.return_value = None
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="file.txt",
                username="testuser",
                password="testpass"
            )

        # No directory operations should occur
        mock_conn.listPath.assert_not_called()
        mock_conn.createDirectory.assert_not_called()

    def test_smb_upload_file_custom_parameters(self, temp_file):
        """Test upload with custom parameters."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        mock_conn.storeFile.return_value = None
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn) as mock_smb_class:
            smb_upload_file(
                local_path=temp_file,
                server_name="customserver",
                server_ip="192.168.1.100",
                share_name="customshare",
                remote_path="file.txt",
                username="customuser",
                password="custompass",
                domain="MYDOMAIN",
                port=139,
                use_ntlm_v2=False
            )

        # Verify SMBConnection was created with custom parameters
        mock_smb_class.assert_called_once_with(
            "customuser", "custompass", "fastapi-smb-relay", "customserver",
            domain="MYDOMAIN", use_ntlm_v2=False
        )
        
        # Verify connection was made to custom IP and port
        mock_conn.connect.assert_called_once_with("192.168.1.100", 139)

    def test_smb_upload_file_directory_creation_permission_error(self, temp_file):
        """Test handling of directory creation permission errors."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.listPath.side_effect = Exception("Directory doesn't exist")
        mock_conn.createDirectory.side_effect = Exception("Permission denied")
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        mock_conn.storeFile.return_value = None
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            # Should not raise exception even if directory creation fails
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="restricted/file.txt",
                username="testuser",
                password="testpass"
            )

        # File upload should still be attempted (might be called multiple times for directory creation)
        assert mock_conn.storeFile.call_count >= 1
        # Verify the actual file upload was called
        actual_file_calls = [call for call in mock_conn.storeFile.call_args_list 
                           if call[0][1] == "restricted/file.txt"]
        assert len(actual_file_calls) == 1

    def test_smb_upload_file_storefile_path_error(self, temp_file):
        """When storeFile raises a path-related error, a ConnectionError explains directory may not exist."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        # Simulate storeFile raising a path-related error
        mock_conn.storeFile.side_effect = Exception("[Errno 2] No such file or directory: 'path not found'")
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            with pytest.raises(ConnectionError) as exc:
                smb_upload_file(
                    local_path=temp_file,
                    server_name="testserver",
                    server_ip="127.0.0.1",
                    share_name="testshare",
                    remote_path="subdir/file.txt",
                    username="testuser",
                    password="testpass",
                )

        assert "Directory path may not exist" in str(exc.value)

    def test_smb_upload_file_storefile_generic_error(self, temp_file):
        """When storeFile raises a generic error, a ConnectionError with original error is raised."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        # Simulate storeFile raising a generic error
        mock_conn.storeFile.side_effect = Exception("Unexpected failure during write")
        mock_conn.close.return_value = None

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            with pytest.raises(ConnectionError) as exc:
                smb_upload_file(
                    local_path=temp_file,
                    server_name="testserver",
                    server_ip="127.0.0.1",
                    share_name="testshare",
                    remote_path="subdir/file.txt",
                    username="testuser",
                    password="testpass",
                )

        assert "Unexpected failure during write" in str(exc.value)

    def test_smb_upload_file_close_raises_does_not_propagate(self, temp_file):
        """If conn.close raises an exception it should be swallowed and not propagate."""
        mock_conn = Mock()
        mock_conn.connect.return_value = True
        mock_conn.getAttributes.side_effect = Exception("File doesn't exist")
        mock_conn.storeFile.return_value = None
        # Simulate close raising an exception
        mock_conn.close.side_effect = Exception("close failed")

        with patch('app.smb.connection.SMBConnection', return_value=mock_conn):
            # Should not raise despite close failing
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="file.txt",
                username="testuser",
                password="testpass",
            )

        # Ensure close was attempted even though it raised
        mock_conn.close.assert_called_once()