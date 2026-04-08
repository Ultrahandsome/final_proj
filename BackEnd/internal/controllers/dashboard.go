package controllers

import (
	"CommentClassifier/internal/db"
	"CommentClassifier/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Dashboard handles the retrieval of dashboard metrics and sends a JSON response with the aggregated data or an error.
func Dashboard(c *gin.Context) {
	resp, err := db.GetDashboardData(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to get dashboard data: "+err.Error()))
		return
	}

	// Prevent frontend crash
	if resp.Pie == nil {
		resp.Pie = make([]utils.PieChartData, 0)
	}
	if resp.Bar == nil {
		resp.Bar = make([]utils.BarChartData, 0)
	}

	c.JSON(http.StatusOK,
		utils.NewSuccessResponse(http.StatusOK, resp))
	return
}
