package mysql

import (
	"database/sql"
	"fmt"
	"golang.org/x/sync/errgroup"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/snowflake"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// GetBannerInfo 查询背景图片信息
func GetBannerInfo(userId string) (fileInfo m.FileTableInfo, err error) {
	sqlStr := `SELECT user_id,filename,mimetype,size FROM banner WHERE user_id = ?;`
	// 查询 数据库
	err = db.Get(&fileInfo, sqlStr, userId)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "GetBannerInfo: sql get fail")
		return
	}

	return
}

// UpdateBannerInfo 更新背景图片信息
func UpdateBannerInfo(fileInfo m.CallBackFileInfo, userId string) (result sql.Result, err error) {

	// 开启一个事务
	tx, err := db.Begin()
	if err != nil {
		err = errors.Wrap(err, "transaction begin failed")
		return
	}

	// 函数关闭时 如果出错 则回滚，没出错则 提交
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			return
		}
	}()
	// size不转 str 阿里云服务器MySQL会出错
	sizeStr := strconv.FormatInt(fileInfo.Size, 10)
	// 更新 banner 的数据
	sqlStr1 := `UPDATE banner set filename = ?,mimetype = ?,size = ? WHERE user_id = ?;`
	// 查询 数据库
	result, err = tx.Exec(sqlStr1, fileInfo.FileName, fileInfo.Type, sizeStr, userId)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "UpdateBannerInfo: sql get fail")
		return
	}

	// 更新 use_profile 数据
	sqlStr2 := `UPDATE user_profile set banner_name = ? WHERE user_id = ?;`
	// 查询 数据库
	result, err = tx.Exec(sqlStr2, fileInfo.FileName, userId)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "UpdateUseProfile: sql get fail")
		return
	}

	return
}

// DeleteBanner 删除背景信息
func DeleteBanner(userId string) (err error) {
	// 开启一个事务
	tx, err := db.Begin()
	if err != nil {
		err = errors.Wrap(err, "transaction begin failed")
		return
	}
	// 函数关闭时 如果出错 则回滚，没出错则 提交
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			return
		}
	}()

	sqlStr1 := `UPDATE user_profile set banner_name = '' WHERE user_id = ?;`
	_, err = tx.Exec(sqlStr1, userId)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "DeleteBanner: sql UPDATE fail")
		return
	}

	sqlStr2 := `UPDATE banner set filename = '',mimetype = '',size = 0 WHERE user_id = ?;`
	_, err = tx.Exec(sqlStr2, userId)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "DeleteBanner: sql reset fail")
		return
	}

	return
}

// UpdateAvatarInfo 更新头像图片信息
func UpdateAvatarInfo(fileInfo m.CallBackFileInfo, userId string) (err error) {
	// 开启一个事务
	tx, err := db.Begin()
	if err != nil {
		err = errors.Wrap(err, "transaction begin failed")
		return
	}

	// 函数关闭时 如果出错 则回滚，没出错则 提交
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			return
		}
	}()
	// size不转 str 阿里云服务器MySQL会出错
	sizeStr := strconv.FormatInt(fileInfo.Size, 10)
	// 更新 avatar表的数据
	sqlStr1 := `UPDATE avatar set filename = ?,mimetype = ?,size = ? WHERE user_id = ?;`
	// 查询 数据库
	_, err = tx.Exec(sqlStr1, fileInfo.FileName, fileInfo.Type, sizeStr, userId)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "UpdateAvatarInfo: sql get fail")
		return
	}

	// 更新 use_profile 数据
	sqlStr2 := `UPDATE user_profile set avatar_name = ? WHERE user_id = ?;`
	// 查询 数据库
	_, err = tx.Exec(sqlStr2, fileInfo.FileName, userId)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "UpdateUseProfile: sql get fail")
		return
	}

	return
}

// CreateArtworkInfo 创建作品信息 保存对应的图片记录
func CreateArtworkInfo(info *m.SaveArtworkInfo) (err error) {

	// 开启一个事务
	tx, err := db.Begin()
	if err != nil {
		err = errors.Wrap(err, "transaction begin failed")
		return
	}
	// 函数关闭时 如果出错 则回滚，没出错则 提交
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			return
		}
	}()

	//统计作品个数
	picCount := len(info.FileList)
	var eg errgroup.Group

	eg.Go(func() error {
		// 创建作品信息
		sqlStr1 := `INSERT INTO artwork (artwork_id,title,user_id,cover,zone,whoSee,pic_count,adults,comment,copyright,first_pic,device) 
				VALUES (?,?,?,?,?,?,?,?,?,?,?,?);`
		_, mErr := tx.Exec(sqlStr1,
			info.ArtworkId, info.Title, info.UserId, info.Cover, info.Zone, info.WhoSee, picCount,
			info.Adults, info.Comment, info.CopyRight, info.FirstPic, info.Device)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreateArtworkInfo sqlStr1 into fail")
		}
		return mErr
	})

	eg.Go(func() error {
		// 构造批量插入 图片数据的SQL语句
		// 存放 (?, ?) 的slice
		fileStrings := make([]string, 0, picCount)
		// 存放values的slice 几个参数 * 几
		fileArgs := make([]interface{}, 0, picCount*4)

		// 遍历users准备相关数据
		for _, file := range info.FileList {
			// 此处占位符要与插入值的个数对应
			fileStrings = append(fileStrings, "(?,?,?,?,?,?,?)")
			fileArgs = append(fileArgs, info.ArtworkId)
			fileArgs = append(fileArgs, file.FileName)
			fileArgs = append(fileArgs, file.Mimetype)
			fileArgs = append(fileArgs, file.Size)
			fileArgs = append(fileArgs, file.Sort)
			fileArgs = append(fileArgs, file.Width)
			fileArgs = append(fileArgs, file.Height)
		}

		// 自行拼接要执行的具体语句 INSERT INTO xx_table (c1,c2,c3) values (???),(???)...
		sqlStr2 := fmt.Sprintf("INSERT INTO artwork_picture (artwork_id,filename,mimetype,size,sort,width,height) VALUES %s",
			strings.Join(fileStrings, ","))
		_, mErr := tx.Exec(sqlStr2, fileArgs...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreateArtworkInfo sqlStr2 into fail")
		}
		return mErr
	})

	eg.Go(func() error {
		// 创建作品统计信息
		sqlStr3 := `INSERT INTO artwork_count (artwork_id,user_id,whoSee) VALUES (?,?,?);`
		_, mErr := tx.Exec(sqlStr3, info.ArtworkId, info.UserId, info.WhoSee)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreateArtworkInfo sqlStr3 into fail")
		}
		return mErr
	})

	eg.Go(func() error {
		// 创建作品描述信息
		sqlStr4 := `INSERT INTO art_intro (artwork_id,description) VALUES (?,?);`
		_, mErr := tx.Exec(sqlStr4, info.ArtworkId, info.Description)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreateArtworkInfo sqlStr4 into fail")
		}
		return mErr
	})

	// 构造批量插入 tags数据的SQL语句
	// 存放 (?, ?) 的slice
	tagStrings := make([]string, 0, len(info.Tags))
	// 存放values的slice 几个参数 * 几
	// tag表参数的存放list
	tagArgs := make([]interface{}, 0, len(info.Tags)*2)
	// tag_artwork 表的参数存放List
	tagArtArgs := make([]interface{}, 0, len(info.Tags)*2)

	// 遍历tags准备相关数据
	for _, tagName := range info.Tags {
		// 此处占位符要与插入值的个数对应
		tagStrings = append(tagStrings, "(?,?)")
		tagId := snowflake.CreateID()
		tagArgs = append(tagArgs, tagId)
		tagArgs = append(tagArgs, tagName)

		tagArtArgs = append(tagArtArgs, info.ArtworkId)
		tagArtArgs = append(tagArtArgs, tagName)
	}

	eg.Go(func() error {
		// 新增的tag插入 tag表创建数据
		// 自行拼接要执行的具体语句 INSERT INTO xx_table (c1,c2,c3) values (???),(???)...
		sqlStr5 := fmt.Sprintf("INSERT IGNORE INTO tag (tag_id, tag_name) VALUES %s",
			strings.Join(tagStrings, ","))
		_, mErr := tx.Exec(sqlStr5, tagArgs...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreateArtworkInfo sqlStr5 into fail")
			return mErr
		}

		//关联tag和作品表
		sqlStr6 := fmt.Sprintf("INSERT INTO tag_artwork (artwork_id, tag_name) VALUES %s",
			strings.Join(tagStrings, ","))
		_, mErr = tx.Exec(sqlStr6, tagArtArgs...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreateArtworkInfo sqlStr6 into fail")
			return mErr
		}

		// 把新添加的标签id 保存 到tag_artwork, 因为有些标签id之前存在过 就不会创建 所以得不到id 需要下面sql 同步
		sqlStr7 := `UPDATE tag_artwork as ta
				SET tag_id = (select tag_id FROM tag WHERE tag_name = ta.tag_name)
				WHERE artwork_id = ? `
		_, mErr = tx.Exec(sqlStr7, info.ArtworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreateArtworkInfo sqlStr7 into fail")
		}
		return mErr
	})

	eg.Go(func() error {
		// 在作品表中 对作品数+1
		sqlStr8 := `UPDATE user_count AS uc SET uc.art_count = uc.art_count + 1 WHERE user_id = ?`
		_, mErr := tx.Exec(sqlStr8, info.UserId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreateArtworkInfo sqlStr8 into fail")
		}
		return mErr
	})

	err = eg.Wait()
	// 记录创建日志
	logger.InfoZapLog("CreateArtworkInfo: creat finish", info.ArtworkId)
	return
}

// SaveTrendInfo 保存动态话题到数据库
func SaveTrendInfo(trend *m.SaveTrendInfo) (err error) {

	// 开启一个事务
	tx, err := db.Begin()
	if err != nil {
		err = errors.Wrap(err, "transaction begin failed")
		return
	}
	// 函数关闭时 如果出错 则回滚，没出错则 提交
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			return
		}
	}()

	sql1 := `UPDATE user_count AS uc SET uc.trend_count = uc.trend_count + 1 WHERE user_id = ?`
	_, err = tx.Exec(sql1, trend.UserId)
	if err != nil {
		err = errors.Wrap(err, "SaveTrendInfo sqlStr1 into fail")
		return
	}

	// 如果没有填写话题
	if trend.Topic.Text == "" {
		return
	}

	// 查看话题是否已经存在
	var needCrate bool
	var topic m.TopicType
	sql2 := `select topic_id,text from topic where text = ?`
	err = db.Get(&topic, sql2, trend.Topic.Text)
	if err != nil {
		// 如果不存在 创建话题信息
		if errors.Cause(err) == sql.ErrNoRows {
			topic.TopicId = strconv.FormatInt(snowflake.CreateID(), 10)
			topic.Text = trend.Topic.Text
			err = nil
			needCrate = true
		} else {
			err = errors.Wrap(err, "SaveTrendInfo sqlStr2 select fail")
			return
		}
	}
	trend.Topic = topic
	// 已经存在 不需要插入
	if !needCrate {
		return
	}

	// 插入到话题表 如果话题存在 不会插入成功
	sql3 := `INSERT IGNORE INTO topic (topic_id, text,user_id) VALUES (?,?,?)`
	_, err = tx.Exec(sql3, topic.TopicId, topic.Text, trend.UserId)
	if err != nil {
		err = errors.Wrap(err, "SaveTrendInfo sqlStr3 into fail")
		return
	}
	return
}
