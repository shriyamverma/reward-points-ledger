package repository

import (
	"context"
	"reward-points-ledger/internal/domain"
	"strings"
	"sync"
	"time"
)

type MemoryRepository struct {
	mu           sync.RWMutex
	members      map[int]*domain.Member
	rewards      []domain.RewardEntry
	nextMemberID int
	nextRewardID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		members:      make(map[int]*domain.Member),
		rewards:      make([]domain.RewardEntry, 0),
		nextMemberID: 1,
		nextRewardID: 1,
	}
}

func (r *MemoryRepository) CreateMember(ctx context.Context, name, email string) (*domain.Member, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, m := range r.members {
		if strings.ToLower(m.Email) == strings.ToLower(email) {
			return nil, domain.ErrDuplicateEmail
		}
	}

	member := &domain.Member{
		MemberID:  r.nextMemberID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	r.members[member.MemberID] = member
	r.nextMemberID++
	return member, nil
}

func (r *MemoryRepository) GetMemberByID(ctx context.Context, id int) (*domain.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	member, exists := r.members[id]
	if !exists {
		return nil, domain.ErrMemberNotFound
	}
	// Return a copy to prevent external mutation
	return &domain.Member{
		MemberID:  member.MemberID,
		Name:      member.Name,
		Email:     member.Email,
		CreatedAt: member.CreatedAt,
	}, nil
}

func (r *MemoryRepository) GetRewardsByMemberID(ctx context.Context, id int) ([]domain.RewardEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.RewardEntry
	for _, rw := range r.rewards {
		if rw.MemberID == id {
			result = append(result, rw)
		}
	}
	return result, nil
}

func (r *MemoryRepository) GetBalance(ctx context.Context, memberID int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	balance := 0
	// Rewards ledger is the source of truth
	for _, rw := range r.rewards {
		if rw.MemberID == memberID {
			balance += rw.Points
		}
	}
	return balance, nil
}

func (r *MemoryRepository) AddRewardEntry(ctx context.Context, memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := domain.RewardEntry{
		RewardID:    r.nextRewardID,
		MemberID:    memberID,
		PointTypeID: pointTypeID,
		Points:      points,
		Description: desc,
		EventDate:   time.Now().UTC().Format(time.RFC3339),
	}
	r.rewards = append(r.rewards, entry)
	r.nextRewardID++
	return &entry, nil
}
