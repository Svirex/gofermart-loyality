package domain

import "time"

type WithdrawData struct {
	OrderNum    string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
