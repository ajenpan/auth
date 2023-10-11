create DATABASE if not exists auth default character set = 'utf8mb4';

use auth;

CREATE TABLE IF NOT EXISTS `users` (
  `uid` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '用户唯一id',
  `uname` varchar(64) CHARACTER SET utf8mb4 NOT NULL COMMENT '用户名',
  `passwd` varchar(64) NOT NULL DEFAULT '' COMMENT '密码',
  `nickname` varchar(64) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` varchar(1024) NOT NULL DEFAULT '' COMMENT '头像',
  `gender` tinyint(4) NOT NULL DEFAULT 0 COMMENT '性别',
  `phone` varchar(32) NOT NULL DEFAULT '' COMMENT '电话号码',
  `email` varchar(64) NOT NULL DEFAULT '' COMMENT '电子邮箱',
  `stat` tinyint(4) NOT NULL DEFAULT 0 COMMENT '状态码',
  `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_at` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `UQE_uname` (`uname`)
) ENGINE = InnoDB AUTO_INCREMENT = 100000 DEFAULT CHARSET = utf8mb4;

-- insert into users (uname,passwd,nickname) values ('test','123456','test');