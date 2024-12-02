package mysql

import (
	"fmt"
	"github.com/pkg/errors"
	m "onpaper-api-go/models"
)

// GetTagArtworkId 获取tag 对应的作品id
func GetTagArtworkId(tagId string, sort string, page uint16) (data []m.ArtIdAndUid, err error) {
	var sqlStr1 string

	if sort == "score" {
		sqlStr1 = `SELECT ac.artwork_id,ac.user_id FROM tag_artwork  as ta
				left join artwork_count as ac on ac.artwork_id = ta.artwork_id
				WHERE tag_id = ? and ac.is_delete = 0 and ac.whoSee = 'public' and ac.state = 0
				ORDER BY ac.score DESC
				LIMIT ?,36`
	} else {
		sqlStr1 = `SELECT ac.artwork_id,ac.user_id FROM tag_artwork  as ta
				left join artwork_count as ac on ac.artwork_id = ta.artwork_id
				WHERE tag_id = ?  and ac.is_delete = 0  and ac.whoSee = 'public' and ac.state = 0
				ORDER BY ta.createAt DESC
				LIMIT ?,36`
	}

	err = db.Select(&data, sqlStr1, tagId, page*36)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("GetTagArtworkId fail sort: %s ", sort))
	}

	return
}

// GetRelevantTags 查找相关标签
func GetRelevantTags(tagId string) (tags []m.ArtworkTag, err error) {
	// 查找 与 tagId 同时出现的次数最多的标签
	sqlStr1 := `SELECT tag_name,tag_id FROM (
				SELECT ta.tag_name,ANY_VALUE(ta.tag_id) as tag_id ,COUNT(*) as num FROM tag_artwork
				left join tag_artwork as ta on ta.artwork_id = tag_artwork.artwork_id
				WHERE tag_artwork.tag_id = ?
				GROUP BY ta.tag_name
				ORDER BY num DESC
				LIMIT 0,15
				) as temp
				WHERE tag_id != ?
				`

	err = db.Select(&tags, sqlStr1, tagId, tagId)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("GetRelevantTags fail tagId: %s ", tagId))
	}
	if len(tags) == 0 {
		tags = make([]m.ArtworkTag, 0, 0)
	}

	return
}

// GetRelevantUser 查找标签相关用户Id
func GetRelevantUser(tagId string) (userData []m.UserBigCard, err error) {

	sqlStr1 := `SELECT user_id FROM (SELECT user_id, SUM(ac.score) as score FROM tag_artwork as ta
			LEFT JOIN artwork_count as ac on ac.artwork_id = ta.artwork_id
			WHERE tag_id = ?
			GROUP BY user_id
			ORDER BY score DESC
			LIMIT 10) as temp`

	var userIds []string
	err = db.Select(&userIds, sqlStr1, tagId)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("GetRelevantUser fail tagId: %s ", tagId))
	}

	userData, err = BatchGetUserAllInfo(userIds, 5)
	if err != nil {
		err = errors.Wrap(err, "GetRelevantUser: BatchGetUserAllInfo get fail")
		return
	}

	return
}

// GetTagArtCount 获取tag 作品统计数
func GetTagArtCount(tagId string) (res m.TagRelevant, err error) {
	sqlStr1 := `select tag_name,art_count from tag where tag_id = ?`

	err = db.Get(&res, sqlStr1, tagId)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("GetTagArtCount fail tagName: %s ", tagId))
	}
	return
}

// GetTagHotRank 获取热门标签
func GetTagHotRank() (tagData []m.HotTagRank, err error) {
	// 查询最新一期 数据
	sqlStr1 := `SELECT tag_id,tag_name,status from rank_tag_day
             WHERE rank_date = (SELECT rank_date FROM rank_tag_day ORDER BY rank_date DESC LIMIT 1)
             order by score desc 
             `

	// 查询 数据库
	err = db.Select(&tagData, sqlStr1)
	if err != nil {
		err = errors.Wrap(err, "GetTagHotRank: sql1 get fail")
	}
	return
}

// GetTopUseTag 获取最多使用的tag
func GetTopUseTag() (tagData []m.SearchTagResult, err error) {
	sqlStr1 := `SELECT tag_id,tag_name,art_count from tag order by art_count desc  limit 20`
	// 查询 数据库
	err = db.Select(&tagData, sqlStr1)
	if err != nil {
		err = errors.Wrap(err, "GetTopUseTag: sql1 get fail")
	}
	return
}

// SearchTagName 按tagName 查找 like%
func SearchTagName(searchText string) (likeData, searchData []m.SearchTagResult, err error) {
	sqlStr1 := `SELECT tag_id,tag_name,art_count from tag where tag_name like ? and tag_name != ?
                order by art_count desc 
                limit 10 `
	// 查询 数据库
	err = db.Select(&likeData, sqlStr1, searchText+"%", searchText)
	if err != nil {
		err = errors.Wrap(err, "GetTagHotRank: sql1 get fail")
	}

	sqlStr2 := `SELECT tag_id,tag_name,art_count from tag where tag_name =? limit 1`
	err = db.Select(&searchData, sqlStr2, searchText)
	if err != nil {
		err = errors.Wrap(err, "GetTagHotRank: sql2 get fail")
	}
	return
}

// SearchRelevantTopic 模糊查询相关话题
func SearchRelevantTopic(searchText string) (topics []m.SearchTopicType, err error) {
	sqlStr1 := `select topic_id,text,trend_count from topic 
                where match(text) against(? in boolean mode)
                order by trend_count desc 
				limit 7
				`

	// 查询 数据库
	err = db.Select(&topics, sqlStr1, searchText+"*")
	if err != nil {
		err = errors.Wrap(err, "SearchRelevantTopic: sql1 get fail")
	}

	return
}

// GetTopicDetail 获取话题详情
func GetTopicDetail(topicId string) (detail m.TopicDetail, err error) {
	sql1 := `select topic_id,text,t.user_id,avatar_name,username,trend_count,intro from topic  as t
			LEFT JOIN user_profile as up
			on t.user_id = up.user_id
			where topic_id = ?`
	err = db.Get(&detail, sql1, topicId)
	if err != nil {
		err = errors.Wrap(err, "GetTopicDetail: sql1 get fail")
	}
	return
}

// GetTopicHotRank 获取热门话题
func GetTopicHotRank() (tagData []m.HotTopicRank, err error) {
	// 查询最新一期 数据
	sqlStr1 := `SELECT rtd.topic_id,topic_name,status,trend_count from rank_topic_day as rtd
             left join topic t on rtd.topic_id = t.topic_id                    
             WHERE rank_date = (SELECT rank_date FROM rank_topic_day ORDER BY rank_date DESC LIMIT 1)
             order by score desc 
             `

	// 查询 数据库
	err = db.Select(&tagData, sqlStr1)
	if err != nil {
		err = errors.Wrap(err, "GetTopicHotRank: sql1 get fail")
	}
	return
}
