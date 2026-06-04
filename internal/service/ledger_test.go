package service

import (
	"errors"
	"testing"

	"reward-points-ledger/internal/domain"
)

// ============================================================================
// MOCK REPOSITORY ARCHITECTURE
// ============================================================================

type mockRepository struct {
	CreateMemberFunc         func(name, email string) (*domain.Member, error)
	GetMemberByIDFunc        func(id int) (*domain.Member, error)
	AddRewardEntryFunc       func(memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error)
	GetRewardsByMemberIDFunc func(id int) ([]domain.RewardEntry, error)
	GetBalanceFunc           func(memberID int) (int, error)
}

func (m *mockRepository) CreateMember(name, email string) (*domain.Member, error) {
	return m.CreateMemberFunc(name, email)
}
func (m *mockRepository) GetMemberByID(id int) (*domain.Member, error) {
	return m.GetMemberByIDFunc(id)
}
func (m *mockRepository) AddRewardEntry(memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error) {
	return m.AddRewardEntryFunc(memberID, pointTypeID, points, desc)
}
func (m *mockRepository) GetRewardsByMemberID(id int) ([]domain.RewardEntry, error) {
	return m.GetRewardsByMemberIDFunc(id)
}
func (m *mockRepository) GetBalance(memberID int) (int, error) {
	return m.GetBalanceFunc(memberID)
}

// ============================================================================
// MEMBER PROFILE CREATION TESTS
// ============================================================================

func TestLedgerService_CreateMember(t *testing.T) {
	tests := []struct {
		name          string
		memberName    string
		memberEmail   string
		mockSetup     func(mock *mockRepository)
		expectedError error
		expectNil     bool
	}{
		{
			name:        "Successful registration",
			memberName:  "Alice Smith",
			memberEmail: "alice@example.com",
			mockSetup: func(m *mockRepository) {
				m.CreateMemberFunc = func(name, email string) (*domain.Member, error) {
					return &domain.Member{MemberID: 1, Name: name, Email: email}, nil
				}
			},
			expectedError: nil,
			expectNil:     false,
		},
		{
			name:        "Rejected duplicate email registration",
			memberName:  "Bob Duplicate",
			memberEmail: "alice@example.com",
			mockSetup: func(m *mockRepository) {
				m.CreateMemberFunc = func(name, email string) (*domain.Member, error) {
					return nil, domain.ErrDuplicateEmail
				}
			},
			expectedError: domain.ErrDuplicateEmail,
			expectNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			svc := NewLedgerService(mock)

			res, err := svc.CreateMember(tt.memberName, tt.memberEmail)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("Expected error type %v, got: %v", tt.expectedError, err)
			}
			if tt.expectNil && res != nil {
				t.Errorf("Expected nil response payload, got struct reference: %v", res)
			}
		})
	}
}

// ============================================================================
// REWARD PROCESSING & TRANSACTION ENGINE TESTS
// ============================================================================

func TestLedgerService_ProcessReward(t *testing.T) {
	tests := []struct {
		name          string
		memberID      int
		pointType     int
		points        int
		description   string
		mockSetup     func(mock *mockRepository)
		expectedError error
	}{
		{
			name:        "Successful earning credit transaction",
			memberID:    1,
			pointType:   domain.TypePurchaseEarning,
			points:      150,
			description: "Store purchase crediting",
			mockSetup: func(m *mockRepository) {
				m.GetMemberByIDFunc = func(id int) (*domain.Member, error) {
					return &domain.Member{MemberID: 1}, nil
				}
				m.AddRewardEntryFunc = func(mID, ptID, pts int, desc string) (*domain.RewardEntry, error) {
					return &domain.RewardEntry{RewardID: 101, Points: pts}, nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "Successful redemption debit transaction",
			memberID:    1,
			pointType:   domain.TypeRedemption,
			points:      50,
			description: "Valid gift card cashout",
			mockSetup: func(m *mockRepository) {
				m.GetMemberByIDFunc = func(id int) (*domain.Member, error) {
					return &domain.Member{MemberID: 1}, nil
				}
				m.GetBalanceFunc = func(memberID int) (int, error) {
					return 100, nil // Sufficient balance available
				}
				m.AddRewardEntryFunc = func(mID, ptID, pts int, desc string) (*domain.RewardEntry, error) {
					return &domain.RewardEntry{RewardID: 102, Points: -pts}, nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "Rejected overdraft redemption debit",
			memberID:    1,
			pointType:   domain.TypeRedemption,
			points:      200,
			description: "Illegal overdraft attempt",
			mockSetup: func(m *mockRepository) {
				m.GetMemberByIDFunc = func(id int) (*domain.Member, error) {
					return &domain.Member{MemberID: 1}, nil
				}
				m.GetBalanceFunc = func(memberID int) (int, error) {
					return 50, nil // 50 available < 200 requested = overdraft!
				}
			},
			expectedError: domain.ErrInsufficientBalance,
		},
		{
			name:        "Nonexistent member validation failure",
			memberID:    999,
			pointType:   domain.TypePurchaseEarning,
			points:      10,
			description: "Points allocation to dark hole",
			mockSetup: func(m *mockRepository) {
				m.GetMemberByIDFunc = func(id int) (*domain.Member, error) {
					return nil, domain.ErrMemberNotFound
				}
			},
			expectedError: domain.ErrMemberNotFound,
		},
		{
			name:          "Negative points amount validation rejection",
			memberID:      1,
			pointType:     domain.TypePurchaseEarning,
			points:        -50,
			description:   "Malformed request payload",
			mockSetup:     func(m *mockRepository) {},
			expectedError: domain.ErrPointsNotPositive, // <-- Matched to your models.go error variable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			svc := NewLedgerService(mock)

			_, err := svc.ProcessReward(tt.memberID, tt.pointType, tt.points, tt.description)

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error boundary mismatch. Want: %v, Got: %v", tt.expectedError, err)
			}
		})
	}
}

// ============================================================================
// PROFILE RETRIEVAL & HYDRATION TESTS
// ============================================================================

func TestLedgerService_GetMember(t *testing.T) {
	mock := &mockRepository{}
	svc := NewLedgerService(mock)

	t.Run("Hydrates profile with balance summation metrics", func(t *testing.T) {
		mock.GetMemberByIDFunc = func(id int) (*domain.Member, error) {
			return &domain.Member{MemberID: 7, Name: "Charlie"}, nil
		}
		mock.GetBalanceFunc = func(memberID int) (int, error) {
			return 420, nil
		}

		member, err := svc.GetMember(7)
		if err != nil {
			t.Fatalf("Expected clean hydration, got error: %v", err)
		}
		if member.PointsBalance != 420 {
			t.Errorf("Expected points balance to be aggregate 420, got: %d", member.PointsBalance)
		}
	})
}
