import pytest
import os
import tempfile
import io
import time
from pathlib import Path
from fastapi.testclient import TestClient
from app.main import app
from app.smb.operations import smb_upload_file
import smbclient


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
            
        max_wait = 30  # seconds in CI
        wait_interval = 1  # seconds
        start_time = time.time()
        
        # Build UNC path for connection test
        unc_path = f"//{smb_config['server_ip']}/{smb_config['share_name']}"
        
        while time.time() - start_time < max_wait:
            try:
                # Register session with smbclient
                smbclient.register_session(
                    server=smb_config["server_ip"],
                    username=smb_config["username"],
                    password=smb_config["password"],
                    port=smb_config["port"]
                )
                
                # Test connection by listing the share
                smbclient.listdir(unc_path)
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
        
        # Verify file was uploaded by checking if it exists
        unc_path = f"//{smb_config['server_ip']}/{smb_config['share_name']}"
        full_remote_path = f"{unc_path}/{remote_path}"
        
        # Register session for verification
        smbclient.register_session(
            server=smb_config["server_ip"],
            username=smb_config["username"],
            password=smb_config["password"],
            port=smb_config["port"]
        )
        
        try:
            # File should exist
            attrs = smbclient.stat(full_remote_path)
            assert attrs is not None
            
            # Clean up
            try:
                smbclient.remove(full_remote_path)
            except:
                pass  # Ignore cleanup errors
                
        except Exception as e:
            # Clean up on error too
            try:
                smbclient.remove(full_remote_path)
            except:
                pass
            raise e

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
        unc_path = f"//{smb_config['server_ip']}/{smb_config['share_name']}"
        full_remote_path = f"{unc_path}/{remote_path}"
        
        smbclient.register_session(
            server=smb_config["server_ip"],
            username=smb_config["username"],
            password=smb_config["password"],
            port=smb_config["port"]
        )
        
        try:
            smbclient.remove(full_remote_path)
        except:
            pass  # Ignore cleanup errors

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
        unc_path = f"//{smb_config['server_ip']}/{smb_config['share_name']}"
        full_remote_path = f"{unc_path}/{remote_path}"
        
        smbclient.register_session(
            server=smb_config["server_ip"],
            username=smb_config["username"],
            password=smb_config["password"],
            port=smb_config["port"]
        )
        
        try:
            # File should exist
            attrs = smbclient.stat(full_remote_path)
            assert attrs is not None
            
            # Clean up
            try:
                smbclient.remove(full_remote_path)
            except:
                pass  # Ignore cleanup errors
                
        except Exception as e:
            # Clean up on error too
            try:
                smbclient.remove(full_remote_path)
            except:
                pass
            raise e

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
        server_ip = os.environ["SMB_SERVER_IP"]
        port = int(os.environ["SMB_PORT"])
        share_name = os.environ["SMB_SHARE_NAME"]
        username = os.environ["SMB_USERNAME"]
        password = os.environ["SMB_PASSWORD"]
        
        unc_path = f"//{server_ip}/{share_name}"
        full_remote_path = f"{unc_path}/{remote_path}"
        
        smbclient.register_session(
            server=server_ip,
            username=username,
            password=password,
            port=port
        )
        
        try:
            # Check file exists
            attrs = smbclient.stat(full_remote_path)
            assert attrs is not None
            
            # Clean up
            try:
                smbclient.remove(full_remote_path)
            except:
                pass  # Ignore cleanup errors
                
        except Exception as e:
            # Clean up on error too
            try:
                smbclient.remove(full_remote_path)
            except:
                pass
            raise e

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
        server_ip = os.environ["SMB_SERVER_IP"]
        port = int(os.environ["SMB_PORT"])
        share_name = os.environ["SMB_SHARE_NAME"]
        username = os.environ["SMB_USERNAME"]
        password = os.environ["SMB_PASSWORD"]
        
        unc_path = f"//{server_ip}/{share_name}"
        full_remote_path = f"{unc_path}/{remote_path}"
        
        smbclient.register_session(
            server=server_ip,
            username=username,
            password=password,
            port=port
        )
        
        try:
            smbclient.remove(full_remote_path)
        except:
            pass  # Ignore cleanup errors

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
        server_ip = os.environ["SMB_SERVER_IP"]
        port = int(os.environ["SMB_PORT"])
        share_name = os.environ["SMB_SHARE_NAME"]
        username = os.environ["SMB_USERNAME"]
        password = os.environ["SMB_PASSWORD"]
        
        unc_path = f"//{server_ip}/{share_name}"
        full_remote_path = f"{unc_path}/{remote_path}"
        
        smbclient.register_session(
            server=server_ip,
            username=username,
            password=password,
            port=port
        )
        
        try:
            # Check file attributes
            attrs = smbclient.stat(full_remote_path)
            assert attrs.st_size == len(large_content)
            
            # Clean up
            try:
                smbclient.remove(full_remote_path)
            except:
                pass  # Ignore cleanup errors
                
        except Exception as e:
            # Clean up on error too
            try:
                smbclient.remove(full_remote_path)
            except:
                pass
            raise e