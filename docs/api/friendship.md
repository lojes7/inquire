## 好友申请

### 加载好友申请列表

```http
GET /api/auth/friendship_requests
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "request_id": "123",
            "sender_id": "456",
            "sender_name": "李四",
            "verification_message": "我是你的大学同学",
            "status": "pending",
            "created_at": "2026-01-15T09:30:00Z"
        }
    ]
}
```

失败返回示例：

```json
{
    "code": 500,
    "message": "服务器错误"
}
```

### 发送好友申请

```http
POST /api/auth/friendship_requests
Content-Type: application/json
Authorization: Bearer <access_token>
```

请求体：

```json
{
    "receiver_id": "67890",
    "sender_name": "我的昵称",
    "verification_message": "你好，我们加个好友吧"
}
```

成功返回：

```json
{
    "code": 201,
    "message": "发送成功",
    "data": null
}
```

失败返回示例：

```json
{
    "code": 409,
    "message": "发送失败，请勿重复发送"
}
```

### 同意好友申请

```http
POST /api/auth/friendship_requests/{request_id}
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 201,
    "message": "success",
    "data": null
}
```

失败返回示例：

```json
{
    "code": 400,
    "message": "requestID错误"
}
```

### 删除好友申请

```http
DELETE /api/auth/friendship_requests/{request_id}
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 201,
    "message": "success",
    "data": null
}
```

---

## 好友关系

### 加载好友列表

```http
GET /api/auth/friendships
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "friendship_id": "12345",
            "friend_id": "67890",
            "friend_remark": "同事"
        }
    ]
}
```

### 删除好友

```http
DELETE /api/auth/friendships/{friend_id}
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 201,
    "message": "success",
    "data": null
}
```

失败返回示例：

```json
{
    "code": 400,
    "message": "好友不存在"
}
```

### 修改好友备注

```http
POST /api/auth/friendships/remark/{friend_id}
Authorization: Bearer <access_token>
Content-Type: application/json
```

请求体：

```json
{
    "remark": "大学同学"
}
```

成功返回：

```json
{
    "code": 200,
    "message": "success",
    "data": null
}
```

















