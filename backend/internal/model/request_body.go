package model

// IDReq 消息ID请求体
type IDReq struct {
	ID uint64 `json:"id,string" binding:"required,gt=0"`
}

type RemarkReq struct {
	Remark string `json:"remark" binding:"required,max=64"`
}

// UidReq 修改微信号请求体
type UidReq struct {
	Uid string `json:"uid" binding:"required,min=1,max=20"`
}

// NameReq 修改用户名请求体
type NameReq struct {
	Name string `json:"name" binding:"required,min=1,max=64"`
}

// RegisterReq 注册请求体
type RegisterReq struct {
	Name        string `json:"name" binding:"required,min=1,max=64"`
	Password    string `json:"password" binding:"required,min=6,max=72"`
	PhoneNumber string `json:"phone_number" binding:"required,len=11,numeric"`
}

// LoginByUidReq 微信号登陆请求体
type LoginByUidReq struct {
	Password string `json:"password" binding:"required,min=6,max=72"`
	Uid      string `json:"uid" binding:"required,min=1,max=20"`
}

// LoginByPhoneReq 手机号登陆请求体
type LoginByPhoneReq struct {
	PhoneNumber string `json:"phone_number" binding:"required,len=11,numeric"`
	Password    string `json:"password" binding:"required,min=6,max=72"`
}

// RevisePasswordReq 修改密码请求体
type RevisePasswordReq struct {
	PrevPassword string `json:"prev_password" binding:"required,min=6,max=72"`
	NewPassword  string `json:"new_password" binding:"required,min=6,max=72"`
}

// AddFriendReq 添加好友请求体
type AddFriendReq struct {
	ReceiverID          uint64 `json:"receiver_id,string" binding:"required,gt=0"`
	SenderName          string `json:"sender_name" binding:"required,max=64"`
	VerificationMessage string `json:"verification_message" binding:"omitempty,max=128"`
}

// SendTextReq 发送消息请求体
type SendTextReq struct {
	ConversationID uint64 `json:"conversation_id,string" binding:"required,gt=0"`
	Content        string `json:"content" binding:"required,max=1024"`
}
