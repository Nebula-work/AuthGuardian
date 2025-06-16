package models

// UserResponse defines a standard user API response
type UserSwaggerResponse struct {
	Success bool         `json:"success"`
	Token   string       `json:"token,omitempty"`
	User    UserResponse `json:"user,omitempty"`
	Error   string       `json:"error,omitempty"`
	Message string       `json:"message,omitempty"`
}

type APIResponse struct {
	Success bool   `json:"success"`         // Indicates if the operation was successful
	Message string `json:"message"`         // A descriptive message about the operation
	Error   string `json:"error,omitempty"` // Error details (optional, only present if there's an error)
}
