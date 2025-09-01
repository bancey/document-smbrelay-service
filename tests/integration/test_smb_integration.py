import pytest
import os
import tempfile
import io
import time
from pathlib import Path
from fastapi.testclient import TestClient
from app.main import app, smb_upload_file


@pytest.mark.integration
@pytest.mark.slow
class TestSMBIntegration:
    """Integration tests with real SMB server."""

    @pytest.fixture(scope="class")
    def smb_config(self):
        """SMB server configuration for CI and local environments."""
        # Check if running in CI with GitHub Actions services
        if os.getenv('CI_SMB_PORT'):
            return {
                "server_name": "localhost",
                "server_ip": os.getenv('CI_SMB_SERVER_IP', 'localhost'),
                "share_name": "testshare",
                "username": "testuser",
                "password": "testpass",
                "domain": "",
                "port": int(os.getenv('CI_SMB_PORT', 1445)),
                "use_ntlm_v2": True
            }
        else:
            # Local development - skip if no Docker available
            pytest.skip("Integration tests require CI environment or manual SMB server setup")

    @pytest.fixture
    def client(self):
        """Create test client."""
        return TestClient(app)

    @pytest.fixture
    def setup_env(self, smb_config):
        """Setup environment variables for tests."""
        env_vars = {
            "SMB_SERVER_NAME": smb_config["server_name"],
            "SMB_SERVER_IP": smb_config["server_ip"],
            "SMB_SHARE_NAME": smb_config["share_name"],
            "SMB_USERNAME": smb_config["username"],
            "SMB_PASSWORD": smb_config["password"],
            "SMB_DOMAIN": smb_config["domain"],
            "SMB_PORT": str(smb_config["port"]),
            "SMB_USE_NTLM_V2": "true"
        }
        
        # Store original values
        original_values = {}
        for key, value in env_vars.items():
            if key in os.environ:
                original_values[key] = os.environ[key]
            os.environ[key] = value
        
        yield
        
        # Restore original values
        for key in env_vars:
            if key in original_values:
                os.environ[key] = original_values[key]
            else:
                del os.environ[key]

    @pytest.fixture
    def wait_for_smb(self, smb_config):
        """Wait for SMB server to be ready in CI environment."""
        if not os.getenv('CI_SMB_PORT'):
            return
            
        from smb.SMBConnection import SMBConnection
        
        max_wait = 30  # seconds in CI
        wait_interval = 1  # seconds
        start_time = time.time()
        
        while time.time() - start_time < max_wait:
            try:
                conn = SMBConnection(
                    smb_config["username"],
                    smb_config["password"],
                    "test-client",
                    smb_config["server_name"],
                    domain=smb_config["domain"],
                    use_ntlm_v2=smb_config["use_ntlm_v2"]
                )
                
                if conn.connect(smb_config["server_ip"], smb_config["port"], timeout=5):
                    conn.close()
                    return  # SMB server is ready
                    
            except Exception:
                pass
            
            time.sleep(wait_interval)
        
        pytest.skip("SMB server not ready within timeout")

    def test_smb_upload_file_integration(self, smb_config, temp_file, wait_for_smb):
        """Test direct SMB upload function with real server."""
        remote_path = "integration_test.txt"
        
        # Upload file
        smb_upload_file(
            local_path=temp_file,
            remote_path=remote_path,
            **smb_config
        )
        
        # Verify file was uploaded by trying to retrieve its attributes
        from smb.SMBConnection import SMBConnection
        
        conn = SMBConnection(
            smb_config["username"],
            smb_config["password"],
            "test-client",
            smb_config["server_name"],
            domain=smb_config["domain"],
            use_ntlm_v2=smb_config["use_ntlm_v2"]
        )
        
        try:
            connected = conn.connect(smb_config["server_ip"], smb_config["port"], timeout=10)
            assert connected, "Could not connect to SMB server"
            
            # File should exist
            attrs = conn.getAttributes(smb_config["share_name"], remote_path)
            assert attrs is not None
            
            # Clean up
            try:
                conn.deleteFiles(smb_config["share_name"], remote_path)
            except:
                pass  # Ignore cleanup errors
            
        finally:
            conn.close()

    def test_smb_upload_file_overwrite_protection(self, smb_config, temp_file, wait_for_smb):
        """Test that overwrite protection works."""
        remote_path = "overwrite_test.txt"
        
        # Upload file first time
        smb_upload_file(
            local_path=temp_file,
            remote_path=remote_path,
            **smb_config
        )
        
        # Try to upload again without overwrite - should fail
        with pytest.raises(FileExistsError):
            smb_upload_file(
                local_path=temp_file,
                remote_path=remote_path,
                overwrite=False,
                **smb_config
            )
        
        # Upload with overwrite should succeed
        smb_upload_file(
            local_path=temp_file,
            remote_path=remote_path,
            overwrite=True,
            **smb_config
        )
        
        # Clean up
        from smb.SMBConnection import SMBConnection
        conn = SMBConnection(
            smb_config["username"],
            smb_config["password"],
            "test-client",
            smb_config["server_name"],
            domain=smb_config["domain"],
            use_ntlm_v2=smb_config["use_ntlm_v2"]
        )
        
        try:
            connected = conn.connect(smb_config["server_ip"], smb_config["port"], timeout=10)
            if connected:
                try:
                    conn.deleteFiles(smb_config["share_name"], remote_path)
                except:
                    pass  # Ignore cleanup errors
        finally:
            conn.close()

    def test_smb_upload_nested_directories(self, smb_config, temp_file, wait_for_smb):
        """Test uploading to nested directories."""
        remote_path = "level1/level2/level3/nested_test.txt"
        
        # Upload file to nested path
        smb_upload_file(
            local_path=temp_file,
            remote_path=remote_path,
            **smb_config
        )
        
        # Verify file exists
        from smb.SMBConnection import SMBConnection
        conn = SMBConnection(
            smb_config["username"],
            smb_config["password"],
            "test-client",
            smb_config["server_name"],
            domain=smb_config["domain"],
            use_ntlm_v2=smb_config["use_ntlm_v2"]
        )
        
        try:
            connected = conn.connect(smb_config["server_ip"], smb_config["port"], timeout=10)
            assert connected
            
            # File should exist
            attrs = conn.getAttributes(smb_config["share_name"], remote_path)
            assert attrs is not None
            
            # Clean up
            try:
                conn.deleteFiles(smb_config["share_name"], remote_path)
            except:
                pass  # Ignore cleanup errors
            
        finally:
            conn.close()

    def test_endpoint_integration(self, client, setup_env, wait_for_smb):
        """Test complete end-to-end upload through the FastAPI endpoint."""
        test_content = b"Integration test file content"
        test_file = io.BytesIO(test_content)
        remote_path = "endpoint_test.txt"
        
        # Upload through endpoint
        response = client.post(
            "/upload",
            files={"file": ("test.txt", test_file, "text/plain")},
            data={"remote_path": remote_path}
        )
        
        assert response.status_code == 200
        assert response.json() == {"status": "ok", "remote_path": remote_path}
        
        # Verify file was uploaded correctly by checking it exists
        from smb.SMBConnection import SMBConnection
        
        # Get config from environment
        server_ip = os.environ["SMB_SERVER_IP"]
        port = int(os.environ["SMB_PORT"])
        share_name = os.environ["SMB_SHARE_NAME"]
        username = os.environ["SMB_USERNAME"]
        password = os.environ["SMB_PASSWORD"]
        
        conn = SMBConnection(username, password, "test-client", "localhost")
        
        try:
            connected = conn.connect(server_ip, port, timeout=10)
            assert connected
            
            # Check file exists
            attrs = conn.getAttributes(share_name, remote_path)
            assert attrs is not None
            
            # Clean up
            try:
                conn.deleteFiles(share_name, remote_path)
            except:
                pass  # Ignore cleanup errors
            
        finally:
            conn.close()

    def test_endpoint_overwrite_conflict(self, client, setup_env, wait_for_smb):
        """Test endpoint overwrite conflict handling."""
        test_content = b"Conflict test content"
        remote_path = "conflict_test.txt"
        
        # Upload file first time
        test_file1 = io.BytesIO(test_content)
        response1 = client.post(
            "/upload",
            files={"file": ("test.txt", test_file1, "text/plain")},
            data={"remote_path": remote_path}
        )
        assert response1.status_code == 200
        
        # Try to upload again without overwrite - should fail with 409
        test_file2 = io.BytesIO(test_content)
        response2 = client.post(
            "/upload",
            files={"file": ("test.txt", test_file2, "text/plain")},
            data={"remote_path": remote_path, "overwrite": "false"}
        )
        assert response2.status_code == 409
        assert "already exists" in response2.json()["detail"]
        
        # Upload with overwrite should succeed
        test_file3 = io.BytesIO(test_content)
        response3 = client.post(
            "/upload",
            files={"file": ("test.txt", test_file3, "text/plain")},
            data={"remote_path": remote_path, "overwrite": "true"}
        )
        assert response3.status_code == 200
        
        # Clean up
        from smb.SMBConnection import SMBConnection
        
        # Get config from environment
        server_ip = os.environ["SMB_SERVER_IP"]
        port = int(os.environ["SMB_PORT"])
        share_name = os.environ["SMB_SHARE_NAME"]
        username = os.environ["SMB_USERNAME"]
        password = os.environ["SMB_PASSWORD"]
        
        conn = SMBConnection(username, password, "test-client", "localhost")
        try:
            connected = conn.connect(server_ip, port, timeout=10)
            if connected:
                try:
                    conn.deleteFiles(share_name, remote_path)
                except:
                    pass  # Ignore cleanup errors
        finally:
            conn.close()

    def test_large_file_upload(self, client, setup_env, wait_for_smb):
        """Test uploading a larger file."""
        # Create a smaller test file for CI (100KB instead of 1MB)
        large_content = b"A" * (100 * 1024)
        test_file = io.BytesIO(large_content)
        remote_path = "large_file_test.bin"
        
        # Upload through endpoint
        response = client.post(
            "/upload",
            files={"file": ("large_test.bin", test_file, "application/octet-stream")},
            data={"remote_path": remote_path}
        )
        
        assert response.status_code == 200
        
        # Verify file size
        from smb.SMBConnection import SMBConnection
        
        # Get config from environment
        server_ip = os.environ["SMB_SERVER_IP"]
        port = int(os.environ["SMB_PORT"])
        share_name = os.environ["SMB_SHARE_NAME"]
        username = os.environ["SMB_USERNAME"]
        password = os.environ["SMB_PASSWORD"]
        
        conn = SMBConnection(username, password, "test-client", "localhost")
        
        try:
            connected = conn.connect(server_ip, port, timeout=10)
            assert connected
            
            # Check file attributes
            attrs = conn.getAttributes(share_name, remote_path)
            assert attrs.file_size == len(large_content)
            
            # Clean up
            try:
                conn.deleteFiles(share_name, remote_path)
            except:
                pass  # Ignore cleanup errors
            
        finally:
            conn.close()