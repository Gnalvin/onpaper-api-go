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

// GetArtworkCollect 获取作品收藏数据
func GetArtworkCollect(userId string, page int) (artIds []m.MsgIdAndUid, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collectTable := Mgo.Collection("user_collect")

	filter := bson.D{{"user_id", userId}, {"is_cancel", false}}

	var limit int64 = 30
	var skip = int64(page) * limit
	opts := options.FindOptions{
		Limit:      &limit,
		Skip:       &skip,
		Sort:       bson.M{"updateAt": -1},
		Projection: bson.D{{"msg_id", 1}, {"author_id", 1}, {"_id", 0}},
	}

	cur, err := collectTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.MsgIdAndUid
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		artIds = append(artIds, result)
	}
	if err = cur.Err(); err != nil {
		return
	}

	return
}

// GetUserALlCollect 获取用户所有的收藏id
func GetUserALlCollect(userId string) (collectIds []m.InitUserData, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collectTable := Mgo.Collection("user_collect")
	filter := bson.D{{"user_id", userId}, {"is_cancel", false}}

	//只在缓存中保存 1000个
	var limit int64 = 1000
	opts := options.FindOptions{
		Sort:       bson.M{"updateAt": -1},
		Limit:      &limit,
		Projection: bson.D{{"msg_id", 1}, {"updateAt", 1}, {"_id", 0}},
	}

	cur, err := collectTable.Find(ctx, filter, &opts)

	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.InitUserData
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		collectIds = append(collectIds, result)
	}
	if err = cur.Err(); err != nil {
		return
	}

	return
}

// SetUserCollect 设置作品收藏
func SetUserCollect(userId string, cData m.PostInteractData) (isChange bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collectTable := Mgo.Collection("user_collect")
	filter := bson.D{{"user_id", userId}, {"msg_id", cData.MsgId}}

	update1 := bson.D{
		{"user_id", userId},
		{"msg_id", cData.MsgId},
		{"author_id", cData.AuthorId},
		{"type", cData.Type},
		{"is_cancel", cData.IsCancel},
	}
	update2 := bson.D{
		{"updateAt", time.Now()},
	}

	dataList := make([]mongo.WriteModel, 0, 2)
	dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(bson.D{{"$set", update1}}).SetUpsert(true))
	dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(bson.D{{"$set", update2}}).SetUpsert(true))
	// 即使一个错了 其他也插入
	opts := options.BulkWrite().SetOrdered(false)
	res, err := collectTable.BulkWrite(ctx, dataList, opts)
	if err != nil {
		err = errors.Wrap(err, "SetUserCollect: mongo set fail")
	}

	// 如有实际更新
	if res.MatchedCount == 2 && res.ModifiedCount == 2 {
		// 之前收藏过 然后改变
		isChange = true
	}
	if res.MatchedCount == 1 && res.UpsertedCount == 1 {
		// 之前没收藏过 新收藏
		isChange = true
	}

	return
}

// GetUserAllLike 获取用户所有的点赞过的id
func GetUserAllLike(userId string) (likeIds []m.InitUserData, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	likeTable := Mgo.Collection("user_like")
	filter := bson.D{{"user_id", userId}, {"is_cancel", false}}

	//只在缓存中保存 1000个
	var limit int64 = 1000
	opts := options.FindOptions{
		Sort:       bson.M{"updateAt": -1},
		Limit:      &limit,
		Projection: bson.D{{"msg_id", 1}, {"updateAt", 1}, {"_id", 0}},
	}

	cur, err := likeTable.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.InitUserData
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		likeIds = append(likeIds, result)
	}
	if err = cur.Err(); err != nil {
		return
	}

	return
}

// SetUserLike 设置作品点赞
func SetUserLike(userId string, lData m.PostInteractData) (isChange bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	likeTable := Mgo.Collection("user_like")

	filter := bson.D{{"user_id", userId}, {"msg_id", lData.MsgId}}

	update1 := bson.D{
		{"user_id", userId},
		{"msg_id", lData.MsgId},
		{"type", lData.Type},
		{"author_id", lData.AuthorId},
		{"is_cancel", lData.IsCancel},
	}
	update2 := bson.D{
		{"updateAt", time.Now()},
	}

	dataList := make([]mongo.WriteModel, 0, 2)
	dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(bson.D{{"$set", update1}}).SetUpsert(true))
	dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(bson.D{{"$set", update2}}).SetUpsert(true))
	// 即使一个错了 其他也插入
	opts := options.BulkWrite().SetOrdered(false)
	res, err := likeTable.BulkWrite(ctx, dataList, opts)
	if err != nil {
		err = errors.Wrap(err, "SetUserLike: mongo set fail")
	}

	// 如有实际更新
	if res.MatchedCount == 2 && res.ModifiedCount == 2 {
		// 之前点赞过 然后改变
		isChange = true
	}
	if res.MatchedCount == 1 && res.UpsertedCount == 1 {
		// 之前没点赞过 新点赞
		isChange = true
	}
	return
}
