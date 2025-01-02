// validator.go
package validator

import (
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
	// 驗證遊戲邏輯
	result := ValidationResult{
		GameID:  gameDetails.Data.GameID,
		IsValid: true,
		Errors:  []string{},
	}
	
	// TODO: 實現驗證邏輯
	
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
