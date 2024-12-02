package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	c "onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"onpaper-api-go/settings"
	"onpaper-api-go/utils/oss"
	"onpaper-api-go/utils/snowflake"
	"strconv"
	"time"
)

// SaveContractPlan 保存接稿计划
func SaveContractPlan(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("contractPlan")
	contractPlan := ctxData.(m.AcceptPlan)

	contractPlan.UserId = userInfo.Id
	contractPlan.UpdateAt = time.Now()
	contractPlan.CreateAt = time.Now()
	// 如果是编辑方案 已经存在 id
	if contractPlan.PlanId == 0 {
		contractPlan.PlanId = snowflake.CreateID()
	}

	err := mongo.SaveAcceptPlan(contractPlan)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	err = mysql.UpdateCommissionStatus(true, userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ctx.Next()
	ResponseSuccess(ctx, gin.H{
		"status": "ok",
	})
}

// SaveInvitePlan 保存接稿邀请
func SaveInvitePlan(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("invitePlan")
	invitePlan := ctxData.(m.InvitePlan)

	invitePlan.UserId = userInfo.Id
	invitePlan.InviteId = snowflake.CreateID()
	invitePlan.UpdateAt = time.Now()
	invitePlan.CreateAt = time.Now()

	err := mysql.UpdateCommissionCuntAndEvaluate(invitePlan.UserId, invitePlan.ArtistId, 0, 0, nil)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	err = mongo.SaveInvitePlan(invitePlan)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"status": "ok",
	})
	// 删除不需要的图片
	if len(invitePlan.FileList) > 0 {
		err = oss.BatchDeleteOssObject(settings.Conf.TempBucket, "commission/"+userInfo.Id+"/", "")
		if err != nil {
			logger.ErrZapLog(err, "SaveInvitePlan BatchDeleteCosObject fail")
		}
	}

	// 设置通知需要的数据
	ctx.Set("planNext", m.PlanNext{InviteId: invitePlan.InviteId})
	ctx.Set("planUser", m.PlanUserInfo{ArtistId: invitePlan.ArtistId, Sender: invitePlan.UserId})

}

// GetAcceptPlan 获取接稿方案
func GetAcceptPlan(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userId")
	queryId := ctxData.(string)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("cache")
	cache := ctxData.(c.CtxCacheVale)

	var res m.UserAcceptPlan
	if cache.HaveCache {
		res = cache.Val.(m.UserAcceptPlan)
		ctx.Abort()
	} else {
		plan, err := mongo.GetAcceptPlan(queryId)
		if err != nil {
			if err == mongodb.ErrNoDocuments {
				ResponseError(ctx, CodeUserNoAcceptPlan)
				return
			}
			err = errors.Wrap(err, "GetAcceptPlan mongodb fail")
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		userMap, err := mysql.GetBatchUserSimpleInfo([]string{queryId})
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		res = m.UserAcceptPlan{AcceptPlan: plan}
		if userInfo, ok := userMap[queryId]; ok {
			res.Avatar = userInfo.Avatar
			res.UserName = userInfo.UserName
			res.VTag = userInfo.VTag
			res.VStatus = userInfo.VStatus
			res.Status = userInfo.Commission
		}

		ctx.Set("plan", res)
	}

	// 如果停止接稿 而且不是登陆用户 不显示接稿内容
	if !res.Status && loginUser.Id != queryId {
		ResponseError(ctx, CodeUserStopCommission)
		return
	}

	// 如果不是登陆用户查询自己的方案
	if loginUser.Id != queryId {
		res.Contact = ""
		res.ContactType = ""
	}

	ResponseSuccess(ctx, res)
}

// GetInvitePlan 获取收到的邀请
func GetInvitePlan(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("query")
	query := ctxData.(m.PlanQuery)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	plans, err := mongo.GetInvitePlanCard(query, "receive")
	if err != nil {
		err = errors.Wrap(err, "GetInvitePlanCard mongodb fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//  当前登陆用户是查询的本人 检查已完成的订单是否评论了
	if (query.Type == 3 || query.Type < 0) && query.UserId == loginUser.Id {
		var checkIds []int64
		for _, plan := range plans {
			if plan.Status != -1 {
				checkIds = append(checkIds, plan.InviteId)
			}
		}
		var noEvaluate map[int64]struct{}
		noEvaluate, err = mysql.GetNoEvaluateID(checkIds, loginUser.Id)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		if len(noEvaluate) != 0 {
			for i, plan := range plans {
				if _, ok := noEvaluate[plan.InviteId]; ok {
					plans[i].NeedEvaluate = true
				}
			}
		}
	}

	ResponseSuccess(ctx, plans)
}

// GetSendPlan 获取发出的邀请
func GetSendPlan(ctx *gin.Context) {
	ctxData, _ := ctx.Get("query")
	query := ctxData.(m.PlanQuery)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	if query.UserId != loginUser.Id {
		ResponseError(ctx, CodeUnPermission)
		return
	}

	plans, err := mongo.GetInvitePlanCard(query, "send")
	if err != nil {
		err = errors.Wrap(err, "GetInvitePlanCard mongodb fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 已完成的订单查询 当前登陆用户是否评论了
	if query.Type == 3 || query.Type < 0 {
		var checkIds []int64
		for _, plan := range plans {
			if plan.Status != -1 {
				checkIds = append(checkIds, plan.InviteId)
			}
		}
		var noEvaluate map[int64]struct{}
		noEvaluate, err = mysql.GetNoEvaluateID(checkIds, loginUser.Id)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		if len(noEvaluate) != 0 {
			for i, plan := range plans {
				if _, ok := noEvaluate[plan.InviteId]; ok {
					plans[i].NeedEvaluate = true
				}
			}
		}
	}

	ResponseSuccess(ctx, plans)
}

// GetPlanDetail 获取约稿计划详情
func GetPlanDetail(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("inviteId")
	inviteId := ctxData.(int64)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("cache")
	cache := ctxData.(c.CtxCacheVale)

	var res m.UserInvitePlan
	if cache.HaveCache {
		res = cache.Val.(m.UserInvitePlan)
		ctx.Abort()
	} else {
		plan, err := mongo.GetPlanDetail(inviteId)
		if err != nil {
			if err == mongodb.ErrNoDocuments {
				ResponseError(ctx, CodeUserNoAcceptPlan)
				return
			}
			err = errors.Wrap(err, "GetPlanDetail mongodb fail")
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		res, err = mysql.GetUserCommissionScore(plan.UserId, plan.ArtistId)
		if err != nil {
			err = errors.Wrap(err, "GetPlanDetail mysql GetUserCommissionScore fail")
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		res.InvitePlan = plan

		// 完成的订单查询评价
		if res.Status == 3 || res.Status == -2 {
			evaluate, _err := mysql.GetPlanEvaluate(strconv.FormatInt(plan.InviteId, 10))
			if _err != nil {
				_err = errors.Wrap(_err, "GetPlanDetail mysql GetUserCommissionScore fail")
				ResponseErrorAndLog(ctx, CodeServerBusy, _err)
				return
			}
			res.Evaluates = evaluate
		}
	}
	//避免缓存的评论唯空
	temp := make([]m.Evaluate, 0)
	temp = res.Evaluates
	// 如果只有一个评价 说明有一方没有填写
	if len(res.Evaluates) == 1 {
		if res.Evaluates[0].InviteOwn == res.Evaluates[0].Sender {
			res.EvaluateStatus = 2 // 约稿者填写了
		} else {
			res.EvaluateStatus = 1 // 画师填写了
		}
		// 双方都完成评价才能显示, 先发表评论的可以看自己的
		if res.Evaluates[0].Sender != loginUser.Id {
			res.Evaluates = make([]m.Evaluate, 0)
		}
	}

	i := 0
	for _, e := range res.Evaluates {
		if !e.IsDelete {
			res.Evaluates[i] = e
			i++
		}
	}
	res.Evaluates = res.Evaluates[:i]

	ResponseSuccess(ctx, res)

	res.Evaluates = temp
	ctx.Set("plan", res)
}

// HandlePlanNext 处理约稿方案下一步
func HandlePlanNext(ctx *gin.Context) {
	ctxData, _ := ctx.Get("planNext")
	planNext := ctxData.(m.PlanNext)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("planUser")
	planUser := ctxData.(m.PlanUserInfo)

	// 更新status
	err := mongo.UpdatePlanStatus(loginUser.Id, planNext)
	if err != nil {
		err = errors.Wrap(err, "HandlePlanNext UpdatePlanStatus fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 更新计数
	err = mysql.UpdateCommissionCuntAndEvaluate(planUser.Sender, planUser.ArtistId, planNext.Status, planUser.NowStatus, nil)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{"status": "ok"})
	ctx.Set("inviteId", planNext.InviteId)
}

// GetUserContact 获取约稿双方联系方式
func GetUserContact(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("inviteId")
	inviteId := ctxData.(int64)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	// 查找约稿方案的两个用户
	userInfo, err := mongo.GetPlanUserInfo(inviteId)
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			ResponseError(ctx, CodeParamsError)
			return
		}
		err = errors.Wrap(err, "GetPlanUserInfo mongodb fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 如果 不是计划中的两个用户不能查看
	if userInfo.Sender != loginUser.Id && userInfo.ArtistId != loginUser.Id {
		ResponseError(ctx, CodeUnPermission)
		return
	}

	artist, sender, err := mongo.GetUserContact(inviteId, userInfo.ArtistId)
	if err != nil {
		err = errors.Wrap(err, "GetUserContact mongo fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	userMap, err := mysql.GetBatchUserSimpleInfo([]string{artist.UserId, sender.UserId})
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	artist.UserName = userMap[artist.UserId].UserName
	artist.Avatar = userMap[artist.UserId].Avatar
	sender.UserName = userMap[sender.UserId].UserName
	sender.Avatar = userMap[sender.UserId].Avatar

	ResponseSuccess(ctx, gin.H{
		"artist": artist,
		"sender": sender,
	})
}

func SaveEvaluate(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("evaluate")
	evaluate := ctxData.(m.Evaluate)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("planUser")
	planUser := ctxData.(m.PlanUserInfo)

	ctx.Set("inviteId", evaluate.InviteId)

	// 仅保存一条评论 不改变其他数据
	if evaluate.Only {
		evaluate.EvaluateId = snowflake.CreateID()
		err := mysql.SaveOneEvaluate(&evaluate)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		ResponseSuccess(ctx, gin.H{
			"inviteId": strconv.FormatInt(evaluate.InviteId, 10),
			"status":   evaluate.Status,
		})
		return
	}

	planNext := m.PlanNext{
		InviteId: evaluate.InviteId,
		Status:   evaluate.Status,
	}

	// 更新方案状态
	err := mongo.UpdatePlanStatus(loginUser.Id, planNext)
	if err != nil {
		err = errors.Wrap(err, "SaveEvaluate UpdatePlanStatus fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// mysql 保存评价和计数
	err = mysql.UpdateCommissionCuntAndEvaluate(planUser.Sender, planUser.ArtistId, evaluate.Status, planUser.NowStatus, &evaluate)
	if err != nil {
		err = errors.Wrap(err, "SaveEvaluate mysql.SaveEvaluate fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"inviteId": strconv.FormatInt(evaluate.InviteId, 10),
		"status":   evaluate.Status,
	})
	ctx.Set("planNext", planNext)
}

func GetUserReceiveEvaluate(ctx *gin.Context) {
	ctxData, _ := ctx.Get("query")
	query := ctxData.(m.EvaluateQuery)

	evaluate, err := mysql.GetUserReceiveEvaluate(query.UserId, query.Page-1)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	var userId []string
	for _, e := range evaluate {
		userId = append(userId, e.UserId)
	}

	userMap, err := mysql.GetBatchUserSimpleInfo(userId)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	for i, e := range evaluate {
		evaluate[i].UserName = userMap[e.UserId].UserName
		evaluate[i].Avatar = userMap[e.UserId].Avatar
	}

	ResponseSuccess(ctx, evaluate)
}

// UpdateCommissionStatus 更新约稿状态
func UpdateCommissionStatus(ctx *gin.Context) {
	ctxData, _ := ctx.Get("status")
	status := ctxData.(bool)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	err := mysql.UpdateCommissionStatus(status, loginUser.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{"status": status})
}

// CheckAcceptPermission 验证用户是否有权限开启约稿
func CheckAcceptPermission(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	ok, artCount, err := mysql.CheckCreatAcceptPermission(loginUser.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{"isOk": ok, "artCount": artCount})

}
