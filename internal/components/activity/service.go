package activity

import (
	"context"

	"github.com/google/uuid"
)

type (
	servicer interface {
		CreateActivity(context.Context, uuid.UUID, CreateActivityIn) (*CreateActivityOut, error)
		GetActivities(context.Context, uuid.UUID, GetActivitiesQuery) (*GetActivitiesResponse, error)
		GetActivityByID(context.Context, uuid.UUID, int) (*CreateActivityOut, error)
		UpdateActivity(context.Context, uuid.UUID, int, UpdateActivityIn) (*CreateActivityOut, error)
		DeleteActivity(context.Context, uuid.UUID, int) error
	}
	service struct {
		repo repoer
	}
)

func NewService(repo repoer) servicer {
	return &service{repo: repo}
}

func (s *service) CreateActivity(ctx context.Context, userID uuid.UUID, req CreateActivityIn) (*CreateActivityOut, error) {
	return s.repo.Create(ctx, userID, req)
}

func (s *service) GetActivities(ctx context.Context, userID uuid.UUID, query GetActivitiesQuery) (*GetActivitiesResponse, error) {
	return s.repo.List(ctx, userID, query)
}

func (s *service) GetActivityByID(ctx context.Context, userID uuid.UUID, id int) (*CreateActivityOut, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *service) UpdateActivity(ctx context.Context, userID uuid.UUID, id int, req UpdateActivityIn) (*CreateActivityOut, error) {
	return s.repo.Update(ctx, userID, id, req)
}

func (s *service) DeleteActivity(ctx context.Context, userID uuid.UUID, id int) error {
	return s.repo.Delete(ctx, userID, id)
}
