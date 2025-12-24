**登陆成功后 后端POST方法返回**

```json
{
    "code": 200,
    "message": 信息,
    "data": 
    {
    	"token_class":
    	{
    		"token": token,
    		"refresh_token": 用于刷新token,
    		"expires_in": token到期时间（以秒为单位）
		},
    	"user_info": 
    	{
    		"uid": 微信号,
    		"name": 昵称
		}
	}
}
```

**refresh_token前端接口**

```http
POST /auth/refresh_token
Authorization: Bearer <refresh_token>
```

**refresh_token 成功后 后端POST方法返回**

```json
{
    "code": ,
    "message": 信息,
    "data":
    {
    	"token": 刷新之后的token,
    	"refresh_token": 新的用于刷新的token,
    	"expires_in": token到期时间（秒）,
	}
}
```

**refresh_token失败，返回**

```json
{
    "code": 状态码,
    "message": 信息
}
```

**refresh_token 状态码**

```html
401: token格式错误或失效
400: 请求参数错误
409: 请求有冲突
```

- **用户修改自己的微信号**

  前端接口

  ```http
  PATCH api/auth/me/uid
  json
  {
  	"new_uid": 新的微信号
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401,
      "message": 信息
  }
  ```

- **打开微信时， 加载聊天会话列表**

  加载会话列表前端接口：

  ```http
  GET api/auth/conversations
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  ```

  失败后端返回：

  ```json
  
  ```

- **用户点击“通讯录”，加载好友列表**

  加载好友列表前端接口：

  ```http
  GET api/auth/friendships
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息,
      "data":[
      	{
              "friendship_id":表主键,
           	"friend_name":好友昵称,
              "friend_id":好友主键，
          },
  		{……}
      ]
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- **用户点击“好友申请”， 加载好友申请列表**

  前端接口：

  ```http
  GET api/auth/friendship_requests
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息,
      "data":[
      	{
              "request_id":表主键,
           	"sender_id":主键,
           	"sender_name":昵称,
           	"verification_message":验证消息,
           	"status":状态,
              "created_at": 好友申请发送时间
          },
  		{……}
      ]
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- **发送好友申请操作**

  前端接口：

  ```http
  POST api/auth/friendship_requests
  json
  {
  	"sender_name": 发送者的昵称,
  	"receiver_id": 谁收到这个好友申请,
  	"verification_message": 发送者写的验证消息
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- **同意/拒绝 好友申请操作**

  前端接口

  ```http
  # 同意
  POST api/auth/friendship_requests/{request_id}
  json
  {
  	"status": "accepted" | "rejected"  #同意还是拒绝
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- **删除好友操作**

  前端接口

  ```http
  DELETE api/auth/friendships/{friend_id #对方的id}
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- **会话详情（表现为用户点击进入聊天窗口）**

  前端接口

  ```http
  GET api/auth/conversations/{conversation_id}
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- **私发消息**

  前端接口

  ```http
  POST api/auth/messages
  json
  {
  	"receiver_id": 给谁发
  	"content": 消息内容
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息,
      "data": message_id（信息表的主键）
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- **撤回消息**

  前端接口

  ```http
  DELETE api/auth/message/recall
  {
  	"message_id":消息表的主键
  }
  Authorization: Bearer <token>
  ```
  
  成功后端返回：
  
  ```json
  {
      "code":  200,
      "message": 信息,
  }
  ```

  失败后端返回：
  
  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```
  
- **删除消息**

  ```http
  DELETE api/auth/message/delete
  {
  	"message_id":消息表的主键
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息,
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

  





















