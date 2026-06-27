package repository

import (
	"context"
	"reward-points-ledger/internal/domain"
)

type MockRepository struct {
	CreateMemberFunc               func(ctx context.Context, name, email string) (*domain.Member, error)
	GetMemberByIDFunc              func(ctx context.Context, memberID int) (*domain.Member, error)
	AddRewardEntryFunc             func(ctx context.Context, memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error)
	GetRewardsByMemberIDFunc       func(ctx context.Context, id int) ([]domain.RewardEntry, error)
	GetBalanceFunc                 func(ctx context.Context, memberID int) (int, error)
	GetMembersFunc                 func(ctx context.Context) ([]domain.Member, error)
	GetRewardsFunc                 func(ctx context.Context) ([]domain.RewardEntry, error)
	GetMemberWithPointCategoryFunc func(ctx context.Context, id int) (*domain.MemberWithPointCategory, error)
	CreatePointsFunc               func(ctx context.Context, pointTypeID int, pointCode string) (*domain.Point, error)
	GetPointDetailsByPointTypeFunc func(ctx context.Context, pointTypeID int) (*domain.Point, error)
	GetAllPointsFunc               func(ctx context.Context) (*domain.Points, error)
	ActivatePointFunc              func(ctx context.Context, pointTypeID int) (*domain.Point, error)
}

func (m *MockRepository) CreateMember(ctx context.Context, name, email string) (*domain.Member, error) {
	return m.CreateMemberFunc(ctx, name, email)
}
func (m *MockRepository) GetMemberByID(ctx context.Context, memberID int) (*domain.Member, error) {
	return m.GetMemberByIDFunc(ctx, memberID)
}
func (m *MockRepository) AddRewardEntry(ctx context.Context, memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error) {
	return m.AddRewardEntryFunc(ctx, memberID, pointTypeID, points, desc)
}
func (m *MockRepository) GetRewardsByMemberID(ctx context.Context, id int) ([]domain.RewardEntry, error) {
	return m.GetRewardsByMemberIDFunc(ctx, id)
}
func (m *MockRepository) GetBalance(ctx context.Context, memberID int) (int, error) {
	return m.GetBalanceFunc(ctx, memberID)
}
func (m *MockRepository) GetMembers(ctx context.Context) ([]domain.Member, error) {
	return m.GetMembersFunc(ctx)
}
func (m *MockRepository) GetRewards(ctx context.Context) ([]domain.RewardEntry, error) {
	return m.GetRewardsFunc(ctx)
}
func (m *MockRepository) GetMemberWithPointCategory(ctx context.Context, memberID int) (*domain.MemberWithPointCategory, error) {
	return m.GetMemberWithPointCategoryFunc(ctx, memberID)
}
func (m *MockRepository) CreatePoints(ctx context.Context, pointTypeID int, pointCode string) (*domain.Point, error) {
	return m.CreatePointsFunc(ctx, pointTypeID, pointCode)
}
func (m *MockRepository) GetPointDetailsByPointType(ctx context.Context, pointTypeID int) (*domain.Point, error) {
	return m.GetPointDetailsByPointTypeFunc(ctx, pointTypeID)
}
func (m *MockRepository) GetAllPoints(ctx context.Context) (*domain.Points, error) {
	return m.GetAllPointsFunc(ctx)
}
func (m *MockRepository) ActivatePoint(ctx context.Context, pointTypeID int) (*domain.Point, error) {
	return m.ActivatePointFunc(ctx, pointTypeID)
}
