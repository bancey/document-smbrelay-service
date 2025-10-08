import pytest
from unittest.mock import Mock, patch
from app.smb.connection import check_smb_health


@pytest.mark.unit
class TestCheckSMBHealth:
    """Test cases for the check_smb_health function."""

    def test_check_smb_health_success(self):
        """Test successful SMB health check."""
        with patch('app.smb.connection.get_conn') as mock_get_conn, \
             patch('app.smb.connection.smbclient.listdir') as mock_listdir:
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_listdir.return_value = []  # Successful listdir call
            
            result = check_smb_health(
                "testserver",
                "127.0.0.1", 
                "testshare",
                "testuser",
                "testpass"
            )
        
        assert result["status"] == "healthy"
        assert result["smb_connection"] == "ok"
        assert result["smb_share_accessible"] is True
        assert result["server"] == "testserver (127.0.0.1:445)"
        assert result["share"] == "testshare"
        assert "error" not in result
        
        # Verify connection was established and tested
        mock_get_conn.assert_called_once_with(
            "testuser",
            "testpass", 
            "testserver",
            "127.0.0.1",
            "",  # domain
            445,  # port
            True,  # use_ntlm_v2
            None,  # auth_protocol
        )
        mock_listdir.assert_called_once_with("//127.0.0.1/testshare/")

    def test_check_smb_health_connection_failure(self):
        """Test SMB health check with connection failure."""
        with patch('app.smb.connection.get_conn') as mock_get_conn:
            mock_get_conn.side_effect = ConnectionError("Could not connect to SMB server")
            
            result = check_smb_health(
                "testserver",
                "127.0.0.1",
                "testshare", 
                "testuser",
                "testpass"
            )
        
        assert result["status"] == "unhealthy"
        assert result["smb_connection"] == "failed"
        assert result["smb_share_accessible"] is False
        assert result["server"] == "testserver (127.0.0.1:445)"
        assert result["share"] == "testshare"
        assert "Could not connect to SMB server" in result["error"]

    def test_check_smb_health_share_access_failure(self):
        """Test SMB health check with share access failure."""
        with patch('app.smb.connection.get_conn') as mock_get_conn, \
             patch('app.smb.connection.smbclient.listdir') as mock_listdir:
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_listdir.side_effect = Exception("Access denied")
            
            result = check_smb_health(
                "testserver",
                "127.0.0.1",
                "testshare",
                "testuser",
                "testpass"
            )
        
        assert result["status"] == "unhealthy"
        assert result["smb_connection"] == "failed"
        assert result["smb_share_accessible"] is False
        assert result["server"] == "testserver (127.0.0.1:445)"
        assert result["share"] == "testshare"
        assert result["error"] == "Access denied"

    def test_check_smb_health_close_exception_ignored(self):
        """Test that connection close exceptions don't affect health check result."""
        # Note: In the new smbprotocol API, connections are managed automatically,
        # so this test verifies that the health check still works even if there are issues
        with patch('app.smb.connection.get_conn') as mock_get_conn, \
             patch('app.smb.connection.smbclient.listdir') as mock_listdir:
            
            mock_get_conn.return_value = {'server': '127.0.0.1', 'port': 445}
            mock_listdir.return_value = []  # Successful listdir call
            
            result = check_smb_health(
                "testserver",
                "127.0.0.1", 
                "testshare",
                "testuser",
                "testpass"
            )
        
        # Should still report success even if close fails
        assert result["status"] == "healthy"
        assert result["smb_connection"] == "ok"
        assert result["smb_share_accessible"] is True

    def test_check_smb_health_custom_parameters(self):
        """Test SMB health check with custom domain, port, and NTLM settings."""
        with patch('app.smb.connection.get_conn') as mock_get_conn, \
             patch('app.smb.connection.smbclient.listdir') as mock_listdir:
            
            mock_get_conn.return_value = {'server': '192.168.1.100', 'port': 139}
            mock_listdir.return_value = []
            
            result = check_smb_health(
                "myserver",
                "192.168.1.100",
                "documents",
                "domain_user", 
                "secret_pass",
                domain="MYDOMAIN",
                port=139,
                use_ntlm_v2=False
            )
        
        assert result["status"] == "healthy"
        assert result["server"] == "myserver (192.168.1.100:139)"
        assert result["share"] == "documents"
        
        # Verify custom parameters were passed correctly
        mock_get_conn.assert_called_once_with(
            "domain_user",
            "secret_pass",
            "myserver", 
            "192.168.1.100",
            "MYDOMAIN",  # domain
            139,         # port
            False,       # use_ntlm_v2
            None,        # auth_protocol
        )

    def test_check_smb_health_generic_error(self):
        """Test SMB health check with generic error."""
        with patch('app.smb.connection.get_conn') as mock_get_conn:
            mock_get_conn.side_effect = Exception("Unexpected error occurred")
            
            result = check_smb_health(
                "testserver",
                "127.0.0.1",
                "testshare",
                "testuser",
                "testpass"
            )
        
        assert result["status"] == "unhealthy"
        assert result["smb_connection"] == "failed"
        assert result["smb_share_accessible"] is False
        assert result["error"] == "Unexpected error occurred"