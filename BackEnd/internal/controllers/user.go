package controllers

import (
	"CommentClassifier/internal/db"
	"CommentClassifier/internal/utils"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

// GetUserFromContext extracts the user from the JWT claims in the context.
func GetUserFromContext(c *gin.Context) (string, error) {
	claims, exists := c.Get("claims")
	if !exists {
		return "", errors.New("no JWT claims found in context")
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid JWT claims format")
	}

	username, ok := mapClaims["username"].(string)
	if !ok {
		return "", errors.New("user ID not found in token")
	}

	return username, nil
}

func GetUserFromToken(c *gin.Context) {
	username, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized,
			utils.NewErrorResponse(http.StatusUnauthorized, err.Error()))
		return
	}

	u, err := db.FindUserByUsername(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK,
		utils.NewSuccessResponse(http.StatusOK, u))
	return
}

// Login handles user authentication by validating username and password, generates a JWT token, and responds with it.
func Login(c *gin.Context) {
	var req utils.UserLoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Find user by username
	user, err := db.FindUserByUsername(c, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized,
			utils.NewErrorResponse(http.StatusUnauthorized, err.Error()))
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized,
			utils.NewErrorResponse(http.StatusUnauthorized, "Invalid username or password"))
		return
	}

	// Generate JWT token
	expiryTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"id":       user.ID.Hex(),
		"username": user.Username,
		"role":     user.Role,
		"exp":      expiryTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(utils.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK,
		utils.NewSuccessResponse(http.StatusOK, utils.LoginTokenResponse{Token: tokenString}))
	return
}

// CreateUser handles the creation of a new user in the system by validating input and saving the user to the database.
func CreateUser(c *gin.Context) {
	var req utils.CreateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Role is "moderator" if not specified
	if req.Role == "" {
		req.Role = utils.RoleModerator
	}

	// Add user to MongoDB
	if err := db.RegisterUser(c, req.Username, req.Password, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	// No response body
	c.JSON(http.StatusCreated,
		utils.NewSuccessResponse(http.StatusCreated, struct{}{}))
	return
}

// DeleteUser handles a request to delete a user by their unique ID and returns a status response.
func DeleteUser(c *gin.Context) {
	var req utils.DeleteUserRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "Invalid request body"))
		return
	}

	id, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	if err := db.DeleteUserByID(c, id); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	// No response body
	c.JSON(http.StatusOK,
		utils.NewSuccessResponse(http.StatusOK, struct{}{}))
	return
}

// GetUsers handles the request for retrieving a paginated list of users based on the provided query parameters.
func GetUsers(c *gin.Context) {
	var req utils.GetUsersRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "Invalid request body"))
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	// Calculate total pages
	totalUsers, err := db.CountUsers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	totalPages := (totalUsers + req.Limit - 1) / req.Limit

	// Get all users
	skip := (req.Page - 1) * req.Limit
	users, err := db.GetAllUsers(c, skip, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK,
		utils.NewSuccessResponse(http.StatusOK, utils.GetUserResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalPages: totalPages,
			TotalUsers: totalUsers,
			Users:      users,
		}))
	return
}
