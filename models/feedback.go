package models

type PostReport struct {
	MsgId      int64  `json:"msgId,string" binding:"required"`               //举报的消息id
	MsgType    string `json:"msgType" binding:"oneof=aw tr cm usr ac in ev"` // 消息属于的类型 aw,tr,cm（评论）,usr(用户)，ac(接稿计划)，in(邀请计划)，ev(评价)
	ReportType string `json:"reportType" binding:"required"`
	Describe   string `json:"describe"`                     // 其他描述
	PostUser   string `json:"postUser"`                     // 提交人
	Defendant  string `json:"defendant" binding:"required"` // 被告人
}
