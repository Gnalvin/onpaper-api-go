package models

import (
	"database/sql"
	"sort"
	"strconv"
	"time"
)

// UserTableInfo 用户表的字段
type UserTableInfo struct {
	SnowId   string         `json:"snowId" db:"snow_id"`
	UserName string         `json:"userName"  db:"userName"`
	Password string         `json:"-"  db:"password"`
	Avatar   string         `json:"avatar" db:"avatar_name"`
	Email    sql.NullString `json:"-" db:"email"`
	Phone    string         `json:"-" db:"phone"`
	Forbid   uint8          `json:"forbid" db:"forbid"` // 是否封号
}

// UserHomeProfile 用户首页资料
type UserHomeProfile struct {
	Owner   bool
	IsFocus uint8
	Profile UserProfileTableInfo
}

// UserProfileTableInfo 用户资料表字段
type UserProfileTableInfo struct {
	UserName    string       `json:"userName"  db:"username"`
	UserId      string       `json:"userId" db:"user_id"`
	Sex         string       `json:"sex" db:"sex"`
	Birthday    sql.NullTime `json:"birthday" db:"birthday"`
	Following   int64        `json:"following" db:"following"`
	WorkEmail   string       `json:"workEmail" db:"work_email"`
	Email       *string      `json:"email,omitempty" db:"email"`
	QQ          string       `json:"QQ" db:"QQ"`
	Weibo       string       `json:"Weibo" db:"Weibo"`
	Twitter     string       `json:"Twitter" db:"Twitter"`
	Pixiv       string       `json:"Pixiv" db:"Pixiv"`
	WeChat      string       `json:"WeChat" db:"WeChat"`
	Bilibili    string       `json:"Bilibili" db:"Bilibili"`
	Address     string       `json:"address" db:"address"`
	ExpectWork  string       `json:"expectWork" db:"expect_work"`
	Software    string       `json:"software" db:"software"`
	CreateStyle string       `json:"createStyle" db:"create_style"`
	Introduce   string       `json:"introduce" db:"introduce"`
	BannerName  string       `json:"bannerName" db:"banner_name"`
	AvatarName  string       `json:"avatarName" db:"avatar_name"`
	VTag        string       `json:"vTag" db:"v_tag"`
	VStatus     int8         `json:"vStatus" db:"v_status"`
	Commission  bool         `json:"commission" db:"commission"`
	CreateTime  time.Time    `json:"createTime" db:"createAt"`
	Count       UserAllCount `json:"count"`
}

// UserSimpleInfo 到数据库查评论用户信息类型
type UserSimpleInfo struct {
	UserId     string `json:"userId"  db:"user_id" bson:"userId"`
	UserName   string `json:"userName"  db:"username" bson:"userName,omitempty"`
	Avatar     string `json:"avatar" db:"avatar_name" bson:"avatar,omitempty"`
	VTag       string `json:"vTag" db:"v_tag" bson:"-"`
	VStatus    int8   `json:"vStatus" db:"v_status" bson:"-"`
	Commission bool   `json:"commission" db:"commission" bson:"-"`
}

type UserCount struct {
	Fans     int `json:"fans" db:"fans"`
	Likes    int `json:"likes" db:"likes"`
	Collects int `json:"collects" db:"collects"`
}

// UserAllCount 用户所有统计
type UserAllCount struct {
	Fans         int `json:"fans" db:"fans"`                  // 粉丝数
	Likes        int `json:"likes" db:"likes"`                // 被点赞数
	Collects     int `json:"collects" db:"collects"`          // 被收藏数
	Following    int `json:"following" db:"following"`        // 主动关注的用户数
	TrendCount   int `json:"trendCount" db:"trend_count"`     // 发布的动态数
	ArtCount     int `json:"artCount" db:"art_count"`         // 发布的作品数
	CollectCount int `json:"collectCount" db:"collect_count"` // 收藏的作品数
}

type UserAllCountAndId struct {
	UserId string `json:"-" db:"user_id"`
	UserAllCount
}

type UserSimpleInfoCount struct {
	UserSimpleInfo
	UserCount
}

// LoginForm 用户注册需要的数据
type LoginForm struct {
	SnowId     int64
	UserName   string
	Phone      string `json:"phone" binding:"required"`
	Password   string `json:"password"`
	VerifyCode string `json:"verifyCode" binding:"required"`
	InviteCode string `json:"inviteCode"`
	IsRegister bool
}

// SnsLinkData sns上传时的数据格式
type SnsLinkData struct {
	QQ       string `json:"QQ" db:"QQ"`
	Weibo    string `json:"Weibo" db:"Weibo"`
	Twitter  string `json:"Twitter" db:"Twitter"`
	Pixiv    string `json:"Pixiv" db:"Pixiv"`
	WeChat   string `json:"WeChat" db:"WeChat"`
	Bilibili string `json:"Bilibili" db:"Bilibili"`
}

// UpdateProfileData 更新用户上传资料需要的数据
type UpdateProfileData struct {
	ProfileType string      `json:"profileType" binding:"required"`
	Profile     string      `json:"profile"`
	SnsData     SnsLinkData `json:"snsData"`
}

// UserNavData 导航栏用户数据
type UserNavData struct {
	UserId        string `json:"userId" db:"user_id"`
	UserName      string `json:"userName" db:"username"`
	Avatar        string `json:"avatar" db:"avatar_name"`
	Banner        string `json:"banner" db:"banner_name"`
	Following     int    `json:"following" db:"following"`
	Fans          int    `json:"fans" db:"fans"`
	Likes         int    `json:"likes" db:"likes"`
	NotifyUnread  int    `json:"notifyUnread"`
	MessageUnread int    `json:"messageUnread"`
}

// HotUser 推荐用户数据
type HotUser struct {
	UserId   string         `json:"userId"`
	Avatar   string         `json:"avatar" `
	UserName string         `json:"userName" `
	Artworks []ArtworkCover `json:"artworks"`
	Count    UserAllCount   `json:"count"`
	Tags     string         `json:"tags"`
	IsFocus  bool           `json:"isFocus"`
	VTag     string         `json:"vTag" `
	VStatus  int8           `json:"vStatus" `
}

// VerifyUserFocus 发送的关注请求 需要的数据
type VerifyUserFocus struct {
	FocusId  string `json:"focusId" binding:"required"`
	IsCancel bool   `json:"isCancel"`
}

// UserArtActivity 用户最近的作品动态
type UserArtActivity struct {
	ArtworkId int64 `json:"artwork_id" db:"artwork_id"`
}

// UserSimpleInfoAndLink 头像昵称+sns链接
type UserSimpleInfoAndLink struct {
	UserSimpleInfo
	WorkEmail string `json:"workEmail" db:"work_email"`
	SnsLinkData
	Intro string `json:"intro"  db:"introduce"`
}

// UserInfoAndArtwork 用户和用户最近的作品类型
type UserInfoAndArtwork struct {
	UserSimpleInfo
	Intro    string          `json:"intro,omitempty"`
	IsFocus  uint8           `json:"isFocus"`
	Artworks []ArtSimpleInfo `json:"artworks"`
}

// UserBigCard 用户和用户最近的作品类型和链接
type UserBigCard struct {
	UserInfoAndArtwork
	WorkEmail string `json:"workEmail" db:"work_email"`
	SnsLinkData
	Score  string            `json:"score,omitempty"`
	Active string            `json:"active,omitempty"`
	Count  UserAllCountAndId `json:"count"`
}

type UserBigCardList []UserBigCard

func (u UserBigCardList) SortByUserIdDesc() {
	sort.Slice(u, func(i, j int) bool {
		userId1, _ := strconv.Atoi(u[i].UserId)
		userId2, _ := strconv.Atoi(u[j].UserId)
		return userId1 > userId2
	})
}

func (u UserBigCardList) SortByScoreDesc() {
	sort.Slice(u, func(i, j int) bool {
		score1, _ := strconv.ParseFloat(u[i].Score, 64)
		score2, _ := strconv.ParseFloat(u[j].Score, 64)
		return score1 > score2
	})
}

func (u UserBigCardList) SortByActiveDesc() {
	sort.Slice(u, func(i, j int) bool {
		artId1, _ := strconv.Atoi(u[i].Active)
		artId2, _ := strconv.Atoi(u[j].Active)
		return artId1 > artId2
	})
}

type UserSmallCard struct {
	UserId    string       `json:"userId"  db:"user_id"`
	UserName  string       `json:"userName"  db:"username"`
	Avatar    string       `json:"avatar" db:"avatar_name"`
	VTag      string       `json:"vTag" db:"v_tag"`
	VStatus   int8         `json:"vStatus" db:"v_status"`
	Introduce string       `json:"introduce" db:"introduce"`
	Count     UserAllCount `json:"count"`
	IsFocus   uint8        `json:"isFocus"`
}

// QueryUserRank 查询作品排序 需要传递的数据
type QueryUserRank struct {
	RankType string `form:"type" binding:"oneof=new girl boy like collect "`
}

// VerifyEmail 邮箱验证需要的参数
type VerifyEmail struct {
	Email string `form:"email" binding:"required"`
}

// VerifyPhone 手机验证需要的参数
type VerifyPhone struct {
	Phone string `form:"phone" binding:"required"`
	Code  string `form:"code"`
}

// VerifyUserId 验证用户id
type VerifyUserId struct {
	UserId string `form:"id" binding:"required,numeric,gt=0"`
}

type VerifySafetyCode struct {
	Code string `form:"code" binding:"required"`
}

// ChangeEmailForm 修改邮箱需要的表单
type ChangeEmailForm struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
	Token string `json:"token" binding:"required"`
}

// ChangePasswordForm 修改密码需要的表单
type ChangePasswordForm struct {
	Password string `json:"password" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

// ChangePhoneForm 修改手机绑定需要参数
type ChangePhoneForm struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
	Token string `json:"token" binding:"required"`
}

// UserPanel 用户面板信息
type UserPanel struct {
	UserId     string         `json:"userId" db:"user_id"`
	UserName   string         `json:"userName" db:"username"`
	Avatar     string         `json:"avatar" db:"avatar_name"`
	Banner     string         `json:"banner" db:"banner_name"`
	Collects   int            `json:"collects" db:"collects"`
	Fans       int            `json:"fans" db:"fans"`
	Likes      int            `json:"likes" db:"likes"`
	Intro      string         `json:"intro" db:"introduce"`
	VTag       string         `json:"vTag" db:"v_tag"`
	VStatus    int8           `json:"vStatus" db:"v_status"`
	Commission bool           `json:"commission" db:"commission"`
	HavePlan   bool           `json:"havePlan" db:"have_plan"`
	Rating     float64        `json:"rating" db:"rating"`
	IsOwner    bool           `json:"isOwner"`
	Artworks   []ArtworkCover `json:"artworks"`
}

// UserIsFocus 用户是否关注
type UserIsFocus struct {
	UserId  string `json:"userId" db:"user_id"`
	IsFocus uint8  `json:"isFocus" db:"is_focus"`
}

type InvitationCode struct {
	UserId   string `db:"used" json:"userId"`
	UserName string `db:"userName" json:"userName"`
	Avatar   string `db:"avatar" json:"avatar"`
	Code     string `db:"code" json:"code"`
}

// PostInteractData 点赞收藏时上传的数据
type PostInteractData struct {
	AuthorId string `json:"authorId" binding:"required"`
	MsgId    string `json:"msgId" binding:"required"`
	IsCancel bool   `json:"isCancel"`
	Type     string `json:"type" binding:"oneof=aw tr"`
	Force    bool   `json:"force"`
	Action   string
}

// PostFeedback 提交的反馈内容
type PostFeedback struct {
	UserId       string    `bson:"userId"`
	FeedbackType string    `json:"type" bson:"type" binding:"oneof=suggest bug use feature"`
	Describe     string    `json:"describe" bson:"describe" binding:"required"`
	Contact      string    `json:"contact" bson:"contact,omitempty"`
	CreateAt     time.Time `bson:"createAt"`
}

// InitUserData 初始化用户收藏点赞的数据
type InitUserData struct {
	MsgId string    `bson:"msg_id"`
	Time  time.Time `bson:"updateAt"`
}

// AllUserShowQuery 全站用户展示查询参数
type AllUserShowQuery struct {
	Type string `form:"type" binding:"oneof=hot new active"`
	Next string `form:"next" binding:"numeric"`
}

// BigCardUserId 用户id 和 分数
type BigCardUserId struct {
	UserId string `db:"user_id"`
	Score  string `db:"score"`
	Active string `db:"createAT"`
}
