package mongo

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	m "onpaper-api-go/models"
	tools "onpaper-api-go/utils/formatTools"
	"strconv"
	"time"
)

// SaveOneComment 添加一条评论
func SaveOneComment(cid int64, userId string, commentData m.PostCommentData) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("comment")

	comment := bson.D{
		{"cid", cid},
		{"own_id", commentData.OwnId},
		{"own_type", commentData.Type},
		{"user_id", userId},
		{"content", commentData.Text},
		{"reply_id", commentData.ReplyId},
		{"reply_user", commentData.ReplyUserId},
		{"root_id", commentData.RootId},
		{"root_count", 0},
		{"likes", 0},
		{"is_delete", false},
		{"createAt", time.Now()},
		{"updateAt", time.Now()},
	}

	dataList := make([]mongo.WriteModel, 0, 2)
	dataList = append(dataList, mongo.NewInsertOneModel().SetDocument(comment))
	// 如果是属于根回复下面的子回复 需要对根回复 +1
	if commentData.RootId != 0 {
		filter := bson.D{{"cid", commentData.RootId}}
		dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(bson.D{{"$inc", bson.M{"root_count": 1}}}))
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = table.BulkWrite(ctx, dataList, opts)
	if err != nil {
		err = errors.Wrap(err, "SaveOneComment: mongo set fail")
	}
	return
}

// GetRootCommentInfo 获取根评论详情
func GetRootCommentInfo(ownId string, cid int64) (rootComment []m.ReturnComment, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("comment")

	var limit int64 = 20
	opts := options.FindOptions{
		Sort:       bson.M{"cid": -1},
		Limit:      &limit,
		Projection: bson.D{{"_id", 0}, {"reply_id", 0}, {"reply_user", 0}, {"updateAt", 0}},
	}
	var cur *mongo.Cursor
	filter := bson.D{{"own_id", ownId}, {"root_id", 0}, {"is_delete", false}}
	if cid != 0 {
		filter = append(filter, bson.E{Key: "cid", Value: bson.M{"$lt": cid}})
	}

	cur, err = table.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.ReturnComment
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		rootComment = append(rootComment, result)
	}
	if err = cur.Err(); err != nil {
		return
	}

	return
}

// GetChildComment 获取作品的子评论
func GetChildComment(rootID int64, cid int64) (childComments []m.Comment, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("comment")

	var limit int64 = 20
	opts := options.FindOptions{
		Sort:       bson.M{"cid": 1},
		Limit:      &limit,
		Projection: bson.D{{"_id", 0}, {"root_count", 0}, {"updateAt", 0}},
	}

	var cur *mongo.Cursor
	filter := bson.D{{"root_id", rootID}, {"is_delete", false}}
	if cid != 0 {
		filter = append(filter, bson.E{Key: "cid", Value: bson.M{"$gt": cid}})
	}

	cur, err = table.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.Comment
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		childComments = append(childComments, result)
	}

	if err = cur.Err(); err != nil {
		return
	}

	return
}

// GetTwoChildComment 获取两个子评论
func GetTwoChildComment(rootIds []int64) (childComments []m.FirstShowChildComment, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("comment")

	matchStage := bson.D{{"$match", bson.D{{"root_id", bson.M{"$in": rootIds}}, {"is_delete", false}}}}
	//隐藏字段
	unsetStage := bson.D{{"$unset", bson.A{"_id", "root_count", "updateAt"}}}
	sortStage := bson.D{{"$sort", bson.D{{"cid", -1}}}}
	//分组
	groupStage := bson.D{
		{"$group", bson.D{{"_id", "$root_id"}, {"comment", bson.M{"$push": "$$ROOT"}}}},
	}
	// 每个根评论只要前面两条 重命名 _id
	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 0},
			{"rootId", "$_id"},
			{"comment", bson.M{"$slice": bson.A{"$comment", 2}}},
		}},
	}

	cur, err := table.Aggregate(ctx, mongo.Pipeline{matchStage, unsetStage, sortStage, groupStage, projectStage})
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.FirstShowChildComment
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		childComments = append(childComments, result)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// GetRootComment 获取根评论
func GetRootComment(ownId, cid string) (rootComment []m.ReturnComment, childComment []m.FirstShowChildComment, userIds []string, err error) {
	//获取前20个根评论信息
	intCid, _ := strconv.ParseInt(cid, 10, 64)
	rootComment, err = GetRootCommentInfo(ownId, intCid)
	if err != nil {
		err = errors.Wrap(err, "GetRootCommentInfo: mongo get fail")
		return
	}

	// 如果没有评论
	if len(rootComment) == 0 {
		return
	}
	// 需要查询的 所有评论id列表
	var cIdList []int64
	// 需要查询的 所有用户列表
	var userIdList []string
	// 存放 select语句的 切片
	for _, comment := range rootComment {
		cIdList = append(cIdList, comment.CId)
		userIdList = append(userIdList, comment.UserId)
	}

	childComment, err = GetTwoChildComment(cIdList)
	if err != nil {
		err = errors.Wrap(err, "GetTwoChildComment: mongo get fail")
		return
	}

	// cIdList 的顺序是 rootId 在前 childrenID 在后（因为先循环 rootComment 在循环 childComment ）
	// 将查询到的 子评论id保存
	for _, item := range childComment {
		for _, comment := range item.Comment {
			userIdList = append(userIdList, comment.UserId)
			if comment.ReplyId != 0 {
				//如果子评论是有回复其他人的 需要把其他人的 userID 也加入切片 待查询
				userIdList = append(userIdList, comment.ReplyUserId)
			}
		}
	}

	// 对 uidList 去重
	userIds, _ = tools.RemoveSliceDuplicate(userIdList)

	return
}

// GetCommentReply 获取根评论的子评论信息
func GetCommentReply(rootId, cid string) (childComment []m.Comment, userIds []string, err error) {
	//获取前20个子评论信息
	intCid, _ := strconv.ParseInt(cid, 10, 64)
	intRootId, _ := strconv.ParseInt(rootId, 10, 64)

	childComment, err = GetChildComment(intRootId, intCid)
	if err != nil {
		err = errors.Wrap(err, "GetCommentReply: mongo get fail")
		return
	}

	// 如果没有评论
	if len(childComment) == 0 {
		return
	}

	// 需要查询的 所有用户列表
	var uIds []string
	// 存放 select语句的 切片
	for _, item := range childComment {
		uIds = append(uIds, item.UserId)
	}

	//查询 发布评论的用户信息
	// 对 uidList 去重
	userIds, _ = tools.RemoveSliceDuplicate(uIds)

	return
}

// GetOneComment 获取一条评论
func GetOneComment(cid int64) (res m.Comment, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("comment")
	filter := bson.D{{"cid", cid}}
	opts := options.FindOneOptions{
		Projection: bson.D{{"_id", 0}},
	}
	err = table.FindOne(ctx, filter, &opts).Decode(&res)

	return
}

func DelOneComment(comment m.Comment, userId string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("comment")
	filter1 := bson.D{{"cid", comment.CId}}

	update := bson.D{
		{"is_delete", true},
		{"operation.deleted_at", time.Now()},
		{"operation.deleted_by", userId},
	}

	dataList := make([]mongo.WriteModel, 0, 2)
	dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter1).SetUpdate(bson.D{{"$set", update}}))

	// 如果是属于根回复下面的子回复 需要对根回复 -1
	if comment.RootId != 0 {
		filter2 := bson.D{{"cid", comment.RootId}}
		dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter2).SetUpdate(bson.D{{"$inc", bson.M{"root_count": -1}}}))
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err = table.BulkWrite(ctx, dataList, opts)
	if err != nil {
		err = errors.Wrap(err, "SaveOneComment: mongo set fail")
	}
	return
}

// BatchGetComment 批量获取评论
func BatchGetComment(cIds []int64) (commentMap map[int64]m.Comment, err error) {
	if len(cIds) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("comment")

	filter := bson.D{{"cid", bson.M{"$in": cIds}}}
	opts := options.FindOptions{
		Projection: bson.D{{"_id", 0}},
	}
	cur, err := table.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	commentMap = make(map[int64]m.Comment, 0)
	for cur.Next(ctx) {
		var result m.Comment
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		commentMap[result.CId] = result
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}
