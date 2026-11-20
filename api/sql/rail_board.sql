CREATE TABLE IF NOT EXISTS rail_board_runs (
  player_id VARCHAR(160) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  run_id CHAR(32) CHARACTER SET ascii NOT NULL,
  counters JSON NOT NULL,
  PRIMARY KEY (player_id, run_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS rail_board_totals (
  period CHAR(10) CHARACTER SET ascii NOT NULL,
  metric VARCHAR(96) CHARACTER SET ascii NOT NULL,
  subject VARCHAR(160) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  label VARCHAR(80) CHARACTER SET utf8mb4 NOT NULL,
  amount BIGINT NOT NULL,
  PRIMARY KEY (period, metric, subject),
  INDEX board_rank (period, metric, amount)
) ENGINE=InnoDB;
