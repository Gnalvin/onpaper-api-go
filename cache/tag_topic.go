package cache

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	m "onpaper-api-go/models"
	"time"
)

// SetTagRelevant 设置相关tag 标签缓存
func SetTagRelevant(data m.TagRelevant, tagId string) (err error) {
	key := fmt.Sprintf(TagRelevant, tagId)
	err = SetOneStringValue(key, data, time.Hour*2)
	if err != nil {
		err = errors.Wrap(err, "SetTagRelevant fail")
	}
	return
}

// SetTagUser 设置相关tag 用户缓存
func SetTagUser(data []m.UserBigCard, tagId string) (err error) {
	key := fmt.Sprintf(TagUser, tagId)
	err = SetOneStringValue(key, data, time.Hour*12)
	if err != nil {
		err = errors.Wrap(err, "SetTagUser fail")
	}
	return
}

// SetTagArtId  设置相关tag 作品id的缓存
func SetTagArtId(tagId, page, sort string, data []m.ArtIdAndUid) (err error) {
	key := fmt.Sprintf(TagArtworkAndPage, tagId, page+"&"+sort)

	err = SetOneStringValue(key, data, time.Hour*12)
	if err != nil {
		err = errors.Wrap(err, "SetTagArtId fail")
	}
	return
}

// SetHotTag 设置热门tag 缓存
func SetHotTag(rankData []m.HotTagRank) (err error) {
	key := fmt.Sprintf(RankTag, "hours")
	err = SetOneStringValue(key, rankData, 3*time.Hour)
	if err != nil {
		err = errors.Wrap(err, "setHotTag fail")
	}
	return
}

// SetHotTopicIncr 给topic hot排名加分
func SetHotTopicIncr(topic string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf(RankTopicHot)
	err = Rdb.ZIncrBy(ctx, key, 1, topic).Err()
	if err != nil {
		err = errors.Wrap(err, "SetHotTopicIncr ZIncrBy Cache Fail")
	}
	return
}

// SetHotTopic 设置热门topic 缓存
func SetHotTopic(rankData []m.HotTopicRank) (err error) {
	key := fmt.Sprintf(RankTopic, "hours")
	err = SetOneStringValue(key, rankData, 3*time.Hour)
	if err != nil {
		err = errors.Wrap(err, "setHotTag fail")
	}
	return
}

// SetTopicDetail 设置话题详情
func SetTopicDetail(detail m.TopicDetail) (err error) {
	key := fmt.Sprintf(TopicProfile, detail.TopicId)
	err = SetOneStringValue(key, detail, 1*time.Hour)
	if err != nil {
		err = errors.Wrap(err, "SetTopicDetail fail")
	}
	return
}
