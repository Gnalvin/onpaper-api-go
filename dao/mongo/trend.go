package mongo

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	m "onpaper-api-go/models"
	"time"
)

// SaveTrendInfo 保存上传的动态信息
func SaveTrendInfo(trend m.SaveTrendInfo) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trendTable := Mgo.Collection("trend")
	_, err = trendTable.InsertOne(ctx, trend)
	return
}

// GetUserRecentlyTrendId 查询最近用户的动态
func GetUserRecentlyTrendId(userId string) (trendIds []int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trendTable := Mgo.Collection("trend")

	filter := bson.D{{"user_id", userId}, {"is_delete", false}}

	//最近30个
	var limit int64 = 30
	opts := options.FindOptions{
		Sort:       bson.M{"trend_id": -1},
		Limit:      &limit,
		Projection: bson.D{{"_id", 0}, {"trend_id", 1}},
	}

	cur, err := trendTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result bson.M
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		trendIds = append(trendIds, result["trend_id"].(int64))
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// GetMoreTrendInfo 获取多个动态信息
func GetMoreTrendInfo(trendIds []int64) (trendData m.TrendList, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trendTable := Mgo.Collection("trend")
	filter := bson.D{{"trend_id", bson.M{"$in": trendIds}}, {"state", 0}}

	opts := options.FindOptions{
		Projection: bson.D{
			{"_id", 0},
			{"trend_id", 1},
			{"comment", 1},
			{"whoSee", 1},
			{"forward_info", 1},
			{"pics.fileName", 1},
			{"pics.sort", 1},
			{"pics.width", 1},
			{"pics.height", 1},
			{"count", 1},
			{"text", 1},
			{"topic", 1},
			{"createAt", 1},
			{"user_id", 1},
			{"is_delete", 1}},
	}
	cur, err := trendTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.TrendShowInfo
		result.Type = "tr"
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		trendData = append(trendData, result)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// GetOneUserTrend 获取某个用户发过的所有动态
func GetOneUserTrend(userId string, nextId int64) (msgData []m.MongoFeed, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feedTable := Mgo.Collection("feed")
	filter := bson.D{{"accept_id", userId}, {"send_id", userId}}
	if nextId != 0 {
		filter = bson.D{{"accept_id", userId}, {"send_id", userId}, {"msg_id", bson.M{"$lt": nextId}}}
	}

	//最近30个
	var limit int64 = 30
	opts := options.FindOptions{
		Sort:       bson.M{"msg_id": -1},
		Limit:      &limit,
		Projection: bson.D{{"_id", 0}, {"accept_id", 0}},
	}
	cur, err := feedTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.MongoFeed
		err = cur.Decode(&result)
		if err != nil {
			return
		}

		msgData = append(msgData, result)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// VerifyTrendOwner 验证动态所有权
func VerifyTrendOwner(trendId int64, userId string) (isOwner bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feedTable := Mgo.Collection("feed")
	filter := bson.D{{"accept_id", userId}, {"send_id", userId}, {"msg_id", trendId}}

	var result bson.M
	err = feedTable.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		// 没有找到 说明不是他的动态
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		err = errors.Wrap(err, "FindChatId fail")
		return
	}
	// 作品动态不许通过这个接口删除
	tType := result["type"].(string)
	if tType == "aw" {
		return false, nil
	}
	isOwner = true
	return
}

// DeleteTrend 标记删除一条动态
func DeleteTrend(trendId int64) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trendTable := Mgo.Collection("trend")
	filter := bson.D{{"trend_id", trendId}}
	update := bson.D{
		{"is_delete", true},
		{"updateAt", time.Now()},
	}
	_, err = trendTable.UpdateOne(ctx, filter, bson.D{{"$set", update}})
	return
}

// UpdateTrendPermission 更新动态权限
func UpdateTrendPermission(permission m.TrendPermission) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trendTable := Mgo.Collection("trend")
	filter := bson.D{{"trend_id", permission.TrendId}}
	update := bson.D{{"$set", bson.D{
		{"comment", permission.Comment},
		{"whoSee", permission.WhoSee},
		{"updateAt", time.Now()},
	}}}
	_, err = trendTable.UpdateOne(ctx, filter, update)

	return
}

// GetTopicTrend 获取话题动态
func GetTopicTrend(topicId string, sortType string, page uint8) (result []m.TrendIdAndUserId, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{
		{"topic.topic_id", topicId},
		{"is_delete", false},
		{"state", 0},
		{"whoSee", "public"},
	}

	var sort bson.M
	if sortType == "hot" {
		sort = bson.M{"score": -1}
	} else {
		sort = bson.M{"trend_id": -1}
	}
	var limit int64 = 20
	var skip = int64(page) * limit
	opts := options.FindOptions{
		Sort:       sort,
		Skip:       &skip,
		Limit:      &limit,
		Projection: bson.D{{"_id", 0}, {"trend_id", 1}, {"user_id", 1}},
	}

	trendTable := Mgo.Collection("trend")

	cur, err := trendTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var temp m.TrendIdAndUserId
		err = cur.Decode(&temp)
		if err != nil {
			return
		}
		result = append(result, temp)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

func GetNewTrend(nextId int64) (msgData []m.MongoFeed, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feedTable := Mgo.Collection("feed")
	filter := bson.D{{"owner", true}}
	if nextId != 0 {
		filter = append(filter, bson.E{Key: "msg_id", Value: bson.M{"$lt": nextId}})
	}

	//最近30个
	var limit int64 = 30
	opts := options.FindOptions{
		Sort:       bson.M{"msg_id": -1},
		Limit:      &limit,
		Projection: bson.D{{"_id", 0}, {"accept_id", 0}},
	}
	cur, err := feedTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.MongoFeed
		err = cur.Decode(&result)
		if err != nil {
			return
		}

		msgData = append(msgData, result)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}
