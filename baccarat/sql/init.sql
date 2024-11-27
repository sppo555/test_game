CREATE DATABASE IF NOT EXISTS baccarat_db;
USE baccarat_db;

-- 遊戲記錄表
CREATE TABLE IF NOT EXISTS game_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    game_id VARCHAR(36) NOT NULL,  -- UUID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    player_initial_cards VARCHAR(100) NOT NULL,  -- 閒家初始牌
    banker_initial_cards VARCHAR(100) NOT NULL,  -- 莊家初始牌
    player_third_card VARCHAR(50),               -- 閒家補牌
    banker_third_card VARCHAR(50),               -- 莊家補牌
    player_final_score INT NOT NULL,             -- 閒家最終點數
    banker_final_score INT NOT NULL,             -- 莊家最終點數
    winner ENUM('Player', 'Banker', 'Tie') NOT NULL,
    is_lucky_six BOOLEAN DEFAULT FALSE,
    lucky_six_type VARCHAR(10),                  -- '2cards' 或 '3cards'
    player_payout DECIMAL(10,2),                 -- 閒家賠率
    banker_payout DECIMAL(10,2),                 -- 莊家賠率
    tie_payout DECIMAL(10,2),                    -- 和局賠率
    lucky_six_payout DECIMAL(10,2),              -- 幸運6賠率
    INDEX idx_game_id (game_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 詳細牌記錄表
CREATE TABLE IF NOT EXISTS card_details (
    id INT AUTO_INCREMENT PRIMARY KEY,
    game_id VARCHAR(36) NOT NULL,
    position VARCHAR(20) NOT NULL,
    suit INT NOT NULL,
    value INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES game_records(game_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
