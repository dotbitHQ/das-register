CREATE DATABASE `das_register_db`;
USE `das_register_db`;

-- DROP TABLE IF EXISTS `t_block_parser_info`;
-- DROP TABLE IF EXISTS `t_register_pending_info`;
-- DROP TABLE IF EXISTS `t_das_order_info`;
-- DROP TABLE IF EXISTS `t_das_order_pay_info`;

-- t_block_parser_info
CREATE TABLE `t_block_parser_info`
(
    `id`           BIGINT(20) UNSIGNED                                           NOT NULL AUTO_INCREMENT COMMENT '',
    `parser_type`  SMALLINT                                                      NOT NULL DEFAULT '0' COMMENT 'das-99 ckb-0 eth-1 btc-2 tron-3 bsc-5 4-wx polygon-6',
    `block_number` BIGINT(20) UNSIGNED                                           NOT NULL DEFAULT '0' COMMENT '',
    `block_hash`   VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '',
    `parent_hash`  VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '',
    `created_at`   TIMESTAMP                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '',
    `updated_at`   TIMESTAMP                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_parser_number` (parser_type, block_number) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT ='for block parser';

-- t_register_pending_info

CREATE TABLE `t_register_pending_info`
(
    `id`              BIGINT(20) UNSIGNED                                           NOT NULL AUTO_INCREMENT COMMENT '',
    `block_number`    BIGINT(20) UNSIGNED                                           NOT NULL DEFAULT '0' COMMENT '',
    `account`         VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '',
    `action`          VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '',
    `chain_type`      SMALLINT(6)                                                   NOT NULL DEFAULT '0' COMMENT '',
    `address`         VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '',
    `capacity`        BIGINT(20) unsigned                                           NOT NULL DEFAULT '0' COMMENT '',
    `outpoint`        VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '',
    `block_timestamp` BIGINT(20) unsigned                                           NOT NULL DEFAULT '0' COMMENT '',
    `status`          SMALLINT(6)                                                   NOT NULL DEFAULT '0' COMMENT '0-default 1-rejected',
    `created_at`      TIMESTAMP                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '',
    `updated_at`      TIMESTAMP                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_a_o` (`action`, `outpoint`) USING BTREE,
    KEY `k_a_a` (`account`) USING BTREE,
    KEY `k_ct_a` (`chain_type`, `address`) USING BTREE,
    KEY `k_block_number` (block_number) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT ='pending tx info';

-- t_das_order_info
CREATE TABLE `t_das_order_info`
(
    `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '',
    `order_type`          SMALLINT        NOT NULL DEFAULT '0' COMMENT '1-self 2-other',
    `order_id`            VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `account_id`          VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `account`             VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `action`              VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `chain_type`          SMALLINT        NOT NULL DEFAULT '0' COMMENT 'order chain type',
    `address`             VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT 'order address',
    `timestamp`           BIGINT          NOT NULL DEFAULT '0' COMMENT 'order time',
    `pay_token_id`        VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `pay_type`            VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '',
    `pay_amount`          DECIMAL(60)     NOT NULL DEFAULT '0' COMMENT '',
    `content`             TEXT            NOT NULL COMMENT 'order detail',
    `pay_status`          SMALLINT        NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok',
    `hedge_status`        SMALLINT        NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok',
    `pre_register_status` SMALLINT        NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok',
    `register_status`     SMALLINT        NOT NULL DEFAULT '0' COMMENT '1-6',
    `order_status`        SMALLINT        NOT NULL DEFAULT '0' COMMENT '1-closed',
    `created_at`          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '',
    `updated_at`          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_id` (`order_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT ='das order info';

-- t_das_order_tx_info
CREATE TABLE `t_das_order_tx_info`
(
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '',
    `order_id`   VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `hash`       VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `action`     VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `status`     SMALLINT        NOT NULL DEFAULT '0' COMMENT '0-default 1-confirm',
    `timestamp`  BIGINT          NOT NULL DEFAULT '0' COMMENT '',
    `created_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '',
    `updated_at` TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_id_hash` (`order_id`, `hash`),
    KEY `k_hash` (`hash`),
    KEY `k_action` (`action`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT ='order tx info';

-- t_das_order_pay_info
CREATE TABLE `t_das_order_pay_info`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '',
    `hash`          VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `order_id`      VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `chain_type`    SMALLINT        NOT NULL DEFAULT '0' COMMENT '',
    `address`       VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `account_id`    VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `status`        SMALLINT        NOT NULL DEFAULT '0' COMMENT '0-default 1-confirm',
    `timestamp`     BIGINT          NOT NULL DEFAULT '0' COMMENT '',
    `refund_status` SMALLINT        NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok',
    `refund_hash`   VARCHAR(255)    NOT NULL DEFAULT '' COLLATE utf8mb4_0900_ai_ci COMMENT '',
    `created_at`    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '',
    `updated_at`    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_hash` (`hash`),
    KEY `k_order_id` (`order_id`),
    KEY `k_address` (`chain_type`, `address`),
    KEY `k_account_id` (account_id)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT ='order pay info';
