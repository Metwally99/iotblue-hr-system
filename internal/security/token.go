package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type Claims struct {
	Role model.Role `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret, issuer string, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must contain at least 32 characters")
	}
	if ttl <= 0 {
		return nil, errors.New("JWT TTL must be positive")
	}
	return &TokenManager{secret: []byte(secret), issuer: issuer, ttl: ttl, now: time.Now}, nil
}

func (m *TokenManager) Generate(employee model.Employee) (string, time.Time, error) {
	now := m.now().UTC()
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		Role: employee.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   employee.ID,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *TokenManager) Verify(raw string) (Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid || claims.Subject == "" || !claims.Role.Valid() {
		return Claims{}, errors.New("invalid access token")
	}
	return claims, nil
}
