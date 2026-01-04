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
      "message": "success",
      "data":[
      	{
              "friendship_id":"表主键",
           	"friend_remark": "好友备注",
              "friend_id":"好友主键",
          },
  		{……}
      ]
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  500,
      "message": "服务器错误"
  }
  ```

- **用户点击“好友申请”， 加载好友申请列表**

  前端接口：

  ```http
  GET /api/auth/friendship_requests
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  200,
      "message": 信息,
      "data":[
      	{
              "request_id":"表主键",
           	"sender_id":"主键",
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
  POST /api/auth/friendship_requests
  Content-Type: application/json

  {
  	"sender_name": "发送者的昵称",
  	"receiver_id": "接收者ID",
  	"verification_message": "发送者写的验证消息"
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  201,
      "message": "发送成功"
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400,
      "message": "发送失败"
  }
  ```

- **同意好友申请操作**

  前端接口

  ```http
  POST /api/auth/friendship_requests/{request_id}
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  201,
      "message": "success",
      "data": null
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  500,
      "message": "服务器错误"
  }
  ```

- **拒绝/删除好友申请操作**

  前端接口

  ```http
  DELETE /api/auth/friendship_requests/{request_id}
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  201,
      "message": "success",
      "data": null
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  500,
      "message": "服务器错误"
  }
  ```

- **删除好友操作**

  前端接口

  ```http
  DELETE /api/auth/friendships/{friend_id}
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":  201,
      "message": "success"
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400,
      "message": "好友不存在"
  }
  ```

  













