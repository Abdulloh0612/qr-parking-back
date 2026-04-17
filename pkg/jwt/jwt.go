package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserID       uuid.UUID `json:"user_id"`
	IsAdmin      bool      `json:"is_admin"`
	AdminSession bool      `json:"admin_session"` // true только при входе в админку логином/паролем
	AdminRole    string    `json:"admin_role"`    // "admin" | "super_admin" для admin_session; иначе пусто
	jwt.RegisteredClaims
}

type Manager struct {
	secret          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:          secret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// NewManagerFromVars creates a Manager from string duration values (e.g. "15m", "720h").
// Falls back to safe defaults if parsing fails.
func NewManagerFromVars(secret, accessTTLStr, refreshTTLStr string) *Manager {
	accessTTL, err := time.ParseDuration(accessTTLStr)
	if err != nil || accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	refreshTTL, err := time.ParseDuration(refreshTTLStr)
	if err != nil || refreshTTL <= 0 {
		refreshTTL = 720 * time.Hour
	}
	return NewManager(secret, accessTTL, refreshTTL)
}

// GenerateToken creates a single JWT valid for the given TTL.
func (m *Manager) GenerateToken(userID uuid.UUID, isAdmin bool, adminSession bool, ttl time.Duration) (string, error) {
	return m.generateToken(userID, isAdmin, adminSession, "", ttl)
}

// GenerateTokenPair creates an access + refresh token pair (used for admin sessions).
func (m *Manager) GenerateTokenPair(userID uuid.UUID, isAdmin bool, adminSession bool, adminRole string) (*TokenPair, error) {
	accessToken, err := m.generateToken(userID, isAdmin, adminSession, adminRole, m.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := m.generateToken(userID, isAdmin, adminSession, adminRole, m.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (m *Manager) generateToken(userID uuid.UUID, isAdmin bool, adminSession bool, adminRole string, ttl time.Duration) (string, error) {
	claims := &Claims{
		UserID:       userID,
		IsAdmin:      isAdmin,
		AdminSession: adminSession,
		AdminRole:    adminRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
