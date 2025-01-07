// validator.go
package validator

import (
	"fmt"
	"github.com/letron/verify/internal/api"
	"github.com/letron/verify/internal/config"
	"strings"
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
		TotalGames:    1,
		ValidGames:    1,
		InvalidGames:  0,
		InvalidGameIDs: []string{},
		ErrorDetails:  []string{},
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
				if gameDetails.Data.IsLuckySix {
					if gameDetails.Data.LuckySixType.Valid && gameDetails.Data.LuckySixType.String == "3cards" {
						expectedPayout = bet.BetAmount * v.config.BankerLucky6_3CardsPayout
					} else {
						expectedPayout = bet.BetAmount * v.config.BankerLucky6_2CardsPayout
					}
				} else {
					expectedPayout = bet.BetAmount * v.config.BankerPayout
				}
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
			result.ValidGames = 0
			result.InvalidGames = 1
			result.InvalidGameIDs = append(result.InvalidGameIDs, gameDetails.Data.GameID)
			result.ErrorDetails = append(result.ErrorDetails, 
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

// ValidationResult represents the result of validation
type ValidationResult struct {
	TotalGames    int      `json:"total_games"`
	ValidGames    int      `json:"valid_games"`
	InvalidGames  int      `json:"invalid_games"`
	InvalidGameIDs []string `json:"invalid_game_ids"` 
	ErrorDetails  []string `json:"error_details"`     
}

// String returns a string representation of the validation result
func (vr *ValidationResult) String() string {
    var sb strings.Builder
    sb.WriteString("Validation Summary:\n")
    sb.WriteString(fmt.Sprintf("Total games: %d\n", vr.TotalGames))
    sb.WriteString(fmt.Sprintf("Valid games: %d\n", vr.ValidGames))
    sb.WriteString(fmt.Sprintf("Invalid games: %d\n", vr.InvalidGames))
    
    if len(vr.InvalidGameIDs) > 0 {
        sb.WriteString("\nInvalid Game IDs:\n")
        for i, gameID := range vr.InvalidGameIDs {
            sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, gameID))
        }
    }

    if len(vr.ErrorDetails) > 0 {
        sb.WriteString("\nError Details:\n")
        for i, detail := range vr.ErrorDetails {
            sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, detail))
        }
    }

    return sb.String()
}
