import os
from typing import Dict, Any, Tuple, List


def _parse_bool_env(value: str, default: bool = True) -> bool:
    """Parse boolean-like environment variables.

    Treats '1', 'true', 'yes' (case-insensitive) as True. Anything else is False.
    """
    if value is None:
        return default
    return str(value).lower() in ("1", "true", "yes")


def load_smb_config_from_env() -> Tuple[Dict[str, Any], List[str]]:
    """Load SMB config from environment variables and return (config, missing_keys).

    Keeps the same config keys as the previous single-file implementation.
    """
    server_name = os.environ.get("SMB_SERVER_NAME")
    server_ip = os.environ.get("SMB_SERVER_IP")
    share_name = os.environ.get("SMB_SHARE_NAME")
    username = os.environ.get("SMB_USERNAME")
    password = os.environ.get("SMB_PASSWORD")
    domain = os.environ.get("SMB_DOMAIN", "")
    port = int(os.environ.get("SMB_PORT", "445"))
    use_ntlm_v2 = _parse_bool_env(os.environ.get("SMB_USE_NTLM_V2"), True)

    missing = [
        k
        for k, v in (
            ("SMB_SERVER_NAME", server_name),
            ("SMB_SERVER_IP", server_ip),
            ("SMB_SHARE_NAME", share_name),
            ("SMB_USERNAME", username),
            ("SMB_PASSWORD", password),
        )
        if not v
    ]

    config = {
        "server_name": server_name,
        "server_ip": server_ip,
        "share_name": share_name,
        "username": username,
        "password": password,
        "domain": domain,
        "port": port,
        "use_ntlm_v2": use_ntlm_v2,
    }

    return config, missing
