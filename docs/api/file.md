## 文件相关

### 发送文件（http）

```http
POST /api/auth/messages/file
Authorization: Bearer <access_token>
multipart/form-data
```

| 字段名          | 类型   | 说明               |
| --------------- | ------ | ------------------ |
| conversation_id | string | 会话ID             |
| file            | File   | 用户想要发送的文件 |

成功返回：

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "message_id": "123456",
        "file_name": "有点意思",
        "file_size": "114514", （单位：字节）
        "file_type": "application/pdf"
    }
}
```

失败返回示例：

```json
{
  "code":400 / 401 / 404 / 500,
  "message":"错误信息"
}
```

