- **打开微信时， 加载会话列表**

  根据user_id查conversation_users表，需要注意删除的会话就隐藏，

  前端需要拿到：会话备注、最后一条消息预览，会话ID 排序：会话时间、是否置顶

  加载会话列表前端接口：

  ```http
  GET api/auth/conversations
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":
      "message":
      "data":
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

  根据conversation_id查messages表，需要根据时间排

  所以要返回每个消息的 sender_name  content  id  status

  需要注意撤回的消息和删除的消息，status = 2是系统消息

  前端接口

  ```http
  GET api/auth/conversations/{conversation_id}
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  { 
      "code": 200
      "message":
      "data": {
  		[
      		{update_time, sender_name, content, id, status},
  			{},
      	]    
  	}
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- 创建私聊

  前端接口

  ```http
  POST /api/auth/conversations/private
  Content-Type: application/json
  
  {
  	"id": 好友id
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code": 201
      "message": "success"
      "data": null
  }
  ```

  失败后端返回：

  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```

- 创建群聊

  前端接口

  ```http
  POST /api/auth/conversations/group
  Content-Type: application/json
  
  {
  
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":
      "message":
      "data":
  }
  ```

  失败后端返回：

- **发送消息**

  根据sender_id conversation_id content 写进messages表

  同时，需要更新conversation_users表的last_message_id and unread_count字段

  需要返回给前端该消息的ID

  前端接口

  ```http
  POST /api/auth/messages
  Content-Type: application/json
  
  {
  	"conversation_id": "会话ID",
  	"content": "消息内容"
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":
      "message":
      "data":
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

  根据message_id 更新messages表中的status字段

  同时创建一个系统级消息为“对方撤回了一条消息”

  同时更新conversation_users表的last_message_id字段指向新创建的系统级消息

  前端接口

  ```http
  DELETE /api/auth/messages/recall
  Content-Type: application/json
  
  {
  	"id": "消息ID"
  }
  Authorization: Bearer <token>
  ```

  成功后端返回：

  ```json
  {
      "code":
      "message":
      "data":
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

  根据message_id和user_id更新message_users表的is_deleted字段为true

  同时需要注意修改conversation_users表中的last_message_id
  
  前端接口

  ```http
  DELETE /api/auth/messages/delete
  Content-Type: application/json
  
  {
  	"id": "消息ID"
  }
  Authorization: Bearer <token>
  ```
  
  成功后端返回：
  
  ```json
  {
      "code":
      "message":
      "data":
  }
  ```
  
  失败后端返回：
  
  ```json
  {
      "code":  400 / 401 / 500 / 409,
      "message": 信息
  }
  ```
  
  





















