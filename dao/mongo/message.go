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

// FindChatId 查找会话id
func FindChatId(sender, receiver string) (chatId int64, isExits bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	table := Mgo.Collection("chat_relation")
	filter := bson.D{{"sender", sender}, {"receiver", receiver}}
	opts := options.FindOneOptions{
		Projection: bson.D{{"_id", 0}, {"chat_id", 1}},
	}
	var result bson.M
	err = table.FindOne(ctx, filter, &opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, false, nil
		}
		err = errors.Wrap(err, "FindChatId fail")
		return
	}

	chatId = result["chat_id"].(int64)
	isExits = true
	return
}

// SetChatRelation 设置会话关系
func SetChatRelation(msg m.MessageBody, isExits bool) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dataList := make([]mongo.WriteModel, 0, 3)
	chatTable := Mgo.Collection("chat_relation")

	//如果会话关系不存在 需要先建立关系
	if !isExits {
		// 发送者的会话列表
		sender := bson.D{
			{"sender", msg.Sender},
			{"receiver", msg.Receiver},
			{"chat_id", msg.ChatId},
			{"unread", 0},
			{"last_msg", msg.MsgId},
		}
		// 接收者的会话列表
		receiver := bson.D{
			{"sender", msg.Receiver},
			{"receiver", msg.Sender},
			{"chat_id", msg.ChatId},
			{"unread", 0},
			{"last_msg", msg.MsgId},
		}

		dataList = append(dataList, mongo.NewInsertOneModel().SetDocument(sender))
		dataList = append(dataList, mongo.NewInsertOneModel().SetDocument(receiver))
	}
	// 发送者最后一条消息
	filter1 := bson.D{{"sender", msg.Sender}, {"receiver", msg.Receiver}}
	dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter1).SetUpdate(bson.D{{"$set", bson.D{{"last_msg", msg.MsgId}}}}))

	// 接受者添加一条未读和最后一条消息
	filter2 := bson.D{{"sender", msg.Receiver}, {"receiver", msg.Sender}}
	dataList = append(dataList, mongo.NewUpdateOneModel().SetFilter(filter2).SetUpdate(bson.D{{"$set", bson.D{{"last_msg", msg.MsgId}}}, {"$inc", bson.M{"unread": 1}}}))

	// 按顺序执行
	opts := options.BulkWrite().SetOrdered(true)
	_, err = chatTable.BulkWrite(ctx, dataList, opts)
	if err != nil {
		err = errors.Wrap(err, "SetChatRelation fail")
	}
	return
}

// SaveMsg 保存聊天记录
func SaveMsg(msg m.MessageBody) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 聊天记录表
	recordTable := Mgo.Collection("chat_records")

	_, err = recordTable.InsertOne(ctx, msg)
	if err != nil {
		err = errors.Wrap(err, "SaveMsg fail")
	}
	return
}

// GetChatList 获取会话列表
func GetChatList(userId string, nextId int64) (chatList []m.ChatRelation, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chatTable := Mgo.Collection("chat_relation")

	// 匹配字段
	match := bson.D{{"sender", userId}}
	if nextId != 0 {
		match = bson.D{{"sender", userId}, {"last_msg", bson.D{{"$lt", nextId}}}}
	}
	matchStage := bson.D{{"$match", match}}
	//排序字段
	sortStage := bson.D{{"$sort", bson.D{{"last_msg", -1}}}}
	//切割
	limitStage := bson.D{{"$limit", 30}}
	// left jon
	lookUpStage := bson.D{
		{"$lookup",
			bson.D{
				{"from", "chat_records"},
				{"localField", "last_msg"},
				{"foreignField", "msg_id"},
				{"as", "message"},
			},
		},
	}
	// 隐藏字段
	projectStage := bson.D{
		{"$project",
			bson.D{
				{"_id", 0},
				{"last_msg", 0},
				{"message._id", 0},
				{"message.chat_id", 0},
			},
		},
	}

	cur, err := chatTable.Aggregate(ctx, mongo.Pipeline{matchStage, sortStage, limitStage, lookUpStage, projectStage})

	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.ChatRelation
		err = cur.Decode(&result)
		if err != nil {
			return
		}

		chatList = append(chatList, result)
	}
	if err = cur.Err(); err != nil {
		return
	}

	if len(chatList) == 0 {
		chatList = make([]m.ChatRelation, 0)
	}

	return
}

// GetChatRecord 获取聊天记录
func GetChatRecord(chatId, nextId int64) (messages []m.MessageBody, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	recordTable := Mgo.Collection("chat_records")

	//最近30个
	var limit int64 = 20
	opts := options.FindOptions{
		Sort:       bson.M{"msg_id": -1},
		Limit:      &limit,
		Projection: bson.D{{"_id", 0}},
	}

	var filter bson.D
	if nextId == 0 {
		filter = bson.D{{"chat_id", chatId}}
	} else {
		filter = bson.D{{"chat_id", chatId}, {"msg_id", bson.M{"$lt": nextId}}}
	}

	cur, err := recordTable.Find(ctx, filter, &opts)
	if err != nil {
		err = errors.Wrap(err, "GetChatRecord fail")
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result m.MessageBody
		err = cur.Decode(&result)
		if err != nil {
			return
		}

		messages = append(messages, result)
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}

// AckChatUnread 取消未读消息
func AckChatUnread(sender, receiver string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	relationTable := Mgo.Collection("chat_relation")
	filter := bson.D{{"sender", sender}, {"receiver", receiver}}
	update := bson.D{{"$set", bson.D{{"unread", 0}}}}
	_, err = relationTable.UpdateOne(ctx, filter, update)
	if err != nil {
		err = errors.Wrap(err, "AckChatUnread fail")
	}
	return
}

// GetUserUnreadCount 获取用户message 未读消息
func GetUserUnreadCount(receiver string) (result m.UserUnreadCount, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	relationTable := Mgo.Collection("chat_relation")
	matchStage := bson.D{
		{"$match", bson.D{
			{"sender", receiver},
		}},
	}
	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", nil},
			{"totalUnread", bson.D{{"$sum", "$unread"}}},
		},
		},
	}
	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 0},
			{"userId", "$_id.sender"},
			{"totalUnread", 1}},
		},
	}
	cur, err := relationTable.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		err = cur.Decode(&result)
		if err != nil {
			return
		}
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}
