import pytest
from unittest.mock import Mock, patch
from app.smb.connection import check_smb_health


@pytest.mark.unit
class TestCheckSMBHealth:
    """Test cases for the check_smb_health function."""

    def test_check_smb_health_success(self):
        """Test successful SMB health check."""
        with patch('app.smb.connection.get_conn') as mock_get_conn:
            mock_conn = Mock()
            mock_conn.listPath.return_value = []  # Successful listPath call
            mock_conn.close.return_value = None
            mock_get_conn.return_value = mock_conn
            
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
        )
        mock_conn.listPath.assert_called_once_with("testshare", "/")
        mock_conn.close.assert_called_once()

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
        assert result["error"] == "Could not connect to SMB server"

    def test_check_smb_health_share_access_failure(self):
        """Test SMB health check with share access failure."""
        with patch('app.smb.connection.get_conn') as mock_get_conn:
            mock_conn = Mock()
            mock_conn.listPath.side_effect = Exception("Access denied")
            mock_conn.close.return_value = None
            mock_get_conn.return_value = mock_conn
            
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
        
        # Connection should still be closed even on error
        mock_conn.close.assert_called_once()

    def test_check_smb_health_close_exception_ignored(self):
        """Test that exceptions during connection close are ignored."""
        with patch('app.smb.connection.get_conn') as mock_get_conn:
            mock_conn = Mock()
            mock_conn.listPath.return_value = []
            mock_conn.close.side_effect = Exception("Close failed")
            mock_get_conn.return_value = mock_conn
            
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
        with patch('app.smb.connection.get_conn') as mock_get_conn:
            mock_conn = Mock()
            mock_conn.listPath.return_value = []
            mock_conn.close.return_value = None
            mock_get_conn.return_value = mock_conn
            
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