package driver_cards

import "time"

type DriverCard struct {
  ID            string     `json:"id"`
  DriverID      string     `json:"driver_id"`
  CardNumber    string     `json:"card_number"`
  CardType      string     `json:"card_type"`
  CurrentBalance float64   `json:"current_balance"`
  IsActive      bool       `json:"is_active"`
  IsBlocked     bool       `json:"is_blocked"`
  IssuedAt      time.Time  `json:"issued_at"`
  BlockedAt     *time.Time `json:"blocked_at"`
  BlockedBy     *string    `json:"blocked_by"`
  BlockReason   *string    `json:"block_reason"`
  Notes         *string    `json:"notes"`
  CreatedAt     time.Time  `json:"created_at"`
  UpdatedAt     time.Time  `json:"updated_at"`
}

type DriverCardTransaction struct {
  ID              string     `json:"id"`
  CardID          string     `json:"card_id"`
  TransactionType string     `json:"transaction_type"`
  Amount          float64    `json:"amount"`
  BalanceBefore   float64    `json:"balance_before"`
  BalanceAfter    float64    `json:"balance_after"`
  Description     *string    `json:"description"`
  TripExpenseID   *string    `json:"trip_expense_id"`
  PerformedBy     *string    `json:"performed_by"`
  CreatedAt       time.Time  `json:"created_at"`
}

type CreateDriverCardInput struct {
  DriverID       string   `json:"driver_id"`
  CardNumber     string   `json:"card_number"`
  CardType       string   `json:"card_type"`
  CurrentBalance *float64 `json:"current_balance"`
  Notes          *string  `json:"notes"`
}

type UpdateDriverCardInput struct {
  CardNumber *string `json:"card_number"`
  CardType   *string `json:"card_type"`
  IsActive   *bool   `json:"is_active"`
  Notes      *string `json:"notes"`
}

type BlockCardInput struct {
  Reason *string `json:"reason"`
}

type CreateCardTransactionInput struct {
  TransactionType string  `json:"transaction_type"`
  Amount          float64 `json:"amount"`
  Description     *string `json:"description"`
}

type ListFilter struct {
  DriverID  string
  IsActive  *bool
  IsBlocked *bool
  Limit     int
  Offset    int
}

type TransactionListFilter struct {
  CardID string
  Limit  int
  Offset int
}
