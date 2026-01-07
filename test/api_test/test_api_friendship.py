from .conftest import api_ok, ensure_user_registered_and_logged_in


def test_friend_request_accept_and_delete(session, api_config):
    sender = ensure_user_registered_and_logged_in(session, api_config, name_prefix="pytest_sender")
    receiver = ensure_user_registered_and_logged_in(session, api_config, name_prefix="pytest_receiver")

    # 1) sender 查 receiver 的 id（通过 uid）
    receiver_info = api_ok(
        session,
        api_config,
        "GET",
        f"/auth/info/strangers/uid/{receiver.uid}",
        token=sender.access_token,
        expected_code=200,
    )
    assert isinstance(receiver_info, dict) and receiver_info.get("id")
    receiver_id = str(receiver_info["id"])

    # 2) sender 发送好友申请（重复发送允许 409；这里用最简单方式：失败就不继续 assert）
    try:
        api_ok(
            session,
            api_config,
            "POST",
            "/auth/friendship_requests",
            token=sender.access_token,
            json_body={
                "sender_name": sender.name,
                "receiver_id": receiver_id,
                "verification_message": "pytest 好友申请",
            },
            expected_code=201,
        )
    except Exception:
        # 可能之前跑过同一对用户导致 409，但我们用的是随机用户，理论不会进这里；留一个容错兜底。
        pass

    # 3) receiver 拉申请列表，找到来自 sender 的请求
    reqs = api_ok(session, api_config, "GET", "/auth/friendship_requests", token=receiver.access_token, expected_code=200) or []
    assert isinstance(reqs, list)

    target = None
    for r in reqs:
        if isinstance(r, dict) and str(r.get("sender_name")) == sender.name:
            target = r
            break

    assert target is not None, f"未找到好友申请，reqs={reqs}"

    # 4) receiver 接受（如果已经 accepted，则跳过）
    if str(target.get("status")) != "accepted":
        api_ok(
            session,
            api_config,
            "POST",
            f"/auth/friendship_requests/{target['request_id']}",
            token=receiver.access_token,
            expected_code=201,
        )

    # 5) 通过 uid 获取 sender 的数据库 id，然后验证好友列表包含它
    sender_info = api_ok(
        session,
        api_config,
        "GET",
        f"/auth/info/strangers/uid/{sender.uid}",
        token=receiver.access_token,
        expected_code=200,
    )
    sender_id = str(sender_info["id"])

    friends = api_ok(session, api_config, "GET", "/auth/friendships", token=receiver.access_token, expected_code=200) or []
    assert isinstance(friends, list)
    assert any(str(f.get("friend_id")) == sender_id for f in friends if isinstance(f, dict)), friends

    # 6) 删除好友
    api_ok(session, api_config, "DELETE", f"/auth/friendships/{sender_id}", token=receiver.access_token, expected_code=201)

    # 7) 删除后不应再存在
    friends2 = api_ok(session, api_config, "GET", "/auth/friendships", token=receiver.access_token, expected_code=200) or []
    assert isinstance(friends2, list)
    assert not any(str(f.get("friend_id")) == sender_id for f in friends2 if isinstance(f, dict)), friends2

    # 8) 再删一次应返回“好友不存在”(HTTP 400)
    # 后端 Fail 用 http=400，json.code=400
    from .conftest import api_fail

    api_fail(session, api_config, "DELETE", f"/auth/friendships/{sender_id}", token=receiver.access_token, expected_http=400, expected_code=400)
