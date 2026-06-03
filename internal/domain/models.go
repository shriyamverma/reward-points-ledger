package domain

import "errors"

// Predefined Point Types
const (
	TypePurchaseEarning = 1
	TypeReferralBonus   = 2
	TypeCashback        = 3
	TypeRedemption      = 4
)

// Shared Application Errors
var (
	ErrDuplicateEmail      = errors.New("a member with this email already exists")
	ErrMemberNotFound      = errors.New("member not found")
	ErrInvalidPointType    = errors.New("invalid point type id; must be 1-4")
	ErrPointsNotPositive   = errors.New("points must be a positive number")
	ErrInsufficientBalance = errors.New("insufficient balance for redemption")
)

type Member struct {
	MemberID      int    `json:"member_id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	PointsBalance int    `json:"points_balance,omitempty"` // populated on GET
	CreatedAt     string `json:"created_at"`
}

type RewardEntry struct {
	RewardID    int    `json:"reward_id"`
	MemberID    int    `json:"member_id"`
	PointTypeID int    `json:"point_type_id"`
	Points      int    `json:"points"` // Stored with applied sign internally
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
}
