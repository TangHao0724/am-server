package handler

import (
	"amserver/internal/models"
	"amserver/internal/repository"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Name string `json:"name"`
}
type Claims struct {
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

func Register(c *gin.Context, db *gorm.DB) {
	ctx := c.Request.Context()
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	name := req.Name
	fmt.Println(name)
	userRepos := repository.NewUserRepository(db)
	user := &models.User{Name: name}
	if err := userRepos.Create(ctx, user); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	token, err := GenerateJWT(user.ID, user.Name)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

var jwtSecret = []byte("aquamarina")

func GenerateJWT(userID uint, userName string) (string, error) {
	claims := Claims{
		UserID: userID,
		Name:   userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}
