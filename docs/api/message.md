- **打开微信时，加载会话列表**

    前端接口：

    ```http
    GET /api/auth/conversations
    Authorization: Bearer <token>
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "success",
        "data": [
            {
                "conversation_id": "c_1001",
                "title": "张三",
                "last_message_preview": "你好，今天有空吗？",
                "last_message_time": "2026-01-15T09:40:00Z",
                "unread_count": 2,
                "is_pinned": false
            }
        ]
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 401,
        "message": "未授权"
    }
    ```

- **会话详情（用户点击进入聊天窗口）**

    前端接口：

    ```http
    GET /api/auth/conversations/{conversation_id}
    Authorization: Bearer <token>
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "success",
        "data": [
            {
                "id": "m_2001",
                "sender_id": "u_111",
                "sender_name": "李四",
                "content": "下午一起吃饭？",
                "status": 0,
                "created_at": "2026-01-15T09:35:00Z"
            },
            {
                "id": "m_2002",
                "sender_id": "u_100",
                "sender_name": "我",
                "content": "可以，几点？",
                "status": 0,
                "created_at": "2026-01-15T09:36:00Z"
            }
        ]
    }
    ```

    字段说明：`status` = 0 正常消息，1 已撤回，2 系统消息。

    失败后端返回（示例）：

    ```json
    {
        "code": 404,
        "message": "会话不存在"
    }
    ```

- **创建私聊**

    前端接口：

    ```http
    POST /api/auth/conversations/private
    Content-Type: application/json
    Authorization: Bearer <token>
    ```

    请求体示例：

    ```json
    {
        "friend_id": "u_67890"
    }
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 201,
        "message": "会话已创建",
        "data": {
            "conversation_id": "c_1002"
        }
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 400,
        "message": "参数错误"
    }
    ```

- **创建群聊**

    前端接口：

    ```http
    POST /api/auth/conversations/group
    Content-Type: application/json
    Authorization: Bearer <token>
    ```

    请求体示例：

    ```json
    {
        "name": "项目讨论组",
        "member_ids": ["u_111", "u_222", "u_333"]
    }
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 201,
        "message": "群聊已创建",
        "data": { "conversation_id": "c_2001" }
    }
    ```

- **发送消息**

    前端接口：

    ```http
    POST /api/auth/messages
    Content-Type: application/json
    Authorization: Bearer <token>
    ```

    请求体示例：

    ```json
    {
        "conversation_id": "c_1001",
        "content": "大家下午见！"
    }
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 201,
        "message": "消息已发送",
        "data": {
            "message_id": "m_3001",
            "created_at": "2026-01-15T10:00:00Z"
        }
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 400,
        "message": "发送失败"
    }
    ```

- **撤回消息**

    前端接口：

    ```http
    DELETE /api/auth/messages/recall
    Content-Type: application/json
    Authorization: Bearer <token>
    ```

    请求体示例：

    ```json
    { "id": "m_3001" }
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "消息已撤回",
        "data": {
            "system_message_id": "m_sys_1"
        }
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 403,
        "message": "不可撤回（超时或无权限）"
    }
    ```

- **删除消息（沙箱级别的用户删除，仅对该用户隐藏）**

    前端接口：

    ```http
    DELETE /api/auth/messages/delete
    Content-Type: application/json
    Authorization: Bearer <token>
    ```

    请求体示例：

    ```json
    { "id": "m_3001" }
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "已删除",
        "data": null
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 500,
        "message": "服务器错误"
    }
    ```























