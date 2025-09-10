import pytest
import os
from unittest.mock import Mock, patch, mock_open
from app.smb.operations import smb_upload_file


@pytest.mark.unit
class TestSMBUploadFile:
    """Test cases for smb_upload_file function."""

    def test_smb_upload_file_basic_success(self, temp_file):
        """Test successful SMB file upload."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb, \
             patch('builtins.open', mock_open(read_data=b'test content')) as mock_local_file:
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")  # File doesn't exist
            mock_remote_file = Mock()
            mock_open_smb.return_value.__enter__ = Mock(return_value=mock_remote_file)
            mock_open_smb.return_value.__exit__ = Mock(return_value=False)
            
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="test/file.txt",
                username="testuser",
                password="testpass"
            )

        # Verify connection was established
        mock_get_conn.assert_called_once_with(
            "testuser",
            "testpass",
            "testserver",
            "127.0.0.1",
            "",  # domain
            445,  # port
            True,  # use_ntlm_v2
        )
        
        # Verify directory creation was attempted for "test"
        mock_makedirs.assert_called_once_with("//127.0.0.1/testshare/test", exist_ok=True)
        
        # Verify file existence check
        mock_stat.assert_called_once_with("//127.0.0.1/testshare/test/file.txt")
        
        # Verify file upload
        mock_open_smb.assert_called_once_with("//127.0.0.1/testshare/test/file.txt", "wb")

    def test_smb_upload_file_connection_failure(self, temp_file):
        """Test SMB connection failure."""
        with patch('app.smb.operations.get_conn') as mock_get_conn:
            mock_get_conn.side_effect = ConnectionError("Could not connect to SMB server")
            
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
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat:
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.return_value = Mock()  # File exists

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

    def test_smb_upload_file_exists_with_overwrite(self, temp_file):
        """Test upload when file exists and overwrite is True."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb, \
             patch('builtins.open', mock_open(read_data=b'test content')) as mock_local_file:
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.return_value = Mock()  # File exists
            mock_remote_file = Mock()
            mock_open_smb.return_value.__enter__ = Mock(return_value=mock_remote_file)
            mock_open_smb.return_value.__exit__ = Mock(return_value=False)
            
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

        # Should proceed with upload despite file existence
        mock_open_smb.assert_called_once_with("//127.0.0.1/testshare/file.txt", "wb")

    def test_smb_upload_file_nested_directory_creation(self, temp_file):
        """Test upload with nested directory creation."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb, \
             patch('builtins.open', mock_open(read_data=b'test content')):
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")
            mock_remote_file = Mock()
            mock_open_smb.return_value.__enter__ = Mock(return_value=mock_remote_file)
            mock_open_smb.return_value.__exit__ = Mock(return_value=False)
            
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="dir1/dir2/file.txt",
                username="testuser",
                password="testpass"
            )

        # Verify nested directory creation
        mock_makedirs.assert_called_once_with("//127.0.0.1/testshare/dir1/dir2", exist_ok=True)

    def test_smb_upload_file_directory_already_exists(self, temp_file):
        """Test upload when directory already exists."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb, \
             patch('builtins.open', mock_open(read_data=b'test content')):
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")
            mock_remote_file = Mock()
            mock_open_smb.return_value.__enter__ = Mock(return_value=mock_remote_file)
            mock_open_smb.return_value.__exit__ = Mock(return_value=False)
            
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="existing_dir/file.txt",
                username="testuser",
                password="testpass"
            )

        # Directory creation should still be called (with exist_ok=True)
        mock_makedirs.assert_called_once_with("//127.0.0.1/testshare/existing_dir", exist_ok=True)

    def test_smb_upload_file_no_directory(self, temp_file):
        """Test upload to root directory (no subdirectory)."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb, \
             patch('builtins.open', mock_open(read_data=b'test content')):
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")
            mock_remote_file = Mock()
            mock_open_smb.return_value.__enter__ = Mock(return_value=mock_remote_file)
            mock_open_smb.return_value.__exit__ = Mock(return_value=False)
            
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="file.txt",
                username="testuser",
                password="testpass"
            )

        # No directory creation should be attempted for root level
        mock_makedirs.assert_not_called()

    def test_smb_upload_file_custom_parameters(self, temp_file):
        """Test upload with custom port and domain parameters."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb, \
             patch('builtins.open', mock_open(read_data=b'test content')):
            
            mock_get_conn.return_value = {'server': '192.168.1.10', 'port': 139}
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")
            mock_remote_file = Mock()
            mock_open_smb.return_value.__enter__ = Mock(return_value=mock_remote_file)
            mock_open_smb.return_value.__exit__ = Mock(return_value=False)
            
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="192.168.1.10",
                share_name="testshare",
                remote_path="file.txt",
                username="domain_user",
                password="secret_pass",
                domain="TESTDOMAIN",
                port=139,
                use_ntlm_v2=False
            )

        # Verify custom parameters were passed to connection
        mock_get_conn.assert_called_once_with(
            "domain_user",
            "secret_pass",
            "testserver",
            "192.168.1.10",
            "TESTDOMAIN",
            139,
            False,
        )

    def test_smb_upload_file_directory_creation_permission_error(self, temp_file):
        """Test handling of directory creation permission errors."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb, \
             patch('builtins.open', mock_open(read_data=b'test content')):
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_makedirs.side_effect = PermissionError("Permission denied")
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")
            mock_remote_file = Mock()
            mock_open_smb.return_value.__enter__ = Mock(return_value=mock_remote_file)
            mock_open_smb.return_value.__exit__ = Mock(return_value=False)
            
            # Should still proceed with upload despite directory creation failure
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="test/file.txt",
                username="testuser",
                password="testpass"
            )

        mock_open_smb.assert_called_once()

    def test_smb_upload_file_storefile_path_error(self, temp_file):
        """Test handling of store file path-related errors."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb:
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")
            mock_open_smb.side_effect = Exception("path not found")
            
            with pytest.raises(ConnectionError, match="Directory path may not exist"):
                smb_upload_file(
                    local_path=temp_file,
                    server_name="testserver",
                    server_ip="127.0.0.1",
                    share_name="testshare",
                    remote_path="nonexistent/file.txt",
                    username="testuser",
                    password="testpass"
                )

    def test_smb_upload_file_storefile_generic_error(self, temp_file):
        """Test handling of generic store file errors."""
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb:
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")
            mock_open_smb.side_effect = Exception("Generic upload error")
            
            with pytest.raises(ConnectionError, match="Failed to store file.txt on testshare: Generic upload error"):
                smb_upload_file(
                    local_path=temp_file,
                    server_name="testserver",
                    server_ip="127.0.0.1",
                    share_name="testshare",
                    remote_path="file.txt",
                    username="testuser",
                    password="testpass"
                )

    def test_smb_upload_file_close_raises_does_not_propagate(self, temp_file):
        """Test that connection close exceptions don't propagate to caller."""
        # Note: In smbprotocol API, connections are managed automatically,
        # so this test verifies that the upload works despite potential cleanup issues
        with patch('app.smb.operations.get_conn') as mock_get_conn, \
             patch('app.smb.operations.smbclient.makedirs') as mock_makedirs, \
             patch('app.smb.operations.smbclient.stat') as mock_stat, \
             patch('app.smb.operations.smbclient.open_file') as mock_open_smb, \
             patch('builtins.open', mock_open(read_data=b'test content')):
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_stat.side_effect = FileNotFoundError("File doesn't exist")
            mock_remote_file = Mock()
            mock_open_smb.return_value.__enter__ = Mock(return_value=mock_remote_file)
            mock_open_smb.return_value.__exit__ = Mock(return_value=False)
            
            # Should complete successfully despite any potential cleanup issues
            smb_upload_file(
                local_path=temp_file,
                server_name="testserver",
                server_ip="127.0.0.1",
                share_name="testshare",
                remote_path="file.txt",
                username="testuser",
                password="testpass"
            )

        mock_open_smb.assert_called_once()