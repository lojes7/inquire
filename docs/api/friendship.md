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
      "code":  201,
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
      "code":  201,
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

  













