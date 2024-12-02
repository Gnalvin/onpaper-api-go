package models

import "time"

type TrendShowInfo struct {
	TrendInfo `bson:",inline"`
	Forward   *TrendInfo `json:"forward,omitempty"`
}

type TrendInfo struct {
	TrendId     int64             `json:"trendId,string" bson:"trend_id" db:"artwork_id"`
	UserId      string            `json:"userId" bson:"user_id" db:"user_id"`
	UserName    string            `json:"userName"`
	Avatar      string            `json:"avatar"`
	IsDelete    bool              `json:"isDelete" db:"is_delete" bson:"is_delete"`
	Pics        []PicsType        `json:"pics" bson:"pics"`
	Count       TrendCount        `json:"count" bson:"count"`
	Intro       string            `json:"intro" bson:"text" db:"description"`
	Type        string            `json:"type"`
	ForwardInfo ForwardInfo       `json:"forwardInfo" bson:"forward_info"`
	Topic       TopicType         `json:"topic" bson:"topic"`
	Interact    UserTrendInteract `json:"interact"`
	Comment     string            `json:"comment" bson:"comment" db:"comment"`
	WhoSee      string            `json:"whoSee" bson:"whoSee" db:"whoSee"`
	IsOwner     bool              `json:"isOwner"`
	VTag        string            `json:"vTag" db:"v_tag"`
	VStatus     int8              `json:"vStatus" db:"v_status"`
	CreateAT    time.Time         `json:"createAt" bson:"createAt" db:"createAt"`
}

type ForwardInfo struct {
	Id   int64  `json:"id,string" bson:"id"`
	Type string `json:"type" bson:"type"`
}

// UserTrendInteract 用户与动态的互动信息
type UserTrendInteract struct {
	IsFocusAuthor uint8 `json:"isFocusAuthor"`
	IsLike        bool  `json:"isLike"`
}

type TopicType struct {
	TopicId string `json:"topicId"  bson:"topic_id" db:"topic_id"`
	Text    string `json:"text" bson:"text" db:"text"`
}

type SearchTopicType struct {
	TopicId string `json:"topicId"  db:"topic_id"`
	Text    string `json:"text" db:"text"`
	Count   int    `json:"count" db:"trend_count"`
}

type TrendList []TrendShowInfo

func (s TrendList) Len() int {
	return len(s)
}

// Less 高到低排序
func (s TrendList) Less(i, j int) bool {
	return s[i].TrendId > s[j].TrendId
}

func (s TrendList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// TrendCount 动态数据
type TrendCount struct {
	Likes    int `json:"likes" bson:"likes"`
	Comments int `json:"comments" bson:"comments"`
	Forwards int `json:"forwards" bson:"forwards"`
	Collects int `json:"collects" bson:"collects"`
	Views    int `json:"views" bson:"views"`
}

// TrendQuery 查询单个动态详情
type TrendQuery struct {
	TrendId int64  `json:"trendId,string" form:"id" binding:"required"`
	Type    string `json:"type" form:"type" binding:"oneof=tr aw"`
}

type TrendUserQuery struct {
	UserId string `form:"uid" binding:"required,numeric,gt=0"`
	NextId *int64 `form:"next" binding:"required"`
}

// TrendPermission 动态权限
type TrendPermission struct {
	TrendId int64  `json:"trendId,string"  binding:"required"`
	Comment string `json:"comment" binding:"oneof=public onlyFans close" `
	WhoSee  string `json:"whoSee" binding:"oneof=public onlyFans privacy"`
}
