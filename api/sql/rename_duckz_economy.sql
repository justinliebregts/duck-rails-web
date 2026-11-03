-- Duck Rails only. Run once after the original create_duckz_economy.sql.
-- Stop API traffic, rename, deploy the matching API, then resume traffic.
-- Do NOT run create_duckrails_economy.sql first: destinations must not exist.
-- Preserves rows, receipt IDs and indexes; the shared transactions ledger is unchanged.
-- Intentionally fails if old tables are missing or destination tables already exist.
USE `duckrails_tencent_global`;

RENAME TABLE
  `duckz_coin_deposits` TO `duckrails_coin_deposits`,
  `duckz_store_purchases` TO `duckrails_store_purchases`,
  `duckz_entitlements` TO `duckrails_entitlements`,
  `duckz_loadouts` TO `duckrails_loadouts`,
  `rails_wallet_grants` TO `duckrails_wallet_grants`;
