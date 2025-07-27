package activity

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/hlog"
)

type (
	ActivityRouter struct {
		service activitySrvc
	}
)

func NewActivityRouter(service activitySrvc) chi.Router {
	router := &ActivityRouter{service: service}
	return router.Routes()
}

func (r *ActivityRouter) Routes() chi.Router {
	router := chi.NewRouter()

	router.Post("/", r.CreateActivity)
	router.Post("/bulk", r.BulkCreateActivities)
	router.Get("/", r.GetActivities)
	router.Get("/{id}", r.GetActivityByID)
	router.Put("/{id}", r.UpdateActivity)
	router.Delete("/{id}", r.DeleteActivity)

	return router
}

// CreateActivity godoc
// @Summary Create a new activity
// @Description Create a new activity for the authenticated user
// @Tags activities
// @Accept json
// @Produce json
// @Param activity body CreateActivityIn true "Activity to create"
// @Success 201 {object} CreateActivityOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /activities [post]
func (r *ActivityRouter) CreateActivity(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	// TODO: Extract user_name from JWT context
	userName := "mock_user"

	var body CreateActivityIn
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body for activity creation")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := r.service.CreateActivity(ctx, userName, body)
	if err != nil {
		logger.Error().Err(err).Msg("Error creating activity")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error().Err(err).Msg("Failed to encode create activity response")
	}
}

// BulkCreateActivities godoc
// @Summary Bulk create activities
// @Description Create multiple activities for the authenticated user
// @Tags activities
// @Accept json
// @Produce json
// @Param activities body BulkCreateActivityIn true "Activities to create"
// @Success 201 {object} BulkCreateActivityOut
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /activities/bulk [post]
func (r *ActivityRouter) BulkCreateActivities(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	// TODO: Extract user_name from JWT context
	userName := "mock_user"

	var body BulkCreateActivityIn
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body for bulk activity creation")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := r.service.BulkCreateActivities(ctx, userName, body)
	if err != nil {
		logger.Error().Err(err).Msg("Error bulk creating activities")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error().Err(err).Msg("Failed to encode bulk create response")
	}
}

// GetActivities godoc
// @Summary Get all activities
// @Description Get all activities for the authenticated user with pagination
// @Tags activities
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} GetActivitiesResponse
// @Failure 500 {object} map[string]string
// @Router /activities [get]
func (r *ActivityRouter) GetActivities(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	// TODO: Extract user_name from JWT context
	userName := "mock_user"

	query := GetActivitiesQuery{
		Page:  1,
		Limit: 20,
	}

	// Parse query parameters
	if p := req.URL.Query().Get("page"); p != "" {
		if page, err := strconv.Atoi(p); err == nil && page > 0 {
			query.Page = page
		}
	}
	if l := req.URL.Query().Get("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil && limit > 0 && limit <= 100 {
			query.Limit = limit
		}
	}

	resp, err := r.service.GetActivities(ctx, userName, query)
	if err != nil {
		logger.Error().Err(err).Msg("Error getting activities")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error().Err(err).Msg("Failed to encode get activities response")
	}
}

// GetActivityByID godoc
// @Summary Get activity by ID
// @Description Get a specific activity by ID for the authenticated user
// @Tags activities
// @Produce json
// @Param id path int true "Activity ID"
// @Success 200 {object} CreateActivityOut
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /activities/{id} [get]
func (r *ActivityRouter) GetActivityByID(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	// TODO: Extract user_name from JWT context
	userName := "mock_user"

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn().Str("id", idStr).Msg("Invalid activity ID")
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	resp, err := r.service.GetActivityByID(ctx, userName, id)
	if err != nil {
		logger.Error().Err(err).Int("id", id).Msg("Error getting activity by ID")
		http.Error(w, "Activity not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error().Err(err).Msg("Failed to encode get activity response")
	}
}

// UpdateActivity godoc
// @Summary Update activity
// @Description Update an activity for the authenticated user (partial updates supported)
// @Tags activities
// @Accept json
// @Produce json
// @Param id path int true "Activity ID"
// @Param activity body UpdateActivityIn true "Activity fields to update"
// @Success 200 {object} CreateActivityOut
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /activities/{id} [put]
func (r *ActivityRouter) UpdateActivity(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	// TODO: Extract user_name from JWT context
	userName := "mock_user"

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn().Str("id", idStr).Msg("Invalid activity ID")
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	var body UpdateActivityIn
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body for activity update")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := r.service.UpdateActivity(ctx, userName, id, body)
	if err != nil {
		logger.Error().Err(err).Int("id", id).Msg("Error updating activity")
		http.Error(w, "Activity not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error().Err(err).Msg("Failed to encode update activity response")
	}
}

// DeleteActivity godoc
// @Summary Delete activity
// @Description Delete an activity for the authenticated user
// @Tags activities
// @Param id path int true "Activity ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /activities/{id} [delete]
func (r *ActivityRouter) DeleteActivity(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	// TODO: Extract user_name from JWT context
	userName := "mock_user"

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn().Str("id", idStr).Msg("Invalid activity ID")
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	err = r.service.DeleteActivity(ctx, userName, id)
	if err != nil {
		logger.Error().Err(err).Int("id", id).Msg("Error deleting activity")
		http.Error(w, "Activity not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}