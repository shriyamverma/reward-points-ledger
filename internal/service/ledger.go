package service

import (
	"context"
	"reward-points-ledger/internal/domain"
	"reward-points-ledger/internal/repository"
)

type LedgerService struct {
	repo repository.Repository
}

func NewLedgerService(repo repository.Repository) *LedgerService {
	return &LedgerService{repo: repo}
}

func (s *LedgerService) CreateMember(ctx context.Context, name, email string) (*domain.Member, error) {
	return s.repo.CreateMember(ctx, name, email)
}

func (s *LedgerService) GetMember(ctx context.Context, id int) (*domain.Member, error) {
	member, err := s.repo.GetMemberByID(ctx, id)
	if err != nil {
		return nil, err
	}
	balance, err := s.repo.GetBalance(ctx, id)
	if err != nil {
		return nil, err
	}
	member.PointsBalance = balance
	return member, nil
}

func (s *LedgerService) GetRewards(ctx context.Context, memberID int) ([]domain.RewardEntry, error) {
	// Verify user profile exists first
	if _, err := s.repo.GetMemberByID(ctx, memberID); err != nil {
		return nil, err
	}
	return s.repo.GetRewardsByMemberID(ctx, memberID)
}

func (s *LedgerService) ProcessReward(ctx context.Context, memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error) {
	if pointTypeID < 1 || pointTypeID > 4 {
		return nil, domain.ErrInvalidPointType
	}
	if points <= 0 {
		return nil, domain.ErrPointsNotPositive
	}

	// Verify target identity exists
	if _, err := s.repo.GetMemberByID(ctx, memberID); err != nil {
		return nil, err
	}

	calculatedPoints := points
	if pointTypeID == domain.TypeRedemption {
		calculatedPoints = -points

		currentBalance, err := s.repo.GetBalance(ctx, memberID)
		if err != nil {
			return nil, err
		}
		if currentBalance+calculatedPoints < 0 {
			return nil, domain.ErrInsufficientBalance
		}
	}

	return s.repo.AddRewardEntry(ctx, memberID, pointTypeID, calculatedPoints, desc)
}
