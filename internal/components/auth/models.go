package auth

import "github.com/google/uuid"

type (
	User struct {
		ID           uuid.UUID `json:"id"`
		Username     string    `json:"username"`
		PasswordHash string    `json:"-"` // Never serialize password hash
	}

	LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	LoginResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
	}
)