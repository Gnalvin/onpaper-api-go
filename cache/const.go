package cache

const WxAccessToken = "wx:token" // 微信服务器接口调用凭证

const ServeUserId = "serve:uid"
const CompressStreamName = "compress" // 图片压缩流的名称

const NotifyConfig = "notify:config:%s" // 用户通知配置

const ActiveDay = "active:day:%s"     // 日活统计
const ActiveMonth = "active:month:%s" // 月活统计
const ActiveTime = "active:time:user" // 用户最后活跃时间

const AuthPhone = "auth:phone:%s" // 手机验证码
const AuthEmail = "auth:email:%s" // 邮箱验证码
const AuthToken = "auth:token:%s" // 保存token 有效性

const TrendProfile = "trend:profile:%s"     // 动态详情
const ArtworkProfile = "artwork:profile:%s" // 作品详情
const UserProfile = "user:profile:%s"       // 用户资料详情
const TopicProfile = "topic:profile:%s"     // 话题详情
const ArtworkBasic = "artwork:basic:%s"     // 作品封面基本展示信息
const UserPanel = "user:panel:%s"           // 用户简单资料面板
const UserBigCard = "user:bigCard:%s"       // 用户大卡片展示资料
const UserSmallCard = "user:smallCard:%s"   // 用户小卡片展示资料

const UserFollow = "user:follow:%s"   // 用户关注列表
const UserCollect = "user:collect:%s" // 用户收藏列表
const UserLike = "user:like:%s"       // 用户喜欢列表

const CommentRootList = "comment:rootList:%s"   // 根评论列表
const CommentChildList = "comment:childList:%s" // 某个根回复里面子回复列表
const CommentDetail = "comment:detail:%s"       // 评论详情

const ArtworkCount = "artwork:count:%s" // 作品点赞收藏转发统计
const TrendCount = "trend:count:%s"     // 动态点赞收藏转发统计
const UserCount = "user:count:%s"       // 用户数据统计
const ArtworkHLog = "artwork:HLog:%s"   // 作品浏览量
const CommentCount = "comment:count:%s" // 评论点赞统计

const TagRelevant = "tag:relevant:%s"         // tag相关的其他标签
const TagUser = "tag:user:%s"                 // tag相关用户
const TagArtworkAndPage = "tag:artwork:%s:%s" // tag 相关作品 按分页缓存

const HotTrendAll = "hot:trend:all"     // 热门动态列表
const HotArtworkAll = "hot:artwork:all" // 热门作品列表
const HotArtwork = "hot:artwork:%s"     // 单个热门作品资料
const HotUserAll = "hot:user:all"       // 热门用户列表
const HotUser = "hot:user:%s"           // 单个热门用户信息
const HotZone = "hot:zone:%s"           // 分区热门作品

const RankTag = "rank:tag:%s"         // tag排行榜
const RankUser = "rank:user:%s"       // 用户排行榜
const RankArtwork = "rank:artwork:%s" // 作品排行
const RankTopic = "rank:topic:%s"     // 话题排行榜
const RankTopicHot = "rank:topic:hot" // 发布话题时 热度+1
const RankTagHot = "rank:tag:hot"     // 发布作品 tag时 热度+1
const RankTopUseTag = "rank:tag:use"  // 使用最多的tag排行

const AcceptPlan = "plan:accept:%s" //接稿计划详情
const InvitePlan = "plan:invite:%d" // 邀请计划详情
