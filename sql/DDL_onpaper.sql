/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = analyse   */
/******************************************/
CREATE TABLE `analyse` (
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `day_active` int unsigned DEFAULT '0' COMMENT '日活',
  `month_active` int unsigned DEFAULT '0' COMMENT '本月月活',
  `new_user` int unsigned DEFAULT '0' COMMENT '新用户数',
  `new_art` int unsigned DEFAULT '0' COMMENT '新作品数',
  `new_trend` int unsigned DEFAULT '0' COMMENT '新动态数',
  `likes` int unsigned DEFAULT '0' COMMENT '点赞数',
  `collects` int unsigned DEFAULT '0' COMMENT '收藏数',
  `comments` int unsigned DEFAULT '0' COMMENT '评论数',
  `following` int unsigned DEFAULT '0' COMMENT '关注数',
  `all_user` int unsigned DEFAULT '0' COMMENT '所有用户数',
  `all_art` int unsigned DEFAULT '0' COMMENT '所有作品数',
  PRIMARY KEY (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = art_intro   */
/******************************************/
CREATE TABLE `art_intro` (
  `artwork_id` bigint unsigned NOT NULL COMMENT '作品ID',
  `description` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '作品描述',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`artwork_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品介绍表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = artwork   */
/******************************************/
CREATE TABLE `artwork` (
  `artwork_id` bigint unsigned NOT NULL COMMENT '作品ID',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '作品名字',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `pic_count` tinyint unsigned NOT NULL COMMENT '图片数',
  `cover` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '作品封面文件名',
  `first_pic` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '首张图片',
  `zone` char(6) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '分区',
  `whoSee` enum('public','onlyFans','privacy') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'privacy' COMMENT '查看范围',
  `adults` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否敏感内容',
  `comment` enum('public','onlyFans','close') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'public' COMMENT '作品评论权限',
  `copyright` enum('BY','BY-ND','BY-NC','BY-NC-ND','OWNER') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'OWNER' COMMENT '作品版权',
  `is_delete` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否删除',
  `state` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态',
  `device` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'PC' COMMENT '上传设备',
  `createAT` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`artwork_id`),
  KEY `user_id` (`user_id`),
  KEY `permission` (`is_delete`,`whoSee`,`state`),
  KEY `createAt` (`createAT`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = artwork_count   */
/******************************************/
CREATE TABLE `artwork_count` (
  `artwork_id` bigint unsigned NOT NULL COMMENT '作品ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户id',
  `likes` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '喜欢数',
  `collects` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '收藏数',
  `comments` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '评论个数',
  `forwards` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '转发数',
  `views` int unsigned NOT NULL DEFAULT '0' COMMENT '浏览数',
  `score` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '综合分数',
  `is_delete` tinyint unsigned DEFAULT '0' COMMENT '是否删除',
  `whoSee` enum('public','onlyFans','privacy') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'privacy' COMMENT '查看范围',
  `state` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '状态',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`artwork_id`),
  KEY `user_id` (`user_id`),
  KEY `permission` (`is_delete`,`whoSee`,`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品数据统计表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = artwork_picture   */
/******************************************/
CREATE TABLE `artwork_picture` (
  `artwork_id` bigint unsigned NOT NULL COMMENT '作品ID',
  `filename` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '文件名',
  `sort` tinyint unsigned NOT NULL COMMENT '图片排序',
  `mimetype` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '文件类型',
  `height` int unsigned NOT NULL DEFAULT '0' COMMENT '图片高度',
  `width` int unsigned NOT NULL DEFAULT '0' COMMENT '图片宽度',
  `size` int unsigned NOT NULL COMMENT '文件大小',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`artwork_id`,`filename`),
  KEY `artwork_id` (`artwork_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品图片表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = avatar   */
/******************************************/
CREATE TABLE `avatar` (
  `user_id` bigint unsigned NOT NULL COMMENT '雪花id',
  `filename` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '文件名',
  `mimetype` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'image/jpeg' COMMENT '文件类型',
  `size` int unsigned NOT NULL DEFAULT '0' COMMENT '文件大小',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='头像图片表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = banner   */
/******************************************/
CREATE TABLE `banner` (
  `user_id` bigint unsigned NOT NULL COMMENT '雪花id',
  `filename` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '文件名',
  `mimetype` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'image/jpeg' COMMENT '文件类型',
  `size` int unsigned NOT NULL DEFAULT '0' COMMENT '文件大小',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='头像图片表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = commission_count   */
/******************************************/
CREATE TABLE `commission_count` (
  `user_id` int unsigned NOT NULL,
  `rating` decimal(3,2) unsigned NOT NULL DEFAULT '0.00' COMMENT '约稿评价分',
  `receive_wait` int unsigned NOT NULL DEFAULT '0' COMMENT '收到约稿_待接受个数',
  `receive_talk` int unsigned NOT NULL DEFAULT '0' COMMENT '收到约稿_沟通中个数',
  `receive_ing` int unsigned NOT NULL DEFAULT '0' COMMENT '收到约稿_创作个数',
  `receive_finish` int unsigned NOT NULL DEFAULT '0' COMMENT '收到约稿_完成个数',
  `receive_close` int unsigned NOT NULL DEFAULT '0' COMMENT '收到约稿_关闭个数',
  `send_wait` int unsigned NOT NULL DEFAULT '0' COMMENT '发出的约稿_待接受个数',
  `send_talk` int unsigned NOT NULL DEFAULT '0' COMMENT '发出的约稿_沟通中个数',
  `send_ing` int unsigned NOT NULL DEFAULT '0' COMMENT '发出的约稿_创作中个数',
  `send_finish` int unsigned NOT NULL DEFAULT '0' COMMENT '发出的约稿_完成个数',
  `send_close` int unsigned NOT NULL DEFAULT '0' COMMENT '发出的约稿_关闭个数',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = commission_evaluate   */
/******************************************/
CREATE TABLE `commission_evaluate` (
  `evaluate_id` bigint unsigned DEFAULT '0' COMMENT '评价id',
  `invite_id` bigint NOT NULL,
  `invite_own` int unsigned NOT NULL COMMENT '约稿者',
  `sender` int unsigned NOT NULL COMMENT '评价发出者',
  `receiver` int unsigned NOT NULL COMMENT '评价接收者',
  `text` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '系统默认好评' COMMENT '评论',
  `rate_1` tinyint unsigned NOT NULL DEFAULT '5' COMMENT '准时交稿/付款及时',
  `rate_2` tinyint unsigned NOT NULL DEFAULT '5' COMMENT '沟通能力/反馈及时',
  `rate_3` tinyint unsigned NOT NULL DEFAULT '5' COMMENT '作品质量/需求明确',
  `total_rating` decimal(3,2) unsigned NOT NULL DEFAULT '5.00' COMMENT '总得分',
  `is_delete` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否删除',
  `updateAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `createAT` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`invite_id`,`sender`,`receiver`) USING BTREE,
  KEY `sender` (`sender`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = invite_code   */
/******************************************/
CREATE TABLE `invite_code` (
  `owner` bigint unsigned NOT NULL COMMENT '邀请码拥有者',
  `code` varchar(7) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `used` bigint unsigned DEFAULT '0' COMMENT '邀请码使用者',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`code`,`owner`),
  KEY `code` (`code`,`used`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = notify_config   */
/******************************************/
CREATE TABLE `notify_config` (
  `user_id` bigint unsigned NOT NULL,
  `comment` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '收到评论，0 不接受提醒，1 所有人  2 关注的人',
  `like` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '收到点赞，0 不接受提醒，1 所有人  2 关注的人',
  `collect` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '收到收藏，0 不接受提醒，1 所有人  2 关注的人',
  `follow` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '收到关注，0 不接受提醒，1 所有人  2 关注的人',
  `message` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '收到关注，0 不接受提醒，1 所有人  2 关注的人',
  `at` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '收到at，0 不接受提醒，1 所有人  2 关注的人',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = notify_unread_count   */
/******************************************/
CREATE TABLE `notify_unread_count` (
  `user_id` bigint unsigned NOT NULL,
  `comment` int unsigned NOT NULL DEFAULT '0',
  `like` int unsigned NOT NULL DEFAULT '0',
  `collect` int unsigned NOT NULL DEFAULT '0',
  `follow` int unsigned NOT NULL DEFAULT '0',
  `at` int unsigned NOT NULL DEFAULT '0',
  `commission` int unsigned DEFAULT '0',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_art_1_day   */
/******************************************/
CREATE TABLE `rank_art_1_day` (
  `artwork_id` bigint unsigned NOT NULL COMMENT '作品ID',
  `user_id` bigint unsigned NOT NULL COMMENT '作者id',
  `likes` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '喜欢数',
  `collects` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '收藏数',
  `views` mediumint NOT NULL COMMENT '浏览数',
  `score` decimal(10,3) NOT NULL DEFAULT '0.000' COMMENT '热度',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`artwork_id`),
  KEY `score` (`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品30天排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_art_30_day   */
/******************************************/
CREATE TABLE `rank_art_30_day` (
  `artwork_id` bigint unsigned NOT NULL COMMENT '作品ID',
  `user_id` bigint unsigned NOT NULL COMMENT '作者id',
  `likes` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '喜欢数',
  `collects` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '收藏数',
  `views` mediumint NOT NULL COMMENT '浏览数',
  `score` decimal(10,3) NOT NULL DEFAULT '0.000' COMMENT '热度',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`artwork_id`),
  KEY `score` (`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品30天排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_art_7_day   */
/******************************************/
CREATE TABLE `rank_art_7_day` (
  `artwork_id` bigint unsigned NOT NULL COMMENT '作品ID',
  `user_id` bigint unsigned NOT NULL COMMENT '作者id',
  `likes` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '喜欢数',
  `collects` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '收藏数',
  `views` mediumint NOT NULL COMMENT '浏览数',
  `score` decimal(10,3) NOT NULL DEFAULT '0.000' COMMENT '热度',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`artwork_id`),
  KEY `score` (`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品30天排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_tag_day   */
/******************************************/
CREATE TABLE `rank_tag_day` (
  `ranks` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '当期排行',
  `tag_id` bigint unsigned NOT NULL COMMENT 'tagId',
  `tag_name` varchar(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '标签名字',
  `score` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '分数',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `status` enum('up','down','keep','new') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'keep' COMMENT '相对上一期的状态',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品30天排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_topic_day   */
/******************************************/
CREATE TABLE `rank_topic_day` (
  `ranks` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '当期排行',
  `topic_id` bigint unsigned NOT NULL COMMENT 'topicId',
  `topic_name` varchar(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '话题名字',
  `score` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '分数',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `status` enum('up','down','keep','new') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'keep' COMMENT '相对上一期的状态',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`topic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='作品30天排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_user_boy   */
/******************************************/
CREATE TABLE `rank_user_boy` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `fans` mediumint unsigned DEFAULT '0' COMMENT '粉丝数',
  `likes` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '喜欢数',
  `collects` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '收藏数',
  `score` decimal(10,3) NOT NULL DEFAULT '0.000' COMMENT '热度',
  `art_count` smallint unsigned DEFAULT '0' COMMENT '作品数',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`user_id`),
  KEY `score` (`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='男生用户综合人气排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_user_collect   */
/******************************************/
CREATE TABLE `rank_user_collect` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户id',
  `collects` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '近7日获得的收藏数',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='近7日收到最多收藏的排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_user_girl   */
/******************************************/
CREATE TABLE `rank_user_girl` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `fans` mediumint unsigned DEFAULT '0' COMMENT '粉丝数',
  `likes` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '喜欢数',
  `collects` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '收藏数',
  `score` decimal(10,3) NOT NULL DEFAULT '0.000' COMMENT '热度',
  `art_count` smallint unsigned DEFAULT '0' COMMENT '作品数',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`user_id`),
  KEY `score` (`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='女生用户综合人气排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_user_like   */
/******************************************/
CREATE TABLE `rank_user_like` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户id',
  `likes` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '近7日获得的喜欢数',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='近7日收到最多赞的排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = rank_user_new   */
/******************************************/
CREATE TABLE `rank_user_new` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `fans` mediumint unsigned DEFAULT '0' COMMENT '粉丝数',
  `likes` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '喜欢数',
  `collects` mediumint unsigned NOT NULL DEFAULT '0' COMMENT '收藏数',
  `score` decimal(10,3) NOT NULL DEFAULT '0.000' COMMENT '热度',
  `art_count` smallint unsigned DEFAULT '0' COMMENT '作品数',
  `rank_date` datetime NOT NULL COMMENT '排行日期',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  PRIMARY KEY (`rank_date`,`user_id`),
  KEY `score` (`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='新用户（90天）综合人气排行榜'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = suggested_user   */
/******************************************/
CREATE TABLE `suggested_user` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户id',
  `fans` int unsigned DEFAULT '0' COMMENT '粉丝数',
  `likes` int unsigned DEFAULT '0' COMMENT '被赞数',
  `collects` int unsigned DEFAULT '0' COMMENT '被收藏数',
  `art_count` smallint unsigned DEFAULT '0' COMMENT '作品数',
  `score` decimal(10,3) NOT NULL DEFAULT '0.000' COMMENT '热度',
  `user_createAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `row_createAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`,`row_createAt`),
  KEY `score` (`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='推荐用户表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = tag   */
/******************************************/
CREATE TABLE `tag` (
  `tag_id` bigint unsigned NOT NULL COMMENT '标签ID',
  `tag_name` varchar(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin NOT NULL COMMENT '标签名',
  `art_count` int unsigned NOT NULL DEFAULT '1' COMMENT '标签的作品统计',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`tag_id`),
  UNIQUE KEY `tag_name` (`tag_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='标签表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = tag_artwork   */
/******************************************/
CREATE TABLE `tag_artwork` (
  `artwork_id` bigint unsigned NOT NULL COMMENT '作品ID',
  `tag_name` varchar(25) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin NOT NULL COMMENT '标签名',
  `tag_id` bigint unsigned DEFAULT NULL COMMENT '标签ID',
  `is_delete` tinyint unsigned DEFAULT '0' COMMENT '是否删除了标签',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`artwork_id`,`tag_name`),
  KEY `tag_name` (`tag_name`),
  KEY `tag_id` (`tag_id`),
  KEY `createAt` (`createAt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='标签作品关系表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = topic   */
/******************************************/
CREATE TABLE `topic` (
  `topic_id` bigint unsigned NOT NULL COMMENT '话题id',
  `user_id` bigint DEFAULT NULL COMMENT '话题发起人',
  `text` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin NOT NULL COMMENT '话题名称',
  `trend_count` int unsigned NOT NULL DEFAULT '1' COMMENT '话题动态统计',
  `intro` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '话题介绍',
  `recommend` tinyint unsigned DEFAULT '0' COMMENT '是否推荐话题',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`topic_id`),
  UNIQUE KEY `text_unique` (`text`),
  FULLTEXT KEY `text_fulltext` (`text`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='话题表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = user   */
/******************************************/
CREATE TABLE `user` (
  `snow_id` bigint unsigned NOT NULL COMMENT '雪花id',
  `password` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '密码',
  `username` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `email` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '邮箱',
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '手机',
  `forbid` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否被封',
  `active` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后活跃时间',
  `ip` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'ip地址',
  `createAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`snow_id`),
  UNIQUE KEY `phone` (`phone`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `create` (`createAt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='用户主表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = user_count   */
/******************************************/
CREATE TABLE `user_count` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户id',
  `following` int unsigned NOT NULL DEFAULT '0' COMMENT '关注数',
  `fans` int unsigned NOT NULL DEFAULT '0' COMMENT '粉丝数',
  `likes` int unsigned NOT NULL DEFAULT '0' COMMENT '被赞数',
  `collects` int unsigned NOT NULL DEFAULT '0' COMMENT '被收藏数',
  `trend_count` smallint unsigned NOT NULL DEFAULT '0' COMMENT '动态数',
  `art_count` smallint unsigned NOT NULL DEFAULT '0' COMMENT '作品数',
  `collect_count` smallint unsigned NOT NULL DEFAULT '0' COMMENT '收藏作品数',
  `score` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '综合分数',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='用户数据统计表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = user_focus   */
/******************************************/
CREATE TABLE `user_focus` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户id',
  `focus_id` bigint unsigned NOT NULL COMMENT '被关注的用户id',
  `is_cancel` tinyint unsigned DEFAULT '0' COMMENT '是否取消关注',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`,`focus_id`),
  KEY `create` (`createAt`),
  KEY `is_cancel` (`is_cancel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = user_intro   */
/******************************************/
CREATE TABLE `user_intro` (
  `user_id` bigint unsigned NOT NULL COMMENT '雪花id',
  `introduce` varchar(4000) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '个人介绍',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='用户介绍表'
;

/******************************************/
/*   DatabaseName = onpaper   */
/*   TableName = user_profile   */
/******************************************/
CREATE TABLE `user_profile` (
  `user_id` bigint unsigned NOT NULL COMMENT '雪花id',
  `username` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `sex` enum('man','woman','privacy') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT 'privacy' COMMENT '性别',
  `birthday` date DEFAULT NULL COMMENT '生日',
  `work_email` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '工作邮箱',
  `QQ` varchar(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT 'qq',
  `Weibo` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT 'weibo',
  `Twitter` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT 'twitter',
  `Pixiv` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT 'pixiv',
  `Bilibili` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT 'b站',
  `WeChat` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '微信',
  `banner_name` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '背景文件名',
  `avatar_name` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '头像文件名',
  `address` varchar(35) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '所在地区',
  `create_style` varchar(35) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '创作风格',
  `software` varchar(35) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '常用软件',
  `expect_work` enum('全职工作','约稿创作','项目外包','暂不考虑','') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '' COMMENT '期望工作类型',
  `v_tag` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'v认证',
  `v_status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '1是用户推荐之类，2是官方认证',
  `commission` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否开始接稿',
  `have_plan` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否有接稿计划',
  `createAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updateAt` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `userName` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='用户资料表'
;
