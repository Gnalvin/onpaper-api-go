package models

import "time"

// AcceptPlan 接稿方案类型
type AcceptPlan struct {
	PlanId      int64     `json:"planId,string" bson:"plan_id"`
	UserId      string    `json:"userId" bson:"user_id"`
	Name        string    `json:"name" bson:"name" binding:"required,max=25"`                                //方案名称
	Intro       string    `json:"intro" bson:"intro" binding:"required,max=700,min=10"`                      // 方案介绍
	Preference  string    `json:"preference"  bson:"preference" binding:"required,max=100"`                  // 偏好类型
	Refuse      string    `json:"refuse"  bson:"refuse" binding:"required,max=100"`                          // 不接类型
	Money       string    `json:"money" bson:"money" binding:"required,max=20"`                              //期望金额
	Change      int       `json:"change"  bson:"change" binding:"min=1,max=5"`                               // 可修改次数
	Contact     string    `json:"contact,omitempty" bson:"contact" binding:"required,max=25"`                // 联系方式
	ContactType string    `json:"contactType,omitempty" bson:"contact_type" binding:"oneof=QQ Phone WeChat"` // 联系方式类型
	Payment     string    `json:"payment" bson:"payment" binding:"oneof=1 2 3 4 5"`                          // 支付方式
	Finish      int       `json:"finish"  bson:"finish" binding:"required,max=100"`                          // 完成时间天数
	FileType    []string  `json:"fileType"  bson:"file_type" binding:"required"`                             // 提供文件类型
	Status      bool      `json:"status" bson:"-"`                                                           // 0 正常 1 不接单
	IsDelete    bool      `json:"isDelete" bson:"is_delete"`
	UpdateAt    time.Time `json:"updateAt" bson:"updateAt"`
	CreateAt    time.Time `json:"createAt" bson:"createAt"`
}

// UserAcceptPlan 某一用户接稿方案
type UserAcceptPlan struct {
	UserName string `json:"userName"`
	Avatar   string `json:"avatar"`
	VTag     string `json:"vTag"`
	VStatus  int8   `json:"vStatus"`
	AcceptPlan
}

// InvitePlan 约稿邀请计划
type InvitePlan struct {
	InviteId    int64      `json:"inviteId,string" bson:"invite_id"`
	UserId      string     `json:"userId"  bson:"user_id"`
	PlanId      int64      `json:"planId,string"  bson:"plan_id" binding:"required"`
	ArtistId    string     `json:"artistId"  bson:"artist_id" binding:"numeric,max=10"`
	Category    string     `json:"category" bson:"category" binding:"min=2,max=6"`
	Name        string     `json:"name" bson:"name" binding:"required,max=25"`
	Intro       string     `json:"intro" bson:"intro" binding:"required,max=650,min=10"`
	FileList    []PicsType `json:"fileList" bson:"file_list"`
	Purpose     string     `json:"purpose" bson:"purpose" binding:"required,max=20"`
	FileSize    string     `json:"fileSize" bson:"file_size" binding:"oneof=game weibo pc a4 diy square"`
	Color       string     `json:"color" bson:"color" binding:"oneof=RGB CMYK"`
	FileType    []string   `json:"fileType" bson:"file_type" binding:"required"`
	Date        string     `json:"date" bson:"date" binding:"required"`
	Money       string     `json:"money" bson:"money" binding:"required,max=20"`
	Payment     string     `json:"payment" bson:"payment" binding:"oneof=1 2 3 4 5"`
	OpenOption  string     `json:"openOption" bson:"open_option" binding:"oneof=open appoint privacy"`
	ContactType string     `json:"contactType,omitempty" bson:"contact_type" binding:"oneof=QQ Phone WeChat"`
	Contact     string     `json:"contact,omitempty" bson:"contact" binding:"required,max=25"`
	Status      int8       `json:"status" bson:"status"` // 0 未接受 1 沟通中  2 创作中 3 已完成  -1 画师/约稿人关闭(待接稿阶段和沟通阶段关闭) -2 退出（创作中散伙））
	FeedBack    uint8      `json:"feedBack" bson:"feedBack " binding:"oneof=0 3 5 7 15"`
	IsDelete    bool       `json:"isDelete" bson:"is_delete"`
	UpdateAt    time.Time  `json:"updateAt" bson:"updateAt"`
	CreateAt    time.Time  `json:"createAt" bson:"createAt"`
}

// InvitePlanCard 简略的显示约稿计划
type InvitePlanCard struct {
	ArtistId     string     `json:"artistId"  bson:"artist_id"`
	UserId       string     `json:"userId"  bson:"user_id"`
	InviteId     int64      `json:"inviteId,string" bson:"invite_id"`
	Name         string     `json:"name" bson:"name"`
	Intro        string     `json:"intro" bson:"intro"`
	Date         string     `json:"date" bson:"date"`
	Status       int8       `json:"status" bson:"status"` // 0 未接受 1 沟通中  2 创作中 3 已完成  -1 画师/约稿人关闭(待接稿阶段和沟通阶段关闭) -2 退出（创作中中散伙）
	NeedEvaluate bool       `json:"needEvaluate"`
	Money        string     `json:"money" bson:"money"`
	Category     string     `json:"category" bson:"category"`
	UpdateAt     time.Time  `json:"updateAt" bson:"updateAt"`
	FileList     []PicsType `json:"fileList" bson:"file_list"`
}

type PlanQuery struct {
	UserId string `form:"uid" binding:"required,numeric,gt=0"`
	NextId *int64 `form:"next" binding:"required"`
	Type   int8   `form:"type" binding:"oneof=0 1 2 3 -1 -2"`
}

type PlanIdQuery struct {
	InviteId int64 `form:"id" binding:"required"`
}

// UserInvitePlan 某一用户约稿方案
type UserInvitePlan struct {
	Sender UserScoreInfo `json:"sender"`
	Artist UserScoreInfo `json:"artist"`
	InvitePlan
	Evaluates      []Evaluate `json:"evaluates"`
	EvaluateStatus uint8      `json:"evaluateStatus"` // 1 画师完成评价 2 约稿方完成评价 0 都完成评价
}

type UserScoreInfo struct {
	UserId   string  `json:"userId"  db:"user_id" `
	UserName string  `json:"userName"  db:"username" `
	Avatar   string  `json:"avatar" db:"avatar_name" `
	Score    float64 `json:"score" db:"rating"`
	Finish   uint16  `json:"finish" db:"finish"`
}

// Evaluate 评价数据结构
type Evaluate struct {
	EvaluateId int64     `json:"evaluateId"`
	InviteId   int64     `json:"inviteId,string"  binding:"required" db:"invite_id"` // 约稿的邀请ID
	InviteOwn  string    `json:"inviteOwn" db:"invite_own"`
	Sender     string    `json:"sender"  db:"sender" binding:"numeric"`             // 评价发布人
	Receiver   string    `json:"receiver"  db:"receiver" binding:"numeric"`         // 被评价人
	Text       string    `json:"text"  db:"text" binding:"required,max=150,min=10"` // 评价内容
	Rate1      uint8     `json:"rate1"  db:"rate_1" binding:"oneof=1 2 3 4 5"`
	Rate2      uint8     `json:"rate2"  db:"rate_2" binding:"oneof=1 2 3 4 5"`
	Rate3      uint8     `json:"rate3"  db:"rate_3" binding:"oneof=1 2 3 4 5"` // 评分
	Status     int8      `json:"status"  binding:"oneof=3 -2"`                 // 修改状态
	Only       bool      `json:"only,omitempty"`
	Score      float64   `json:"score" db:"total_rating" db:"score"`
	IsDelete   bool      `json:"isDelete" db:"is_delete"`
	CreateAt   time.Time `json:"createAt" db:"createAT"`
}

type PlanNext struct {
	InviteId int64 `json:"inviteId,string" binding:"required"` // 约稿的邀请ID
	Status   int8  `json:"status" binding:"oneof=1 2 -1 "`
}

type EvaluateQuery struct {
	UserId string `form:"uid" binding:"required,numeric,gt=0"`
	Page   uint8  `form:"page" binding:"numeric"`
}

// EvaluateShow 评价展示
type EvaluateShow struct {
	EvaluateId string  `json:"evaluateId" db:"evaluate_id"`
	InviteId   string  `json:"inviteId" db:"invite_id"`
	UserId     string  `json:"userId" db:"sender"`
	UserName   string  `json:"userName"`
	Avatar     string  `json:"avatar"`
	Text       string  `json:"text" db:"text"`
	CreateAt   string  `json:"createAt" db:"createAT"`
	Score      float64 `json:"score" db:"total_rating"`
}

type PlanUserInfo struct {
	ArtistId  string `bson:"artist_id" `
	Sender    string `bson:"user_id"`
	NowStatus int8   `bson:"status"`
}

type PlanContact struct {
	UserId      string `json:"userId" bson:"user_id"`
	UserName    string `json:"userName"`
	Avatar      string `json:"avatar"`
	ContactType string `json:"contactType" bson:"contact_type"`
	Contact     string `json:"contact" bson:"contact"`
}

type CommissionStatus struct {
	Status bool `json:"status"`
}
