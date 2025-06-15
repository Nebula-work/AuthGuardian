package models

type Role struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	OrganizationID  string   `json:"organizationId,omitempty"`
	PermissionIDs   []string `json:"permissionIds"`
	IsSystemDefault bool     `json:"isSystemDefault"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}
