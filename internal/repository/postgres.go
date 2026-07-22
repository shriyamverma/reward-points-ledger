package repository

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgconn"

	"reward-points-ledger/internal/domain"

	"github.com/jackc/pgx/v5"
)

type PostgreSQLPool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PostgresRepository struct {
	pool PostgreSQLPool
}

func NewPostgresRepository(pool PostgreSQLPool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) logQuery(ctx context.Context, op, query string, args pgx.NamedArgs) {
	slog.Debug("executing database raw query",
		"request_id", middleware.GetReqID(ctx),
		"op", op,
		"query", query,
		"args", args,
	)
}

func (r *PostgresRepository) CreateMember(ctx context.Context, name, email string) (*domain.Member, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

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
	r.logQuery(ctx, "CreateMember", query, args)

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
	r.logQuery(ctx, "GetMemberByID", query, args)

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
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Lock the member row for update to ensure serialized execution
	var dummy int
	err = tx.QueryRow(ctx, "SELECT member_id FROM members WHERE member_id = @member_id FOR UPDATE", pgx.NamedArgs{"member_id": memberID}).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}

	// 2. Perform double-spend validation if redemption
	if points < 0 {
		var balance int
		err = tx.QueryRow(ctx, "SELECT COALESCE(SUM(points), 0) FROM rewards WHERE member_id = @member_id", pgx.NamedArgs{"member_id": memberID}).Scan(&balance)
		if err != nil {
			return nil, err
		}
		if balance+points < 0 {
			return nil, domain.ErrInsufficientBalance
		}
	}

	// 3. Atomically check if point type is active and insert the reward record
	query := `INSERT INTO rewards (member_id, point_type_id, points, description, event_date)
              SELECT @member_id, @point_type_id, @points, @description, NOW()
              WHERE EXISTS (
                  SELECT 1 FROM points WHERE point_type_id = @point_type_id AND is_active = true
              )
              RETURNING reward_id, event_date`

	args := pgx.NamedArgs{
		"member_id":     memberID,
		"point_type_id": pointTypeID,
		"points":        points,
		"description":   desc,
	}
	r.logQuery(ctx, "AddRewardEntry", query, args)

	var rewardID int
	var dbEventDate time.Time
	err = tx.QueryRow(ctx, query, args).Scan(&rewardID, &dbEventDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidPointType
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
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

func (r *PostgresRepository) GetRewardsByMemberID(ctx context.Context, id int, limit, cursorID int) ([]domain.RewardEntry, error) {
	query := `SELECT reward_id, member_id, point_type_id, points, description, event_date FROM rewards WHERE member_id = @member_id`
	if cursorID > 0 {
		query += ` AND reward_id < @cursor_id`
	}
	query += ` ORDER BY reward_id DESC LIMIT @limit`

	args := pgx.NamedArgs{
		"member_id": id,
		"limit":     limit,
	}
	if cursorID > 0 {
		args["cursor_id"] = cursorID
	}
	r.logQuery(ctx, "GetRewardsByMemberID", query, args)

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
	r.logQuery(ctx, "GetBalance", query, args)

	var balance int
	err := r.pool.QueryRow(ctx, query, args).Scan(&balance)
	return balance, err
}

func (r *PostgresRepository) GetAllMembers(ctx context.Context, limit, offset int) ([]domain.Member, error) {
	query := `SELECT member_id, name, email, created_at FROM members ORDER BY member_id LIMIT @limit OFFSET @offset`

	args := pgx.NamedArgs{
		"limit":  limit,
		"offset": offset,
	}
	r.logQuery(ctx, "GetAllMembers", query, args)

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allMembers []domain.Member
	for rows.Next() {
		var member domain.Member
		var createdAtTime time.Time
		if err := rows.Scan(&member.MemberID, &member.Name, &member.Email, &createdAtTime); err != nil {
			return nil, err
		}
		member.CreatedAt = createdAtTime.Format(time.RFC3339)
		allMembers = append(allMembers, member)
	}
	return allMembers, rows.Err()
}

func (r *PostgresRepository) GetAllRewards(ctx context.Context, limit, cursorID int) ([]domain.RewardEntry, error) {
	query := `SELECT reward_id, member_id, point_type_id, points, description, event_date FROM rewards`
	if cursorID > 0 {
		query += ` WHERE reward_id > @cursor_id`
	}
	query += ` ORDER BY reward_id ASC LIMIT @limit`

	args := pgx.NamedArgs{
		"limit": limit,
	}
	if cursorID > 0 {
		args["cursor_id"] = cursorID
	}
	r.logQuery(ctx, "GetAllRewards", query, args)

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

func (r *PostgresRepository) GetMemberPointSummary(ctx context.Context, id int) (*domain.MemberPointSummary, error) {
	query := `SELECT point_type_id, COALESCE(SUM(points), 0) FROM rewards WHERE member_id = @member_id GROUP BY point_type_id ORDER BY point_type_id`

	args := pgx.NamedArgs{"member_id": id}
	r.logQuery(ctx, "GetMemberPointSummary", query, args)

	summary := &domain.MemberPointSummary{MemberID: id}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category domain.CategoryBalance
		if err := rows.Scan(&category.PointTypeID, &category.Balance); err != nil {
			return nil, err
		}
		summary.Categories = append(summary.Categories, category)
		summary.PointsBalance += category.Balance
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(summary.Categories) == 0 {
		return nil, domain.ErrRewardNotFound
	}

	return summary, nil
}

func (r *PostgresRepository) CreatePoints(ctx context.Context, pointTypeID int, pointCode string) (*domain.Point, error) {
	query := `INSERT INTO points(point_type_id, point_code) VALUES (@point_type_id, @point_code)
				RETURNING point_id, is_active, created_at`

	args := pgx.NamedArgs{
		"point_type_id": pointTypeID,
		"point_code":    pointCode,
	}
	r.logQuery(ctx, "CreatePoints", query, args)

	point := &domain.Point{PointTypeID: pointTypeID, PointCode: pointCode}
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, query, args).Scan(&point.PointID, &point.IsActive, &createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicatePointTypeID
		}
		return nil, err
	}
	point.CreatedAt = createdAt.Format(time.RFC3339)
	return point, nil
}

func (r *PostgresRepository) GetPointDetailsByPointType(ctx context.Context, pointTypeID int) (*domain.Point, error) {
	query := `SELECT point_id, point_type_id, point_code, is_active, created_at FROM points WHERE point_type_id = @point_type_id`

	args := pgx.NamedArgs{"point_type_id": pointTypeID}
	r.logQuery(ctx, "GetPointDetailsByPointType", query, args)

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
	query := `SELECT point_id, point_type_id, point_code, is_active, created_at FROM points ORDER BY point_type_id`

	args := pgx.NamedArgs{}
	r.logQuery(ctx, "GetAllPoints", query, args)

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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &points, nil
}

func (r *PostgresRepository) SetPointActive(ctx context.Context, pointTypeID int, active bool) (*domain.Point, error) {
	query := `UPDATE points SET is_active = @is_active WHERE point_type_id = @point_type_id
				RETURNING point_id, point_type_id, point_code, is_active, created_at`
	args := pgx.NamedArgs{
		"point_type_id": pointTypeID,
		"is_active":     active,
	}

	r.logQuery(ctx, "SetPointActive", query, args)

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
