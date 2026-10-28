-- Duck Rails schema (filename retained from the Classic template).
-- Create this database on the existing TencentDB instance.
-- Run only for the separate Duck Rails database, never Classic, Gun Raiders, or Duckz.

CREATE DATABASE IF NOT EXISTS `duckrails_tencent_global` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `duckrails_tencent_global`;

SET NAMES utf8mb4;
SET time_zone = '+00:00';

CREATE TABLE IF NOT EXISTS `players` (
  `id` varchar(48) NOT NULL,
  `name` varchar(128) NOT NULL,
  `platform` enum('OCULUS','STEAM','IQIYI','EMAIL','PICO','APPLE','GOOGLE','ROOMBOT','DEBUGGER','YVR') NOT NULL,
  `ip` varchar(45) NOT NULL,
  `region` varchar(24) NOT NULL,
  `version` varchar(16) NOT NULL,
  `device` varchar(128) NOT NULL,
  `device_uuid` varchar(128) NOT NULL,
  `is_converted` tinyint(1) NOT NULL DEFAULT '0',
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `is_converted` (`is_converted`),
  KEY `version` (`version`),
  KEY `device` (`device`),
  KEY `name_platform` (`name`,`platform`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `platform_session_entries` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `player_id` varchar(128) NOT NULL,
  `platform` enum('OCULUS','STEAM','IQIYI','APPLE','EMAIL','PICO','GOOGLE','ROOMBOT','DEBUGGER','YVR') NOT NULL DEFAULT 'OCULUS',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- One row per owned skin or recorded platform purchase. debit and credit stay 0.
-- There is no coin balance.
CREATE TABLE IF NOT EXISTS `transactions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `player_id` varchar(48) NOT NULL,
  `item_id` int DEFAULT NULL,
  `debit` int NOT NULL DEFAULT '0',
  `credit` int NOT NULL DEFAULT '0',
  `platform` varchar(24) DEFAULT NULL,
  `platform_account` varchar(512) DEFAULT NULL,
  `parent_id` int DEFAULT NULL,
  `platform_txn_id` varchar(512) DEFAULT NULL,
  `platform_txn_time` varchar(512) DEFAULT NULL,
  `is_deleted` tinyint(1) NOT NULL DEFAULT '0',
  `date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `platform_txn_id_unique` (`platform_txn_id`),
  KEY `player_id_index` (`player_id`),
  KEY `item_id_index` (`item_id`) USING BTREE,
  KEY `platform_index` (`platform`),
  KEY `platform_account_index` (`platform`,`platform_account`),
  KEY `date` (`date`),
  KEY `is_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


CREATE TABLE IF NOT EXISTS `service_bans` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `target_type` enum('device','platform') NOT NULL,
  `target_id` varchar(128) NOT NULL DEFAULT '',
  `ban_type` enum('game','mic','automated') NOT NULL,
  `reason` json NOT NULL,
  `start_datetime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `end_datetime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `target_type` (`target_type`) USING BTREE,
  KEY `target_id` (`target_id`) USING BTREE,
  KEY `ban_type` (`ban_type`) USING BTREE,
  KEY `datetime_start_end` (`start_datetime`,`end_datetime`) USING BTREE,
  KEY `target_type_id` (`target_type`,`target_id`) USING BTREE,
  KEY `start_datetime` (`start_datetime`) USING BTREE,
  KEY `end_datetime` (`end_datetime`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `ip_blacklists` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `ip` varchar(45) NOT NULL,
  `reason` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ip` (`ip`),
  KEY `idx_ip_blacklists_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `admin_privileges` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `player_id` varchar(48) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_privileges_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


CREATE TABLE IF NOT EXISTS `build_hash` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `version` varchar(64) NOT NULL,
  `hash` varchar(128) NOT NULL,
  `obb_hash` json DEFAULT NULL,
  `date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `version_index` (`version`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Ids must match the C# skin style enum. Set id on create when the enum value is known.
CREATE TABLE IF NOT EXISTS `skin_styles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(48) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `item_types` (
  `id` tinyint NOT NULL,
  `type` varchar(24) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `item_types` (`id`, `type`) VALUES
  (2, 'weapon'),
  (3, 'avatar')
ON DUPLICATE KEY UPDATE `type` = VALUES(`type`);

-- Ids match the Gun Raiders weapon type ids used by gun skins.
CREATE TABLE IF NOT EXISTS `weapons` (
  `id` smallint unsigned NOT NULL,
  `name` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `weapons` (`id`, `name`) VALUES
  (1, 'Burst Pistol'),
  (2, 'Dual Pistol'),
  (4, 'Knife'),
  (5, 'Minigun'),
  (7, 'Revolver'),
  (8, 'Rifle'),
  (9, 'Shotgun'),
  (10, 'Sniper'),
  (11, 'SMG'),
  (13, 'Shield'),
  (15, 'Projectile Launcher'),
  (19, 'Dual Uzi'),
  (20, 'Glove'),
  (21, 'Misc Item'),
  (22, 'Katana'),
  (24, 'Gauntlet'),
  (30, 'Rocket Launcher'),
  (34, 'Bow'),
  (37, 'Crossbow'),
  (39, 'Fusion Rifle'),
  (40, 'Riot Shield'),
  (42, 'Scythe'),
  (43, 'Grapple Hook'),
  (45, 'Shuriken'),
  (46, 'Flamethrower'),
  (52, 'Lightning Gun'),
  (53, 'Crystal Wand'),
  (56, 'Blaster'),
  (57, 'Hammer'),
  (59, 'Scout'),
  (60, 'Grenade'),
  (61, 'Saw Launcher')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

CREATE TABLE IF NOT EXISTS `items` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `rarity_id` tinyint unsigned NOT NULL DEFAULT '0',
  `type` tinyint NOT NULL,
  `weapon_type` smallint unsigned DEFAULT NULL,
  `asset_id` varchar(64) DEFAULT NULL,
  `skin_style_id` int NOT NULL DEFAULT '0',
  `price` int NOT NULL DEFAULT '0',
  `is_default` tinyint(1) NOT NULL DEFAULT '0',
  `dev_owned` tinyint(1) NOT NULL DEFAULT '0',
  `can_own_multiple` tinyint(1) NOT NULL DEFAULT '0',
  `gr_ready` tinyint(1) NOT NULL DEFAULT '0',
  `created_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `type_index` (`type`),
  KEY `items_skin_style_id` (`skin_style_id`),
  KEY `weapon_type_index` (`weapon_type`),
  KEY `is_default` (`is_default`),
  KEY `dev_owned_index` (`dev_owned`),
  KEY `asset_id` (`asset_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
