CREATE TABLE `test` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '姓名',
  `card_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '身份证',
  `sex` varchar(10) NOT NULL DEFAULT '' COMMENT '性别',
  `birthday` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '出生年月日',
  `status` tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '状态 ',
  `create_time` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_card_id` (`card_id`) USING BTREE,
  KEY `abc` (`name`,`card_id`,`birthday`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8 COMMENT='测试';

CREATE TABLE "honeypot"."test"
(
    "id" BIGINT IDENTITY(3, 1) NOT NULL,
    "name" VARCHAR(128) DEFAULT '' NOT NULL,
    "card_id" DECIMAL DEFAULT 0 NOT NULL,
    "sex" VARCHAR(10) DEFAULT '' NOT NULL,
    "birthday" DECIMAL DEFAULT 0 NOT NULL,
    "status" INT DEFAULT 0 NOT NULL,
    "create_time" DECIMAL DEFAULT 0 NOT NULL,
    "update_time" TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    NOT CLUSTER PRIMARY KEY("id"),
    CONSTRAINT "uk_card_id" UNIQUE("card_id"),
    CHECK("card_id" >= 0)
    ,CHECK("birthday" >= 0)
    ,CHECK("status" >= 0)
    ,CHECK("create_time" >= 0)) STORAGE(ON "MAIN", CLUSTERBTR) ;

COMMENT ON TABLE "honeypot"."test" IS '测试';
COMMENT ON COLUMN "honeypot"."test"."birthday" IS '出生年月日';
COMMENT ON COLUMN "honeypot"."test"."card_id" IS '身份证';
COMMENT ON COLUMN "honeypot"."test"."create_time" IS '创建时间';
COMMENT ON COLUMN "honeypot"."test"."name" IS '姓名';
COMMENT ON COLUMN "honeypot"."test"."sex" IS '性别';
COMMENT ON COLUMN "honeypot"."test"."status" IS '状态';
COMMENT ON COLUMN "honeypot"."test"."update_time" IS '更新时间';


CREATE  INDEX "abc" ON "honeypot"."test"("name" ASC,"card_id" ASC,"birthday" ASC) STORAGE(ON "MAIN", CLUSTERBTR) ;