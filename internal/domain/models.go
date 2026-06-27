package domain

import "errors"

// Shared Application Errors
var (
	ErrDuplicateEmail       = errors.New("a member with this email already exists")
	ErrMemberNotFound       = errors.New("member not found")
	ErrInvalidPointType     = errors.New("invalid point type id, or point type is inactive")
	ErrPointsNotPositive    = errors.New("points must be a positive number")
	ErrInsufficientBalance  = errors.New("insufficient balance for redemption")
	ErrRewardNotFound       = errors.New("reward not found")
	ErrDuplicatePointTypeID = errors.New("a point type with this ID already exists")
	ErrPointNotFound        = errors.New("point not found")
)

// Predefined Point Types
const (
	TypePurchaseEarning int = 1
	TypeReferralBonus   int = 2
	TypeCashback        int = 3
	TypeRedemption      int = 4
)

// Point instantiation for testing
// Add below data into points table using '/points' POST API -> CreatePoint
// 1,PurchaseEarning
// 2,ReferralBonus
// 3,Cashback
// 4,Redemption
type Point struct {
	PointID     int    `json:"point_id"`
	PointTypeID int    `json:"point_type_id"`
	PointCode   string `json:"point_code"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

type Points []*Point

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

type MemberWithPointCategory struct {
	MemberID        int `json:"member_id"`
	PurchaseEarning int `json:"purchase_earning"`
	ReferralBonus   int `json:"referral_bonus"`
	Cashback        int `json:"cashback"`
	Redemption      int `json:"redemption"`
	PointsBalance   int `json:"points_balance"`
}
