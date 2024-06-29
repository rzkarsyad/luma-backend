package handler

import (
	"context"
	"net/http"
	"os"
	"time"

	"luma-backend/model"
	"luma-backend/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleoauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

type OAuthHandler struct {
	MongoRepo *repository.MongoRepository
}

func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	oauth2Service, err := googleoauth.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OAuth2 service: " + err.Error()})
		return
	}

	userinfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	user := model.User{
		Email:   userinfo.Email,
		Name:    userinfo.Name,
		Picture: userinfo.Picture,
	}

	existingUser, err := h.MongoRepo.FindUserByEmail(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user: " + err.Error()})
		return
	}

	if existingUser == nil {
		err = h.MongoRepo.InsertUser(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user: " + err.Error()})
			return
		}
	}

	jwtToken, err := generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
	})
}

func generateJWT(user model.User) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"email":   user.Email,
		"name":    user.Name,
		"picture": user.Picture,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func (h *OAuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
