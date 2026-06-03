package service

import (
	"reward-points-ledger/internal/domain"
	"reward-points-ledger/internal/repository"
)

type LedgerService struct {
	repo repository.LedgerRepository
}

func NewLedgerService(repo repository.LedgerRepository) *LedgerService {
	return &LedgerService{repo: repo}
}

func (s *LedgerService) CreateMember(name, email string) (*domain.Member, error) {
	return s.repo.CreateMember(name, email)
}

func (s *LedgerService) GetMember(id int) (*domain.Member, error) {
	member, err := s.repo.GetMemberByID(id)
	if err != nil {
		return nil, err
	}
	balance, err := s.repo.GetBalance(id)
	if err != nil {
		return nil, err
	}
	member.PointsBalance = balance
	return member, nil
}

func (s *LedgerService) GetRewards(memberID int) ([]domain.RewardEntry, error) {
	// Verify user profile exists first
	if _, err := s.repo.GetMemberByID(memberID); err != nil {
		return nil, err
	}
	return s.repo.GetRewardsByMemberID(memberID)
}

func (s *LedgerService) ProcessReward(memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error) {
	if pointTypeID < 1 || pointTypeID > 4 {
		return nil, domain.ErrInvalidPointType
	}
	if points <= 0 {
		return nil, domain.ErrPointsNotPositive
	}

	// Verify target identity exists
	if _, err := s.repo.GetMemberByID(memberID); err != nil {
		return nil, err
	}

	calculatedPoints := points
	if pointTypeID == domain.TypeRedemption {
		calculatedPoints = -points

		currentBalance, err := s.repo.GetBalance(memberID)
		if err != nil {
			return nil, err
		}
		if currentBalance+calculatedPoints < 0 {
			return nil, domain.ErrInsufficientBalance
		}
	}

	return s.repo.AddRewardEntry(memberID, pointTypeID, calculatedPoints, desc)
}
