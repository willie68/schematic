package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/willie68/schematics2/backend/internal/domain/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreateUser adds a new user to the database
func (s *MongoStore) CreateUser(ctx context.Context, user model.User) error {
	if s.usersCol == nil {
		return errors.New("mongodb users collection not initialised")
	}

	// Only set Created/Updated if not already set by caller
	if user.Created == 0 {
		user.Created = time.Now().UTC().Unix()
	}
	if user.Updated == 0 {
		user.Updated = user.Created
	}

	_, err := s.usersCol.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user with this email already exists")
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (s *MongoStore) GetUserByEmail(ctx context.Context, email string) (*model.User, bool) {
	if s.usersCol == nil {
		return nil, false
	}

	var user model.User
	err := s.usersCol.FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, false
	}
	if err != nil {
		s.logger.Error("get user failed", "error", err, "email", email)
		return nil, false
	}

	return &user, true
}

// UpdateUser updates an existing user in the database
func (s *MongoStore) UpdateUser(ctx context.Context, user model.User) error {
	if s.usersCol == nil {
		return errors.New("mongodb users collection not initialised")
	}

	result, err := s.usersCol.ReplaceOne(ctx, bson.D{{Key: "_id", Value: user.ID}}, user)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// ListAllUsers retrieves all users from the database
func (s *MongoStore) ListAllUsers(ctx context.Context) ([]model.User, error) {
	if s.usersCol == nil {
		return nil, errors.New("mongodb users collection not initialised")
	}

	opts := options.Find().SetSort(bson.D{{Key: "email", Value: 1}})

	cursor, err := s.usersCol.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}

	return users, nil
}

// GetUserByID retrieves a user by ID
func (s *MongoStore) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	if s.usersCol == nil {
		return nil, errors.New("mongodb users collection not initialised")
	}

	var user model.User
	err := s.usersCol.FindOne(ctx, bson.D{{Key: "_id", Value: userID}}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}

	return &user, nil
}
