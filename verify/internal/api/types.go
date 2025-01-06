package api

type NullString struct {
	String string
	Valid  bool
}

type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

type GameDetailsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		GameID          string     `json:"game_id"`
		Winner          string     `json:"winner"`
		PlayerScore     int        `json:"player_score"`
		BankerScore     int        `json:"banker_score"`
		IsLuckySix      bool       `json:"is_lucky_six"`
		LuckySixType    NullString `json:"lucky_six_type"`
		PlayerCards     string     `json:"player_cards"`
		BankerCards     string     `json:"banker_cards"`
		PlayerThirdCard NullString `json:"player_third_card"`
		BankerThirdCard NullString `json:"banker_third_card"`
		PlayerPayout    NullFloat64 `json:"player_payout"`
		BankerPayout    NullFloat64 `json:"banker_payout"`
		TiePayout       NullFloat64 `json:"tie_payout"`
		LuckySixPayout  NullFloat64 `json:"lucky_six_payout"`
		Bets            []Bet      `json:"bets"`
		TotalBets       float64    `json:"total_bets"`
		TotalPayouts    float64    `json:"total_payouts"`
	} `json:"data"`
}

type Bet struct {
	Username   string      `json:"username"`
	BetType    string      `json:"bet_type"`
	BetAmount  float64     `json:"bet_amount"`
	Payout     NullFloat64 `json:"payout"`
}
