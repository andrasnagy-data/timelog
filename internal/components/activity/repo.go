package activity

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	repoer interface {
		Create(ctx context.Context, userID uuid.UUID, req CreateActivityIn) (*CreateActivityOut, error)
		Delete(ctx context.Context, userID uuid.UUID, id int) error
		List(ctx context.Context, userID uuid.UUID, query GetActivitiesQuery) (*GetActivitiesResponse, error)
		GetByID(ctx context.Context, userID uuid.UUID, id int) (*CreateActivityOut, error)
		Update(ctx context.Context, userID uuid.UUID, id int, req UpdateActivityIn) (*CreateActivityOut, error)
	}

	repo struct {
		pool *pgxpool.Pool
	}
)

func NewRepo(pool *pgxpool.Pool) repoer {
	return &repo{pool: pool}
}

func (r *repo) Create(ctx context.Context, userID uuid.UUID, req CreateActivityIn) (*CreateActivityOut, error) {
	activityOut := new(CreateActivityOut)

	stmt := `
	INSERT INTO activities (
		user_id, activity_name, duration, date
	)
	VALUES (
		$1, $2, $3, $4
	)
	RETURNING *`

	err := r.pool.QueryRow(
		ctx,
		stmt,
		userID,
		req.ActivityName,
		req.Duration,
		req.Date,
	).Scan(
		&activityOut.ID,
		&activityOut.UserID,
		&activityOut.ActivityName,
		&activityOut.Duration,
		&activityOut.Date,
		&activityOut.CreatedAt,
		&activityOut.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return activityOut, nil
}

// List retrieves paginated activities for a user with optional date filtering.
// It dynamically builds WHERE clauses based on provided date filters and returns
// both the activities and pagination metadata. Results are ordered by date DESC, then created_at DESC.
func (r *repo) List(ctx context.Context, userID uuid.UUID, query GetActivitiesQuery) (*GetActivitiesResponse, error) {
	offset := (query.Page - 1) * query.Limit

	// Build dynamic WHERE clause for date filtering
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if query.StartDate != nil {
		whereClause += fmt.Sprintf(" AND date >= $%d", argIndex)
		args = append(args, *query.StartDate)
		argIndex++
	}

	if query.EndDate != nil {
		whereClause += fmt.Sprintf(" AND date <= $%d", argIndex)
		args = append(args, *query.EndDate)
		argIndex++
	}

	// Get total count with same filters
	countStmt := fmt.Sprintf("SELECT COUNT(*) FROM activities %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countStmt, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Get activities with same filters
	stmt := fmt.Sprintf(`
	SELECT id, user_id, activity_name, duration, date, created_at, updated_at
	FROM activities 
	%s
	ORDER BY date DESC, created_at DESC
	LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	// Add LIMIT and OFFSET args
	args = append(args, query.Limit, offset)

	rows, err := r.pool.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []CreateActivityOut
	for rows.Next() {
		var activity CreateActivityOut
		err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.ActivityName,
			&activity.Duration,
			&activity.Date,
			&activity.CreatedAt,
			&activity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return &GetActivitiesResponse{
		Activities: activities,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
	}, nil
}

func (r *repo) GetByID(ctx context.Context, userID uuid.UUID, id int) (*CreateActivityOut, error) {
	stmt := `
	SELECT id, user_id, activity_name, duration, date, created_at, updated_at
	FROM activities 
	WHERE id = $1 AND user_id = $2`

	var activity CreateActivityOut
	err := r.pool.QueryRow(ctx, stmt, id, userID).Scan(
		&activity.ID,
		&activity.UserID,
		&activity.ActivityName,
		&activity.Duration,
		&activity.Date,
		&activity.CreatedAt,
		&activity.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &activity, nil
}

// Update performs partial updates on activities by dynamically building SET clauses
// only for non-nil fields. If no fields are provided for update, it returns the current activity.
// The updated_at timestamp is automatically set for any update.
func (r *repo) Update(ctx context.Context, userID uuid.UUID, id int, req UpdateActivityIn) (*CreateActivityOut, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{id, userID}
	argIndex := 3

	if req.ActivityName != nil {
		setParts = append(setParts, fmt.Sprintf("activity_name = $%d", argIndex))
		args = append(args, *req.ActivityName)
		argIndex++
	}
	if req.Duration != nil {
		setParts = append(setParts, fmt.Sprintf("duration = $%d", argIndex))
		args = append(args, *req.Duration)
		argIndex++
	}
	if req.Date != nil {
		setParts = append(setParts, fmt.Sprintf("date = $%d", argIndex))
		args = append(args, *req.Date)
		argIndex++
	}

	if len(setParts) == 0 {
		// No fields to update, just return current activity
		return r.GetByID(ctx, userID, id)
	}

	setParts = append(setParts, "updated_at = NOW()")

	stmt := fmt.Sprintf(`
	UPDATE activities 
	SET %s
	WHERE id = $1 AND user_id = $2
	RETURNING id, user_id, activity_name, duration, date, created_at, updated_at`,
		strings.Join(setParts, ", "))

	var activity CreateActivityOut
	err := r.pool.QueryRow(ctx, stmt, args...).Scan(
		&activity.ID,
		&activity.UserID,
		&activity.ActivityName,
		&activity.Duration,
		&activity.Date,
		&activity.CreatedAt,
		&activity.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &activity, nil
}

func (r *repo) Delete(ctx context.Context, userID uuid.UUID, id int) error {
	stmt := `DELETE FROM activities WHERE id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, stmt, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("activity not found")
	}

	return nil
}
