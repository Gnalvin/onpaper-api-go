package models

// MongoFeed 的数据结构
type MongoFeed struct {
	AcceptId string `bson:"accept_id"`
	MsgID    int64  `bson:"msg_id"`
	SendId   string `bson:"send_id"`
	Type     string `bson:"type"`
}

// VerifyNextId 翻页需要携带的 msgId
type VerifyNextId struct {
	NextId *int64 `form:"nextid" binding:"required"`
}

// UploadArtOrTrend 上传的作品或动态 feed
type UploadArtOrTrend struct {
	MsgID  int64
	SendId string
	Type   string
}
