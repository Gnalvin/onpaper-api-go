package models

import "time"

// Artwork 作品信息
type Artwork struct {
	ArtworkId  string    `json:"artworkId"  db:"artwork_id"`
	Title      string    `json:"title"  db:"title"`
	UserId     string    `json:"userId"  db:"user_id"`
	PicCount   uint8     `json:"picCount" db:"pic_count"`
	Cover      string    `json:"cover" db:"cover"`
	Zone       string    `json:"zone"  db:"zone"`
	WhoSee     string    `json:"whoSee"  db:"whoSee"`
	Adults     bool      `json:"adults"  db:"adults"`
	ComSetting string    `json:"comSetting" db:"comment"`
	Copyright  string    `json:"copyright" db:"copyright"`
	Likes      int       `json:"likes" db:"likes"`
	Views      int       `json:"views" db:"views"`
	Comments   int       `json:"comments" db:"comments"`
	Collects   int       `json:"collects" db:"collects"`
	Forwards   int       `json:"forwards" db:"forwards"`
	IsDelete   bool      `json:"-"  db:"is_delete"`
	CreateAT   time.Time `json:"createAT"  db:"createAT"`
}

// ArtworkCount 作品统计信息
type ArtworkCount struct {
	ArtworkId string `json:"artworkId"  db:"artwork_id"`
	UserId    string `json:"userId" db:"user_id"`
	Likes     int    `json:"likes" db:"likes"`
	Views     int    `json:"views" db:"views"`
	Collects  int    `json:"collects" db:"collects"`
	Forwards  int    `json:"forwards" db:"forwards"`
	Comments  int    `json:"comments" db:"comments"`
	WhoSee    string `json:"whoSee,omitempty" db:"whoSee"`
}

type ArtSimpleInfo struct {
	ArtworkId      string    `json:"artworkId"  db:"artwork_id"`
	Title          string    `json:"title"  db:"title"`
	UserId         string    `json:"userId"  db:"user_id"`
	PicCount       int       `json:"picCount" db:"pic_count"`
	Cover          string    `json:"cover" db:"cover"`
	FirstPic       string    `json:"firstPic" db:"first_pic"`
	FirstPicWidth  uint16    `json:"width" db:"width"`
	FirstPicHeight uint16    `json:"height" db:"height"`
	Adults         bool      `json:"adults"  db:"adults"`
	WhoSee         string    `json:"whoSee" db:"whoSee"`
	IsDelete       bool      `json:"isDelete"  db:"is_delete"`
	IsOwner        bool      `json:"isOwner"`
	CreateAT       time.Time `json:"createAT"  db:"createAT"`
}

// BasicArtwork 作品基本的数据类型
type BasicArtwork struct {
	ArtSimpleInfo
	UserName   string `json:"userName"`
	UserAvatar string `json:"userAvatar"`
}

// ArtworkRank 作品排名返回的数据
type ArtworkRank struct {
	ArtworkId  string `json:"artworkId"  db:"artwork_id"`
	Title      string `json:"title"  db:"title"`
	UserId     string `json:"userId"  db:"user_id"`
	UserName   string `json:"userName"`
	UserAvatar string `json:"userAvatar"`
	PicCount   int    `json:"picCount" db:"pic_count"`
	Cover      string `json:"cover" db:"cover"`
	Adults     bool   `json:"adults"  db:"adults"`
	Likes      int    `json:"likes" db:"likes"`
	Collects   int    `json:"collects" db:"collects"`
}

// BasicArtworkAndLike 基本作品信息和点赞数
type BasicArtworkAndLike struct {
	BasicArtwork
	Likes  int  `json:"likes" db:"likes"`
	IsLike bool `json:"isLike"`
}

// ArtworkQueryPage 查询作品带分页
type ArtworkQueryPage struct {
	Artworks []BasicArtwork `json:"artworks"`
	Total    int16          `json:"total"`
}

// VerifyUserHomeArtwork 验证获取主页作品的接口参数
type VerifyUserHomeArtwork struct {
	UId  string `form:"id" binding:"required,numeric,gt=0"`
	Page int    `form:"page" binding:"gt=0"`
	Sort string `form:"sort" binding:"oneof=like collect now "`
}

// VerifyUserAndPage 写的查询用户和分页的 数据验证
type VerifyUserAndPage struct {
	UId  string `form:"uid" binding:"required,numeric,gt=0"`
	Page int    `form:"page" binding:"gt=0"`
}

type VerifyUserFollow struct {
	VerifyUserAndPage
	Type string `form:"type" binding:"oneof=follower following"`
}

// VerifyOneArtworkId 验证获取作品的query id
type VerifyOneArtworkId struct {
	ArtId string `form:"artid" binding:"required,numeric,gt=0"`
}

// ShowArtworkInfo 详细展示的作品信息
type ShowArtworkInfo struct {
	Intro string `json:"intro" db:"description"`
	Artwork
	UserName     string               `json:"userName"  db:"username"`
	AvatarName   string               `json:"avatarName" db:"avatar_name"`
	VTag         string               `json:"vTag"`
	VStatus      int8                 `json:"vStatus"`
	OtherArtwork []AuthorOtherArtwork `json:"otherArtwork"`
	Picture      []ArtworkPicture     `json:"picture"`
	Tag          []ArtworkTag         `json:"tag"`
	AuthorCount  UserCount            `json:"userCount"`
	Interact     UserArtworkInteract  `json:"interact"`
	IsOwner      bool                 `json:"isOwner"`
}

// ArtworkPicture 作品对应的图片信息
type ArtworkPicture struct {
	FileName string `json:"fileName"  db:"filename"`
	Sort     uint8  `json:"sort"  db:"sort"`
	Size     uint   `json:"size" db:"size"`
	Width    uint16 `json:"width" db:"width"`
	Height   uint16 `json:"height" db:"height"`
}

// AuthorOtherArtwork 作者的其他作品信息
type AuthorOtherArtwork struct {
	ArtworkId string `json:"artworkId"  db:"artwork_id"`
	Cover     string `json:"cover" db:"cover"`
}

// UserArtworkInteract 用户与作品的互动信息
type UserArtworkInteract struct {
	IsCollect     bool  `json:"isCollect"`
	IsFocusAuthor uint8 `json:"isFocusAuthor"`
	IsLike        bool  `json:"isLike"`
}

// QueryArtworkRank 查询作品排序 需要传递的数据
type QueryArtworkRank struct {
	RankType string `form:"type" binding:"oneof=today month week "`
}

// QueryChanelType 查询分类作品
type QueryChanelType struct {
	NextId string `form:"next" binding:"numeric,min=0"`
	Zone   string `form:"zone" binding:"required"`
	Sort   string `form:"sort" binding:"oneof=new hot"`
	Page   int    `form:"page" binding:"gt=0"`
}

type ArtIdAndUid struct {
	ArtworkId string `bson:"art_id" db:"artwork_id"`
	AuthorId  string `bson:"author_id" db:"user_id"`
}

type MsgIdAndUid struct {
	MsgId    string `bson:"msg_id"`
	AuthorId string `bson:"author_id"`
}

// HotArtworkData 热门作品数据结构
type HotArtworkData struct {
	ArtworkId string `json:"artworkId" db:"artwork_id"`
	Likes     int    `json:"likes" db:"likes"`
	Collects  int    `json:"collects" db:"collects"`
	UserId    string `json:"userId" db:"user_id"`
	UserName  string `json:"userName" db:"user_name"`
	Cover     string `json:"cover" db:"cover"`
	Avatar    string `json:"userAvatar" db:"avatar"`
	Title     string `json:"title" db:"title"`
	PicCount  int8   `json:"picCount" db:"pic_count"`
	FirstPic  string `json:"firstPic" db:"first_pic"`
	Width     uint16 `json:"width" db:"width"`
	Height    uint16 `json:"height" db:"height"`
	IsLike    bool   `json:"isLike"`
}

// UpdateArtInfo 更新作品的信息
type UpdateArtInfo struct {
	ArtworkId   string   `json:"artworkId" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"intro"`
	Zone        string   `json:"zone" binding:"required"`
	Tags        []string `json:"tags" binding:"required"`
	WhoSee      string   `json:"whoSee" binding:"required"`
	Adult       bool     `json:"adult"`
	Comment     string   `json:"comment" binding:"oneof=close public onlyFans"`
	CopyRight   string   `json:"copyRight" binding:"required"`
}

// ArtIntro 作品介绍
type ArtIntro struct {
	ArtworkId string `db:"artwork_id"`
	Intro     string `db:"description"`
}

// ArtPic 作品图片
type ArtPic struct {
	ArtworkId string `db:"artwork_id"`
	Pic       string `db:"filename"`
	Sort      uint8  `db:"sort"`
	Width     uint16 `db:"width"`
	Height    uint16 `db:"height"`
}

// ArtworkCover 作品封面数据
type ArtworkCover struct {
	ArtworkId string `json:"artworkId" db:"artwork_id"`
	UserId    string `json:"userId" db:"user_id"`
	Cover     string `json:"cover" db:"cover"`
}

// ZoneIndex 查询分区热门数据
type ZoneIndex struct {
	Zone int `form:"type" binding:"min=0,max=8"`
}
