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

// SetTheUserFeed 添加指定用户feed
func SetTheUserFeed(msgIds []int64, feedType, sendId, acceptId string) (err error) {
	// 空直接返回
	if len(msgIds) == 0 {
		return
	}

	feeds := make([]mongo.WriteModel, 0, len(msgIds))
	for _, id := range msgIds {
		update := bson.D{
			{"accept_id", acceptId},
			{"msg_id", id},
			{"send_id", sendId},
			{"type", feedType},
			{"owner", sendId == acceptId}}
		//不要修改顺序 不然没有索引
		filter := bson.D{{"accept_id", acceptId}, {"send_id", sendId}, {"msg_id", id}}

		feeds = append(feeds, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(bson.D{{"$set", update}}).SetUpsert(true))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	feedTable := Mgo.Collection("feed")
	// 即使一个错了 其他也插入
	opts := options.BulkWrite().SetOrdered(false)
	_, err = feedTable.BulkWrite(ctx, feeds, opts)
	if err != nil {
		err = errors.Wrap(err, "SetTheUserFeed get fail")
	}
	return
}

// DelTheUserFeed 删除指定用户feed
func DelTheUserFeed(acceptId, sendId string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	//不要修改顺序 不然没有索引
	filter := bson.D{{"accept_id", acceptId}, {"send_id", sendId}}
	feedTable := Mgo.Collection("feed")
	_, err = feedTable.DeleteMany(ctx, filter)
	if err != nil {
		err = errors.Wrap(err, "DelRecentlyFeed get fail")
	}
	return
}

// GetFeed 获取feed流
func GetFeed(msgId int64, acceptId, feedType string) (msgData []m.MongoFeed, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feedTable := Mgo.Collection("feed")

	//最近30个
	var limit int64 = 30
	opts := options.FindOptions{
		Sort:       bson.M{"msg_id": -1},
		Limit:      &limit,
		Projection: bson.D{{"_id", 0}, {"accept_id", 0}},
	}

	var filter bson.D
	if msgId == 0 {
		if feedType == "all" {
			filter = bson.D{{"accept_id", acceptId}}
		} else {
			filter = bson.D{{"accept_id", acceptId}, {"type", feedType}}
		}
	} else {
		if feedType == "all" {
			filter = bson.D{{"accept_id", acceptId}, {"msg_id", bson.M{"$lt": msgId}}}
		} else {
			filter = bson.D{{"accept_id", acceptId}, {"msg_id", bson.M{"$lt": msgId}}, {"type", feedType}}

		}
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

// DeleteOneFeed 删除一条feed
func DeleteOneFeed(msgId int64, sendId, acceptId string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	feedTable := Mgo.Collection("feed")

	filter := bson.D{{"accept_id", acceptId}, {"send_id", sendId}, {"msg_id", msgId}}
	_, err = feedTable.DeleteOne(ctx, filter)
	if err != nil {
		err = errors.Wrap(err, "DeleteOneFeed fail")
		return
	}
	return
}

// SetFansFeed 给粉丝添加feed
func SetFansFeed(fans []string, msgId int64, sendId, feedType string) (err error) {
	// 空直接返回
	if len(fans) == 0 {
		return
	}

	feeds := make([]mongo.WriteModel, 0, len(fans))

	for _, id := range fans {
		update := bson.D{{"accept_id", id}, {"msg_id", msgId}, {"send_id", sendId}, {"type", feedType}}
		//不要修改顺序 不然没有索引
		filter := bson.D{{"accept_id", id}, {"send_id", sendId}, {"msg_id", id}}
		feeds = append(feeds, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(bson.D{{"$set", update}}).SetUpsert(true))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	feedTable := Mgo.Collection("feed")
	// 即使一个错了 其他也插入
	opts := options.BulkWrite().SetOrdered(false)
	_, err = feedTable.BulkWrite(ctx, feeds, opts)
	if err != nil {
		err = errors.Wrap(err, "SetFansFeed get fail")
	}

	return
}
