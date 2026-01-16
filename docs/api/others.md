## 基础说明

**响应格式（成功）**

```json
{
    "code": 200,
    "message": "success",
    "data": null
}
```

成功响应的 HTTP 状态码恒为 200，但 `code` 字段会按业务返回（如 201）。

**响应格式（失败）**

```json
{
    "code": 400,
    "message": "错误信息"
}
```

失败响应的 HTTP 状态码与 `code` 一致。

**鉴权**

`/api/auth/**` 需要 `Authorization: Bearer <access_token>`。

`/api/auth/refresh_token` 需要 `Authorization: Bearer <refresh_token>`。

---

## 注册与登录

### 注册

```http
POST /api/register
Content-Type: application/json
```

请求体：

```json
{
    "name": "xiaoming",
    "password": "P@ssw0rd",
    "phone_number": "13712345678"
}
```

成功返回：

```json
{
    "code": 201,
    "message": "注册成功",
    "data": null
}
```

失败返回示例：

```json
{
    "code": 400,
    "message": "手机号已存在"
}
```

### 微信号登录

```http
POST /api/login/uid
Content-Type: application/json
```

请求体：

```json
{
    "uid": "V_abcd123",
    "password": "P@ssw0rd"
}
```

成功返回：

```json
{
    "code": 200,
    "message": "登陆成功",
    "data": {
        "user_info": {
            "name": "xiaoming",
            "uid": "V_abcd123"
        },
        "token_class": {
            "token": "eyJhbGci...",
            "refresh_token": "eyJhbGci...",
            "expires_in": 3600
        }
    }
}
```

失败返回示例：

```json
{
    "code": 400,
    "message": "登陆失败 微信号或密码错误"
}
```

### 手机号登录

```http
POST /api/login/phone_number
Content-Type: application/json
```

请求体：

```json
{
    "phone_number": "13712345678",
    "password": "P@ssw0rd"
}
```

成功/失败返回与微信号登录相同结构，失败消息可能为：

```json
{
    "code": 400,
    "message": "登陆失败 手机号或密码错误"
}
```

### 刷新 Token

```http
POST /api/auth/refresh_token
Authorization: Bearer <refresh_token>
```

成功返回：

```json
{
    "code": 201,
    "message": "success",
    "data": {
        "token": "eyJhbGci...",
        "refresh_token": "eyJhbGci...",
        "expires_in": 3600
    }
}
```

失败返回示例：

```json
{
    "code": 403,
    "message": "不是refresh_token"
}
```

---

## 个人资料

### 修改微信号

```http
POST /api/auth/me/uid
Authorization: Bearer <access_token>
Content-Type: application/json
```

请求体：

```json
{
    "uid": "V_newuid"
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

失败返回示例：

```json
{
    "code": 400,
    "message": "微信号重复"
}
```

### 修改密码

```http
POST /api/auth/me/password
Authorization: Bearer <access_token>
Content-Type: application/json
```

请求体：

```json
{
    "prev_password": "旧密码",
    "new_password": "新密码"
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

失败返回示例：

```json
{
    "code": 500,
    "message": "新密码与旧密码不能相同"
}
```

### 修改昵称

```http
POST /api/auth/me/name
Authorization: Bearer <access_token>
Content-Type: application/json
```

请求体：

```json
{
    "name": "新的昵称"
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

## 用户信息查询

### 查看好友信息（通过 ID）

```http
GET /api/auth/info/friends/id/{id}
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "123456",
        "friend_remark": "备注",
        "name": "好友昵称",
        "uid": "V_abc123"
    }
}
```

失败返回示例：

```json
{
    "code": 400,
    "message": "找不到好友"
}
```

### 查看陌生人信息（通过 ID）

```http
GET /api/auth/info/strangers/id/{id}
Authorization: Bearer <access_token>
```

成功返回：

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "123456",
        "name": "陌生人昵称"
    }
}
```

### 查看好友信息（通过 Uid）

```http
GET /api/auth/info/friends/uid/{uid}
Authorization: Bearer <access_token>
```

成功返回结构与“通过 ID”一致。

### 查看陌生人信息（通过 Uid）

```http
GET /api/auth/info/strangers/uid/{uid}
Authorization: Bearer <access_token>
```

成功返回结构与“通过 ID”一致。




