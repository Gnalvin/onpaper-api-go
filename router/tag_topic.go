package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"
)

func tagAndTopicRouter(router *gin.Engine) {
	//	标签
	tag := router.Group("/tag", hm.VerifyAuth)
	// 查找tagId 对应的 作品
	tag.GET("/search/art", hm.HandleQueryTagArt, cm.GetTagArtwork, ctl.GetTagArtwork, cm.BatchSetBasicArt, cm.SetTagArtwork, cm.BatchSetArtViews)
	//查询tag 相关的tags
	tag.GET("/relevant", hm.HandleQueryTag, cm.GetRelevantTags, ctl.GetRelevantTags, cm.SetTagRelevant)
	//查询tag 相关的users
	tag.GET("/user", hm.HandleQueryTag, cm.GetTagUser, ctl.GetRelevantUsers, cm.SetTagUser)
	// 获取热门tag
	tag.GET("/hot", cm.GetHotTagRank, ctl.GetHotTagRank, cm.SetTagHotRank)
	// 获取最多使用的tag
	tag.GET("/top_use", cm.GetTopUseTagRank, ctl.GetTopUseTag, cm.SetTopUseTagRank)
	//精确搜索标签
	tag.GET("/search", hm.HandleQueryTag, ctl.GetLikeNameTag)

	// 话题
	topic := router.Group("/topic", hm.VerifyAuth)
	//模糊查找 话题
	topic.GET("/relevant", hm.HandleQueryTopic, ctl.GetRelevantTopic)
	// 查找话题先关动态
	topic.GET("/trend", hm.HandleQueryTopicTrend, ctl.GetTopicTrend, cm.BatchSetTrend)
	//话题详情
	topic.GET("/detail", hm.HandleQueryTopic, cm.GetTopicDetail, ctl.GetTopicDetail, cm.SetTopicDetail)
	// 获取热门topic
	topic.GET("/hot", cm.GetHotTopicRank, ctl.GetHotTopicRank, cm.SetHotTopicRank)
}
