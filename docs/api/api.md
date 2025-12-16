**登陆成功后后端返回**

```json
{
    "code": http状态码
    "message": 信息
    "data": {
    	"token": token
    	"refresh_token": 用于刷新token
    	"expires_in": token到期时间
    	"user_info": {
    		"uid": 微信号
    		"name": 昵称
		}
	}
}
```

**refresh_token** 的接口

```http
/refresh_token
```

