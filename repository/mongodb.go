package repository

import (
	"context"
	"time"

	"luma-backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongoRepository(uri, dbName string) (*MongoRepository, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	return &MongoRepository{Client: client, DB: db}, nil
}

func (r *MongoRepository) InsertUser(user model.User) error {
	collection := r.DB.Collection("users")
	_, err := collection.InsertOne(context.Background(), bson.M{
		"email":           user.Email,
		"name":            user.Name,
		"profile_picture": user.Picture,
	})
	return err
}

func (r *MongoRepository) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	filter := bson.D{{Key: "email", Value: email}}
	collection := r.DB.Collection("users")
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
