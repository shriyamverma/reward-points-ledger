package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"reward-points-ledger/internal/domain"
)

type PostgreSQLPool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	pool PostgreSQLPool
}

func NewPostgresRepository(pool PostgreSQLPool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateMember(name, email string) (*domain.Member, error) {
	ctx := context.Background()
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	// Let Postgres handle NOW() inside the CTE layer
	query := `
       WITH new_member AS (
          SELECT @name AS name, @email AS email, NOW() AS created_at
       )
       INSERT INTO members (name, email, created_at)
       SELECT name, email, created_at 
       FROM new_member
       WHERE NOT EXISTS (
          SELECT 1 FROM members WHERE members.email = new_member.email
       )
       RETURNING member_id, created_at;
    `

	args := pgx.NamedArgs{
		"name":  name,
		"email": cleanEmail,
	}

	var memberID int
	var dbCreatedAt time.Time
	err := r.pool.QueryRow(ctx, query, args).Scan(&memberID, &dbCreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDuplicateEmail
		}
		return nil, err
	}

	return &domain.Member{
		MemberID:  memberID,
		Name:      name,
		Email:     email,
		CreatedAt: dbCreatedAt.Format(time.RFC3339),
	}, nil
}

func (r *PostgresRepository) GetMemberByID(id int) (*domain.Member, error) {
	ctx := context.Background()
	query := `SELECT member_id, name, email, created_at FROM members WHERE member_id = @id`

	args := pgx.NamedArgs{"id": id}

	var m domain.Member
	var createdAtTime time.Time
	err := r.pool.QueryRow(ctx, query, args).Scan(&m.MemberID, &m.Name, &m.Email, &createdAtTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}
	m.CreatedAt = createdAtTime.Format(time.RFC3339)
	return &m, nil
}

func (r *PostgresRepository) AddRewardEntry(memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error) {
	ctx := context.Background()

	// Uniformly updated to use NOW() on insert and scan it right back out
	query := `INSERT INTO rewards (member_id, point_type_id, points, description, event_date) 
              VALUES (@member_id, @point_type_id, @points, @description, NOW()) 
              RETURNING reward_id, event_date`

	args := pgx.NamedArgs{
		"member_id":     memberID,
		"point_type_id": pointTypeID,
		"points":        points,
		"description":   desc,
	}

	var rewardID int
	var dbEventDate time.Time
	err := r.pool.QueryRow(ctx, query, args).Scan(&rewardID, &dbEventDate)
	if err != nil {
		return nil, err
	}

	return &domain.RewardEntry{
		RewardID:    rewardID,
		MemberID:    memberID,
		PointTypeID: pointTypeID,
		Points:      points,
		Description: desc,
		EventDate:   dbEventDate.Format(time.RFC3339),
	}, nil
}

func (r *PostgresRepository) GetRewardsByMemberID(id int) ([]domain.RewardEntry, error) {
	ctx := context.Background()
	query := `SELECT reward_id, member_id, point_type_id, points, description, event_date FROM rewards WHERE member_id = @member_id`

	args := pgx.NamedArgs{"member_id": id}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.RewardEntry
	for rows.Next() {
		var rw domain.RewardEntry
		var eventTime time.Time
		if err := rows.Scan(&rw.RewardID, &rw.MemberID, &rw.PointTypeID, &rw.Points, &rw.Description, &eventTime); err != nil {
			return nil, err
		}
		rw.EventDate = eventTime.Format(time.RFC3339)
		results = append(results, rw)
	}
	return results, nil
}

func (r *PostgresRepository) GetBalance(memberID int) (int, error) {
	ctx := context.Background()
	query := `SELECT COALESCE(SUM(points), 0) FROM rewards WHERE member_id = @member_id`

	args := pgx.NamedArgs{"member_id": memberID}

	var balance int
	err := r.pool.QueryRow(ctx, query, args).Scan(&balance)
	return balance, err
}
