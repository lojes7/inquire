import pytest

from .conftest import api_fail, api_ok, ensure_user_registered_and_logged_in


def test_protected_endpoint_requires_auth(session, api_config):
    # 不带 token 调用受保护接口，应返回 401（而不是数据库错误）
    api_fail(session, api_config, "GET", "/auth/friendships", expected_http=401, expected_code=401)


def test_register_login_refresh_token(session, api_config):
    user = ensure_user_registered_and_logged_in(session, api_config, name_prefix="pytest_auth")

    # refresh_token 只能用于 /auth/refresh_token
    data = api_ok(
        session,
        api_config,
        "POST",
        "/auth/refresh_token",
        token=user.refresh_token,
        expected_code=201,
    )
    assert isinstance(data, dict)
    assert data.get("token"), data

    # refresh_token 不能当 access token 用
    api_fail(session, api_config, "GET", "/auth/friendships", token=user.refresh_token, expected_http=403, expected_code=403)
