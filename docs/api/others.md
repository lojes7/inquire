**注册**

前端接口：

```http
POST /api/register
Content-Type: application/json

{
    "name": "xiaoming",
    "password": "P@ssw0rd",
    "phone_number": "+8613712345678"
}
```

成功后端返回（示例）：

```json
{
    "code": 201,
    "message": "注册成功",
    "data": {
        "user_id": "u_1001"
    }
}
```

失败后端返回（示例）：

```json
{
    "code": 400,
    "message": "手机号已存在"
}
```

**微信号登录**

前端接口：

```http
POST /api/login/uid
Content-Type: application/json

{
    "uid": "wx_abc123",
    "password": "P@ssw0rd"
}
```

成功后端返回（示例）：

```json
{
    "code": 200,
    "message": "登陆成功",
    "data": {
        "user_info": {
            "name": "xiaoming",
            "uid": "wx_abc123",
            "id": "u_1001"
        },
        "token_class": {
            "token": "eyJhbGci...",
            "refresh_token": "rft_abc...",
            "expires_in": 3600
        }
    }
}
```

失败后端返回（示例）：

```json
{
    "code": 401,
    "message": "账号或密码错误"
}
```

**手机号登录**

前端接口：

```http
POST /api/login/phone_number
Content-Type: application/json

{
    "phone_number": "+8613712345678",
    "password": "P@ssw0rd"
}
```

成功/失败返回与微信号登录相同格式。

**刷新Token**

前端接口：

```http
POST /api/auth/refresh_token
Content-Type: application/json
Authorization: Bearer <refresh_token>

{
    "refresh_token": "rft_abc..."
}
```

成功后端返回（示例）：

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "token": "new_jwt_token",
        "refresh_token": "new_refresh_token",
        "expires_in": 3600
    }
}
```

失败后端返回（示例）：

```json
{
    "code": 401,
    "message": "refresh_token 无效或已过期"
}
```

**修改微信号**

前端接口：

```http
POST /api/auth/me/uid
Authorization: Bearer <token>
Content-Type: application/json

{
    "uid": "new_wx_id"
}
```

成功后端返回（示例）：

```json
{
    "code": 200,
    "message": "微信号已更新",
    "data": null
}
```

失败后端返回（示例）：

```json
{
    "code": 400,
    "message": "微信号重复"
}
```

**修改昵称**

前端接口：

```http
POST /api/auth/me/name
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "新的昵称"
}
```

成功后端返回（示例）：

```json
{
    "code": 200,
    "message": "昵称已更新",
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

**修改密码**

前端接口：

```http
POST /api/auth/me/password
Authorization: Bearer <token>
Content-Type: application/json

{
    "prev_password": "旧密码",
    "new_password": "新密码"
}
```

成功后端返回（示例）：

```json
{
    "code": 200,
    "message": "密码已修改",
    "data": null
}
```

失败后端返回（示例）：

```json
{
    "code": 400,
    "message": "新密码与旧密码不能相同"
}
```

**查看好友信息（通过 ID）**

前端接口：

```http
GET /api/auth/info/friends/id/{id}
Authorization: Bearer <token>
```

成功后端返回（示例）：

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "u_67890",
        "friend_remark": "备注",
        "name": "好友昵称",
        "uid": "wx_67890"
    }
}
```

失败后端返回（示例）：

```json
{
    "code": 404,
    "message": "找不到好友"
}
```

**查看陌生人信息（通过 ID 或 Uid）**

前端接口示例：

```http
GET /api/auth/info/strangers/id/{id}
GET /api/auth/info/strangers/uid/{uid}
Authorization: Bearer <token>
```

成功后端返回（示例）：

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "u_99999",
        "name": "陌生人昵称"
    }
}
```

失败后端返回（示例）：

```json
{
    "code": 404,
    "message": "找不到此人"
}
```




