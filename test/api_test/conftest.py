import os
import time
import random
from dataclasses import dataclass
from typing import Any, Dict, Optional, Tuple

import pytest
import requests


DEFAULT_BASE_URL = "http://localhost:8080/api"


@dataclass(frozen=True)
class ApiConfig:
    base_url: str
    timeout: float


class ApiResponseError(RuntimeError):
    def __init__(self, *, http_status: int, code: Optional[int], message: str, payload: Any):
        super().__init__(f"HTTP {http_status} code={code} message={message}")
        self.http_status = http_status
        self.code = code
        self.message = message
        self.payload = payload


@pytest.fixture(scope="session")
def api_config() -> ApiConfig:
    base_url = os.getenv("VVECHAT_BASE_URL", DEFAULT_BASE_URL).rstrip("/")
    timeout = float(os.getenv("VVECHAT_TIMEOUT", "8"))
    return ApiConfig(base_url=base_url, timeout=timeout)


@pytest.fixture()
def session() -> requests.Session:
    s = requests.Session()
    yield s
    s.close()


def _api_json(
    session: requests.Session,
    cfg: ApiConfig,
    method: str,
    path: str,
    *,
    token: Optional[str] = None,
    json_body: Optional[Dict[str, Any]] = None,
) -> Tuple[int, Dict[str, Any]]:
    url = f"{cfg.base_url}{path}"
    headers: Dict[str, str] = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"

    try:
        res = session.request(method=method, url=url, headers=headers, json=json_body, timeout=cfg.timeout)
    except requests.RequestException as e:
        raise RuntimeError(
            f"无法连接后端：{e}. 请先启动 Go 服务，并确认 VVECHAT_BASE_URL={cfg.base_url}"
        )

    try:
        payload = res.json()
    except Exception as e:
        raise RuntimeError(f"响应不是 JSON: {e}; body={res.text[:300]}")

    if not isinstance(payload, dict) or "code" not in payload:
        raise RuntimeError(f"响应结构异常: http={res.status_code} payload={payload}")

    return res.status_code, payload


def api_ok(
    session: requests.Session,
    cfg: ApiConfig,
    method: str,
    path: str,
    *,
    token: Optional[str] = None,
    json_body: Optional[Dict[str, Any]] = None,
    expected_code: int,
) -> Any:
    http_status, payload = _api_json(session, cfg, method, path, token=token, json_body=json_body)
    if http_status != 200:
        raise ApiResponseError(
            http_status=http_status,
            code=payload.get("code"),
            message=str(payload.get("message", "")),
            payload=payload,
        )

    code = payload.get("code")
    if code != expected_code:
        raise ApiResponseError(
            http_status=http_status,
            code=code,
            message=str(payload.get("message", "")),
            payload=payload,
        )

    return payload.get("data")


def api_fail(
    session: requests.Session,
    cfg: ApiConfig,
    method: str,
    path: str,
    *,
    token: Optional[str] = None,
    json_body: Optional[Dict[str, Any]] = None,
    expected_http: int,
    expected_code: Optional[int] = None,
) -> Dict[str, Any]:
    http_status, payload = _api_json(session, cfg, method, path, token=token, json_body=json_body)
    assert http_status == expected_http, f"期望 HTTP {expected_http}, 实际 {http_status}, payload={payload}"
    if expected_code is not None:
        assert payload.get("code") == expected_code, f"期望 code={expected_code}, 实际 {payload.get('code')}, payload={payload}"
    return payload


def gen_unique_phone() -> str:
    # 11 位“手机号”字符串：尽量避免重复，保证可重复跑、也避免数据库脏数据冲突
    # 后 10 位 = (时间戳毫秒 + 随机扰动) mod 1e10
    now_ms = int(time.time() * 1000)
    jitter = random.randint(0, 9999)
    suffix = (now_ms + jitter) % (10**10)
    return "1" + f"{suffix:010d}"


@dataclass
class TestUser:
    name: str
    phone_number: str
    password: str
    uid: Optional[str] = None
    access_token: Optional[str] = None
    refresh_token: Optional[str] = None


def ensure_user_registered_and_logged_in(
    session: requests.Session,
    cfg: ApiConfig,
    *,
    name_prefix: str = "pytest",
    password: str = "password123",
    max_attempts: int = 5,
) -> TestUser:
    last_err: Optional[Exception] = None
    for attempt in range(1, max_attempts + 1):
        phone = gen_unique_phone()
        user = TestUser(name=f"{name_prefix}_{phone[-6:]}", phone_number=phone, password=password)

        http_status, payload = _api_json(
            session,
            cfg,
            "POST",
            "/register",
            json_body={"name": user.name, "password": user.password, "phone_number": user.phone_number},
        )

        if http_status == 200 and payload.get("code") == 201:
            # ok
            pass
        elif http_status == 400 and payload.get("message") == "手机号已存在":
            # 极小概率撞库，重试
            continue
        else:
            last_err = ApiResponseError(
                http_status=http_status,
                code=payload.get("code"),
                message=str(payload.get("message", "")),
                payload=payload,
            )
            break

        # login
        data = api_ok(
            session,
            cfg,
            "POST",
            "/login/phone_number",
            json_body={"phone_number": user.phone_number, "password": user.password},
            expected_code=200,
        )

        assert isinstance(data, dict)
        user_info = data.get("user_info")
        token_class = data.get("token_class")
        assert isinstance(user_info, dict) and isinstance(token_class, dict)
        user.uid = str(user_info.get("uid"))
        user.access_token = str(token_class.get("token"))
        user.refresh_token = str(token_class.get("refresh_token"))
        assert user.uid and user.access_token and user.refresh_token
        return user

    raise RuntimeError(f"创建测试用户失败（尝试 {max_attempts} 次）: {last_err}")
