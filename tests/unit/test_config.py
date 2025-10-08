import pytest
import os
from app.config.smb_config import load_smb_config_from_env


@pytest.mark.unit
class TestSMBConfig:
    """Test cases for SMB configuration loading."""

    def test_load_config_with_defaults(self):
        """Test loading config with default values."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_USERNAME"] = "testuser"
        os.environ["SMB_PASSWORD"] = "testpass"
        
        # Clear optional vars
        for key in ["SMB_DOMAIN", "SMB_PORT", "SMB_USE_NTLM_V2", "SMB_AUTH_PROTOCOL"]:
            if key in os.environ:
                del os.environ[key]
        
        config, missing = load_smb_config_from_env()
        
        assert missing == []
        assert config["server_name"] == "testserver"
        assert config["server_ip"] == "127.0.0.1"
        assert config["share_name"] == "testshare"
        assert config["username"] == "testuser"
        assert config["password"] == "testpass"
        assert config["domain"] == ""
        assert config["port"] == 445
        assert config["use_ntlm_v2"] is True
        assert config["auth_protocol"] == "ntlm"  # Default from use_ntlm_v2
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_USERNAME", "SMB_PASSWORD"]:
            del os.environ[key]

    def test_load_config_with_auth_protocol_negotiate(self):
        """Test loading config with auth_protocol set to negotiate."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_USERNAME"] = "testuser"
        os.environ["SMB_PASSWORD"] = "testpass"
        os.environ["SMB_AUTH_PROTOCOL"] = "negotiate"
        
        config, missing = load_smb_config_from_env()
        
        assert missing == []
        assert config["auth_protocol"] == "negotiate"
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_USERNAME", "SMB_PASSWORD", "SMB_AUTH_PROTOCOL"]:
            del os.environ[key]

    def test_load_config_with_auth_protocol_ntlm(self):
        """Test loading config with auth_protocol set to ntlm."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_USERNAME"] = "testuser"
        os.environ["SMB_PASSWORD"] = "testpass"
        os.environ["SMB_AUTH_PROTOCOL"] = "ntlm"
        
        config, missing = load_smb_config_from_env()
        
        assert missing == []
        assert config["auth_protocol"] == "ntlm"
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_USERNAME", "SMB_PASSWORD", "SMB_AUTH_PROTOCOL"]:
            del os.environ[key]

    def test_load_config_with_auth_protocol_kerberos(self):
        """Test loading config with auth_protocol set to kerberos."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_AUTH_PROTOCOL"] = "kerberos"
        
        # Note: username and password are NOT set for Kerberos
        
        config, missing = load_smb_config_from_env()
        
        # For Kerberos, username and password are optional
        assert missing == []
        assert config["auth_protocol"] == "kerberos"
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_AUTH_PROTOCOL"]:
            if key in os.environ:
                del os.environ[key]

    def test_load_config_kerberos_with_credentials(self):
        """Test loading config for Kerberos with explicit credentials."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_USERNAME"] = "testuser"
        os.environ["SMB_PASSWORD"] = "testpass"
        os.environ["SMB_AUTH_PROTOCOL"] = "kerberos"
        
        config, missing = load_smb_config_from_env()
        
        assert missing == []
        assert config["auth_protocol"] == "kerberos"
        assert config["username"] == "testuser"
        assert config["password"] == "testpass"
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_USERNAME", "SMB_PASSWORD", "SMB_AUTH_PROTOCOL"]:
            del os.environ[key]

    def test_load_config_ntlm_requires_credentials(self):
        """Test that NTLM requires username and password."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_AUTH_PROTOCOL"] = "ntlm"
        
        # Note: username and password are NOT set
        
        config, missing = load_smb_config_from_env()
        
        # For NTLM, username and password are required
        assert "SMB_USERNAME" in missing
        assert "SMB_PASSWORD" in missing
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_AUTH_PROTOCOL"]:
            del os.environ[key]

    def test_load_config_invalid_auth_protocol(self):
        """Test that invalid auth_protocol falls back to default."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_USERNAME"] = "testuser"
        os.environ["SMB_PASSWORD"] = "testpass"
        os.environ["SMB_AUTH_PROTOCOL"] = "invalid_protocol"
        
        config, missing = load_smb_config_from_env()
        
        assert missing == []
        # Should fall back to default based on use_ntlm_v2
        assert config["auth_protocol"] == "ntlm"
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_USERNAME", "SMB_PASSWORD", "SMB_AUTH_PROTOCOL"]:
            del os.environ[key]

    def test_load_config_auth_protocol_case_insensitive(self):
        """Test that auth_protocol is case-insensitive."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_AUTH_PROTOCOL"] = "KERBEROS"
        
        config, missing = load_smb_config_from_env()
        
        assert missing == []
        assert config["auth_protocol"] == "kerberos"
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_AUTH_PROTOCOL"]:
            del os.environ[key]

    def test_load_config_backward_compatibility(self):
        """Test backward compatibility with use_ntlm_v2 setting."""
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_USERNAME"] = "testuser"
        os.environ["SMB_PASSWORD"] = "testpass"
        os.environ["SMB_USE_NTLM_V2"] = "false"
        
        # Clear SMB_AUTH_PROTOCOL to test backward compatibility
        if "SMB_AUTH_PROTOCOL" in os.environ:
            del os.environ["SMB_AUTH_PROTOCOL"]
        
        config, missing = load_smb_config_from_env()
        
        assert missing == []
        assert config["use_ntlm_v2"] is False
        assert config["auth_protocol"] == "negotiate"  # When use_ntlm_v2 is False
        
        # Cleanup
        for key in ["SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME", "SMB_USERNAME", "SMB_PASSWORD", "SMB_USE_NTLM_V2"]:
            del os.environ[key]
