package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ErrNotFound is returned when a user is not found.
var ErrNotFound = errors.New("user not found")

// Role represents a user role.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// User represents a registered user in the system.
type User struct {
	ID           uuid.UUID `bson:"_id" json:"id"`
	Username     string    `bson:"username" json:"username"`
	Email        string    `bson:"email" json:"email"`
	PasswordHash string    `bson:"password_hash" json:"-"`
	Role         Role      `bson:"role" json:"role"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}

// NewUser creates a new user with hashed password.
func NewUser(username, email, password string) (*User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	return &User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         RoleUser,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}, nil
}

// CheckPassword verifies the provided password against the hash.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Repository defines the interface for user persistence.
// Defined here (Consumer-side) usually, but common to define validation interface with Entity in go clean arch if it's core.
// Actually, instructions say "Interfaces: Define them where they are USED (Consumer-side pattern)".
// So I will define it in the Usecase or where needed.
// However, domain entity usually doesn't need Repo.
// But we need a place to export the interface for the implementation to implement?
// No, the implementation just implements the methods. The Interface lives in the Usecase package usually.
// Or if it's a generic Domain Service, it might live here.
// Current project structure seems to have `player/player.go` and `ranking/service.go`.
// Let's check `game/game.go` to see conventions used in this repo.
