package activity

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/andrasnagy-data/timelog/internal/shared/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/hlog"
)

type (
	Router struct {
		service servicer
	}
)

func NewRouter(service servicer) chi.Router {
	router := &Router{service: service}
	return router.Routes()
}

func (r *Router) Routes() chi.Router {
	router := chi.NewRouter()

	router.Get("/", r.GetActivities)
	router.Get("/filter", r.GetActivities)
	router.Post("/add", r.CreateActivity)
	router.Get("/{id}/edit", r.GetEditForm)
	router.Put("/{id}", r.UpdateActivity)
	router.Get("/{id}/cancel-edit", r.CancelEdit)
	router.Delete("/{id}", r.DeleteActivity)
	router.Get("/export", r.ExportCSV)

	return router
}

// HTMX template data structure
type ActivitiesTemplateData struct {
	Activities []CreateActivityOut
	Total      int
	Page       int
	PageSize   int
	TotalPages int
	NextPage   int
	PrevPage   int
	StartDate  string
	EndDate    string
}

// GetActivities returns activities as HTML table for HTMX
func (r *Router) GetActivities(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	userID := middleware.GetUserID(ctx)

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
	if l := req.URL.Query().Get("page_size"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil && limit > 0 && limit <= 100 {
			query.Limit = limit
		}
	}

	// Parse date filters
	if startDate := req.URL.Query().Get("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			query.StartDate = &date
		}
	}
	if endDate := req.URL.Query().Get("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			query.EndDate = &date
		}
	}

	resp, err := r.service.GetActivities(ctx, userID, query)
	if err != nil {
		logger.Error().Err(err).Msg("Error getting activities")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate pagination data
	totalPages := int(math.Ceil(float64(resp.Total) / float64(resp.Limit)))
	nextPage := resp.Page + 1
	if nextPage > totalPages {
		nextPage = totalPages
	}
	prevPage := resp.Page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	templateData := ActivitiesTemplateData{
		Activities: resp.Activities,
		Total:      resp.Total,
		Page:       resp.Page,
		PageSize:   resp.Limit,
		TotalPages: totalPages,
		NextPage:   nextPage,
		PrevPage:   prevPage,
		StartDate:  req.URL.Query().Get("start_date"),
		EndDate:    req.URL.Query().Get("end_date"),
	}

	tmpl, err := template.ParseFiles("templates/activities_table.html")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse activities table template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, templateData); err != nil {
		logger.Error().Err(err).Msg("Failed to execute activities table template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// CreateActivity creates activity and returns success message for HTMX
func (r *Router) CreateActivity(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	userID := middleware.GetUserID(ctx)

	// Parse form data
	if err := req.ParseForm(); err != nil {
		logger.Warn().Err(err).Msg("Failed to parse form")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Invalid form data</div>`)
		return
	}

	activityName := req.FormValue("activity_name")
	durationStr := req.FormValue("duration")
	dateStr := req.FormValue("date")

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Invalid duration</div>`)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Invalid date format</div>`)
		return
	}

	body := CreateActivityIn{
		ActivityName: activityName,
		Duration:     duration,
		Date:         date,
	}

	_, err = r.service.CreateActivity(ctx, userID, body)
	if err != nil {
		logger.Error().Err(err).Msg("Error creating activity")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Failed to create activity</div>`)
		return
	}

	// Return success message and trigger table refresh
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("HX-Trigger", "refreshTable")
	fmt.Fprint(w, `<div class="success">Activity added successfully!</div>`)
}

// GetEditForm returns edit form for a specific activity
func (r *Router) GetEditForm(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	userID := middleware.GetUserID(ctx)

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn().Str("id", idStr).Msg("Invalid activity ID")
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	activity, err := r.service.GetActivityByID(ctx, userID, id)
	if err != nil {
		logger.Error().Err(err).Int("id", id).Msg("Error getting activity by ID")
		http.Error(w, "Activity not found", http.StatusNotFound)
		return
	}

	editFormHTML := fmt.Sprintf(`
		<div class="actions" id="edit-form-%d">
			<form hx-put="/api/activities/%d" hx-target="#edit-form-%d" hx-swap="outerHTML" style="display: inline-flex; gap: 5px; align-items: center;">
				<input type="text" name="activity_name" value="%s" style="width: 120px; padding: 4px; border: 1px solid #ddd; border-radius: 3px;">
				<input type="number" name="duration" value="%d" style="width: 60px; padding: 4px; border: 1px solid #ddd; border-radius: 3px;">
				<input type="date" name="date" value="%s" style="padding: 4px; border: 1px solid #ddd; border-radius: 3px;">
				<button type="submit" class="btn btn-success btn-sm">Save</button>
				<button type="button" class="btn btn-secondary btn-sm" 
						hx-get="/api/activities/%d/cancel-edit" 
						hx-target="#edit-form-%d" 
						hx-swap="outerHTML">Cancel</button>
			</form>
		</div>`,
		activity.ID, activity.ID, activity.ID,
		activity.ActivityName, activity.Duration, activity.Date.Format("2006-01-02"),
		activity.ID, activity.ID)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, editFormHTML)
}

// DeleteActivity deletes activity and returns success message for HTMX
func (r *Router) DeleteActivity(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	userID := middleware.GetUserID(ctx)

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn().Str("id", idStr).Msg("Invalid activity ID")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Invalid activity ID</div>`)
		return
	}

	err = r.service.DeleteActivity(ctx, userID, id)
	if err != nil {
		logger.Error().Err(err).Int("id", id).Msg("Error deleting activity")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Failed to delete activity</div>`)
		return
	}

	// Return success message and trigger table refresh
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("HX-Trigger", "refreshTable")
	fmt.Fprint(w, `<div class="success">Activity deleted successfully!</div>`)
}

// ExportCSV exports activities as CSV file
func (r *Router) ExportCSV(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	userID := middleware.GetUserID(ctx)

	// Get all activities (no pagination for export)
	query := GetActivitiesQuery{
		Page:  1,
		Limit: 10000, // Large limit to get all activities
	}

	resp, err := r.service.GetActivities(ctx, userID, query)
	if err != nil {
		logger.Error().Err(err).Msg("Error getting activities for export")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=activities.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV header
	if err := writer.Write([]string{"Activity", "Duration (minutes)", "Date", "Created At"}); err != nil {
		logger.Error().Err(err).Msg("Error writing CSV header")
		return
	}

	// Write activity data
	for _, activity := range resp.Activities {
		record := []string{
			activity.ActivityName,
			strconv.Itoa(activity.Duration),
			activity.Date.Format("2006-01-02"),
			activity.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			logger.Error().Err(err).Msg("Error writing CSV record")
			return
		}
	}
}

// UpdateActivity updates activity and returns updated row for HTMX
func (r *Router) UpdateActivity(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	userID := middleware.GetUserID(ctx)

	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn().Str("id", idStr).Msg("Invalid activity ID")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Invalid activity ID</div>`)
		return
	}

	// Parse form data
	if err := req.ParseForm(); err != nil {
		logger.Warn().Err(err).Msg("Failed to parse form")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Invalid form data</div>`)
		return
	}

	body := UpdateActivityIn{}

	if activityName := req.FormValue("activity_name"); activityName != "" {
		body.ActivityName = &activityName
	}

	if durationStr := req.FormValue("duration"); durationStr != "" {
		if duration, err := strconv.Atoi(durationStr); err == nil {
			body.Duration = &duration
		}
	}

	if dateStr := req.FormValue("date"); dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			body.Date = &date
		}
	}

	updatedActivity, err := r.service.UpdateActivity(ctx, userID, id, body)
	if err != nil {
		logger.Error().Err(err).Int("id", id).Msg("Error updating activity")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="error">Failed to update activity</div>`)
		return
	}

	// Return updated action buttons
	actionsHTML := fmt.Sprintf(`
		<div class="actions" id="edit-form-%d">
			<button class="btn btn-warning btn-sm" 
					hx-get="/api/activities/%d/edit"
					hx-target="#edit-form-%d"
					hx-swap="outerHTML">
				Edit
			</button>
			<button class="btn btn-danger btn-sm" 
					hx-delete="/api/activities/%d"
					hx-confirm="Are you sure you want to delete this activity?"
					hx-target="#messages"
					hx-swap="innerHTML">
				Delete
			</button>
		</div>`,
		updatedActivity.ID, updatedActivity.ID, updatedActivity.ID, updatedActivity.ID)

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("HX-Trigger", "refreshTable")
	fmt.Fprint(w, actionsHTML)
}

// CancelEdit returns normal action buttons for HTMX
func (r *Router) CancelEdit(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	// Return normal action buttons
	actionsHTML := fmt.Sprintf(`
		<div class="actions" id="edit-form-%d">
			<button class="btn btn-warning btn-sm" 
					hx-get="/api/activities/%d/edit"
					hx-target="#edit-form-%d"
					hx-swap="outerHTML">
				Edit
			</button>
			<button class="btn btn-danger btn-sm" 
					hx-delete="/api/activities/%d"
					hx-confirm="Are you sure you want to delete this activity?"
					hx-target="#messages"
					hx-swap="innerHTML">
				Delete
			</button>
		</div>`,
		id, id, id, id)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, actionsHTML)
}
