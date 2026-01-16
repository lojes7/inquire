## 会话

### 加载会话列表

```http
GET /api/auth/conversations
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "remark": "张三",
            "conversation_id": "123456",
            "unread_count": 2,
            "content": "你好，今天有空吗？"
        }
    ]
}
```

### 加载聊天记录

```http
GET /api/auth/conversations/{conversation_id}
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "message_id": "2001",
            "sender_id": "111",
            "sender_name": "李四",
            "status": 0,
            "updated_at": "2026-01-15T09:35:00Z",
            "content": "下午一起吃饭？"
        },
        {
            "message_id": "2002",
            "sender_id": "100",
            "sender_name": "我",
            "status": 3,
            "updated_at": "2026-01-15T09:36:00Z",
            "content": {
                "file_name": "doc.pdf",
                "file_url": "https://example.com/doc.pdf",
                "file_size": 12345,
                "file_type": "application/pdf"
            }
        }
    ]
}
```

`status` 枚举值：

- 0：文本消息
- 1：撤回（查询结果中已过滤）
- 2：系统消息
- 3：文件消息

### 创建私聊

```http
POST /api/auth/conversations/private
Content-Type: application/json
Authorization: Bearer <access_token>
```

请求体：

```json
{
    "id": "67890"
}
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

## 消息

### 发送文本消息

```http
POST /api/auth/messages
Content-Type: application/json
Authorization: Bearer <access_token>
```

请求体：

```json
{
    "conversation_id": "123456",
    "content": "大家下午见！"
}
```

成功返回：

```json
{
    "code": 201,
    "message": "success",
    "data": 3001
}
```

### 撤回消息

```http
DELETE /api/auth/messages/recall
Content-Type: application/json
Authorization: Bearer <access_token>
```

请求体：

```json
{
    "id": "3001"
}
```

成功返回：

```json
{
    "code": 201,
    "message": "success",
    "data": 4001
}
```

### 删除消息（仅对当前用户隐藏）

```http
DELETE /api/auth/messages/delete
Content-Type: application/json
Authorization: Bearer <access_token>
```

请求体：

```json
{
    "id": "3001"
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























