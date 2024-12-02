package cache

import (
	"context"
	"fmt"
	"time"
)

// DelCommissionStatus 删除约稿状态相关缓存
func DelCommissionStatus(userId string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	panelKey := fmt.Sprintf(UserPanel, userId)
	profileKey := fmt.Sprintf(UserProfile, userId)
	planKey := fmt.Sprintf(AcceptPlan, userId)

	pipe := Rdb.Pipeline()

	pipe.Del(ctx, panelKey)
	pipe.Del(ctx, profileKey)
	pipe.Del(ctx, planKey)

	_, err = pipe.Exec(ctx)
	return
}

// DelInvitePlanStatus 删除约稿计划相关缓存
func DelInvitePlanStatus(invitedId int64) (err error) {
	key := fmt.Sprintf(InvitePlan, invitedId)
	err = DelOneCache(key)
	return
}
