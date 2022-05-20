use jnuwechat;

-- ----------------------------
-- Table structure for commit_info
-- ----------------------------
DROP TABLE IF EXISTS `commit_info`;
CREATE TABLE `commit_info` (
                               `id` int(10) NOT NULL AUTO_INCREMENT COMMENT '主键',
                               `from_uid` int(10) NOT NULL COMMENT '提交者的ID',
                               `to_uid` int(10) NOT NULL COMMENT '发布者的ID',
                               `order_id` varchar(255) NOT NULL COMMENT '表单号',
                               `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
                               `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
                               `is_delete` int(10) NOT NULL DEFAULT '0' COMMENT '0:正常 1:无效(双方都可修改)',
                               `status` int(10) NOT NULL DEFAULT '0' COMMENT '0: 我要填 1:我填好了',
                               `hex_id` varchar(255) DEFAULT NULL COMMENT 'mongoId',
                               PRIMARY KEY (`id`),
                               KEY `find_from_uid` (`from_uid`) USING BTREE COMMENT '方便找该用户填写了哪些表格',
                               KEY `find_to_uid` (`to_uid`) USING BTREE COMMENT '方便找一起指向了哪位用户',
                               KEY `find_order` (`order_id`) USING BTREE COMMENT '通过订单找'
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for commit_json
-- ----------------------------
DROP TABLE IF EXISTS `commit_json`;
CREATE TABLE `commit_json` (
                               `id` int(10) NOT NULL AUTO_INCREMENT COMMENT '主键',
                               `created_at` timestamp NULL DEFAULT NULL,
                               `updated_at` timestamp NULL DEFAULT NULL,
                               `form` varchar(8192) NOT NULL COMMENT '压缩后的信息',
                               `out` int(10) NOT NULL COMMENT '外键',
                               PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for form_done_json
-- ----------------------------
DROP TABLE IF EXISTS `form_done_json`;
CREATE TABLE `form_done_json` (
                                  `id` int(11) NOT NULL AUTO_INCREMENT,
                                  `created_at` timestamp NULL DEFAULT NULL,
                                  `updated_at` timestamp NULL DEFAULT NULL,
                                  `form` varchar(8192) NOT NULL COMMENT '压缩后的信息',
                                  `title` varchar(255) NOT NULL COMMENT '标题',
                                  `tip` varchar(255) NOT NULL COMMENT '提示',
                                  `out` int(10) NOT NULL COMMENT '外键',
                                  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for form_uid
-- ----------------------------
DROP TABLE IF EXISTS `form_uid`;
CREATE TABLE `form_uid` (
                            `id` int(11) NOT NULL AUTO_INCREMENT,
                            `random_id` varchar(255) NOT NULL COMMENT '表单编号',
                            `uid` int(10) NOT NULL COMMENT '用户id',
                            `created_at` timestamp NULL DEFAULT NULL,
                            `updated_at` timestamp NULL DEFAULT NULL,
                            `status` int(10) DEFAULT '0' COMMENT '0:创建 1:完成提交',
                            `is_delete` int(10) DEFAULT '0' COMMENT '0:正常 1:已经删除',
                            `is_ban` int(10) DEFAULT '0' COMMENT '0:正常填写 1:不能填写',
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `find_id_by_order` (`random_id`) USING HASH COMMENT '通过订单来找UID',
                            KEY `find_order_by_id` (`uid`) USING HASH COMMENT '通过人找创建了的表单'
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for home
-- ----------------------------
DROP TABLE IF EXISTS `home`;
CREATE TABLE `home` (
                        `id` int(11) NOT NULL AUTO_INCREMENT,
                        `img` varchar(255) DEFAULT NULL COMMENT '图片URL',
                        `created_at` timestamp NULL DEFAULT NULL COMMENT '为以后管理打下基础',
                        `updated_at` timestamp NULL DEFAULT NULL,
                        `jump` varchar(255) DEFAULT NULL COMMENT '跳转链接',
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for share_info
-- ----------------------------
DROP TABLE IF EXISTS `share_info`;
CREATE TABLE `share_info` (
                              `id` int(10) NOT NULL AUTO_INCREMENT COMMENT '主键',
                              `share_id` varchar(255) NOT NULL COMMENT '分享号',
                              `created_at` timestamp NULL DEFAULT NULL,
                              `updated_at` timestamp NULL DEFAULT NULL,
                              `out` int(10) NOT NULL COMMENT '外键',
                              `parent` int(10) NOT NULL COMMENT '创建者',
                              PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for user_info
-- ----------------------------
DROP TABLE IF EXISTS `user_info`;
CREATE TABLE `user_info` (
                             `id` int(11) NOT NULL AUTO_INCREMENT,
                             `avatar_url` varchar(255) NOT NULL COMMENT '头像地址',
                             `city` varchar(128) DEFAULT NULL COMMENT '城市',
                             `country` varchar(128) DEFAULT NULL COMMENT '国家',
                             `gender` tinyint(10) DEFAULT '0' COMMENT '性别',
                             `nick_name` varchar(255) NOT NULL COMMENT '昵称',
                             `province` varchar(128) DEFAULT NULL COMMENT '省会',
                             `created_at` datetime DEFAULT NULL COMMENT '创建时间',
                             `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
                             `is_ban` tinyint(10) DEFAULT '0' COMMENT '是否在黑名单中 0:正常 1:禁止',
                             `identity` tinyint(10) DEFAULT '0' COMMENT '身份 0:普通 1:超级',
                             `open_id` varchar(255) NOT NULL COMMENT '固定的openId',
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `use_openid` (`open_id`) USING HASH COMMENT '因为不使用范围直接使用hash更快'
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;