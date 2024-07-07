package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are currently not logged in. Please log in to access this feature."})
			c.Redirect(http.StatusFound, os.Getenv("FRONTEND_URL"))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Redirect(http.StatusFound, os.Getenv("FRONTEND_URL"))
			c.Abort()
			return
		}

		secretKey := os.Getenv("JWT_SECRET")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Redirect(http.StatusFound, os.Getenv("FRONTEND_URL"))
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("email", claims["email"])
			c.Set("name", claims["name"])
			c.Set("picture", claims["picture"])
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Redirect(http.StatusFound, os.Getenv("FRONTEND_URL"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func CheckLoginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				secretKey := os.Getenv("JWT_SECRET")
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, jwt.ErrSignatureInvalid
					}
					return []byte(secretKey), nil
				})
				if err == nil {
					if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
						c.JSON(http.StatusOK, gin.H{"message": "User is logged in", "status": "success"})
						c.Redirect(http.StatusFound, os.Getenv("FRONTEND_CHAT"))
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}
