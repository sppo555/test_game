package game

import (
    "database/sql"
    "log"
    "time"
    "github.com/google/uuid"
    "os"
    "strconv"
)

type AutoGameService struct {
    db *sql.DB
    enabled bool
    interval time.Duration
    betCloseTime time.Duration
}

func NewAutoGameService(db *sql.DB, enabled bool, interval, betCloseTime time.Duration) *AutoGameService {
    return &AutoGameService{
        db: db,
        enabled: enabled,
        interval: interval,
        betCloseTime: betCloseTime,
    }
}

func (s *AutoGameService) Start() {
    if !s.enabled {
        log.Println("Auto game service is disabled")
        return
    }

    go s.run()
}

func (s *AutoGameService) run() {
    ticker := time.NewTicker(s.interval)
    defer ticker.Stop()

    for range ticker.C {
        if err := s.createNewGame(); err != nil {
            log.Printf("Error creating new game: %v", err)
            continue
        }
    }
}

func (s *AutoGameService) createNewGame() error {
    gameID := uuid.New().String()
    now := time.Now()
    startTime := now
    endTime := now.Add(s.interval - s.betCloseTime)

    // 創建新遊戲記錄
    _, err := s.db.Exec(`
        INSERT INTO auto_game_records 
        (game_id, betting_start_time, betting_end_time, game_status)
        VALUES (?, ?, ?, 'betting')
    `, gameID, startTime, endTime)

    if err != nil {
        return err
    }

    // 設定定時器，在適當的時間執行相應的操作
    go s.scheduleGameEvents(gameID, endTime)

    return nil
}

func (s *AutoGameService) scheduleGameEvents(gameID string, endTime time.Time) {
    // 等待到達結束下注時間
    time.Sleep(time.Until(endTime))

    // 更新遊戲狀態為已關閉下注
    s.db.Exec("UPDATE auto_game_records SET game_status = 'closed' WHERE game_id = ?", gameID)

    // 立即執行開牌
    s.drawGame(gameID)
}

func (s *AutoGameService) drawGame(gameID string) {
    // 更新遊戲狀態為開牌中
    s.db.Exec("UPDATE auto_game_records SET game_status = 'drawing' WHERE game_id = ?", gameID)

    // 使用原本的遊戲邏輯
    game := NewBaccaratGame()
    result := game.Play()

    // 開始資料庫交易
    tx, err := s.db.Begin()
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        return
    }

    // 更新遊戲結果
    _, err = tx.Exec(`
        UPDATE auto_game_records 
        SET 
            player_initial_cards = ?,
            banker_initial_cards = ?,
            player_third_card = ?,
            banker_third_card = ?,
            player_final_score = ?,
            banker_final_score = ?,
            winner = ?,
            is_lucky_six = ?,
            lucky_six_type = ?,
            game_status = 'completed'
        WHERE game_id = ?
    `,
        result.PlayerInitialCards,
        result.BankerInitialCards,
        result.PlayerThirdCard,
        result.BankerThirdCard,
        result.PlayerFinalScore,
        result.BankerFinalScore,
        result.Winner,
        result.IsLuckySix,
        result.LuckySixType,
        gameID,
    )

    if err != nil {
        tx.Rollback()
        log.Printf("Error updating game result: %v", err)
        return
    }

    // 處理所有下注的派彩
    err = s.processPayouts(tx, gameID, result)
    if err != nil {
        tx.Rollback()
        log.Printf("Error processing payouts: %v", err)
        return
    }

    if err = tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
        return
    }
}

func (s *AutoGameService) processPayouts(tx *sql.Tx, gameID string, result *GameResult) error {
    // 從環境變數獲取賠率
    playerPayout, _ := strconv.ParseFloat(os.Getenv("PLAYER_PAYOUT"), 64)
    bankerPayout, _ := strconv.ParseFloat(os.Getenv("BANKER_PAYOUT"), 64)
    tiePayout, _ := strconv.ParseFloat(os.Getenv("TIE_PAYOUT"), 64)
    lucky6_2cardsPayout, _ := strconv.ParseFloat(os.Getenv("LUCKY6_2CARDS_PAYOUT"), 64)
    lucky6_3cardsPayout, _ := strconv.ParseFloat(os.Getenv("LUCKY6_3CARDS_PAYOUT"), 64)

    // 獲取所有待處理的下注
    rows, err := tx.Query(`
        SELECT id, user_id, bet_type, amount 
        FROM auto_game_bets 
        WHERE game_id = ? AND status = 'pending'
    `, gameID)
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var betID int
        var userID int
        var betType string
        var amount float64

        err := rows.Scan(&betID, &userID, &betType, &amount)
        if err != nil {
            return err
        }

        // 計算派彩金額
        var payout float64
        switch betType {
        case "Player":
            if result.Winner == "Player" {
                payout = amount * (1 + playerPayout)
            }
        case "Banker":
            if result.Winner == "Banker" {
                payout = amount * (1 + bankerPayout)
            }
        case "Tie":
            if result.Winner == "Tie" {
                payout = amount * (1 + tiePayout)
            }
        case "Lucky6":
            if result.IsLuckySix {
                if result.LuckySixType == "2cards" {
                    payout = amount * (1 + lucky6_2cardsPayout)
                } else {
                    payout = amount * (1 + lucky6_3cardsPayout)
                }
            }
        }

        // 更新用戶餘額
        if payout > 0 {
            _, err = tx.Exec(
                "UPDATE users SET balance = balance + ? WHERE id = ?",
                payout, userID,
            )
            if err != nil {
                return err
            }
        }

        // 更新下注狀態
        _, err = tx.Exec(
            "UPDATE auto_game_bets SET status = 'completed' WHERE id = ?",
            betID,
        )
        if err != nil {
            return err
        }
    }

    return nil
}
