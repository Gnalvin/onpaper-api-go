package models

// TagQueryArtParam Tag 查询作品的参数
type TagQueryArtParam struct {
	TagName string `form:"tag" binding:"required"`
	TagId   string `form:"query" binding:"required,numeric,gt=0"`
	Sort    string `form:"sort" binding:"oneof=score time"`
	Page    uint16 `form:"page" binding:"required,lte=10"`
}

// TagQueryParam tag相关查询
type TagQueryParam struct {
	TagName string `form:"tag" binding:"required"`
	TagId   string `form:"query" binding:"required,numeric,gt=0"`
}

// ArtworkTag 作品tag
type ArtworkTag struct {
	TagName string `json:"tagName"  db:"tag_name"`
	TagId   string `json:"tagId" db:"tag_id" `
}

// TagRelevant tag相关tags 数据
type TagRelevant struct {
	TagName string       `json:"tagName" db:"tag_name"`
	Tags    []ArtworkTag `json:"tags"`
	Total   int          `json:"total" db:"art_count"`
}

// SearchTagResult 标签搜索的结果
type SearchTagResult struct {
	ArtworkTag
	ArtCount int16 `json:"artCount" db:"art_count"`
}

// HotTagRank 热门标签数据
type HotTagRank struct {
	ArtworkTag
	Status string `json:"status" db:"status"`
}

// TopicQueryParam 查询话题的参数
type TopicQueryParam struct {
	TopicName string `form:"topic" binding:"required"`
	TopicId   string `form:"id" binding:"required,numeric,gt=0"`
}

// TopicQueryTrendParam 查询话题相关动态
type TopicQueryTrendParam struct {
	Sort    string `form:"sort" binding:"oneof=new hot"`
	TopicId string `form:"id" binding:"required,numeric,gt=0"`
	Page    uint8  `form:"page" binding:"required,numeric,min=1,max=26"`
}

type TrendIdAndUserId struct {
	TrendId int64  `bson:"trend_id"`
	UserId  string `bson:"user_id"`
}

// TopicDetail 话题详情
type TopicDetail struct {
	TopicId string `json:"topicId" db:"topic_id"`
	Text    string `json:"text" db:"text"`
	Intro   string `json:"intro" db:"intro"`
	Count   int    `json:"count" db:"trend_count"`
	UserSimpleInfo
}

// HotTopicRank 热门话题数据
type HotTopicRank struct {
	TopicName string `json:"topicName"  db:"topic_name"`
	TopicId   string `json:"topicId" db:"topic_id" `
	Status    string `json:"status" db:"status"`
	Count     int    `json:"count" db:"trend_count"`
}
