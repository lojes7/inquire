import argparse
import concurrent.futures
import json
import os
import random
import time
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Tuple

import requests


DEFAULT_BASE_URL = "http://localhost:8080/api"


def _now_iso() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%S", time.localtime())


def _ensure_out_dir(base_dir: str) -> str:
    out_dir = os.path.join(base_dir, "out")
    os.makedirs(out_dir, exist_ok=True)
    return out_dir


def _is_http_ok_for_json(res: requests.Response) -> bool:
    # 后端成功响应永远 http 200，失败才用 4xx/5xx
    return res.status_code in (200, 400, 401, 403, 409, 500)


def _parse_api_json(res: requests.Response) -> Dict[str, Any]:
    if not _is_http_ok_for_json(res):
        raise RuntimeError(f"Unexpected HTTP status: {res.status_code}: {res.text[:300]}")
    try:
        payload = res.json()
    except Exception as e:
        raise RuntimeError(f"Response is not JSON: {e}; body={res.text[:300]}")
    if not isinstance(payload, dict) or "code" not in payload:
        raise RuntimeError(f"Unexpected response shape: {payload}")
    return payload


def api_call(
    session: requests.Session,
    base_url: str,
    method: str,
    path: str,
    token: Optional[str] = None,
    json_body: Optional[Dict[str, Any]] = None,
    expected_code: Optional[int] = None,
    timeout: float = 8.0,
) -> Any:
    url = f"{base_url}{path}"
    headers: Dict[str, str] = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"

    res = session.request(method=method, url=url, headers=headers, json=json_body, timeout=timeout)
    payload = _parse_api_json(res)

    code = payload.get("code")
    message = payload.get("message", "")
    if res.status_code != 200:
        # 失败：后端 Fail() 会用 http=code，并且 json 里也有 code/message
        raise RuntimeError(f"HTTP {res.status_code} (code={code}): {message}")

    if expected_code is not None and code != expected_code:
        raise RuntimeError(f"Unexpected code: got {code}, expected {expected_code}; message={message}")

    # 成功：统一返回 data
    return payload.get("data")


def gen_phone(seed: int, index: int) -> str:
    # 生成可重复、11位、看起来像手机号的字符串
    suffix = (seed * 1000 + index) % (10**10)
    return "1" + f"{suffix:010d}"  # 1 + 10 位数字 = 11 位


@dataclass
class TestUser:
    name: str
    phone_number: str
    password: str
    uid: Optional[str] = None
    access_token: Optional[str] = None
    refresh_token: Optional[str] = None


def assert_login_shape(login_data: Any) -> Tuple[str, str, str]:
    """校验登录返回结构，并提取 (uid, access_token, refresh_token)。"""
    if not isinstance(login_data, dict):
        raise RuntimeError(f"Login data is not object: {login_data}")

    user_info = login_data.get("user_info")
    token_class = login_data.get("token_class")
    if not isinstance(user_info, dict) or not isinstance(token_class, dict):
        raise RuntimeError(f"Login missing user_info/token_class: {login_data}")

    uid = user_info.get("uid")
    access_token = token_class.get("token")
    refresh_token = token_class.get("refresh_token")
    if not uid or not access_token or not refresh_token:
        raise RuntimeError(f"Login missing uid/token/refresh_token: {login_data}")

    return str(uid), str(access_token), str(refresh_token)

def register_user(session: requests.Session, base_url: str, user: TestUser) -> None:
    payload = {"name": user.name, "password": user.password, "phone_number": user.phone_number}
    try:
        api_call(session, base_url, "POST", "/register", json_body=payload, expected_code=201)
    except Exception as e:
        # 支持重复运行：如果手机号已存在，继续
        msg = str(e)
        if "手机号已存在" in msg:
            return
        raise

def login_user(session: requests.Session, base_url: str, user: TestUser) -> None:
    payload = {"phone_number": user.phone_number, "password": user.password}
    login_data = api_call(session, base_url, "POST", "/login/phone_number", json_body=payload, expected_code=200)
    uid, access_token, refresh_token = assert_login_shape(login_data)
    user.uid = uid
    user.access_token = access_token
    user.refresh_token = refresh_token

def refresh_access_token(session: requests.Session, base_url: str, refresh_token: str) -> str:
    # refresh_token 专用中间件：必须是 claims.Type == "refresh"
    data = api_call(session, base_url, "POST", "/auth/refresh_token", token=refresh_token, expected_code=201)
    if not isinstance(data, dict) or not data.get("token"):
        raise RuntimeError(f"Unexpected refresh response data: {data}")
    return str(data["token"])


def lookup_stranger_id_by_uid(session: requests.Session, base_url: str, access_token: str, uid: str) -> str:
    data = api_call(session, base_url, "GET", f"/auth/info/strangers/uid/{uid}", token=access_token, expected_code=200)
    if not isinstance(data, dict) or not data.get("id"):
        raise RuntimeError(f"Unexpected stranger info response: {data}")
    return str(data["id"])


def send_friend_request(
    session: requests.Session,
    base_url: str,
    access_token: str,
    receiver_id: str,
    sender_name: str,
    message: str,
) -> None:
    payload = {
        "receiver_id": str(receiver_id),
        "sender_name": sender_name,
        "verification_message": message,
    }
    try:
        api_call(session, base_url, "POST", "/auth/friendship_requests", token=access_token, json_body=payload, expected_code=201)
    except Exception as e:
        # 重复发送会走 409
        if "请勿重复发送" in str(e) or "409" in str(e):
            return
        raise


def list_friend_requests(session: requests.Session, base_url: str, access_token: str) -> List[Dict[str, Any]]:
    data = api_call(session, base_url, "GET", "/auth/friendship_requests", token=access_token, expected_code=200)
    if data is None:
        return []
    if not isinstance(data, list):
        raise RuntimeError(f"Unexpected friend_requests response: {data}")
    return data


def accept_friend_request(session: requests.Session, base_url: str, access_token: str, request_id: str) -> None:
    api_call(session, base_url, "POST", f"/auth/friendship_requests/{request_id}", token=access_token, expected_code=201)


def list_friends(session: requests.Session, base_url: str, access_token: str) -> List[Dict[str, Any]]:
    data = api_call(session, base_url, "GET", "/auth/friendships", token=access_token, expected_code=200)
    if data is None:
        return []
    if not isinstance(data, list):
        raise RuntimeError(f"Unexpected friendships response: {data}")
    return data


def run_pair_flow(
    session: requests.Session,
    base_url: str,
    sender: TestUser,
    receiver: TestUser,
    message: str,
) -> Dict[str, Any]:
    if not sender.access_token or not receiver.access_token or not sender.uid or not receiver.uid:
        raise RuntimeError("Users must be logged in")

    # 1) sender 查 receiver 的 id（通过 uid）
    receiver_id = lookup_stranger_id_by_uid(session, base_url, sender.access_token, receiver.uid)
    # 2) sender 发送好友申请
    send_friend_request(session, base_url, sender.access_token, receiver_id, sender.name, message)

    # 3) receiver 拉申请列表，找到对应请求
    reqs = list_friend_requests(session, base_url, receiver.access_token)
    target_req_id: Optional[str] = None
    for r in reqs:
        if not isinstance(r, dict):
            continue
        # 后端字段：request_id, sender_name, sender_id, status...
        if str(r.get("sender_name", "")) == sender.name and str(r.get("status", "")) != "accepted":
            target_req_id = str(r.get("request_id"))
            break

    accepted = False
    if target_req_id:
        accept_friend_request(session, base_url, receiver.access_token, target_req_id)
        accepted = True

    # 4) 校验好友列表：receiver 是否包含 sender
    sender_id = lookup_stranger_id_by_uid(session, base_url, receiver.access_token, sender.uid)
    friends = list_friends(session, base_url, receiver.access_token)
    is_friend = any(str(f.get("friend_id")) == str(sender_id) for f in friends if isinstance(f, dict))

    return {
        "sender": {"name": sender.name, "phone": sender.phone_number, "uid": sender.uid},
        "receiver": {"name": receiver.name, "phone": receiver.phone_number, "uid": receiver.uid},
        "receiver_id": receiver_id,
        "sender_id": sender_id,
        "request_found": target_req_id is not None,
        "accepted": accepted,
        "friendship_verified": is_friend,
        "friendship_count_receiver": len(friends),
    }


def main():
    parser = argparse.ArgumentParser(description="vvechat API 自动化测试 (可重复/可扩展)")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--users", type=int, default=2, help="创建并登录的用户数量 (>=2)")
    parser.add_argument("--password", default="password123")
    parser.add_argument("--seed", type=int, default=42, help="影响手机号生成，保证可重复")
    parser.add_argument("--concurrency", type=int, default=4, help="注册/登录并发度")
    parser.add_argument("--no-refresh", action="store_true", help="跳过 refresh_token 测试")
    parser.add_argument("--pairing", choices=["chain", "single"], default="chain", help="users>2 时：链式互加 or 只测前2人")
    parser.add_argument("--timeout", type=float, default=8.0)
    args = parser.parse_args()

    if args.users < 2:
        raise SystemExit("--users 必须 >= 2")

    random.seed(args.seed)

    base_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    out_dir = _ensure_out_dir(base_dir)

    session = requests.Session()

    users: List[TestUser] = []
    for i in range(args.users):
        users.append(
            TestUser(
                name=f"TestUser{i+1}",
                phone_number=gen_phone(args.seed, i + 1),
                password=args.password,
            )
        )

    report: Dict[str, Any] = {
        "timestamp": _now_iso(),
        "base_url": args.base_url,
        "config": {
            "users": args.users,
            "seed": args.seed,
            "concurrency": args.concurrency,
            "pairing": args.pairing,
            "timeout": args.timeout,
            "refresh_enabled": not args.no_refresh,
        },
        "steps": [],
        "pairs": [],
        "ok": True,
        "errors": [],
    }

    print("=" * 70)
    print("Starting vvechat API tests")
    print(f"Base URL: {args.base_url}")
    print(f"Users: {args.users} | Seed: {args.seed} | Pairing: {args.pairing}")
    print("=" * 70)

    # 1) Register (并发)
    try:
        with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrency) as ex:
            futs = [ex.submit(register_user, session, args.base_url, u) for u in users]
            for f in concurrent.futures.as_completed(futs):
                f.result()
        report["steps"].append({"name": "register", "ok": True})
        print("[OK] register")
    except Exception as e:
        report["steps"].append({"name": "register", "ok": False, "error": str(e)})
        report["ok"] = False
        report["errors"].append(f"register: {e}")
        print(f"[FAIL] register: {e}")
        # 注册失败通常会影响后续，直接落盘
        out_file = os.path.join(out_dir, "api_test_results.json")
        with open(out_file, "w", encoding="utf-8") as fp:
            json.dump(report, fp, ensure_ascii=False, indent=2)
        raise

    # 2) Login (并发)
    try:
        with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrency) as ex:
            futs = [ex.submit(login_user, session, args.base_url, u) for u in users]
            for f in concurrent.futures.as_completed(futs):
                f.result()
        report["steps"].append({"name": "login", "ok": True})
        print("[OK] login")
    except Exception as e:
        report["steps"].append({"name": "login", "ok": False, "error": str(e)})
        report["ok"] = False
        report["errors"].append(f"login: {e}")
        print(f"[FAIL] login: {e}")
        out_file = os.path.join(out_dir, "api_test_results.json")
        with open(out_file, "w", encoding="utf-8") as fp:
            json.dump(report, fp, ensure_ascii=False, indent=2)
        raise

    # 3) Refresh token (只测第一个用户)
    if not args.no_refresh:
        try:
            assert users[0].refresh_token
            new_access = refresh_access_token(session, args.base_url, users[0].refresh_token)
            report["steps"].append({"name": "refresh_token", "ok": True})
            print("[OK] refresh_token")
            # 不强制替换主 token，仅记录验证通过
            report["refresh_sample"] = {"user": users[0].name, "new_access_token_prefix": new_access[:16]}
        except Exception as e:
            report["steps"].append({"name": "refresh_token", "ok": False, "error": str(e)})
            report["ok"] = False
            report["errors"].append(f"refresh_token: {e}")
            print(f"[FAIL] refresh_token: {e}")

    # 4) Friend request flows
    pairs: List[Tuple[TestUser, TestUser]] = []
    if args.pairing == "single" or args.users == 2:
        pairs = [(users[0], users[1])]
    else:
        # chain: 0->1, 1->2, ...
        pairs = [(users[i], users[i + 1]) for i in range(args.users - 1)]

    for sender, receiver in pairs:
        try:
            result = run_pair_flow(session, args.base_url, sender, receiver, message="API自动化测试好友申请")
            report["pairs"].append({"ok": True, **result})
            print(f"[OK] friend_flow {sender.name} -> {receiver.name} | verified={result['friendship_verified']}")
            if not result["friendship_verified"]:
                report["ok"] = False
                report["errors"].append(f"friendship_not_verified: {sender.name}->{receiver.name}")
        except Exception as e:
            report["pairs"].append({"ok": False, "sender": sender.name, "receiver": receiver.name, "error": str(e)})
            report["ok"] = False
            report["errors"].append(f"friend_flow {sender.name}->{receiver.name}: {e}")
            print(f"[FAIL] friend_flow {sender.name} -> {receiver.name}: {e}")

    out_file = os.path.join(out_dir, "api_test_results.json")
    with open(out_file, "w", encoding="utf-8") as fp:
        json.dump(report, fp, ensure_ascii=False, indent=2)

    print("=" * 70)
    print(f"DONE. ok={report['ok']} | report={out_file}")
    if report["errors"]:
        print("Errors:")
        for err in report["errors"]:
            print(f"  - {err}")

if __name__ == "__main__":
    main()
