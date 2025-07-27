package activity

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	activitySrvc struct {
		pool *pgxpool.Pool
	}
)

func NewActivitySrvc(pool *pgxpool.Pool) activitySrvc {
	return activitySrvc{pool: pool}
}

func (s *activitySrvc) CreateActivity(ctx context.Context, userName string, req CreateActivityIn) (*CreateActivityOut, error) {
	activityOut := new(CreateActivityOut)

	stmt := `
	INSERT INTO activities (
		user_name, activity_name, duration, date)
	VALUES (
		$1, $2, $3, $4
	)
	RETURNING *`

	err := s.pool.QueryRow(
		ctx,
		stmt,
		userName,
		req.ActivityName,
		req.Duration,
		req.Date,
	).Scan(
		&activityOut.ID,
		&activityOut.UserName,
		&activityOut.ActivityName,
		&activityOut.Duration,
		&activityOut.Date,
		&activityOut.CreatedAt,
		&activityOut.UpdatedAt,
	)
	if err != nil {
		return activityOut, err
	}
	return activityOut, nil
}

func (s *activitySrvc) BulkCreateActivities(ctx context.Context, userName string, req BulkCreateActivityIn) (*BulkCreateActivityOut, error) {
	result := &BulkCreateActivityOut{
		Activities: make([]CreateActivityOut, 0, len(req.Activities)),
	}
	// TODO actually do it in bulk
	for _, activity := range req.Activities {
		created, err := s.CreateActivity(ctx, userName, activity)
		if err != nil {
			return result, err
		}
		result.Activities = append(result.Activities, *created)
	}

	result.Count = len(result.Activities)
	return result, nil
}

func (s *activitySrvc) GetActivities(ctx context.Context, userName string, query GetActivitiesQuery) (*GetActivitiesResponse, error) {
	offset := (query.Page - 1) * query.Limit

	// Get total count
	countStmt := `SELECT COUNT(*) FROM activities WHERE user_name = $1`
	var total int
	err := s.pool.QueryRow(ctx, countStmt, userName).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Get activities
	stmt := `
	SELECT id, user_name, activity_name, duration, date, created_at, updated_at
	FROM activities 
	WHERE user_name = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3`

	rows, err := s.pool.Query(ctx, stmt, userName, query.Limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []CreateActivityOut
	for rows.Next() {
		var activity CreateActivityOut
		err := rows.Scan(
			&activity.ID,
			&activity.UserName,
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

func (s *activitySrvc) GetActivityByID(ctx context.Context, userName string, id int) (*CreateActivityOut, error) {
	stmt := `
	SELECT id, user_name, activity_name, duration, date, created_at, updated_at
	FROM activities 
	WHERE id = $1 AND user_name = $2`

	var activity CreateActivityOut
	err := s.pool.QueryRow(ctx, stmt, id, userName).Scan(
		&activity.ID,
		&activity.UserName,
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

func (s *activitySrvc) UpdateActivity(ctx context.Context, userName string, id int, req UpdateActivityIn) (*CreateActivityOut, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{id, userName}
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
		return s.GetActivityByID(ctx, userName, id)
	}

	setParts = append(setParts, "updated_at = NOW()")

	stmt := fmt.Sprintf(`
	UPDATE activities 
	SET %s
	WHERE id = $1 AND user_name = $2
	RETURNING id, user_name, activity_name, duration, date, created_at, updated_at`,
		strings.Join(setParts, ", "))

	var activity CreateActivityOut
	err := s.pool.QueryRow(ctx, stmt, args...).Scan(
		&activity.ID,
		&activity.UserName,
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

func (s *activitySrvc) DeleteActivity(ctx context.Context, userName string, id int) error {
	stmt := `DELETE FROM activities WHERE id = $1 AND user_name = $2`

	result, err := s.pool.Exec(ctx, stmt, id, userName)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("activity not found")
	}

	return nil
}