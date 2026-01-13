package model

import "time"

// IDResp 通用返回体
// 返回一个uint64的ID
type IDResp struct {
	ID uint64 `json:"id,string"`
}

// UserInfoResp 用户信息返回体
type UserInfoResp struct {
	Name string `json:"name"`
	Uid  string `json:"uid"`
}

// FriendInfoResp 好友信息返回体
type FriendInfoResp struct {
	ID           uint64 `json:"id,string"`
	FriendRemark string `json:"friend_remark"`
	Name         string `json:"name"`
	Uid          string `json:"uid"`
}

// StrangerInfoResp 陌生人信息返回体
type StrangerInfoResp struct {
	ID   uint64 `json:"id,string"`
	Name string `json:"name"`
}

// TokenResp 刷新token操作返回体
type TokenResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    uint64 `json:"expires_in"`
}

// LoginResp 登陆操作返回体
type LoginResp struct {
	UserInfo   UserInfoResp `json:"user_info"`
	TokenClass TokenResp    `json:"token_class"`
}

// FriendRequestListResp 好友申请列表返回体
type FriendRequestListResp struct {
	RequestID           uint64    `gorm:"column:id" json:"request_id,string"`
	SenderID            uint64    `json:"sender_id,string"`
	SenderName          string    `json:"sender_name"`
	VerificationMessage string    `json:"verification_message"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
}

// FriendshipListResp 好友列表返回体
type FriendshipListResp struct {
	FriendshipID uint64 `gorm:"column:id" json:"friendship_id,string"` // friend_ships 表的主键
	FriendRemark string `json:"friend_remark"`
	FriendID     uint64 `json:"friend_id,string"`
}

type EnterConversationResp struct {
	SenderName string `json:"sender_name,string"`
	Content    string `json:"content"`
	ID         uint64 `json:"id,string"`
	Status     uint8  `json:"status"`
}

type ConversationListResp struct {
	Remark         string `json:"remark"`
	Content        string `json:"content"`
	ConversationID uint64 `json:"conversation_id,string"`
}
