package controllers

import (
	"CommentClassifier/internal/data"
	"CommentClassifier/internal/db"
	"CommentClassifier/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetAllCategories returns all categories in DB
func GetAllCategories(c *gin.Context) {
	categories, err := db.GetAllCategories(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to get categories: "+err.Error()))
		return
	}

	if categories == nil {
		categories = make([]data.Category, 0)
	}

	c.JSON(http.StatusOK,
		utils.NewSuccessResponse(http.StatusOK, categories))
	return
}
