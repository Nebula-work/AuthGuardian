package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CustomClaims represents the claims in the JWT token
type CustomClaims struct {
	UserID          primitive.ObjectID   `json:"userId"`
	Username        string               `json:"username"`
	Email           string               `json:"email"`
	RoleIDs         []primitive.ObjectID `json:"roleIds"`
	OrganizationIDs []primitive.ObjectID `json:"organizationIds"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token for a user
func GenerateToken(
	userID primitive.ObjectID,
	username, email string,
	roleIDs, orgIDs []primitive.ObjectID,
	secret string,
	expirationTime time.Duration,
) (string, error) {
	// Create the claims
	claims := CustomClaims{
		UserID:          userID,
		Username:        username,
		Email:           email,
		RoleIDs:         roleIDs,
		OrganizationIDs: orgIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "rbac-system",
			Subject:   userID.Hex(),
		},
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, secret string) (*CustomClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate the token and return the claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func GenerateRefreshToken(
	userID primitive.ObjectID,
	secret string,
	expirationTime time.Duration,
) (string, error) {
	// Create the claims for the refresh token
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "rbac-system-refresh",
		Subject:   userID.Hex(),
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
func ValidateRefreshToken(tokenString, secret string) (*jwt.RegisteredClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate the token and return the claims
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}
