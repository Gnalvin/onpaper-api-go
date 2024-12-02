package models

import "time"

// SendMessage 发送消息需要携带的参数
type SendMessage struct {
	Sender   string `json:"sender"  binding:"required" bson:"sender"`
	Receiver string `json:"receiver"  binding:"required" bson:"receive"`
	Content  string `json:"content"  binding:"required" bson:"content"`
	MsgType  string `json:"msgType"  binding:"required" bson:"msg_type"`
	Width    int    `json:"width"   bson:"width"`
	Height   int    `json:"height"  bson:"height"`
}

// MessageBody 消息字段
type MessageBody struct {
	ChatId   int64     `bson:"chat_id" json:"chatId,omitempty,string"`
	MsgId    int64     `bson:"msg_id" json:"msgId,string"`
	Sender   string    `bson:"sender" json:"sender"`
	Receiver string    `bson:"receiver" json:"receiver"`
	Content  string    `bson:"content" json:"content"`
	MsgType  string    `bson:"msg_type" json:"msgType"`
	SendTime time.Time `bson:"send_time" json:"sendTime"`
	Width    int       `bson:"width,omitempty" json:"width"`
	Height   int       `bson:"height,omitempty" json:"height"`
}

// ChatRelation 会话关系表
type ChatRelation struct {
	SenderId   string         `bson:"sender" json:"-"`
	Sender     UserSimpleInfo `bson:"-" json:"sender"`
	ReceiverId string         `bson:"receiver" json:"-"`
	Receiver   UserSimpleInfo `bson:"-" json:"receiver"`
	ChatId     int64          `bson:"chat_id" json:"chatId,string"`
	Unread     int            `bson:"unread" json:"unread"`
	Message    []MessageBody  `bson:"message" json:"message"`
}

// ReceiveMessage 接收消息需要携带的参数
type ReceiveMessage struct {
	Sender   string `form:"sender" binding:"required,numeric,gt=0"`
	Receiver string `form:"receiver" binding:"required,numeric,gt=0"`
	NextId   *int64 `form:"nextid" binding:"required"`
	ChatId   *int64 `form:"chatid" binding:"required"`
}

// UserUnreadCount 用户未读
type UserUnreadCount struct {
	UserId      string `bson:"userId"`
	TotalUnread int    `bson:"totalUnread"`
}
