package models

// TokenType defines the type of token
type TokenType string

const (
	// TokenTypeRefresh is a refresh token
	TokenTypeRefresh TokenType = "refresh"
	// TokenTypeReset is a password reset token
	TokenTypeReset TokenType = "reset"
	// TokenTypeVerification is an email verification token
	TokenTypeVerification TokenType = "verification"
	// TokenTypeRevoked is a revoked token (for JWT blacklist)
	TokenTypeRevoked TokenType = "revoked"
)

type TokenClaims struct {
	UserID    string   `json:"userId"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	RoleIDs   []string `json:"roleIds"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
}

// Token represents a stored token
type Token struct {
	ID         string    `bson:"_id,omitempty" json:"id"`
	UserID     string    `bson:"userId" json:"userId"`
	TokenType  TokenType `bson:"tokenType" json:"tokenType"`
	TokenValue string    `bson:"tokenValue" json:"tokenValue"`
	CreatedAt  string    `bson:"createdAt" json:"createdAt"`
	ExpiresAt  string    `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
}
