package activity

import (
	"time"

	"github.com/google/uuid"
)

type (
	CreateActivityIn struct {
		ActivityName string    `json:"activity_name"`
		Duration     int       `json:"duration"` // minutes
		Date         time.Time `json:"date"`
	}

	CreateActivityOut struct {
		ID           int       `json:"id"`
		UserID       uuid.UUID `json:"user_id"`
		ActivityName string    `json:"activity_name"`
		Duration     int       `json:"duration"` // minutes
		Date         time.Time `json:"date"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}


	GetActivitiesQuery struct {
		Page      int        `json:"page"`
		Limit     int        `json:"limit"`
		StartDate *time.Time `json:"start_date,omitempty"`
		EndDate   *time.Time `json:"end_date,omitempty"`
	}

	GetActivitiesResponse struct {
		Activities []CreateActivityOut `json:"activities"`
		Total      int                 `json:"total"`
		Page       int                 `json:"page"`
		Limit      int                 `json:"limit"`
	}

	UpdateActivityIn struct {
		ActivityName *string    `json:"activity_name,omitempty"`
		Duration     *int       `json:"duration,omitempty"`
		Date         *time.Time `json:"date,omitempty"`
	}
)
