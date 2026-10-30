-- Additive migration: run only against duckrails_tencent_global before deploying the API.
CREATE TABLE IF NOT EXISTS diagnostic_reports (
 id varchar(32) NOT NULL PRIMARY KEY,
 reporter varchar(128) NOT NULL,
 created_at datetime(3) NOT NULL,
 category varchar(100) NOT NULL,
 summary varchar(600) NOT NULL,
 build varchar(100) NOT NULL,
 platform varchar(50) NOT NULL,
 status varchar(16) NOT NULL DEFAULT 'open',
 resolution varchar(1000) NOT NULL DEFAULT '',
 payload mediumtext NOT NULL,
 INDEX diagnostic_status (status, created_at),
 INDEX diagnostic_reporter (reporter)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
