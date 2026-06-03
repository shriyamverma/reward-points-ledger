package service

import (
	"errors"
	"reward-points-ledger/internal/domain"
	"reward-points-ledger/internal/repository"
	"testing"
)

func TestLedgerService_ValidationRules(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := NewLedgerService(repo)

	// Test 1: Create valid profile
	m, err := svc.CreateMember("Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}

	// Test 2: Duplicate email rejection
	_, err = svc.CreateMember("Bob", "alice@example.com")
	if !errors.Is(err, domain.ErrDuplicateEmail) {
		t.Errorf("Expected ErrDuplicateEmail, got %v", err)
	}

	// Test 3: Earn transaction credits
	_, err = svc.ProcessReward(m.MemberID, domain.TypePurchaseEarning, 100, "Store purchase")
	if err != nil {
		t.Fatalf("Expected credit to succeed, got %v", err)
	}

	// Test 4: Reject overdraft debit transactions
	_, err = svc.ProcessReward(m.MemberID, domain.TypeRedemption, 150, "Overdraft cash-out")
	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Errorf("Expected ErrInsufficientBalance, got %v", err)
	}

	// Test 5: Allow valid balance debit redemptions
	_, err = svc.ProcessReward(m.MemberID, domain.TypeRedemption, 40, "Valid cash-out")
	if err != nil {
		t.Fatalf("Expected debit step to succeed, got %v", err)
	}

	// Check final remaining value calculation
	updatedMember, _ := svc.GetMember(m.MemberID)
	if updatedMember.PointsBalance != 60 {
		t.Errorf("Expected balance remaining to be 60, got %d", updatedMember.PointsBalance)
	}
}
