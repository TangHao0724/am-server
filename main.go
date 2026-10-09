package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"amserver/internal/database"
	"amserver/internal/model"
	"amserver/internal/repository"
)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	router := gin.Default()
	// However, this one will match /user/john/ and also /user/john/send
	// If no other routers match /user/john, it will redirect to /user/john/
	router.GET("/user/:id/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		message := name + " is " + action
		c.String(http.StatusOK, message)
	})

	router.POST("/user", func(c *gin.Context) {
		ctx := c.Request.Context()
		var req CreateUserRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		name := req.Name
		email := req.Email
		fmt.Println(name)
		fmt.Println(email)
		message := name + " and " + email
		db, err := database.Connect()
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		user := model.User{Name: name, Email: &email}
		userRepos := repository.NewUserRepository(db)
		userRepos.Create(ctx, &user)
		c.String(http.StatusOK, message)
	})
	router.GET("/user/:id", func(c *gin.Context) {
		ctx := c.Request.Context()
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if db, err := database.Connect(); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		} else {
			userRepos := repository.NewUserRepository(db)
			userRepos.GetByID(ctx, id)
		}
	})
	if err := router.Run(":8080"); err != nil {
		panic(fmt.Sprintf("failed to start server: %v", err))
	}
}
