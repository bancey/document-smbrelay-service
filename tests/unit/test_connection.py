import pytest
from unittest.mock import Mock, patch, MagicMock
from app.smb.connection import get_conn


@pytest.mark.unit
class TestGetConn:
    """Test cases for the get_conn function."""

    def test_get_conn_with_server_ip(self):
        """Test get_conn uses server_ip when provided."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            result = get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="192.168.1.10",
                domain="",
                port=445,
                use_ntlm_v2=True,
                auth_protocol=None,
            )
        
        # Should use server_ip (192.168.1.10) not server_name
        mock_register.assert_called_once_with(
            server="192.168.1.10",
            username="testuser",
            password="testpass",
            port=445,
            auth_protocol='ntlm',
        )
        
        assert result['server'] == "192.168.1.10"
        assert result['port'] == 445
        assert result['session'] == mock_session
        assert result['server_name'] == "testserver"
        assert result['server_ip'] == "192.168.1.10"

    def test_get_conn_without_server_ip(self):
        """Test get_conn uses server_name when server_ip not provided."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            result = get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="",
                domain="",
                port=445,
                use_ntlm_v2=True,
                auth_protocol=None,
            )
        
        # Should use server_name when server_ip is empty
        mock_register.assert_called_once_with(
            server="testserver",
            username="testuser",
            password="testpass",
            port=445,
            auth_protocol='ntlm',
        )
        
        assert result['server'] == "testserver"

    def test_get_conn_auth_protocol_none_with_ntlm_v2_true(self):
        """Test auth_protocol defaults to 'ntlm' when use_ntlm_v2 is True."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="",
                port=445,
                use_ntlm_v2=True,
                auth_protocol=None,
            )
        
        # Should default to 'ntlm' when use_ntlm_v2 is True
        args, kwargs = mock_register.call_args
        assert kwargs['auth_protocol'] == 'ntlm'

    def test_get_conn_auth_protocol_none_with_ntlm_v2_false(self):
        """Test auth_protocol defaults to 'negotiate' when use_ntlm_v2 is False."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="",
                port=445,
                use_ntlm_v2=False,
                auth_protocol=None,
            )
        
        # Should default to 'negotiate' when use_ntlm_v2 is False
        args, kwargs = mock_register.call_args
        assert kwargs['auth_protocol'] == 'negotiate'

    def test_get_conn_explicit_auth_protocol(self):
        """Test explicit auth_protocol overrides use_ntlm_v2."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="",
                port=445,
                use_ntlm_v2=True,
                auth_protocol='kerberos',
            )
        
        # Should use explicit auth_protocol
        args, kwargs = mock_register.call_args
        assert kwargs['auth_protocol'] == 'kerberos'

    def test_get_conn_with_domain_and_username_ntlm(self):
        """Test username is formatted with domain for NTLM."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="TESTDOMAIN",
                port=445,
                use_ntlm_v2=True,
                auth_protocol='ntlm',
            )
        
        # Should format username as DOMAIN\username for NTLM
        mock_register.assert_called_once_with(
            server="127.0.0.1",
            username="TESTDOMAIN\\testuser",
            password="testpass",
            port=445,
            auth_protocol='ntlm',
        )

    def test_get_conn_with_domain_and_username_kerberos(self):
        """Test username is not formatted with domain for Kerberos."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="TESTDOMAIN",
                port=445,
                use_ntlm_v2=True,
                auth_protocol='kerberos',
            )
        
        # Should NOT format username with domain for Kerberos
        mock_register.assert_called_once_with(
            server="127.0.0.1",
            username="testuser",  # Not TESTDOMAIN\testuser
            password="testpass",
            port=445,
            auth_protocol='kerberos',
        )

    def test_get_conn_without_domain(self):
        """Test username is passed as-is when no domain provided."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="",
                port=445,
                use_ntlm_v2=True,
                auth_protocol='ntlm',
            )
        
        # Should pass username as-is without domain
        mock_register.assert_called_once_with(
            server="127.0.0.1",
            username="testuser",
            password="testpass",
            port=445,
            auth_protocol='ntlm',
        )

    def test_get_conn_without_username(self):
        """Test connection without username (for Kerberos)."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username=None,
                password=None,
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="",
                port=445,
                use_ntlm_v2=True,
                auth_protocol='kerberos',
            )
        
        # Should pass None as username for Kerberos
        mock_register.assert_called_once_with(
            server="127.0.0.1",
            username=None,
            password=None,
            port=445,
            auth_protocol='kerberos',
        )

    def test_get_conn_connection_error(self):
        """Test get_conn raises ConnectionError on failure."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_register.side_effect = Exception("Connection failed")
            
            with pytest.raises(ConnectionError, match="Could not connect to SMB server: Connection failed"):
                get_conn(
                    username="testuser",
                    password="testpass",
                    server_name="testserver",
                    server_ip="127.0.0.1",
                    domain="",
                    port=445,
                    use_ntlm_v2=True,
                    auth_protocol=None,
                )

    def test_get_conn_custom_port(self):
        """Test get_conn with custom port."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            result = get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="",
                port=139,
                use_ntlm_v2=True,
                auth_protocol=None,
            )
        
        # Should use custom port 139
        mock_register.assert_called_once_with(
            server="127.0.0.1",
            username="testuser",
            password="testpass",
            port=139,
            auth_protocol='ntlm',
        )
        
        assert result['port'] == 139

    def test_get_conn_negotiate_protocol(self):
        """Test get_conn with negotiate protocol."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username="testuser",
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="",
                port=445,
                use_ntlm_v2=True,
                auth_protocol='negotiate',
            )
        
        # Should use negotiate protocol
        args, kwargs = mock_register.call_args
        assert kwargs['auth_protocol'] == 'negotiate'

    def test_get_conn_with_domain_no_username(self):
        """Test connection with domain but no username (edge case)."""
        with patch('app.smb.connection.smbclient.register_session') as mock_register:
            mock_session = Mock()
            mock_register.return_value = mock_session
            
            get_conn(
                username=None,
                password="testpass",
                server_name="testserver",
                server_ip="127.0.0.1",
                domain="TESTDOMAIN",
                port=445,
                use_ntlm_v2=True,
                auth_protocol='kerberos',
            )
        
        # Should pass None as username (no domain formatting when username is None)
        mock_register.assert_called_once_with(
            server="127.0.0.1",
            username=None,
            password="testpass",
            port=445,
            auth_protocol='kerberos',
        )
