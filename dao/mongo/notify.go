package mongo

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	m "onpaper-api-go/models"
	"time"
)

// SendNotify 发送通知
func SendNotify(notify m.NotifyBody) (isNew bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notifyTable := Mgo.Collection("notify")
	filter := bson.D{
		{"receiverId", notify.ReceiverId},
		{"action", notify.Action},
		{"sender.userId", notify.Sender.UserId},
		{"targetId", notify.TargetId},
	}
	update := bson.D{
		{"type", notify.Type},
		{"targetId", notify.TargetId},
		{"targetType", notify.TargetType},
		{"action", notify.Action},
		{"sender", notify.Sender},
		{"receiverId", notify.ReceiverId},
		{"updateAt", time.Now()},
	}
	if notify.Content != nil {
		update = append(update, bson.E{Key: "content", Value: notify.Content})
	}
	opt := options.Update().SetUpsert(true)
	res, err := notifyTable.UpdateOne(ctx, filter, bson.D{{"$set", update}}, opt)
	if err != nil {
		err = errors.Wrap(err, "mongo set like notify fail")
	}
	// 新增数据
	if res.UpsertedCount == 1 {
		isNew = true
	}
	return
}

// SendRepetitionNotify 可以重复的提醒
func SendRepetitionNotify(notify m.NotifyBody) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notifyTable := Mgo.Collection("notify")
	_, err = notifyTable.InsertOne(ctx, notify)

	return
}

// GetLikeAndCollectNotify 获取消息通知
func GetLikeAndCollectNotify(userId string, nextId string) (notify []m.NotifyBody, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notifyTable := Mgo.Collection("notify")

	//最近20个
	var limit int64 = 20
	opts := options.FindOptions{
		Sort:  bson.M{"_id": -1},
		Limit: &limit,
	}

	// 谁的通知
	filter := bson.D{{"receiverId", userId}}
	if nextId != "0" {
		id, _err := primitive.ObjectIDFromHex(nextId)
		if _err != nil {
			return make([]m.NotifyBody, 0), _err
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$lt": id}})
	}
	var orQuery []bson.M
	// 通知类型
	orQuery = append(orQuery, bson.M{"action": "like"}, bson.M{"action": "collect"})
	filter = append(filter, bson.E{Key: "$or", Value: orQuery})

	cur, err := notifyTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.NotifyBody
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		if result.TargetType == "cm" {
			var content m.LikeCommentNotify
			contentMap := result.Content.(primitive.D).Map()
			byt, _ := bson.Marshal(contentMap)
			err = bson.Unmarshal(byt, &content)
			if err != nil {
				return
			}
			result.Content = content
		}
		notify = append(notify, result)

	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// GetFocusNotify 获取关注提醒
func GetFocusNotify(userId string, nextId string) (notify []m.NotifyBody, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notifyTable := Mgo.Collection("notify")

	//最近30个
	var limit int64 = 20
	opts := options.FindOptions{
		Sort:  bson.M{"_id": -1},
		Limit: &limit,
	}
	// 谁的通知
	filter := bson.D{{"receiverId", userId}, {"action", "focus"}}
	if nextId != "0" {
		id, _err := primitive.ObjectIDFromHex(nextId)
		if _err != nil {
			return make([]m.NotifyBody, 0), _err
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$lt": id}})
	}
	cur, err := notifyTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.NotifyBody
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		notify = append(notify, result)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// GetCommentNotify 获取评论提醒
func GetCommentNotify(userId string, nextId string) (notify []m.CommentNotify, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notifyTable := Mgo.Collection("notify")

	//最近20个
	var limit int64 = 20
	opts := options.FindOptions{
		Sort:  bson.M{"_id": -1},
		Limit: &limit,
	}
	// 谁的通知
	filter := bson.D{{"receiverId", userId}, {"action", "comment"}}
	if nextId != "0" {
		id, _err := primitive.ObjectIDFromHex(nextId)
		if _err != nil {
			return make([]m.CommentNotify, 0), _err
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$lt": id}})
	}
	cur, err := notifyTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.CommentNotify
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		notify = append(notify, result)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// GetNotifyTrendInfo 获取提醒中的动态信息
func GetNotifyTrendInfo(trendIds []int64) (result []m.NotifyArtOrTrendInfo, err error) {
	if len(trendIds) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trendTable := Mgo.Collection("trend")

	filter := bson.D{{"trend_id", bson.M{"$in": trendIds}}}

	opts := options.FindOptions{
		Projection: bson.D{
			{"_id", 0},
			{"trend_id", 1},
			{"user_id", 1},
			{"pics", bson.M{"$slice": 1}},
			{"text", 1},
			{"is_delete", 1}},
	}

	cur, err := trendTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var temp bson.M
		err = cur.Decode(&temp)
		if err != nil {
			return
		}
		var r m.NotifyArtOrTrendInfo
		r.Id = fmt.Sprintf("%d", temp["trend_id"])
		r.IsDelete = temp["is_delete"].(bool)
		r.Author = temp["user_id"].(string)
		r.Text = temp["text"].(string)
		bStr := []rune(r.Text)
		if len(bStr) >= 50 {
			r.Text = string(bStr[0:45]) + "..."
		}
		picList, ok := temp["pics"].(primitive.A)
		if len(picList) != 0 && ok {
			r.Cover = picList[0].(primitive.M)["fileName"].(string)
		}

		result = append(result, r)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// DelFocusNotify 取消关注时撤回消息
func DelFocusNotify(receiver, sender string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notifyTable := Mgo.Collection("notify")

	filter := bson.D{{"receiverId", receiver}, {"action", "focus"}, {"sender.userId", sender}}
	_, err = notifyTable.DeleteOne(ctx, filter)
	return
}

// GetCommissionNotify 获取约稿通知
func GetCommissionNotify(userId string, nextId string) (notify []m.CommissionNotify, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notifyTable := Mgo.Collection("notify")

	//最近20个
	var limit int64 = 20
	opts := options.FindOptions{
		Sort:  bson.M{"_id": -1},
		Limit: &limit,
	}
	// 谁的通知
	filter := bson.D{{"receiverId", userId}, {"action", "update"}}
	if nextId != "0" {
		id, _err := primitive.ObjectIDFromHex(nextId)
		if _err != nil {
			return make([]m.CommissionNotify, 0), _err
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$lt": id}})
	}
	cur, err := notifyTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.CommissionNotify
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		notify = append(notify, result)
	}
	if err = cur.Err(); err != nil {
		return
	}

	return
}
