package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

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

// Common errors
var (
	ErrPasswordTooShort   = errors.New("password is too short (minimum 8 characters)")
	ErrPasswordTooWeak    = errors.New("password is too weak (must contain uppercase, lowercase, number, and special character)")
	ErrPasswordHashFailed = errors.New("failed to hash password")
)

// BcryptPasswordManager implements PasswordManager using bcrypt
type BcryptPasswordManager struct {
	cost int
}

// NewBcryptPasswordManager creates a new bcrypt password manager
func NewBcryptPasswordManager(cost int) PasswordManager {
	// Ensure cost is within valid range
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}

	return &BcryptPasswordManager{
		cost: cost,
	}
}

// HashPassword hashes a password using bcrypt
func (m *BcryptPasswordManager) HashPassword(password string) (string, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), m.cost)
	if err != nil {
		return "", ErrPasswordHashFailed
	}

	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against a hash
func (m *BcryptPasswordManager) VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateRandomPassword generates a random password
func (m *BcryptPasswordManager) GenerateRandomPassword(length int) string {
	// Ensure minimum length
	if length < 8 {
		length = 8
	}

	// Generate random bytes
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to static password if random fails
		return "Chang3M3!"
	}

	// Encode to Base64
	s := base64.StdEncoding.EncodeToString(b)

	// Truncate to requested length
	if len(s) > length {
		s = s[:length]
	}

	// Ensure password meets complexity requirements
	s = ensurePasswordComplexity(s)

	return s
}

// ValidatePasswordStrength validates password strength
func (m *BcryptPasswordManager) ValidatePasswordStrength(password string) error {
	// Check minimum length
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	// Check complexity
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ErrPasswordTooWeak
	}

	return nil
}

// Helper function to ensure password meets complexity requirements
func ensurePasswordComplexity(password string) string {
	result := password

	// Ensure there's at least one uppercase letter
	upperRegex := regexp.MustCompile(`[A-Z]`)
	if !upperRegex.MatchString(result) {
		result = result[:len(result)-1] + "A"
	}

	// Ensure there's at least one lowercase letter
	lowerRegex := regexp.MustCompile(`[a-z]`)
	if !lowerRegex.MatchString(result) {
		result = result[:len(result)-1] + "a"
	}

	// Ensure there's at least one number
	numberRegex := regexp.MustCompile(`[0-9]`)
	if !numberRegex.MatchString(result) {
		result = result[:len(result)-1] + "1"
	}

	// Ensure there's at least one special character
	specialRegex := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !specialRegex.MatchString(result) {
		result = result[:len(result)-1] + "!"
	}

	return result
}
