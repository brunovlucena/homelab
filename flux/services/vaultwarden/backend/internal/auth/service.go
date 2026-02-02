package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	jwtSecret []byte
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewService(jwtSecret string) *Service {
	return &Service{
		jwtSecret: []byte(jwtSecret),
	}
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func (s *Service) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken validates a JWT token and returns the user ID
func (s *Service) ValidateToken(tokenString string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}

// GenerateUserID generates a unique user ID (in production, use UUID)
func (s *Service) GenerateUserID() string {
	return generateUUID()
}

// generateUUID generates a simple UUID (in production, use a proper UUID library)
func generateUUID() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		time.Now().Unix(),
		time.Now().UnixNano()%1000000,
		time.Now().UnixNano()%10000000,
		time.Now().UnixNano()%100000000,
		time.Now().UnixNano()%1000000000,
	)
}
