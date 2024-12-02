package models

import "time"

type BaseNotify struct {
	NotifyId   string         `json:"notifyId" bson:"_id,omitempty"`
	Type       string         `json:"type" bson:"type"`             // 消息类型(公告announce、提醒remind、通知notify)
	TargetId   string         `json:"targetId" bson:"targetId"`     // 目标的ID(比如作品ID 评论ID )
	TargetType string         `json:"targetType" bson:"targetType"` // 目标的类型(tr-动态 aw-作品 usr-用户 -cm -评论 com -约稿)
	Action     string         `json:"action" bson:"action"`         // 动作类型(比如点赞like,收藏collect,关注 focus)
	Sender     UserSimpleInfo `json:"sender" bson:"sender"`         // 发送者
	ReceiverId string         `json:"receiverId" bson:"receiverId"` // 接受者Id 被通知对象
	UpdateAt   time.Time      `json:"updateAt" bson:"updateAt"`     // 创建/更新时间
}

// NotifyBody 通知体
type NotifyBody struct {
	BaseNotify `bson:",inline"`
	Content    interface{} `json:"content,omitempty" bson:"content,omitempty"` // 主体内容 比如评论主体，文章主体
}

// NotifyUnreadCount 通知未读数
type NotifyUnreadCount struct {
	Like       int `json:"like" db:"like"`
	Collect    int `json:"collect" db:"collect"`
	Follow     int `json:"follow" db:"follow"`
	Comment    int `json:"comment" db:"comment"`
	Commission int `json:"commission" db:"commission"`
	At         int `json:"at" db:"at"`
}

func (n NotifyUnreadCount) Count() int {
	return n.At + n.Comment + n.Follow + n.Collect + n.Like + n.Commission
}

// NotifyArtOrTrendInfo 通知中展示的作品和动态需要的信息
type NotifyArtOrTrendInfo struct {
	Id       string `json:"id" db:"artwork_id"`
	Cover    string `json:"cover" db:"cover"`
	Author   string `json:"author"`
	IsDelete bool   `json:"isDelete" db:"is_delete"`
	Text     string `json:"text"`
}

// CommentNotify 评论类型的通知
type CommentNotify struct {
	BaseNotify `bson:",inline"`
	Content    NotifyCommentInfo `json:"content" bson:"content"`
}

// NotifyCommentInfo 被评论提醒
type NotifyCommentInfo struct {
	SendContent    string `json:"sendContent" bson:"-"`                //  发送的评论内容
	SendCId        int64  `json:"sendCId,string" bson:"sendCId"`       //  发送的评论ID
	BeReplyContent string `json:"beReplyContent" bson:"-"`             //  被回复的评论内容
	BeReplyCId     int64  `json:"beReplyCId,string" bson:"beReplyCId"` //  被回复的评论id
	OwnType        string `json:"ownType" bson:"-"`                    //  属于作品还是动态评论 aw tr
	OwnId          string `json:"ownId" bson:"-"`                      // 作品或动态id
	Author         string `json:"author" bson:"-"`
	Cover          string `json:"cover" bson:"-"`      // 封面
	IsLike         bool   `json:"isLike" bson:"-"`     //  是否点赞过
	SendIsDel      bool   `json:"sendIsDel" bson:"-"`  // 发布的评论是否被删除
	OwnerIsDel     bool   `json:"ownerIsDel" bson:"-"` //作品/动态 是否删除
}

// LikeCommentNotify 点赞评论提醒
type LikeCommentNotify struct {
	BeLikeContent string `json:"beLikeContent" bson:"-"`            //  被喜欢的评论内容
	BeLikeCId     int64  `json:"beLikeCId,string" bson:"beLikeCId"` //  被喜欢的评论id
	RootId        int64  `json:"rootId,string" bson:"-"`            // 被点赞的评论根评论id
	Author        string `json:"author" bson:"-"`
	OwnType       string `json:"ownType" bson:"-"`      //  属于作品还是动态评论 aw tr
	OwnId         string `json:"ownId" bson:"-"`        //  作品或动态id
	Cover         string `json:"cover" bson:"-"`        //  封面
	OwnerIsDel    bool   `json:"ownerIsDel" bson:"-"`   //作品/动态 是否删除
	CommentIsDel  bool   `json:"commentIsDel" bson:"-"` // 评论是否删除
}

type NotifyQuery struct {
	NextId string `form:"next" binding:"required"`
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	Comment uint8 `json:"comment" db:"comment" binding:"oneof=0 1 2"`
	Like    uint8 `json:"like" db:"like" binding:"oneof=0 1 2"`
	Collect uint8 `json:"collect" db:"collect" binding:"oneof=0 1 2"`
	Follow  uint8 `json:"follow" db:"follow" binding:"oneof=0 1 2"`
	Message uint8 `json:"message" db:"message" binding:"oneof=0 1 2"`
	At      uint8 `json:"at" db:"at" binding:"oneof=0 1 2"`
}

// CommissionNotify 约稿通知
type CommissionNotify struct {
	BaseNotify `bson:",inline"`
	Content    NotifyCommissionInfo `json:"content" bson:"content"`
}

type NotifyCommissionInfo struct {
	InviteId int64  `json:"inviteId,string" bson:"invite_id,omitempty"`
	Owner    string `json:"owner" bson:"-"`
	Title    string `json:"text" bson:"name,omitempty"`
	Status   int8   `json:"status" bson:"status"` // 0 未接受 1 沟通中  2 创作中 3 已完成  -1 画师/约稿人关闭(待接稿阶段和沟通阶段关闭) -2 退出（创作中中散伙）
	Cover    string `json:"cover" bson:"file_list,omitempty"`
}
