package mysql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"onpaper-api-go/models"
	"onpaper-api-go/utils/snowflake"
	"strings"
	"time"
)

// GetUserHomeArtwork 数据库中查询 用户作品信息
func GetUserHomeArtwork(userId string, pageNum int, sortType string) (artworkCounts []models.ArtworkCount, err error) {
	var sqlStr1 string

	baseSql := `SELECT artwork_id,user_id,likes,whoSee from artwork_count WHERE user_id = ? and is_delete = 0 `
	switch sortType {
	case "now":
		sqlStr1 = baseSql + `ORDER BY artwork_id DESC LIMIT ?,30;`
	case "like":
		sqlStr1 = baseSql + `ORDER BY likes DESC LIMIT ?,30;`
	case "collect":
		sqlStr1 = baseSql + `ORDER BY collects DESC LIMIT ?,30;`
	}

	// 查询 数据库 pageNum 0 是第一页
	err = db.Select(&artworkCounts, sqlStr1, userId, 30*pageNum)
	if err != nil {
		err = errors.Wrap(err, "GetUserHomeArtwork: sqlStr1 get fail")
		return
	}

	return

}

// GetUserRecentlyArtworkId 获取用户最近1年发布的作品ID
func GetUserRecentlyArtworkId(userId string) (artIds []int64, err error) {
	end := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	start := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")

	sqlStr := `SELECT artwork_id from artwork 
              WHERE user_id = ? and is_delete = 0 and whoSee != 'privacy'  
              and createAT between ? and ?
			  ORDER BY artwork_id DESC
			  `

	err = db.Select(&artIds, sqlStr, userId, start, end)
	if err != nil {
		err = errors.Wrap(err, "GetUserRecentlyArtworkId get fail")
	}

	return

}

// GetOneArtwork 查询单个作品信息
func GetOneArtwork(artworkId string) (artwork models.ShowArtworkInfo, err error) {
	var eg errgroup.Group

	eg.Go(func() error {
		//获取作品信息
		sqlStr1 := `SELECT a.artwork_id,a.user_id,title,pic_count,cover,zone,a.whoSee,adults,a.is_delete,
       				views,likes,collects,comments,forwards,comment,copyright,a.createAT
					from artwork as a
					INNER JOIN artwork_count as ac 
					on a.artwork_id = ac.artwork_id
					WHERE a.artwork_id = ? and a.is_delete = 0
					LIMIT 1`
		mErr := db.Get(&artwork, sqlStr1, artworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetOneArtwork: sql1 get fail")
		}
		return mErr
	})

	eg.Go(func() error {
		//获取作品图片信息
		sqlStr3 := `SELECT filename,sort,size,width,height from artwork_picture WHERE artwork_id = ?`
		mErr := db.Select(&artwork.Picture, sqlStr3, artworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetOneArtwork: sql3 get fail")
		}
		return mErr
	})

	eg.Go(func() error {
		//获取作品描述信息
		sqlStr4 := `SELECT description from art_intro WHERE artwork_id = ? limit 1`
		mErr := db.Get(&artwork, sqlStr4, artworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetOneArtwork: sqlStr4 get fail")
		}
		return mErr
	})

	eg.Go(func() error {
		//获取标签信息
		sqlStr5 := `SELECT tag_name,tag_id FROM tag_artwork WHERE artwork_id = ? and is_delete = 0`
		mErr := db.Select(&artwork.Tag, sqlStr5, artworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetOneArtwork: sql5 get fail")

		}
		return mErr
	})

	err = eg.Wait()
	if err != nil {
		return
	}
	author := new(models.UserSimpleInfoCount)
	eg.Go(func() error {
		//获取作者信息
		sqlStr6 := `SELECT avatar_name,username,likes,fans,collects,v_tag,v_status 
					from user_profile as up
					INNER JOIN user_count as uc 
					on up.user_id = uc.user_id
					WHERE up.user_id = ?
					limit 1`
		mErr := db.Get(author, sqlStr6, artwork.UserId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetOneArtwork: sql6 get fail")
		}
		return mErr
	})

	eg.Go(func() error {
		//获取单作品展示时 作者其他作品信息
		sqlStr7 := `(SELECT artwork_id,cover
				from artwork WHERE user_id = ? AND artwork_id >= ? and is_delete =0 and whoSee != 'privacy' and state = 0
				LIMIT 15)
				UNION ALL
				(SELECT artwork_id,cover
				from artwork WHERE user_id = ? AND artwork_id < ? and is_delete =0 and whoSee != 'privacy' and state = 0
				ORDER BY artwork_id DESC
				LIMIT 15)
				ORDER BY artwork_id DESC
				`
		mErr := db.Select(&artwork.OtherArtwork, sqlStr7, artwork.UserId, artworkId, artwork.UserId, artworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetAuthorOtherArtwork: sqlStr7 get fail")
		}
		if artwork.OtherArtwork == nil {
			artwork.OtherArtwork = make([]models.AuthorOtherArtwork, 0)
		}
		return mErr
	})

	err = eg.Wait()
	if err != nil {
		return
	}

	artwork.AvatarName = author.Avatar
	artwork.UserName = author.UserName
	artwork.VTag = author.VTag
	artwork.VStatus = author.VStatus
	artwork.AuthorCount.Collects = author.Collects
	artwork.AuthorCount.Likes = author.Likes
	artwork.AuthorCount.Fans = author.Fans

	return
}

// GetArtworkRank 查询作品排名数据
func GetArtworkRank(rankType string) (artworks []models.BasicArtwork, artCount []models.ArtworkCount, err error) {
	var sqlStr1 string

	switch rankType {
	case "today":
		sqlStr1 = `SELECT artwork_id,likes,collects,user_id from rank_art_1_day 
                   WHERE rank_date = (SELECT rank_date FROM rank_art_1_day ORDER BY rank_date DESC LIMIT 1)
				   order by score desc,collects desc ,likes desc ,artwork_id desc 
				   limit 250 `
	case "week":
		sqlStr1 = `SELECT artwork_id,likes,collects,user_id from rank_art_7_day
                   WHERE rank_date = (SELECT rank_date FROM rank_art_7_day ORDER BY rank_date DESC LIMIT 1)
			   	    order by score desc,collects desc ,likes desc ,artwork_id desc 
				   limit 250 `
	case "month":
		sqlStr1 = `SELECT artwork_id,likes,collects,user_id from rank_art_30_day
                   WHERE rank_date = (SELECT rank_date FROM rank_art_30_day ORDER BY rank_date DESC LIMIT 1)
				    order by score desc,collects desc ,likes desc ,artwork_id desc 
				   limit 250 `
	}
	// 查询 数据库
	err = db.Select(&artCount, sqlStr1)
	if err != nil {
		err = errors.Wrap(err, "GetArtworkRank: sql1 get fail")
		return
	}

	var artIds []string
	var uIds []string
	for _, info := range artCount {
		artIds = append(artIds, info.ArtworkId)
		uIds = append(uIds, info.UserId)
	}

	artworks, err = GetBatchBasicShowArtInfo(artIds, uIds)

	return
}

// GetChannelArtwork 查询最新作品数据
func GetChannelArtwork(query models.QueryChanelType) (dataList []models.ArtIdAndUid, err error) {
	var sqlStr1 string
	var oderStr string
	baseStr := `SELECT a.artwork_id,a.user_id from artwork_count as ac
                left join artwork a on ac.artwork_id = a.artwork_id          
                where a.is_delete = 0 and a.whoSee ='public'
				`
	var args []interface{}
	// 是否单独查询分区
	if query.Zone != "all" {
		baseStr = baseStr + " and zone = ?"
		args = append(args, query.Zone)
	}
	//是否翻页
	if query.NextId != "0" {
		baseStr = baseStr + " and a.artwork_id < ? "
		args = append(args, query.NextId)
	}

	if query.Sort == "new" {
		oderStr = " order by a.artwork_id desc limit 30"
	} else {
		oderStr = " order by score DESC LIMIT ?,30;"
		args = append(args, (query.Page-1)*30)
	}

	sqlStr1 = baseStr + oderStr
	// 查询 数据库
	err = db.Select(&dataList, sqlStr1, args...)
	if err != nil {
		err = errors.Wrap(err, "GetChannelArtwork: sql1 get fail")
		return
	}

	return
}

// UpdateArtInfo 更新作品数据
func UpdateArtInfo(info models.UpdateArtInfo) (err error) {
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

	var eg errgroup.Group
	eg.Go(func() (mErr error) {
		sql1 := `UPDATE artwork_count SET whoSee = ? WHERE artwork_id = ? `
		_, mErr = tx.Exec(sql1, info.WhoSee, info.ArtworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo: sql1 get fail")
			return
		}
		sql2 := `UPDATE artwork SET title = ?,zone = ?,whoSee=?,adults=?,comment=?,copyright=?
			WHERE artwork_id = ? `
		_, mErr = tx.Exec(sql2, info.Title, info.Zone, info.WhoSee, info.Adult, info.Comment, info.CopyRight, info.ArtworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo: sql2 get fail")
		}
		return
	})

	eg.Go(func() (mErr error) {
		sql3 := `UPDATE art_intro SET description = ? WHERE artwork_id = ? `
		_, mErr = tx.Exec(sql3, info.Description, info.ArtworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo: sql3 get fail")
		}
		return
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
		// 先把之前的 标签全标记成删除
		sqlStr8 := fmt.Sprintf("Update tag_artwork set is_delete = 1 where artwork_id = ?")
		_, mErr := tx.Exec(sqlStr8, info.ArtworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo sqlStr8 into fail")
			return mErr
		}

		// 新增的tag插入 tag表创建数据
		sqlStr5 := fmt.Sprintf("INSERT IGNORE INTO tag (tag_id, tag_name) VALUES %s",
			strings.Join(tagStrings, ","))
		_, mErr = tx.Exec(sqlStr5, tagArgs...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo sqlStr5 into fail")
			return mErr
		}

		//新增的 tag和art 关系 插入
		sqlStr6 := fmt.Sprintf("INSERT IGNORE INTO tag_artwork (artwork_id, tag_name) VALUES %s",
			strings.Join(tagStrings, ","))
		_, mErr = tx.Exec(sqlStr6, tagArtArgs...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo sqlStr6 into fail")
			return mErr
		}

		// 更新 新增的tag id
		sqlStr7 := `UPDATE tag_artwork as ta
				SET tag_id = (select tag_id FROM tag WHERE tag_name = ta.tag_name)
				WHERE artwork_id = ? `
		_, mErr = tx.Exec(sqlStr7, info.ArtworkId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo sqlStr7 into fail")
		}

		// 动态填充id
		query, args, mErr := sqlx.In("Update tag_artwork Set is_delete= 0 Where tag_name IN (?)", info.Tags)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo sqlx in  fail")
			return mErr
		}
		// sqlx.In 返回带 `?` bindvar的查询语句, 我们使用Rebind()重新绑定它
		query = db.Rebind(query)
		_, mErr = tx.Exec(query, args...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "UpdateArtInfo sqlStr8 in  fail")
			return mErr
		}
		return mErr
	})

	err = eg.Wait()

	return
}

func GetArtCount[T int64 | string](artworkIds []T) (count []models.ArtworkCount, err error) {
	if len(artworkIds) == 0 {
		return
	}
	sqlStr1 := `(SELECT artwork_id,likes,comments,forwards,collects,views from artwork_count WHERE artwork_id = ? limit 1 )`

	sql1Strings := make([]string, 0, len(artworkIds))
	findIds := make([]interface{}, 0, len(artworkIds))
	for _, aid := range artworkIds {
		sql1Strings = append(sql1Strings, sqlStr1)
		findIds = append(findIds, aid)
	}
	sqlStr1 = strings.Join(sql1Strings, " UNION ALL ")

	err = db.Select(&count, sqlStr1, findIds...)
	if err != nil {
		err = errors.Wrap(err, "GetArtCount: sql1 get fail")
	}

	return
}

// VerifyArtOwner 验证作品所有权
func VerifyArtOwner(userId, artId string) (isOwner bool, err error) {
	var authorId string
	sql1 := `select user_id from artwork where artwork_id = ?`
	err = db.Get(&authorId, sql1, artId)
	if err != nil {
		err = errors.Wrap(err, "VerifyArtOwner: sql1 get fail")
		return
	}
	isOwner = userId == authorId
	return
}

// DeleteArtwork 删除作品
func DeleteArtwork(artId, userId string) (err error) {
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

	sql1 := `UPDATE artwork SET is_delete = 1 WHERE artwork_id = ?`
	_, err = tx.Exec(sql1, artId)
	if err != nil {
		err = errors.Wrap(err, "DeleteArtwork sql1 fail")
		return
	}

	sql2 := `UPDATE artwork_count SET is_delete = 1 WHERE artwork_id = ?`
	_, err = tx.Exec(sql2, artId)
	if err != nil {
		err = errors.Wrap(err, "DeleteArtwork sql2 fail")
		return
	}

	sql3 := `UPDATE user_count SET art_count = art_count -1 WHERE user_id = ?`
	_, err = tx.Exec(sql3, userId)
	if err != nil {
		err = errors.Wrap(err, "DeleteArtwork sql3 fail")
		return
	}

	sql4 := `UPDATE tag_artwork SET is_delete = 1 WHERE artwork_id = ?`
	_, err = tx.Exec(sql4, artId)
	if err != nil {
		err = errors.Wrap(err, "DeleteArtwork sql4 fail")
		return
	}

	return
}
