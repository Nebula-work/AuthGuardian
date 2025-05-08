package models

// Organization represents an organization in the system
type Organization struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Domain      string   `json:"domain,omitempty"`
	Active      bool     `json:"active"`
	AdminIDs    []string `json:"adminIds"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}
