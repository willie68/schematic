package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/willie68/schematics2/backend/internal/domain/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const sharesCollection = "shares"

type mongoShare struct {
	ID         bson.ObjectID `bson:"_id"`
	CreatedAt  time.Time `bson:"createdAt"`
	ValidTo    time.Time `bson:"validTo"`
	Owner      string    `bson:"owner"`
	DocumentID string    `bson:"documentId"`
}

func (s *MongoStore) GetShare(ctx context.Context, link string) (*model.Share, error) {
	if link == "" {
		return nil, errors.New("link is empty")
	}
	objectID, err := bson.ObjectIDFromHex(link)
	if err != nil {
		return nil, fmt.Errorf("parse share id: %w", err)
	}
	var share mongoShare
	err = s.sharesCol.FindOne(ctx, bson.M{"_id": objectID}).Decode(&share)
	if err != nil {
		return nil, fmt.Errorf("get share: %w", err)
	}
	shareID := share.ID.Hex()
	return &model.Share{
		ID:         shareID,
		Link:       shareID,
		CreatedAt:  share.CreatedAt,
		ValidTo:    share.ValidTo,
		Owner:      share.Owner,
		DocumentID: share.DocumentID,
	}, nil
}

func (s *MongoStore) CreateShare(ctx context.Context, share *model.Share) (string, error) {
	if share == nil {
		return "", errors.New("share is nil")
	}

	res, err := s.sharesCol.InsertOne(ctx, bson.M{
		"createdAt":  share.CreatedAt,
		"validTo":    share.ValidTo,
		"owner":      share.Owner,
		"documentId": share.DocumentID,
	})
	if err != nil {
		return "", fmt.Errorf("create share: %w", err)
	}
	if insertedID, ok := res.InsertedID.(bson.ObjectID); ok {
		return insertedID.Hex(), nil
	}
	return "", errors.New("failed to retrieve inserted ID as object id")
}
