package repository

import (
	"context"
	"reward-points-ledger/internal/domain"
)

type MockRepository struct {
	CreateMemberFunc         func(ctx context.Context, name, email string) (*domain.Member, error)
	GetMemberByIDFunc        func(ctx context.Context, id int) (*domain.Member, error)
	AddRewardEntryFunc       func(ctx context.Context, memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error)
	GetRewardsByMemberIDFunc func(ctx context.Context, id int) ([]domain.RewardEntry, error)
	GetBalanceFunc           func(ctx context.Context, memberID int) (int, error)
}

func (m *MockRepository) CreateMember(ctx context.Context, name, email string) (*domain.Member, error) {
	return m.CreateMemberFunc(ctx, name, email)
}
func (m *MockRepository) GetMemberByID(ctx context.Context, id int) (*domain.Member, error) {
	return m.GetMemberByIDFunc(ctx, id)
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
