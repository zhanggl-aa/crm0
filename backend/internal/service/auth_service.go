package service

import (
	"context"
	"errors"
	"time"

	"crm0/backend/internal/config"
	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	cfg        *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, tenantRepo: tenantRepo, cfg: cfg}
}

func (s *AuthService) Login(email, password string) (*model.AuthResponse, error) {
	ctx := context.Background()
	user, err := s.userRepo.GetByEmailGlobal(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, ExpiresAt: expiresAt, User: *user}, nil
}

func (s *AuthService) Register(tenantName, email, password, name string) (*model.AuthResponse, error) {
	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	tenant := &model.Tenant{ID: uuid.New(), Name: tenantName, Plan: "free", Settings: []byte(`{}`), CreatedAt: now, UpdatedAt: now}
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}
	user := &model.User{ID: uuid.New(), TenantID: tenant.ID, Email: email, PasswordHash: string(hash), Name: name, Role: "admin", CreatedAt: now, UpdatedAt: now}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, ExpiresAt: expiresAt, User: *user}, nil
}

func (s *AuthService) RefreshToken(userID uuid.UUID) (*model.AuthResponse, error) {
	ctx := context.Background()
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, ExpiresAt: expiresAt, User: *user}, nil
}

func (s *AuthService) generateToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.JWT.ExpiryHours) * time.Hour)
	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"tenant_id": user.TenantID.String(),
		"email":     user.Email,
		"role":      user.Role,
		"exp":       expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, expiresAt, nil
}
