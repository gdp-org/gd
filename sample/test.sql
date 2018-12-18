use test

CREATE table `test`(
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '姓名',
  `card_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '身份证',
  `sex` varchar(10) NOT NULL DEFAULT '' COMMENT '性别',
  `birthday` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '出生年月日',
  `status` tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '状态 ',
  `create_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_card_id` (`card_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='测试';