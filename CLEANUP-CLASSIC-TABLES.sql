-- Run only after deploying BOTH reduced-schema API and admin images.
-- Destructive: export a backup first. These DROP statements cannot be rolled back.
-- Explicitly targets Classic, not another game's database.
USE `grc_tencent_global`;
SELECT DATABASE() AS cleanup_database;

-- Keep foreign-key checks enabled. If a dependency blocks this, investigate it.
DROP TABLE IF EXISTS
 `player_bounty`, `bounty`, `bp_levels`, `bp`,
 `player_quests`, `quests`, `player_xp`, `xp_types`,
 `friend_blocks`, `friend_follows`,
 `leaderboard_kdr_alltime`, `leaderboard_kdr_weekly`,
 `leaderboard_kills_alltime`, `leaderboard_kills_weekly`,
 `leaderboard_score_alltime`, `leaderboard_score_weekly`,
 `stats_leaderboard_kills_alltime`, `stats_leaderboard_kills_weekly`,
 `stats_leaderboard_score_alltime`, `stats_leaderboard_score_weekly`,
 `stats_daily_deaths`, `stats_daily_games`, `stats_daily_kills`,
 `stats_daily_maps`, `stats_daily_totals`, `stats_daily_weapons`, `stats_daily_xp`,
 `player_events`, `events`, `player_games`, `kills`, `games`, `game_types`,
 `maps`, `teams`, `item_bundles`, `orders`, `photon`, `player_extra`,
 `rarity`, `store_display`, `bans`, `news`, `player_prefabs`;

-- Retained: players, admin_privileges, reports, service_bans, build_hash,
-- platform_session_entries, items, item_types, weapons, skin_styles, transactions.
-- The previous legacy schema may be missing this required table.
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

SHOW TABLES;
