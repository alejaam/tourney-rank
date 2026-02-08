package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/alejaam/tourney-rank/internal/domain/user"
)

const (
	// UsersCollection is the MongoDB collection name for users.
	UsersCollection = "users"
)

// userDocument represents the MongoDB document structure for a user.
type userDocument struct {
	ID           string    `bson:"_id"`
	Username     string    `bson:"username"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"password_hash"`
	Role         string    `bson:"role"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

func (d *userDocument) toDomain() *user.User {
	id, _ := uuid.Parse(d.ID)
	return &user.User{
		ID:           id,
		Username:     d.Username,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		Role:         user.Role(d.Role),
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func fromDomainUser(u *user.User) *userDocument {
	return &userDocument{
		ID:           u.ID.String(),
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// UserRepository implements user.Repository.
type UserRepository struct {
	coll *mongo.Collection
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(client *Client) *UserRepository {
	return &UserRepository{
		coll: client.Collection(UsersCollection),
	}
}

// EnsureIndexes creates necessary indexes for users.
func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := r.coll.Indexes().CreateMany(ctx, models)
	if err != nil {
		return fmt.Errorf("creating user indexes: %w", err)
	}
	return nil
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	doc := fromDomainUser(u)
	_, err := r.coll.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("user already exists: %w", err)
	}
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var doc userDocument
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by ID: %w", err)
	}
	return doc.toDomain(), nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var doc userDocument
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return doc.toDomain(), nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var doc userDocument
	err := r.coll.FindOne(ctx, bson.M{"username": username}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by username: %w", err)
	}
	return doc.toDomain(), nil
}

// GetAll retrieves all users.
func (r *UserRepository) GetAll(ctx context.Context) ([]*user.User, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("finding all users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*user.User
	for cursor.Next(ctx) {
		var doc userDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decoding user document: %w", err)
		}
		users = append(users, doc.toDomain())
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return users, nil
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	result, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	if result.DeletedCount == 0 {
		return user.ErrNotFound
	}
	return nil
}

// UpdateRole updates a user's role.
func (r *UserRepository) UpdateRole(ctx context.Context, id string, role user.Role) error {
	update := bson.M{
		"$set": bson.M{
			"role":       string(role),
			"updated_at": time.Now().UTC(),
		},
	}
	result, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("updating user role: %w", err)
	}
	if result.MatchedCount == 0 {
		return user.ErrNotFound
	}
	return nil
}
