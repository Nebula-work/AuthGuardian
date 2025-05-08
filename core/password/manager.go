package password

// PasswordManager defines the interface for password operations
type PasswordManager interface {
	// HashPassword hashes a password
	HashPassword(password string) (string, error)

	// VerifyPassword verifies a password against a hash
	VerifyPassword(password, hash string) error

	// GenerateRandomPassword generates a random password
	GenerateRandomPassword(length int) string

	// ValidatePasswordStrength validates password strength
	ValidatePasswordStrength(password string) error
}
