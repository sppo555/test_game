// api_client.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GameDetailsResponse 定義 API 回應結構
type GameDetailsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		GameID          string          `json:"game_id"`
		Winner          string          `json:"winner"`
		PlayerScore     int             `json:"player_score"`
		BankerScore     int             `json:"banker_score"`
		IsLuckySix      bool            `json:"is_lucky_six"`
		LuckySixType    json.RawMessage `json:"lucky_six_type"`
		PlayerCards     string          `json:"player_cards"`
		BankerCards     string          `json:"banker_cards"`
		PlayerThirdCard json.RawMessage `json:"player_third_card"`
		BankerThirdCard json.RawMessage `json:"banker_third_card"`
		PlayerPayout    json.RawMessage `json:"player_payout"`
		BankerPayout    json.RawMessage `json:"banker_payout"`
		TiePayout       json.RawMessage `json:"tie_payout"`
		LuckySixPayout  json.RawMessage `json:"lucky_six_payout"`
		Bets            []Bet           `json:"bets"`
		TotalBets       float64         `json:"total_bets"`
		TotalPayouts    float64         `json:"total_payouts"`
	} `json:"data"`
}

type Bet struct {
	Username  string          `json:"username"`
	BetType   string          `json:"bet_type"`
	BetAmount float64         `json:"bet_amount"`
	Payout    json.RawMessage `json:"payout"`
}

// APIClient 用於調用遊戲 API
type APIClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewAPIClient 建立新的 API 客戶端
func NewAPIClient(baseURL, token string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Token:   token,
		Client:  &http.Client{},
	}
}

// GetGameDetails 調用 API 獲取遊戲詳情
func (api *APIClient) GetGameDetails(gameID string) (*GameDetailsResponse, error) {
	url := fmt.Sprintf("%s/api/game/details", api.BaseURL)

	payload := map[string]string{
		"game_id": gameID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Token))

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gameDetails GameDetailsResponse
	err = json.Unmarshal(body, &gameDetails)
	if err != nil {
		return nil, err
	}

	if !gameDetails.Success {
		return nil, fmt.Errorf("API returned unsuccessful response for game_id %s", gameID)
	}

	return &gameDetails, nil
}
