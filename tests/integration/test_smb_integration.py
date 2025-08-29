import pytest
import os
import subprocess
import time
import tempfile
import io
from pathlib import Path
from fastapi.testclient import TestClient
from app.main import app, smb_upload_file


@pytest.mark.integration
@pytest.mark.slow
class TestSMBIntegration:
    """Integration tests with real SMB server."""

    @pytest.fixture(scope="class", autouse=True)
    def smb_server(self):
        """Start SMB server using Docker Compose."""
        compose_file = Path(__file__).parent / "docker-compose.test.yml"
        
        # Check if Docker is available
        try:
            subprocess.run(["docker", "--version"], check=True, capture_output=True)
        except (subprocess.CalledProcessError, FileNotFoundError):
            pytest.skip("Docker not available - skipping integration tests")
        
        # Check if docker-compose is available
        try:
            subprocess.run(["docker-compose", "--version"], check=True, capture_output=True)
        except (subprocess.CalledProcessError, FileNotFoundError):
            try:
                subprocess.run(["docker", "compose", "version"], check=True, capture_output=True)
            except (subprocess.CalledProcessError, FileNotFoundError):
                pytest.skip("Docker Compose not available - skipping integration tests")
        
        # Start the SMB server
        try:
            # Try docker compose first, then docker-compose
            try:
                subprocess.run([
                    "docker", "compose", "-f", str(compose_file), "up", "-d"
                ], check=True, cwd=compose_file.parent)
                compose_cmd = ["docker", "compose"]
            except subprocess.CalledProcessError:
                subprocess.run([
                    "docker-compose", "-f", str(compose_file), "up", "-d"
                ], check=True, cwd=compose_file.parent)
                compose_cmd = ["docker-compose"]
        except subprocess.CalledProcessError as e:
            pytest.skip(f"Failed to start SMB server: {e}")
        
        # Wait for SMB server to be ready
        max_wait = 60  # seconds
        wait_interval = 2  # seconds
        start_time = time.time()
        
        while time.time() - start_time < max_wait:
            try:
                # Check if container is healthy
                result = subprocess.run([
                    "docker", "ps", "--filter", "name=samba", "--format", "{{.Status}}"
                ], capture_output=True, text=True, check=True)
                
                if "healthy" in result.stdout:
                    break
                elif "unhealthy" in result.stdout:
                    pytest.skip("SMB server failed health check")
                    
            except subprocess.CalledProcessError:
                pass
            
            time.sleep(wait_interval)
        else:
            # Cleanup and skip if server didn't start
            try:
                subprocess.run(compose_cmd + ["-f", str(compose_file), "down", "-v"], 
                             cwd=compose_file.parent, capture_output=True)
            except:
                pass
            pytest.skip("SMB server failed to start within timeout")
        
        yield
        
        # Cleanup
        try:
            subprocess.run(compose_cmd + ["-f", str(compose_file), "down", "-v"], 
                         cwd=compose_file.parent, capture_output=True)
        except:
            pass

    @pytest.fixture
    def smb_config(self):
        """SMB server configuration."""
        return {
            "server_name": "localhost",
            "server_ip": "127.0.0.1",
            "share_name": "testshare",
            "username": "testuser",
            "password": "testpass",
            "domain": "",
            "port": 445,
            "use_ntlm_v2": True
        }

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

    def test_smb_upload_file_integration(self, smb_server, smb_config, temp_file):
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
            connected = conn.connect(smb_config["server_ip"], smb_config["port"])
            assert connected, "Could not connect to SMB server"
            
            # File should exist
            attrs = conn.getAttributes(smb_config["share_name"], remote_path)
            assert attrs is not None
            
            # Clean up
            conn.deleteFiles(smb_config["share_name"], remote_path)
            
        finally:
            conn.close()

    def test_smb_upload_file_overwrite_protection(self, smb_server, smb_config, temp_file):
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
            connected = conn.connect(smb_config["server_ip"], smb_config["port"])
            if connected:
                try:
                    conn.deleteFiles(smb_config["share_name"], remote_path)
                except:
                    pass
        finally:
            conn.close()

    def test_smb_upload_nested_directories(self, smb_server, smb_config, temp_file):
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
            connected = conn.connect(smb_config["server_ip"], smb_config["port"])
            assert connected
            
            # File should exist
            attrs = conn.getAttributes(smb_config["share_name"], remote_path)
            assert attrs is not None
            
            # Directories should exist
            for path in ["level1", "level1/level2", "level1/level2/level3"]:
                dir_list = conn.listPath(smb_config["share_name"], path)
                assert len(dir_list) > 0  # Should contain at least . and .. 
            
            # Clean up
            try:
                conn.deleteFiles(smb_config["share_name"], remote_path)
            except:
                pass
            
        finally:
            conn.close()

    def test_endpoint_integration(self, smb_server, client, setup_env):
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
        
        # Verify file was uploaded correctly
        from smb.SMBConnection import SMBConnection
        conn = SMBConnection("testuser", "testpass", "test-client", "localhost")
        
        try:
            connected = conn.connect("127.0.0.1", 445)
            assert connected
            
            # Download and verify content
            with tempfile.NamedTemporaryFile() as tmp:
                conn.retrieveFile("testshare", remote_path, tmp)
                tmp.seek(0)
                downloaded_content = tmp.read()
                assert downloaded_content == test_content
            
            # Clean up
            try:
                conn.deleteFiles("testshare", remote_path)
            except:
                pass
            
        finally:
            conn.close()

    def test_endpoint_overwrite_conflict(self, smb_server, client, setup_env):
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
        conn = SMBConnection("testuser", "testpass", "test-client", "localhost")
        try:
            connected = conn.connect("127.0.0.1", 445)
            if connected:
                try:
                    conn.deleteFiles("testshare", remote_path)
                except:
                    pass
        finally:
            conn.close()

    def test_large_file_upload(self, smb_server, client, setup_env):
        """Test uploading a larger file."""
        # Create a 1MB test file
        large_content = b"A" * (1024 * 1024)
        test_file = io.BytesIO(large_content)
        remote_path = "large_file_test.bin"
        
        # Upload through endpoint
        response = client.post(
            "/upload",
            files={"file": ("large_test.bin", test_file, "application/octet-stream")},
            data={"remote_path": remote_path}
        )
        
        assert response.status_code == 200
        
        # Verify file size and partial content
        from smb.SMBConnection import SMBConnection
        conn = SMBConnection("testuser", "testpass", "test-client", "localhost")
        
        try:
            connected = conn.connect("127.0.0.1", 445)
            assert connected
            
            # Check file attributes
            attrs = conn.getAttributes("testshare", remote_path)
            assert attrs.file_size == len(large_content)
            
            # Clean up
            try:
                conn.deleteFiles("testshare", remote_path)
            except:
                pass
            
        finally:
            conn.close()