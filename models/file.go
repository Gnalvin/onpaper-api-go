package models

import (
	"database/sql"
	"time"
)

// PicsType 上传的文件信息
type PicsType struct {
	FileName string `json:"fileName" bson:"fileName"`
	Sort     uint8  `json:"sort" bson:"sort"`
	Width    uint16 `json:"width" bson:"width"`
	Height   uint16 `json:"height" bson:"height"`
	Mimetype string `json:"-" bson:"type"`
	Size     int64  `json:"-" bson:"size"`
}

// FileTableInfo  数据库中图片的文件信息
type FileTableInfo struct {
	Mimetype string         `db:"mimetype"`
	FileName sql.NullString `db:"filename"`
	Size     int64          `db:"size"`
	UserId   int64          `db:"user_id"`
}

// DeleteBannerInfo 删除背景时 删除post 传递的参数
type DeleteBannerInfo struct {
	FileName string `json:"bannerName"`
}

// CallBackFileInfo 上传的文件信息
type CallBackFileInfo struct {
	FileName string `json:"fileName" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Size     int64  `json:"size" binding:"required"`
}

// CallBackArtworkInfo 上传作品时上传的数据格式
type CallBackArtworkInfo struct {
	Title       string       `json:"title" binding:"required"`
	Description string       `json:"description"`
	FileList    []uploadFile `json:"fileList" binding:"required"`
	Zone        string       `json:"zone" binding:"required"`
	WhoSee      string       `json:"whoSee" binding:"required"`
	Tags        []string     `json:"tags" binding:"required"`
	Adult       bool         `json:"adult" `
	Cover       string       `json:"cover" binding:"required"`
	Comment     string       `json:"comment" binding:"oneof=public onlyFans close"` //评论权限
	CopyRight   string       `json:"copyRight" binding:"required"`
	Device      string       `json:"device" binding:"oneof=PC WeChat"`
}

// CallBackTrendInfo 上传动态时上传的数据格式
type CallBackTrendInfo struct {
	Text        string              `json:"text"`
	FileList    []uploadFile        `json:"fileList"`
	WhoSee      string              `json:"whoSee" binding:"required"`
	Topic       TopicType           `json:"topic" binding:"required"`
	Comment     string              `json:"comment" binding:"oneof=public onlyFans close"` //评论权限
	ForwardInfo CallBackForwardInfo `json:"forwardInfo" binding:"required"`
}

type CallBackForwardInfo struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type uploadFile struct {
	FileName string
	Sort     uint8
	Width    uint16
	Height   uint16
}

// SaveArtworkInfo 需要保存的作品信息
type SaveArtworkInfo struct {
	ArtworkId   int64
	UserId      string
	Title       string
	Description string
	FileList    []*PicsType
	FirstPic    string
	Tags        []string
	Zone        string
	WhoSee      string
	Adults      bool
	Cover       string
	Comment     string
	CopyRight   string
	Device      string
}

// SaveTrendInfo 保存trend需要的信息
type SaveTrendInfo struct {
	TrendId     int64       `json:"trendId,string" bson:"trend_id"`
	UserId      string      `json:"userId" bson:"user_id"`
	Text        string      `json:"text" bson:"text"`
	ForwardInfo ForwardInfo `json:"ForwardInfo" bson:"forward_info"`
	Topic       TopicType   `json:"topic" bson:"topic"`
	Pics        []PicsType  `json:"pics" bson:"pics"`
	Comment     string      `json:"comment" bson:"comment"` //评论权限
	Count       TrendCount  `json:"count" bson:"count"`
	WhoSee      string      `json:"whoSee" bson:"whoSee"`
	PicCount    uint8       `json:"picCount" bson:"pic_count"`
	UpdateAt    time.Time   `json:"updateAt" bson:"updateAt"`
	CreateAt    time.Time   `json:"createAt" bson:"createAt"`
	State       uint8       `json:"state" bson:"state"`
	IsDelete    bool        `json:"isDelete" bson:"is_delete"`
}
