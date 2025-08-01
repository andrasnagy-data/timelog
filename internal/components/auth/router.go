package auth

import (
	"html/template"
	"net/http"

	"github.com/andrasnagy-data/timelog/internal/shared/cookie"
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
	router.Get("/", r.LoginPage)
	router.Post("/", r.HandleLogInFlow)
	return router
}

func (r *Router) HandleLogInFlow(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := hlog.FromRequest(req)

	username := req.FormValue("username")
	password := req.FormValue("password")

	logger.Debug().Str("username", username).Msg("Login attempt")

	user, err := r.service.ValidateCredentials(ctx, username, password)
	if err != nil {
		logger.Warn().Err(err).Str("username", username).Msg("Login failed: invalid credentials")

		// Return HTMX error fragment
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<div id="error" class="error">Invalid username or password</div>`))
		return
	}

	err = cookie.SetCookie(w, user.ID, r.service.GetSecretKey())
	if err != nil {
		logger.Error().Err(err).Str("username", username).Msg("Login failed: could not set cookie")

		// Return HTMX error fragment
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div id="error" class="error">Login failed. Please try again.</div>`))
		return
	}

	logger.Debug().Str("username", username).Str("user_id", user.ID.String()).Msg("Login successful")

	// Redirect to main page
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (r *Router) LoginPage(w http.ResponseWriter, req *http.Request) {
	logger := hlog.FromRequest(req)

	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse login template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute login template")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
