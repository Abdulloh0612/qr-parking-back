package services

import (
	"context"
	"fmt"

	"qr-parking/db/repositories"
	jwtpkg "qr-parking/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	adminRepo repositories.AdminRepository
	jwt       *jwtpkg.Manager
}

func NewAuthService(adminRepo repositories.AdminRepository, jwt *jwtpkg.Manager) *AuthService {
	return &AuthService{adminRepo: adminRepo, jwt: jwt}
}

// AdminLogin authenticates an admin by username/password and returns a JWT pair with admin_session=true.
func (s *AuthService) AdminLogin(ctx context.Context, username, password string) (*jwtpkg.TokenPair, string, error) {
	admin, err := s.adminRepo.GetByUsername(ctx, username)
	if err != nil || admin == nil {
		return nil, "", fmt.Errorf("invalid username or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("invalid username or password")
	}
	role := admin.Role
	if role == "" {
		role = "admin"
	}
	pair, err := s.jwt.GenerateTokenPair(admin.DisplayID, true, true, role)
	if err != nil {
		return nil, "", err
	}
	return pair, role, nil
}
