package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"reward-points-ledger/internal/domain"
)

// ============================================================================
// CREATE MEMBER TESTS
// ============================================================================

func TestPostgresRepository_CreateMember_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	name := "Alice Johnson"
	email := "alice@example.com"
	mockTime := time.Now().UTC()

	// Adjusted arguments payload and columns mapper definition
	mock.ExpectQuery(`INSERT INTO members`).
		WithArgs(pgx.NamedArgs{
			"name":  name,
			"email": email,
		}).
		WillReturnRows(pgxmock.NewRows([]string{"member_id", "created_at"}).AddRow(1, mockTime))

	member, err := repo.CreateMember(name, email)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if member == nil || member.MemberID != 1 {
		t.Errorf("Expected member_id to be 1, got: %v", member)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_CreateMember_DuplicateEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	name := "Alice Duplicate"
	email := "duplicate@example.com"

	mock.ExpectQuery(`INSERT INTO members`).
		WithArgs(pgx.NamedArgs{
			"name":  name,
			"email": email,
		}).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.CreateMember(name, email)
	if !errors.Is(err, domain.ErrDuplicateEmail) {
		t.Errorf("Expected domain.ErrDuplicateEmail, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

// ============================================================================
// GET MEMBER BY ID TESTS
// ============================================================================

func TestPostgresRepository_GetMemberByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	memberID := 42
	mockTime := time.Now().UTC()

	mock.ExpectQuery(`SELECT member_id, name, email, created_at FROM members`).
		WithArgs(pgx.NamedArgs{"id": memberID}).
		WillReturnRows(pgxmock.NewRows([]string{"member_id", "name", "email", "created_at"}).
			AddRow(memberID, "Bob Smith", "bob@example.com", mockTime))

	member, err := repo.GetMemberByID(memberID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if member == nil || member.Name != "Bob Smith" || member.Email != "bob@example.com" {
		t.Errorf("Returned wrong data or structure: %v", member)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_GetMemberByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	memberID := 999

	mock.ExpectQuery(`SELECT member_id, name, email, created_at FROM members`).
		WithArgs(pgx.NamedArgs{"id": memberID}).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetMemberByID(memberID)
	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("Expected domain.ErrMemberNotFound, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

// ============================================================================
// GET REWARDS BY MEMBER ID TESTS
// ============================================================================

func TestPostgresRepository_GetRewardsByMemberID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	memberID := 1
	mockTime := time.Now().UTC()

	rows := pgxmock.NewRows([]string{"reward_id", "member_id", "point_type_id", "points", "description", "event_date"}).
		AddRow(101, memberID, 1, 100, "Sign-up Bonus", mockTime).
		AddRow(102, memberID, 4, -50, "Coffee Purchase", mockTime)

	mock.ExpectQuery(`SELECT reward_id, member_id, point_type_id, points, description, event_date FROM rewards`).
		WithArgs(pgx.NamedArgs{"member_id": memberID}).
		WillReturnRows(rows)

	rewards, err := repo.GetRewardsByMemberID(memberID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(rewards) != 2 {
		t.Errorf("Expected exactly 2 ledger entries, got: %d", len(rewards))
	}
	if rewards[0].RewardID != 101 || rewards[1].Points != -50 {
		t.Errorf("Unexpected array element mapping content: %v", rewards)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_GetRewardsByMemberID_EmptyList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	memberID := 2

	rows := pgxmock.NewRows([]string{"reward_id", "member_id", "point_type_id", "points", "description", "event_date"})
	mock.ExpectQuery(`SELECT reward_id, member_id, point_type_id, points, description, event_date FROM rewards`).
		WithArgs(pgx.NamedArgs{"member_id": memberID}).
		WillReturnRows(rows)

	rewards, err := repo.GetRewardsByMemberID(memberID)
	if err != nil {
		t.Errorf("Expected zero runtime error on empty user ledger search, got: %v", err)
	}
	if len(rewards) != 0 {
		t.Errorf("Expected slice count to be 0, got: %d", len(rewards))
	}
}

// ============================================================================
// GET BALANCE TESTS
// ============================================================================

func TestPostgresRepository_GetBalance_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	memberID := 1

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(points\), 0\)`).
		WithArgs(pgx.NamedArgs{"member_id": memberID}).
		WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(450))

	balance, err := repo.GetBalance(memberID)
	if err != nil {
		t.Errorf("Expected successful balance calculation, got error: %v", err)
	}
	if balance != 450 {
		t.Errorf("Expected balance to be 450, got: %d", balance)
	}
}

// ============================================================================
// ADD REWARD ENTRY TESTS
// ============================================================================

func TestPostgresRepository_AddRewardEntry_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)
	memberID := 1
	pointTypeID := 1
	points := 200
	description := "Referral Credit"
	mockTime := time.Now().UTC()

	// Updated to verify that event_date comes back from native DB RETURNING engine
	mock.ExpectQuery(`INSERT INTO rewards`).
		WithArgs(pgx.NamedArgs{
			"member_id":     memberID,
			"point_type_id": pointTypeID,
			"points":        points,
			"description":   description,
		}).
		WillReturnRows(pgxmock.NewRows([]string{"reward_id", "event_date"}).AddRow(999, mockTime))

	entry, err := repo.AddRewardEntry(memberID, pointTypeID, points, description)
	if err != nil {
		t.Errorf("Expected safe append operation, got error: %v", err)
	}
	if entry == nil || entry.RewardID != 999 {
		t.Errorf("Failed to assign generated transaction ID payload correctly: %v", entry)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestPostgresRepository_GenericDatabaseErrorPropagation(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to initialize mock pool: %v", err)
	}
	defer mock.Close()

	repo := NewPostgresRepository(mock)

	dbFatalErr := errors.New("conn_lost: driver closed unexpected pipe protocol channel")
	mock.ExpectQuery(`SELECT COALESCE`).
		WithArgs(pgx.NamedArgs{"member_id": 1}).
		WillReturnError(dbFatalErr)

	_, err = repo.GetBalance(1)
	if err == nil || err.Error() != "conn_lost: driver closed unexpected pipe protocol channel" {
		t.Errorf("Expected generic driver error to bubble up verbatim, got: %v", err)
	}
}
