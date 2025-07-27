package activity

import "time"

type (
	CreateActivityIn struct {
		ActivityName string    `json:"activity_name"`
		Duration     int       `json:"duration"` // minutes
		Date         time.Time `json:"date"`
	}

	CreateActivityOut struct {
		ID           int       `json:"id"`
		UserName     string    `json:"user_name"`
		ActivityName string    `json:"activity_name"`
		Duration     int       `json:"duration"` // minutes
		Date         time.Time `json:"date"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}

	BulkCreateActivityIn struct {
		Activities []CreateActivityIn `json:"activities"`
	}

	BulkCreateActivityOut struct {
		Activities []CreateActivityOut `json:"activities"`
		Count      int                 `json:"count"`
	}

	GetActivitiesQuery struct {
		Page     int    `json:"page"`
		Limit    int    `json:"limit"`
		UserName string `json:"user_name,omitempty"`
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