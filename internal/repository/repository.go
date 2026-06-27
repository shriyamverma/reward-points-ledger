package repository

import (
	"context"
	"reward-points-ledger/internal/domain"
)

type Repository interface {
	CreateMember(ctx context.Context, name, email string) (*domain.Member, error)
	GetMemberByID(ctx context.Context, id int) (*domain.Member, error)
	AddRewardEntry(ctx context.Context, memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error)
	GetRewardsByMemberID(ctx context.Context, id int) ([]domain.RewardEntry, error)
	GetBalance(ctx context.Context, memberID int) (int, error)

	GetMembers(ctx context.Context) ([]domain.Member, error)
	GetRewards(ctx context.Context) ([]domain.RewardEntry, error)
	GetMemberWithPointCategory(ctx context.Context, id int) (*domain.MemberWithPointCategory, error)

	CreatePoints(ctx context.Context, pointTypeID int, pointCode string) (*domain.Point, error)
	GetPointDetailsByPointType(ctx context.Context, pointTypeID int) (*domain.Point, error)
	GetAllPoints(ctx context.Context) (*domain.Points, error)
	ActivatePoint(context.Context, int) (*domain.Point, error)
}
