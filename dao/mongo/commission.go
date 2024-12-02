package mongo

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	m "onpaper-api-go/models"
	"strconv"
	"time"
)

// SaveAcceptPlan 保存接稿方案
func SaveAcceptPlan(plan m.AcceptPlan) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("commission_accept")
	filter := bson.D{{"plan_id", plan.PlanId}}

	_, err = table.UpdateOne(ctx, filter, bson.D{{"$set", plan}}, options.Update().SetUpsert(true))
	if err != nil {
		err = errors.Wrap(err, "SaveAcceptPlan mongodb fail")
	}
	return
}

func GetAcceptPlan(userId string) (plan m.AcceptPlan, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("commission_accept")

	filter := bson.D{{"user_id", userId}}
	opts := options.FindOneOptions{
		Projection: bson.D{{"_id", 0}},
	}
	err = table.FindOne(ctx, filter, &opts).Decode(&plan)
	return
}

// SaveInvitePlan 保存约稿方案
func SaveInvitePlan(plan m.InvitePlan) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("commission_invite")

	_, err = table.InsertOne(ctx, plan)
	if err != nil {
		err = errors.Wrap(err, "SaveAcceptPlan mongodb fail")
	}
	return
}

// GetInvitePlanCard 获取用户约稿卡片信息
func GetInvitePlanCard(query m.PlanQuery, pType string) (plans []m.InvitePlanCard, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("commission_invite")

	var limit int64 = 10
	opts := options.FindOptions{
		Sort:  bson.M{"updateAt": -1},
		Limit: &limit,
		Projection: bson.D{
			{"_id", 0},
			{"artist_id", 1},
			{"user_id", 1},
			{"invite_id", 1},
			{"name", 1},
			{"intro", 1},
			{"date", 1},
			{"status", 1},
			{"money", 1},
			{"category", 1},
			{"file_list", 1},
			{"updateAt", 1},
		},
	}

	filter := bson.D{}
	if pType == "receive" {
		filter = append(filter, bson.E{Key: "artist_id", Value: query.UserId})
	} else {
		filter = append(filter, bson.E{Key: "user_id", Value: query.UserId})
	}

	if *query.NextId != 0 {
		t := time.Unix(*query.NextId/1000, 0).UTC()
		filter = append(filter, bson.E{Key: "updateAt", Value: bson.M{"$lt": t}})
	}
	// 小于0 要查询 -1 -2
	if query.Type < 0 {
		filter = append(filter, bson.E{Key: "status", Value: bson.M{"$lt": 0}}, bson.E{Key: "is_delete", Value: false})
	} else {
		filter = append(filter, bson.E{Key: "status", Value: query.Type}, bson.E{Key: "is_delete", Value: false})
	}

	var cur *mongo.Cursor
	cur, err = table.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	plans = make([]m.InvitePlanCard, 0)
	for cur.Next(ctx) {
		var result m.InvitePlanCard
		err = cur.Decode(&result)
		if err != nil {
			return
		}
		// 截取字符串长度
		if len([]rune(result.Intro)) > 80 {
			result.Intro = string([]rune(result.Intro)[:80])
		}
		// 只要封面
		if len(result.FileList) > 0 {
			for _, picsType := range result.FileList {
				if picsType.Sort == 0 {
					result.FileList = []m.PicsType{picsType}
					break
				}
			}
		}
		plans = append(plans, result)
	}
	if err = cur.Err(); err != nil {
		return
	}

	return
}

// GetPlanDetail 查找约稿方案详情
func GetPlanDetail(inviteId int64) (plan m.InvitePlan, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("commission_invite")
	filter := bson.D{{"invite_id", inviteId}}
	opts := options.FindOneOptions{
		Projection: bson.D{{"_id", 0}, {"contact", 0}, {"contact_type", 0}},
	}
	err = table.FindOne(ctx, filter, &opts).Decode(&plan)

	return
}

// GetPlanUserInfo 查找计划属于哪两个用户
func GetPlanUserInfo(inviteId int64) (userInfo m.PlanUserInfo, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("commission_invite")

	filter := bson.D{{"invite_id", inviteId}}

	opts := options.FindOneOptions{
		Projection: bson.D{
			{"_id", 0},
			{"artist_id", 1},
			{"user_id", 1},
			{"status", 1}},
	}

	err = table.FindOne(ctx, filter, &opts).Decode(&userInfo)
	return
}

func UpdatePlanStatus(userId string, planNext m.PlanNext) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var table *mongo.Collection
	table = Mgo.Collection("commission_invite")
	filter := bson.D{{"invite_id", planNext.InviteId}}

	set := bson.M{
		"status":   planNext.Status,
		"updateAt": time.Now(),
	}
	// 如果是关闭/完成方案 添加结束时间
	if planNext.Status < 0 || planNext.Status == 3 {
		set = bson.M{
			"status":   planNext.Status,
			"updateAt": time.Now(),
			"overAt":   time.Now()}
	}
	update := bson.M{
		"$push": bson.M{
			"operate": bson.M{
				"user_id": userId,
				"status":  planNext.Status,
				"time":    time.Now(),
			},
		},
		"$set": set,
	}

	_, err = table.UpdateOne(ctx, filter, update)

	return
}

// GetUserContact 获取约稿方和画师的联系方式
func GetUserContact(inviteId int64, ArtistId string) (artist, sender m.PlanContact, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inviteTable := Mgo.Collection("commission_invite")
	acceptTable := Mgo.Collection("commission_accept")

	filter := bson.D{{"invite_id", inviteId}}
	opts := options.FindOneOptions{
		Projection: bson.D{
			{"_id", 0},
			{"user_id", 1},
			{"contact", 1},
			{"contact_type", 1}},
	}
	err = inviteTable.FindOne(ctx, filter, &opts).Decode(&sender)
	if err != nil {
		return
	}
	filter = bson.D{{"user_id", ArtistId}}
	err = acceptTable.FindOne(ctx, filter, &opts).Decode(&artist)

	return
}

// BatchGetCommissionNotifyInfo 批量获取通知需要的信息
func BatchGetCommissionNotifyInfo(inviteIds []int64) (commissionMap map[string]m.NotifyCommissionInfo, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	table := Mgo.Collection("commission_invite")
	filter := bson.D{{"invite_id", bson.M{"$in": inviteIds}}}
	opts := options.FindOptions{
		Projection: bson.D{
			{"_id", 0},
			{"invite_id", 1},
			{"user_id", 1},
			{"name", 1},
			{"file_list", bson.M{"$slice": 1}}},
	}

	var cur *mongo.Cursor
	cur, err = table.Find(ctx, filter, &opts)
	if err != nil {
		return
	}
	defer cur.Close(ctx)

	commissionMap = make(map[string]m.NotifyCommissionInfo, 0)
	for cur.Next(ctx) {
		var temp bson.M
		err = cur.Decode(&temp)
		if err != nil {
			return
		}
		var r m.NotifyCommissionInfo
		r.InviteId = temp["invite_id"].(int64)
		r.Title = temp["name"].(string)
		r.Owner = temp["user_id"].(string)
		picList, ok := temp["file_list"].(primitive.A)
		if len(picList) != 0 && ok {
			r.Cover = picList[0].(primitive.M)["fileName"].(string)
		}
		commissionMap[strconv.FormatInt(r.InviteId, 10)] = r
	}
	if err = cur.Err(); err != nil {
		return
	}
	return
}
