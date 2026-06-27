package repository

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
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

func (r *PostgresRepository) CreateMember(ctx context.Context, name, email string) (*domain.Member, error) {
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
	reqID := middleware.GetReqID(ctx)
	slog.Debug("executing database raw query", "request_id", reqID, "op", "CreateMember", "query", query, "args", args)

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

func (r *PostgresRepository) GetMemberByID(ctx context.Context, memberID int) (*domain.Member, error) {
	query := `SELECT member_id, name, email, created_at FROM members WHERE member_id = @member_id`

	args := pgx.NamedArgs{"member_id": memberID}
	reqID := middleware.GetReqID(ctx)
	slog.Debug("executing database raw query", "request_id", reqID, "op", "GetMemberByID", "query", query, "args", args)

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

func (r *PostgresRepository) AddRewardEntry(ctx context.Context, memberID, pointTypeID, points int, desc string) (*domain.RewardEntry, error) {
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
	reqID := middleware.GetReqID(ctx)
	slog.Debug("executing database raw query", "request_id", reqID, "op", "AddRewardEntry", "query", query, "args", args)

	var rewardID int
	var dbEventDate time.Time
	err := r.pool.QueryRow(ctx, query, args).Scan(&rewardID, &dbEventDate)
	if err != nil {
		return nil, err
	}

	slog.Info("reward points successfully updated in ledger",
		"request_id", middleware.GetReqID(ctx),
		"member_id", memberID,
		"points_added", points,
	)

	return &domain.RewardEntry{
		RewardID:    rewardID,
		MemberID:    memberID,
		PointTypeID: pointTypeID,
		Points:      points,
		Description: desc,
		EventDate:   dbEventDate.Format(time.RFC3339),
	}, nil
}

func (r *PostgresRepository) GetRewardsByMemberID(ctx context.Context, id int) ([]domain.RewardEntry, error) {
	query := `SELECT reward_id, member_id, point_type_id, points, description, event_date FROM rewards WHERE member_id = @member_id`

	args := pgx.NamedArgs{"member_id": id}
	reqID := middleware.GetReqID(ctx)
	slog.Debug("executing database raw query", "request_id", reqID, "op", "GetRewardsByMemberID", "query", query, "args", args)

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

func (r *PostgresRepository) GetBalance(ctx context.Context, memberID int) (int, error) {
	query := `SELECT COALESCE(SUM(points), 0) FROM rewards WHERE member_id = @member_id`

	args := pgx.NamedArgs{"member_id": memberID}
	// Trace the exact SQL footprint
	reqID := middleware.GetReqID(ctx)
	slog.Debug("executing database raw query", "request_id", reqID, "op", "GetBalance", "query", query, "args", args)

	var balance int
	err := r.pool.QueryRow(ctx, query, args).Scan(&balance)
	return balance, err
}

func (r *PostgresRepository) GetMembers(ctx context.Context) ([]domain.Member, error) {
	query := `SELECT member_id, name, email, created_at FROM members`

	args := pgx.NamedArgs{}
	reqID := middleware.GetReqID(ctx)
	slog.Debug("executing database raw query", "request_id", reqID, "op", "GetMemberByID", "query", query, "args", args)

	var allMembers []domain.Member
	rows, err := r.pool.Query(ctx, query, args)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var member domain.Member
		var createdAtTime time.Time
		rows.Scan(&member.MemberID, &member.Name, &member.Email, &createdAtTime)
		member.CreatedAt = createdAtTime.Format(time.RFC3339)
		allMembers = append(allMembers, member)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return allMembers, nil
}

func (r *PostgresRepository) GetRewards(ctx context.Context) ([]domain.RewardEntry, error) {
	query := `SELECT reward_id, member_id, point_type_id, points, description, event_date FROM rewards`

	args := pgx.NamedArgs{}
	reqID := middleware.GetReqID(ctx)
	slog.Debug("executing database raw query", "request_id", reqID, "op", "GetRewardsByMemberID", "query", query, "args", args)

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

func (r *PostgresRepository) GetMemberWithPointCategory(ctx context.Context, id int) (*domain.MemberWithPointCategory, error) {
	query := `SELECT point_type_id, COALESCE(SUM(points), 0) FROM rewards WHERE member_id = @member_id GROUP BY point_type_id`

	args := pgx.NamedArgs{"member_id": id}
	reqID := middleware.GetReqID(ctx)
	slog.Debug("executing database raw query", "request_id", reqID, "op", "GetMemberWithPointCategory", "query", query, "args", args)

	member := domain.MemberWithPointCategory{
		MemberID: id,
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasRows := false

	for rows.Next() {
		hasRows = true

		var pointType int
		sum := 0
		err := rows.Scan(&pointType, &sum)
		if err != nil {
			return nil, err
		}

		switch pointType {
		case domain.TypePurchaseEarning:
			member.PurchaseEarning = sum
		case domain.TypeReferralBonus:
			member.ReferralBonus = sum
		case domain.TypeCashback:
			member.Cashback = sum
		case domain.TypeRedemption:
			member.Redemption = sum
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if !hasRows {
		return nil, domain.ErrRewardNotFound
	}

	member.PointsBalance = member.PurchaseEarning + member.ReferralBonus + member.Cashback + member.Redemption

	return &member, nil
}

func (r *PostgresRepository) CreatePoints(ctx context.Context, pointTypeID int, pointCode string) (*domain.Point, error) {
	query := `INSERT INTO points(point_type_id, point_code) VALUES (@point_type_id, @point_code)
				RETURNING point_id, is_active, created_at`

	args := pgx.NamedArgs{
		"point_type_id": pointTypeID, "point_code": pointCode,
	}
	slog.Debug("executing database raw query", "op", "CreatePoints", "query", query, "args", args)

	point := &domain.Point{}
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, query, args).Scan(&point.PointID, &point.IsActive, &createdAt)
	if err != nil {
		slog.Debug("failed to insert point", "op", "CreatePoints", "query", query, "args", args)
		if strings.ContainsAny(err.Error(), "already exists") {
			return nil, domain.ErrDuplicatePointTypeID
		}
		return nil, err
	}
	point.PointTypeID = pointTypeID
	point.PointCode = pointCode
	point.CreatedAt = createdAt.Format(time.RFC3339)
	return point, nil
}

func (r *PostgresRepository) GetPointDetailsByPointType(ctx context.Context, pointTypeID int) (*domain.Point, error) {
	query := `SELECT point_id, point_type_id, point_code, is_active, created_at FROM points WHERE point_type_id = @point_type_id`

	args := pgx.NamedArgs{"point_type_id": pointTypeID}
	slog.Debug("executing database raw query", "op", "GetPointDetailsByPointType", "query", query, "args", args)

	point := &domain.Point{}
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, query, args).Scan(&point.PointID, &point.PointTypeID, &point.PointCode, &point.IsActive, &createdAt)
	if err != nil {
		return nil, domain.ErrPointNotFound
	}
	point.CreatedAt = createdAt.Format(time.RFC3339)
	return point, nil
}

func (r *PostgresRepository) GetAllPoints(ctx context.Context) (*domain.Points, error) {
	query := `SELECT point_id, point_type_id, point_code, is_active, created_at FROM points`

	args := pgx.NamedArgs{}

	slog.Debug("executing database raw query", "op", "GetAllPoints", "query", query, "args", args)

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := domain.Points{}
	for rows.Next() {
		point := &domain.Point{}
		var createdAt time.Time
		err := rows.Scan(&point.PointID, &point.PointTypeID, &point.PointCode, &point.IsActive, &createdAt)
		if err != nil {
			slog.Warn("Database error reading rows - inside", "error", err)
			return nil, err
		}
		point.CreatedAt = createdAt.Format(time.RFC3339)
		points = append(points, point)
	}
	if err = rows.Err(); err != nil {
		slog.Warn("Database error reading rows - outside", "error", err)
		return nil, err
	}

	slog.Debug("Got response to GetAllPoints", "response", points)

	return &points, nil
}

func (r *PostgresRepository) ActivatePoint(ctx context.Context, pointTypeID int) (*domain.Point, error) {
	query := `UPDATE points SET is_active = true WHERE point_type_id = @point_type_id
				RETURNING point_id, point_type_id, point_code, is_active, created_at`
	args := pgx.NamedArgs{
		"point_type_id": pointTypeID,
	}

	slog.Debug("executing database raw query", "op", "ActivatePoint", "query", query, "args", args)

	point := &domain.Point{}
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, query, args).Scan(&point.PointID, &point.PointTypeID, &point.PointCode, &point.IsActive, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPointNotFound
		}
		return nil, err
	}
	point.CreatedAt = createdAt.Format(time.RFC3339)
	return point, nil
}
