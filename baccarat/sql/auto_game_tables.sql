-- 自動賭局記錄表
CREATE TABLE IF NOT EXISTS auto_game_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    game_id VARCHAR(36) NOT NULL,  -- UUID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    player_initial_cards VARCHAR(100),  -- 閒家初始牌
    banker_initial_cards VARCHAR(100),  -- 莊家初始牌
    player_third_card VARCHAR(50),      -- 閒家補牌
    banker_third_card VARCHAR(50),      -- 莊家補牌
    player_final_score INT,             -- 閒家最終點數
    banker_final_score INT,             -- 莊家最終點數
    winner ENUM('Player', 'Banker', 'Tie'),
    is_lucky_six BOOLEAN DEFAULT FALSE,
    lucky_six_type VARCHAR(10),         -- '2cards' 或 '3cards'
    player_payout DECIMAL(10,2),        -- 閒家賠率
    banker_payout DECIMAL(10,2),        -- 莊家賠率
    tie_payout DECIMAL(10,2),           -- 和局賠率
    lucky_six_payout DECIMAL(10,2),     -- 幸運6賠率
    game_status ENUM('pending', 'betting', 'closed', 'drawing', 'completed', 'cancelled') NOT NULL DEFAULT 'pending',
    betting_start_time TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,  -- 開始下注時間
    betting_end_time TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,    -- 結束下注時間
    INDEX idx_game_id (game_id),
    INDEX idx_created_at (created_at),
    INDEX idx_game_status (game_status),
    UNIQUE KEY uk_game_id (game_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 自動賭局下注記錄表
CREATE TABLE IF NOT EXISTS auto_game_bets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    game_id VARCHAR(36) NOT NULL,
    user_id INT NOT NULL,
    bet_type ENUM('Player', 'Banker', 'Tie', 'Lucky6') NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    status ENUM('pending', 'completed', 'cancelled') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES auto_game_records(game_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_game_id (game_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;