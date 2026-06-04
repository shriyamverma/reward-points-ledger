package repository

import "reward-points-ledger/internal/domain"

type Repository interface {
	CreateMember(name, email string) (*domain.Member, error)
	GetMemberByID(id int) (*domain.Member, error)
	AddRewardEntry(memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error)
	GetRewardsByMemberID(id int) ([]domain.RewardEntry, error)
	GetBalance(memberID int) (int, error)
}
