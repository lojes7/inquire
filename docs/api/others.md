**查看好友信息**

前端接口：

```http
GET api/auth/info/friends/{id}
Authorization: Bearer <token>
```

成功后端返回：

```json
{
    "code":200,
    "message":
    "data": {
    	"id":表主键,
    	"remark":备注，
    	"name":昵称,
    	"uid":微信号
	}
}
```

失败后端返回：

```json
{
    "code": 400/401/404/500
    "message":
}
```

**查看陌生人信息**

前端接口：

```http
GET api/auth/info/strangers/{id}
Authorization: Bearer <token>
```

成功后端返回：

```json
{
    "code":200,
    "message":,
    "data": {
    	"id":表主键,
    	"name":昵称
	}
}
```

失败后端返回：

```json
{
    "code": 400/401/404/500
    "message":
}
```

