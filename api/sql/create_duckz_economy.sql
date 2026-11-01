-- Duck Rails only: apply to duckrails_tencent_global. Reuses Duckz ledger contracts.
-- Duckz economy: duck coin deposits, coin store purchases, durable IAP entitlements, loadouts.
-- Duck coin balance lives in `transactions` (SUM(credit) - SUM(debit)) on the Duckz DB;
-- these tables add idempotency (request_id / platform_txn_id) and an audit trail.
-- Run against Duck Rails after structure_classic.sql. The API expects these tables.

CREATE TABLE IF NOT EXISTS `duckz_coin_deposits` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `player_id` varchar(128) NOT NULL,
  `request_id` varchar(64) NOT NULL,
  `bars` int NOT NULL DEFAULT 0,
  `coins` int NOT NULL DEFAULT 0,
  `transaction_id` int DEFAULT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_duckz_coin_deposits_request_id` (`request_id`),
  KEY `idx_duckz_coin_deposits_player_created` (`player_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `duckz_store_purchases` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `player_id` varchar(128) NOT NULL,
  `item_key` varchar(128) NOT NULL,
  `price` int NOT NULL DEFAULT 0,
  `transaction_id` int DEFAULT NULL,
  `request_id` varchar(64) NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_duckz_store_purchases_request_id` (`request_id`),
  KEY `idx_duckz_store_purchases_player_item` (`player_id`, `item_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `duckz_entitlements` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `player_id` varchar(128) NOT NULL,
  `sku` varchar(128) NOT NULL,
  `platform` varchar(24) NOT NULL DEFAULT '',
  `platform_txn_id` varchar(512) NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_duckz_entitlements_player_sku` (`player_id`, `sku`),
  UNIQUE KEY `uq_duckz_entitlements_platform_txn_id` (`platform_txn_id`),
  KEY `idx_duckz_entitlements_player_id` (`player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- The avatar each player has equipped, so it follows them across reinstalls and headsets.
-- Empty skin_key is the default duck.
CREATE TABLE IF NOT EXISTS `duckz_loadouts` (
  `player_id` varchar(128) NOT NULL,
  `skin_key` varchar(128) NOT NULL DEFAULT '',
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS rails_wallet_grants (
 player_id varchar(128) NOT NULL PRIMARY KEY,
 created_at datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
