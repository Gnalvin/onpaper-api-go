package mongo

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	m "onpaper-api-go/models"
	"time"
)

// SaveReport 保存举报
func SaveReport(report m.PostReport) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reportTable := Mgo.Collection("report")
	filter := bson.D{{"msg_id", report.MsgId}, {"post_user", report.PostUser}}

	update := bson.D{
		{"msg_id", report.MsgId},
		{"msg_type", report.MsgType},
		{"defendant", report.Defendant},
		{"post_user", report.PostUser},
		{"report_type", report.ReportType},
		{"describe", report.Describe},
		{"status", 0},
		{"updateAt", time.Now()},
		{"createAt", time.Now()},
	}

	_, err = reportTable.UpdateOne(ctx, filter, bson.D{{"$set", update}}, options.Update().SetUpsert(true))
	if err != nil {
		err = errors.Wrap(err, "SaveReport fail")
	}
	return
}

// SaveFeedBack 保存反馈
func SaveFeedBack(feedback m.PostFeedback) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	table := Mgo.Collection("feedback")

	_, err = table.InsertOne(ctx, feedback)
	if err != nil {
		err = errors.Wrap(err, "SaveFeedBack mongodb fail")
	}
	return
}
