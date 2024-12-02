package models

import (
	"strconv"
	"time"
)

// Comment 数据库查询到的 评论数据
type Comment struct {
	CId           int64     `json:"cId,string" bson:"cid"`
	OwnId         string    `json:"ownId" bson:"own_id"` // artID 或者 trendID
	OwnType       string    `json:"ownType" bson:"own_type"`
	UserId        string    `json:"userId"  bson:"user_id"`
	Avatar        string    `json:"avatar"`
	UserName      string    `json:"userName"`
	VTag          string    `json:"vTag" bson:"-"`
	VStatus       int8      `json:"vStatus" bson:"-"`
	ReplyId       int64     `json:"replyId,string"   bson:"reply_id"`
	ReplyUserId   string    `json:"replyUserId"   bson:"reply_user"`
	ReplyUserName string    `json:"replyUserName"`
	RootId        int64     `json:"rootId,string"   bson:"root_id"`
	RootCount     int       `json:"rootCount"   bson:"root_count"`
	Text          string    `json:"text" bson:"content"`
	Likes         int       `json:"likes"  bson:"likes"`
	Sore          int64     `json:"sore"`
	IsLike        bool      `json:"isLike"`
	IsDelete      bool      `json:"isDelete"  bson:"is_delete"`
	CreateAT      time.Time `json:"createAT"  bson:"createAt"`
}

type Comments []Comment

func (comments Comments) GetAllCId() []string {
	var checkId []string
	for _, comment := range comments {
		checkId = append(checkId, strconv.FormatInt(comment.CId, 10))
	}
	return checkId
}

// VerifyGetCommentQuery 验证获取作品评论的参数
type VerifyGetCommentQuery struct {
	OwnId   string `form:"ownid" binding:"required,numeric,gt=0"`
	LastCid string `form:"cid" binding:"required"`
	Sore    string `form:"sore" binding:"required"`
	Type    string `form:"type" binding:"oneof=aw tr"`
}

// VerifyGetCommentReply 验证获取单评论里子评论的参数
type VerifyGetCommentReply struct {
	RootId  string `form:"rid" binding:"required,numeric,gt=0"`
	LastCid string `form:"cid" binding:"required"`
	Sore    string `form:"sore" binding:"required"`
}

// ReturnComment 返回的给前端的 评论数据
type ReturnComment struct {
	Comment       `bson:",inline"`
	ChildComments []Comment `json:"childComments"`
}

type ReturnComments []ReturnComment

func (comments ReturnComments) GetAllCId() []string {
	var checkId []string
	for _, comment := range comments {
		checkId = append(checkId, strconv.FormatInt(comment.CId, 10))
		for _, childComment := range comment.ChildComments {
			checkId = append(checkId, strconv.FormatInt(childComment.CId, 10))
		}
	}
	return checkId
}

type FirstShowChildComment struct {
	RootId  int64     `bson:"rootId"`
	Comment []Comment `bson:"comment"`
}

// PostCommentData 上传的作品评论数据
type PostCommentData struct {
	OwnId         string `json:"ownId" binding:"required"`
	Text          string `json:"text" binding:"required"`
	ReplyId       int64  `json:"replyId,string"`
	ReplyUserId   string `json:"replyUserId" binding:"required"`
	RootId        int64  `json:"rootId,string"`
	ReplyUserName string `json:"replyUserName"`
	SenderName    string `json:"senderName" binding:"required"`
	SenderAvatar  string `json:"senderAvatar"`
	Type          string `json:"type" binding:"oneof=aw tr"`
}

type QueryOneRoot struct {
	RootId int64 `form:"rid" binding:"required,numeric,gt=0"`
}

// PostCommentLike 点赞评论
type PostCommentLike struct {
	AuthorId string `json:"authorId" binding:"required"` // 评论发布者
	CId      string `json:"cId" binding:"required"`      // 点赞哪个评论
	IsCancel bool   `json:"isCancel"`
}

type DeleteComment struct {
	CId int64 `json:"cid,string" binding:"required"` // 删除哪个评论
}
