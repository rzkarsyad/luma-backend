package model

type User struct {
	Email   string `bson:"email"`
	Name    string `bson:"name"`
	Picture string `bson:"profile_picture"`
}
