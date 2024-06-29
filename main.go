package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"luma-backend/handler"
	"luma-backend/repository"
	"luma-backend/service"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	file, err := os.Open("data-series.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	table, err := repository.CsvToSlice(string(data))
	if err != nil {
		fmt.Println("Error parsing CSV:", err)
		return
	}

	mongoRepo, err := repository.NewMongoRepository(os.Getenv("MONGODB_URI"), os.Getenv("MONGO_DB"))
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		return
	}

	aiModelConnector := &repository.AIModelConnector{Client: &http.Client{}}
	aiService := &service.AIService{Connector: aiModelConnector}
	aiHandler := &handler.AIHandler{Service: aiService, Table: table}
	oauthHandler := &handler.OAuthHandler{MongoRepo: mongoRepo}

	router := gin.Default()

	router.GET("/auth/google/login", oauthHandler.GoogleLogin)
	router.GET("/auth/google/callback", oauthHandler.GoogleCallback)
	router.POST("/api/ask", aiHandler.HandleRequest)

	router.Run(":8080")
}
