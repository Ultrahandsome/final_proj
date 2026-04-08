package utils

import (
	"CommentClassifier/internal/data"
	"github.com/gin-gonic/gin"
)

// NewSuccessResponse returns a success response that contains response data
func NewSuccessResponse(code int, data interface{}) gin.H {
	return gin.H{
		"code": code,
		"msg":  "",
		"data": data,
	}
}

// NewErrorResponse returns an error response that contains an error message
func NewErrorResponse(code int, msg string) gin.H {
	return gin.H{
		"code": code,
		"msg":  msg,
		"data": nil,
	}
}

// UploadSuccessfulResponse is the response generated when successfully uploaded an Excel or CSV file
type UploadSuccessfulResponse struct {
	NumberOfComments int `json:"numberOfComments"`
}

// GetCommentsResponse represents the response structure for paginated comments.
type GetCommentsResponse struct {
	Page          int64          `json:"page"`
	Limit         int64          `json:"limit"`
	TotalPages    int64          `json:"totalPages"`
	TotalComments int64          `json:"totalComments"`
	Comments      []data.Comment `json:"comments"`
}

// LoginTokenResponse returns JWT token when successfully logged in
type LoginTokenResponse struct {
	Token string `json:"token"`
}

// GetUserResponse represents the response structure for paginated users.
type GetUserResponse struct {
	Page       int64       `json:"page"`
	Limit      int64       `json:"limit"`
	TotalPages int64       `json:"totalPages"`
	TotalUsers int64       `json:"totalUsers"`
	Users      []data.User `json:"users"`
}

type PieChartData struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type BarChartData struct {
	Category          string  `json:"category"`
	AverageConfidence float64 `json:"averageConfidence"`
}

type DashboardResponse struct {
	TotalComments int64          `json:"totalComments"`
	NeedsReview   int64          `json:"needsReview"`
	Pie           []PieChartData `json:"pie"`
	Bar           []BarChartData `json:"bar"`
}
