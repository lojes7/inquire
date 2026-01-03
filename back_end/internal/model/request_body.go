package model

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

// ReviseUidReq 修改微信号请求体
type ReviseUidReq struct {
	NewUid string `json:"new_uid" binding:"required,min=1,max=20"`
}

// AddFriendReq 添加好友请求体
type AddFriendReq struct {
	ReceiverID          uint64 `json:"receiver_id" binding:"required,gt=0"`
	SenderName          string `json:"sender_name" binding:"required,max=64"`
	VerificationMessage string `json:"verification_message" binding:"omitempty,max=128"`
}

// FriendRequestActionReq 通过/拒绝好友申请请求体
type FriendRequestActionReq struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected"`
}

// SendMessageReq 发送消息请求体
type SendMessageReq struct {
	ReceiverID uint64 `json:"receiver_id" binding:"required,gt=0"`
	Content    string `json:"content" binding:"required,max=1024"`
}

// MessageIDReq 消息ID请求体
type MessageIDReq struct {
	MessageID uint64 `json:"message_id" binding:"required,gt=0"`
}
