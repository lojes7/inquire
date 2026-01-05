import requests
import json
import time

# 配置
BASE_URL = "http://localhost:8080/api"
TEST_PHONE = "13800138000"
TEST_PASSWORD = "password123"
TEST_NAME = "TestUser"

def test_register():
    print("\n--- Testing Register ---")
    url = f"{BASE_URL}/register"
    payload = {
        "name": TEST_NAME,
        "password": TEST_PASSWORD,
        "phone_number": TEST_PHONE
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

def test_login():
    print("\n--- Testing Login ---")
    url = f"{BASE_URL}/login/phone_number"
    payload = {
        "phone_number": TEST_PHONE,
        "password": TEST_PASSWORD
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
            else:
                print(f"❌ Refresh Failed: {data['msg']}")
        else:
            print(f"❌ Refresh Failed: {response.text}")

        # 2. Test with invalid token (simulating access token as refresh token)
        # We need an access token first, but let's just use a dummy string or modify the refresh token
        print("2. Testing with INVALID token (tampered)...")
        headers["Authorization"] = f"Bearer {refresh_token}invalid"
        response = requests.post(url, headers=headers)
        if response.status_code == 401:
             print("✅ Correctly rejected invalid token")
        else:
             print(f"❌ Unexpected status for invalid token: {response.status_code}")

    except Exception as e:
        print(f"❌ Request Failed: {e}")

def main():
    print("Starting API Tests...")
    
    # 1. Register
    if not test_register():
        print("Aborting due to register failure")
        # return # Don't return, maybe user exists

    # 2. Login
    tokens = test_login()
    if tokens:
        access_token = tokens.get('access_token')
        refresh_token = tokens.get('refresh_token')
        
        print(f"Got Refresh Token: {refresh_token[:20]}...")
        
        # 3. Refresh Token
        test_refresh_token(refresh_token)
        
    else:
        print("Skipping Refresh Token test due to login failure")

if __name__ == "__main__":
    main()
