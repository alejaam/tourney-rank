package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alejaam/tourney-rank/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidCredentials is returned when login fails.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Service provides authentication operations.
type Service struct {
	userRepo  user.Repository
	jwtSecret string
	tokenTTL  time.Duration
}

// NewService creates a new authentication service.
func NewService(userRepo user.Repository, jwtSecret string, tokenTTL time.Duration) *Service {
	return &Service{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
	}
}

// RegisterRequest represents the data needed to register a user.
type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

// LoginRequest represents the data needed to login.
type LoginRequest struct {
	Email    string
	Password string
}

// AuthResponse contains the token and user info.
type AuthResponse struct {
	Token string     `json:"token"`
	User  *user.User `json:"user"`
}

// Register creates a new user and returns a token.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if user exists
	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, errors.New("email already registered")
	}

	_, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil {
		return nil, errors.New("username already taken")
	}

	// Create user
	u, err := user.NewUser(req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.generateToken(u)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  u,
	}, nil
}

// Login verifies credentials and returns a token.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !u.CheckPassword(req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(u)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  u,
	}, nil
}

func (s *Service) generateToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  u.ID.String(),
		"role": u.Role,
		"exp":  time.Now().Add(s.tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signed, nil
}
