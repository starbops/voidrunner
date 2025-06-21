package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret      []byte
	expiryTime  time.Duration
	revokedMu   sync.RWMutex
	revokedJTIs map[string]bool
}

func NewTokenManager(secret string, expiryTime time.Duration) *TokenManager {
	return &TokenManager{
		secret:      []byte(secret),
		expiryTime:  expiryTime,
		revokedJTIs: make(map[string]bool),
	}
}

func (tm *TokenManager) GenerateToken(userID int, username, email string) (string, error) {
	now := time.Now()
	jti := fmt.Sprintf("%d_%d", userID, now.UnixNano())

	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.expiryTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "voidrunner",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	tm.revokedMu.RLock()
	isRevoked := tm.revokedJTIs[claims.ID]
	tm.revokedMu.RUnlock()

	if isRevoked {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (tm *TokenManager) RevokeToken(tokenString string) error {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secret, nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token for revocation: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return fmt.Errorf("invalid token claims")
	}

	tm.revokedMu.Lock()
	tm.revokedJTIs[claims.ID] = true
	tm.revokedMu.Unlock()

	return nil
}