-- 用戶表
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    balance DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 遊戲記錄表
CREATE TABLE IF NOT EXISTS game_records (
    id INT AUTO_INCREMENT PRIMARY KEY,
    game_id VARCHAR(36) UNIQUE NOT NULL,
    player_initial_cards TEXT NOT NULL,
    banker_initial_cards TEXT NOT NULL,
    player_third_card VARCHAR(10),
    banker_third_card VARCHAR(10),
    player_final_score INT NOT NULL,
    banker_final_score INT NOT NULL,
    winner VARCHAR(10) NOT NULL,
    is_lucky_six BOOLEAN DEFAULT FALSE,
    lucky_six_type VARCHAR(10),
    player_payout DECIMAL(10,2),
    banker_payout DECIMAL(10,2),
    tie_payout DECIMAL(10,2),
    lucky_six_payout DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 投注記錄表
CREATE TABLE IF NOT EXISTS bets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    game_id VARCHAR(36) NOT NULL,
    bet_amount DECIMAL(10,2) NOT NULL,
    bet_type VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (game_id) REFERENCES game_records(game_id)
);

-- 交易記錄表
CREATE TABLE IF NOT EXISTS transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    transaction_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 索引
CREATE INDEX idx_user_id ON bets(user_id);
CREATE INDEX idx_game_id ON bets(game_id);
CREATE INDEX idx_user_transactions ON transactions(user_id);
CREATE INDEX idx_game_records_game_id ON game_records(game_id);
