- **用户点击“通讯录”，加载好友列表**

    加载好友列表前端接口：

    ```http
    GET /api/auth/friendships
    Authorization: Bearer <token>
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "success",
        "data": [
            {
                "friendship_id": "f_12345",
                "friend_id": "u_67890",
                "friend_name": "张三",
                "friend_remark": "同事"
            }
        ]
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 500,
        "message": "服务器错误"
    }
    ```

- **用户点击“好友申请”， 加载好友申请列表**

    前端接口：

    ```http
    GET /api/auth/friendship_requests
    Authorization: Bearer <token>
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "success",
        "data": [
            {
                "request_id": "r_123",
                "sender_id": "u_111",
                "sender_name": "李四",
                "verification_message": "我是你的大学同学",
                "status": "pending",
                "created_at": "2026-01-15T09:30:00Z"
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

- **发送好友申请**

    请求：

    ```http
    POST /api/auth/friendship_requests
    Content-Type: application/json
    Authorization: Bearer <token>
    ```

    请求体示例：

    ```json
    {
        "receiver_id": "u_67890",
        "verification_message": "你好，我们加个好友吧",
        "sender_name": "我的昵称"
    }
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 201,
        "message": "请求已发送",
        "data": {
            "request_id": "r_124"
        }
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 400,
        "message": "请求参数不合法"
    }
    ```

- **同意好友申请**

    请求：

    ```http
    POST /api/auth/friendship_requests/{request_id}/accept
    Authorization: Bearer <token>
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "已成为好友",
        "data": null
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 404,
        "message": "申请不存在或已处理"
    }
    ```

- **拒绝/删除好友申请**

    请求：

    ```http
    DELETE /api/auth/friendship_requests/{request_id}
    Authorization: Bearer <token>
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "已删除申请",
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

- **删除好友**

    请求：

    ```http
    DELETE /api/auth/friendships/{friend_id}
    Authorization: Bearer <token>
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "已删除好友",
        "data": null
    }
    ```

    失败后端返回（示例）：

    ```json
    {
        "code": 400,
        "message": "好友不存在"
    }
    ```

- **修改好友备注**

    说明：`conversation_users` 表中私聊的 `remark` 字段与好友备注相关联，修改备注时应一并同步。

    请求：

    ```http
    POST /api/auth/friendships/remark/{friend_id}
    Authorization: Bearer <token>
    Content-Type: application/json
    ```

    请求体示例：

    ```json
    {
        "remark": "大学同学"
    }
    ```

    成功后端返回（示例）：

    ```json
    {
        "code": 200,
        "message": "备注已更新",
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

















