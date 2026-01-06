import requests
import json
import time
import random
import string
import concurrent.futures

# 配置
BASE_URL = "http://localhost:8080/api"
TEST_PHONE = "13800138000"
TEST_PASSWORD = "password123"
TEST_NAME = "TestUser"

# 用于加好友测试的第二个用户
TEST_PHONE_2 = "13900139000"
TEST_NAME_2 = "TestUser2"

def random_phone():
    """生成随机手机号"""
    return "1" + "".join(random.choices(string.digits, k=10))

def test_register(phone=TEST_PHONE, name=TEST_NAME, password=TEST_PASSWORD):
    print(f"\n--- Testing Register ({phone}) ---")
    url = f"{BASE_URL}/register"
    payload = {
        "name": name,
        "password": password,
        "phone_number": phone
    }
    try:
        response = requests.post(url, json=payload)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.text}")
        
        if response.status_code == 201:
            print("✅ Register Success")
            return True
        elif response.status_code == 400 and "手机号已存在" in response.text:
             print("⚠️ User already exists, continuing...")
             return True
        else:
            print("❌ Register Failed")
            return False
    except Exception as e:
        print(f"❌ Request Failed: {e}")
        return False

def test_login(phone=TEST_PHONE, password=TEST_PASSWORD):
    print(f"\n--- Testing Login ({phone}) ---")
    url = f"{BASE_URL}/login/phone_number"
    payload = {
        "phone_number": phone,
        "password": password
    }
    try:
        response = requests.post(url, json=payload)
        print(f"Status Code: {response.status_code}")
        # print(f"Response: {response.text}")
        
        if response.status_code == 200:
            data = response.json()
            if data['code'] == 200:
                print("✅ Login Success")
                return data['data'] # Returns tokens
            else:
                print(f"❌ Login Failed: {data['msg']}")
                return None
        else:
            print("❌ Login Failed")
            return None
    except Exception as e:
        print(f"❌ Request Failed: {e}")
        return None

def test_refresh_token(refresh_token):
    print("\n--- Testing Refresh Token ---")
    url = f"{BASE_URL}/auth/refresh_token"
    headers = {
        "Authorization": f"Bearer {refresh_token}"
    }
    
    success = False
    try:
        # 1. Test with valid refresh token
        print("1. Testing with VALID refresh token...")
        response = requests.post(url, headers=headers)
        print(f"Status Code: {response.status_code}")
        
        if response.status_code == 201:
            data = response.json()
            if data['code'] == 201:
                print("✅ Refresh Success")
                print(f"New Access Token: {data['data']['access_token'][:20]}...")
                success = True
            else:
                print(f"❌ Refresh Failed: {data['msg']}")
        else:
            print(f"❌ Refresh Failed: {response.text}")

        # 2. Test with invalid token (simulating access token as refresh token)
        print("2. Testing with INVALID token (tampered)...")
        headers["Authorization"] = f"Bearer {refresh_token}invalid"
        response = requests.post(url, headers=headers)
        if response.status_code == 401:
             print("✅ Correctly rejected invalid token")
        else:
             print(f"❌ Unexpected status for invalid token: {response.status_code}")
        
        return success

    except Exception as e:
        print(f"❌ Request Failed: {e}")
        return False


def test_add_friend(access_token, receiver_id, sender_name, message="测试好友申请"):
    """测试发送好友申请"""
    print(f"\n--- Testing Add Friend (to user {receiver_id}) ---")
    url = f"{BASE_URL}/auth/friendship_requests"
    headers = {"Authorization": f"Bearer {access_token}"}
    payload = {
        "receiver_id": str(receiver_id),
        "sender_name": sender_name,
        "verification_message": message
    }
    
    try:
        response = requests.post(url, json=payload, headers=headers)
        print(f"Status Code: {response.status_code}")
        
        if response.status_code in [200, 201]:
            print("✅ Friend Request Sent")
            return True
        elif response.status_code == 400:
            data = response.json()
            if "已发送" in data.get('message', '') or "already" in data.get('message', '').lower():
                print("⚠️ Friend request already sent")
                return True
            print(f"❌ Add Friend Failed: {data}")
            return False
        else:
            print(f"❌ Add Friend Failed: {response.text}")
            return False
    except Exception as e:
        print(f"❌ Request Failed: {e}")
        return False


def test_get_friend_requests(access_token):
    """测试获取好友申请列表"""
    print("\n--- Testing Get Friend Requests ---")
    url = f"{BASE_URL}/auth/friendship_requests"
    headers = {"Authorization": f"Bearer {access_token}"}
    
    try:
        response = requests.get(url, headers=headers)
        print(f"Status Code: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            requests_list = data.get('data', []) or []
            print(f"✅ Got {len(requests_list)} friend requests")
            return requests_list
        else:
            print(f"❌ Get Friend Requests Failed: {response.text}")
            return []
    except Exception as e:
        print(f"❌ Request Failed: {e}")
        return []


def test_accept_friend(access_token, request_id):
    """测试接受好友申请"""
    print(f"\n--- Testing Accept Friend Request ({request_id}) ---")
    url = f"{BASE_URL}/auth/friendship_requests/{request_id}"
    headers = {"Authorization": f"Bearer {access_token}"}
    
    try:
        response = requests.post(url, headers=headers)
        print(f"Status Code: {response.status_code}")
        
        if response.status_code in [200, 201]:
            print("✅ Friend Request Accepted")
            return True
        else:
            print(f"❌ Accept Friend Failed: {response.text}")
            return False
    except Exception as e:
        print(f"❌ Request Failed: {e}")
        return False


def test_get_friends(access_token):
    """测试获取好友列表"""
    print("\n--- Testing Get Friends ---")
    url = f"{BASE_URL}/auth/friendships"
    headers = {"Authorization": f"Bearer {access_token}"}
    
    try:
        response = requests.get(url, headers=headers)
        print(f"Status Code: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            friends_list = data.get('data', []) or []
            print(f"✅ Got {len(friends_list)} friends")
            return friends_list
        else:
            print(f"❌ Get Friends Failed: {response.text}")
            return []
    except Exception as e:
        print(f"❌ Request Failed: {e}")
        return []


def main():
    print("=" * 60)
    print("Starting API Tests...")
    print("=" * 60)
    
    results = {
        "register": False,
        "login": False,
        "refresh_token": False,
        "add_friend": False,
        "accept_friend": False,
        "get_friends": False
    }
    
    # 1. Register User 1
    if test_register():
        results["register"] = True

    # 2. Login User 1
    tokens = test_login()
    if tokens:
        results["login"] = True
        access_token = tokens.get('access_token')
        refresh_token = tokens.get('refresh_token')
        user_id_1 = tokens.get('id')  # Assume login returns user ID
        
        print(f"Got Refresh Token: {refresh_token[:20]}...")
        
        # 3. Refresh Token
        if test_refresh_token(refresh_token):
            results["refresh_token"] = True
        
        # 4. Test Add Friend Flow
        # First, register and login User 2
        test_register(TEST_PHONE_2, TEST_NAME_2, TEST_PASSWORD)
        tokens2 = test_login(TEST_PHONE_2, TEST_PASSWORD)
        
        if tokens2:
            access_token_2 = tokens2.get('access_token')
            user_id_2 = tokens2.get('id')
            
            # Get User 1's ID by searching (if needed)
            if user_id_1 and user_id_2:
                # User 1 sends friend request to User 2
                if test_add_friend(access_token, user_id_2, TEST_NAME):
                    results["add_friend"] = True
                    
                    # User 2 gets friend requests
                    requests_list = test_get_friend_requests(access_token_2)
                    
                    if requests_list and len(requests_list) > 0:
                        # User 2 accepts friend request
                        request_id = requests_list[0].get('id')
                        if request_id and test_accept_friend(access_token_2, request_id):
                            results["accept_friend"] = True
                            
                            # User 2 gets friends list
                            friends = test_get_friends(access_token_2)
                            if friends and len(friends) > 0:
                                results["get_friends"] = True
    else:
        print("Skipping further tests due to login failure")
    
    # Print Summary
    print("\n" + "=" * 60)
    print("TEST SUMMARY")
    print("=" * 60)
    for test_name, passed in results.items():
        status = "✅ PASS" if passed else "❌ FAIL"
        print(f"{test_name:20s} {status}")
    
    total_passed = sum(results.values())
    total_tests = len(results)
    print(f"\nTotal: {total_passed}/{total_tests} tests passed")

if __name__ == "__main__":
    main()
