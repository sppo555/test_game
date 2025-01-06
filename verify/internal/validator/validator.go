// validator.go
package validator

import (
	"fmt"
	"github.com/letron/verify/internal/api"
	"github.com/letron/verify/internal/config"
)

type Validator struct {
	config    *config.Config
	apiClient *api.APIClient
}

func NewValidator(cfg *config.Config, client *api.APIClient) *Validator {
	return &Validator{
		config:    cfg,
		apiClient: client,
	}
}

func (v *Validator) ValidateGame(gameDetails *api.GameDetailsResponse) ValidationResult {
	result := ValidationResult{
		GameID:  gameDetails.Data.GameID,
		IsValid: true,
		Errors:  []string{},
	}

	// 驗證每個下注的派彩
	for _, bet := range gameDetails.Data.Bets {
		expectedPayout := 0.0
		actualPayout := 0.0
		if bet.Payout.Valid {
			actualPayout = bet.Payout.Float64
		}

		switch bet.BetType {
		case "player":
			if gameDetails.Data.Winner == "Player" {
				expectedPayout = bet.BetAmount * v.config.PlayerPayout
			}
		case "banker":
			if gameDetails.Data.Winner == "Banker" {
				expectedPayout = bet.BetAmount * v.config.BankerPayout
			}
		case "tie":
			if gameDetails.Data.Winner == "Tie" {
				expectedPayout = bet.BetAmount * v.config.TiePayout
			}
		case "luckySix":
			if gameDetails.Data.IsLuckySix {
				if gameDetails.Data.LuckySixType.Valid && gameDetails.Data.LuckySixType.String == "3cards" {
					expectedPayout = bet.BetAmount * v.config.Lucky6_3CardsPayout
				} else {
					expectedPayout = bet.BetAmount * v.config.Lucky6_2CardsPayout
				}
			}
		}

		if actualPayout != expectedPayout {
			result.IsValid = false
			result.Errors = append(result.Errors, 
				fmt.Sprintf("Invalid payout for %s bet: expected %.2f, got %.2f", 
					bet.BetType, expectedPayout, actualPayout))
		}
	}

	return result
}

func (v *Validator) CalculateExpectedPayouts(gameDetails *api.GameDetailsResponse) float64 {
	// TODO: 實現支付計算邏輯
	return 0.0
}

type ValidationResult struct {
	GameID  string
	IsValid bool
	Errors  []string
}
