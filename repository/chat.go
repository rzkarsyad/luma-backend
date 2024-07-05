package repository

import (
	"context"
	"time"

	"luma-backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatRepository struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewChatRepository(client *mongo.Client, dbName string) *ChatRepository {
	db := client.Database(dbName)
	return &ChatRepository{
		Client: client,
		DB:     db,
	}
}

func (r *ChatRepository) SaveMessage(sessionID string, message model.Message) error {
	collection := r.DB.Collection("chat_history")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	update := bson.M{"$push": bson.M{"messages": message}}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *ChatRepository) GetMessages(sessionID string) ([]model.Message, error) {
	collection := r.DB.Collection("chat_history")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"session_id": sessionID}
	var result struct {
		Messages []model.Message `bson:"messages"`
	}

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return result.Messages, nil
}
