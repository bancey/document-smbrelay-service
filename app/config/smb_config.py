"""SMB configuration management for the SMB Relay Service."""

import os


def load_smb_config_from_env():
    """Load SMB configuration from environment variables and return a tuple
    (config_dict, missing_list).

    config_dict contains: server_name, server_ip, share_name, username,
    password, domain, port, use_ntlm_v2, auth_protocol
    
    Returns:
        tuple: (config_dict, missing_list) where config_dict contains all 
               configuration values and missing_list contains names of 
               required environment variables that are missing
    """
    server_name = os.environ.get("SMB_SERVER_NAME")
    server_ip = os.environ.get("SMB_SERVER_IP")
    share_name = os.environ.get("SMB_SHARE_NAME")
    username = os.environ.get("SMB_USERNAME")
    password = os.environ.get("SMB_PASSWORD")
    domain = os.environ.get("SMB_DOMAIN", "")
    port = int(os.environ.get("SMB_PORT", "445"))
    use_ntlm_v2 = os.environ.get("SMB_USE_NTLM_V2", "true").lower() in (
        "1",
        "true",
        "yes",
    )
    
    # Auth protocol: negotiate (default), ntlm, or kerberos
    auth_protocol = os.environ.get("SMB_AUTH_PROTOCOL", "").lower()
    if auth_protocol not in ("negotiate", "ntlm", "kerberos", ""):
        auth_protocol = ""
    
    # If not explicitly set, derive from use_ntlm_v2 for backward compatibility
    if not auth_protocol:
        auth_protocol = "ntlm" if use_ntlm_v2 else "negotiate"

    # For Kerberos, username and password are optional (can use cached credentials)
    # For other auth methods, username and password are required
    required_fields = [
        ("SMB_SERVER_NAME", server_name),
        ("SMB_SERVER_IP", server_ip),
        ("SMB_SHARE_NAME", share_name),
    ]
    
    if auth_protocol != "kerberos":
        required_fields.extend([
            ("SMB_USERNAME", username),
            ("SMB_PASSWORD", password),
        ])
    
    missing = [k for k, v in required_fields if not v]

    config = {
        "server_name": server_name,
        "server_ip": server_ip,
        "share_name": share_name,
        "username": username,
        "password": password,
        "domain": domain,
        "port": port,
        "use_ntlm_v2": use_ntlm_v2,
        "auth_protocol": auth_protocol,
    }

    return config, missing