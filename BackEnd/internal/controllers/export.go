package controllers

import (
	"CommentClassifier/internal/db"
	"CommentClassifier/internal/utils"
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
)

// ExportExcel exports selected comments to an Excel xlsx file
// It uses an []primitive.ObjectId as filter. If the filter is nil, it exports all data in DB.
func ExportExcel(c *gin.Context) {
	var req utils.ExportCommentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "Invalid request payload"))
		return
	}

	// Convert ObjectID string to primitive.ObjectID
	var objIDs []primitive.ObjectID
	for _, id := range req.IDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				utils.NewErrorResponse(http.StatusBadRequest, "Invalid ObjectID: "+id))
			return
		}
		objIDs = append(objIDs, objID)
	}

	// Query MongoDB for matching comments
	comments, err := db.FindCommentsByIDs(c, objIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to get comments: "+err.Error()))
		return
	}

	// Create a new Excel file
	f := excelize.NewFile()
	sheet := "Sheet1"
	err = f.SetSheetName("Sheet1", sheet)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to create sheet: "+err.Error()))
		return
	}

	// Add header row
	for i, header := range utils.Headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				utils.NewErrorResponse(http.StatusInternalServerError, "Failed to create header cell: "+err.Error()))
			return
		}
		err = f.SetCellValue(sheet, cell, header)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				utils.NewErrorResponse(http.StatusInternalServerError, "Failed to set header cell value: "+err.Error()))
			return
		}
	}

	// Add data rows
	for rowIdx, comment := range comments {
		// Add AI generated category
		aiCategory := comment.Category
		if len(comment.UpdateHistory) > 1 {
			aiCategory = comment.UpdateHistory[0].Category
		}
		humanCategory := ""
		if len(comment.UpdateHistory) > 1 {
			humanCategory = comment.UpdateHistory[len(comment.UpdateHistory)-1].Category
		}
		// Data starts from row 2
		rowNumber := rowIdx + 2
		values := []string{
			comment.IdentificationNo,
			comment.ModeOfAttendance,
			comment.TypeOfAttendance,
			comment.NESBIndicator,
			comment.Citizenship,
			comment.StudyArea,
			comment.RawComment,
			comment.TrainCategory,
			aiCategory, // ClassifiedCategory
			fmt.Sprintf("%.2f", comment.ConfidenceScore),
			humanCategory,    // HumanCategory
			comment.Category, // FinalClassification
		}
		for colIdx, value := range values {
			cell, err := excelize.CoordinatesToCellName(colIdx+1, rowNumber)
			if err != nil {
				c.JSON(http.StatusInternalServerError,
					utils.NewErrorResponse(http.StatusInternalServerError, "Failed to create data cell: "+err.Error()))
				return
			}
			err = f.SetCellValue(sheet, cell, value)
			if err != nil {
				c.JSON(http.StatusInternalServerError,
					utils.NewErrorResponse(http.StatusInternalServerError, "Failed to set data cell value: "+err.Error()))
				return
			}
		}
	}

	// Write Excel file to an in-memory buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to generate response file: "+err.Error()))
		return
	}

	// Set response headers and return the generated file
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="comments.xlsx"`)
	c.Header("Content-Length", strconv.Itoa(len(buf.Bytes())))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
	return
}

// ExportCSV exports selected comments to a CSV file
// It uses an []primitive.ObjectId as filter. If the filter is nil, it exports all data in DB.
func ExportCSV(c *gin.Context) {
	var req utils.ExportCommentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "Invalid request payload"))
		return
	}

	// Convert ObjectID string to primitive.ObjectID
	var objIDs []primitive.ObjectID
	for _, id := range req.IDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				utils.NewErrorResponse(http.StatusBadRequest, "Invalid ObjectID: "+id))
			return
		}
		objIDs = append(objIDs, objID)
	}

	// Query MongoDB for matching comments
	comments, err := db.FindCommentsByIDs(c, objIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to get comments: "+err.Error()))
		return
	}

	// Create CSV file in memory
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header row
	if err := writer.Write(utils.Headers[:]); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to write csv header: "+err.Error()))
		return
	}

	// Write data rows
	for _, comment := range comments {
		// Add AI generated category
		aiCategory := comment.Category
		if len(comment.UpdateHistory) > 1 {
			aiCategory = comment.UpdateHistory[0].Category
		}
		humanCategory := ""
		if len(comment.UpdateHistory) > 1 {
			humanCategory = comment.UpdateHistory[len(comment.UpdateHistory)-1].Category
		}
		row := []string{
			comment.IdentificationNo,
			comment.ModeOfAttendance,
			comment.TypeOfAttendance,
			comment.NESBIndicator,
			comment.Citizenship,
			comment.StudyArea,
			comment.RawComment,
			comment.TrainCategory,
			aiCategory, // ClassifiedCategory
			fmt.Sprintf("%.2f", comment.ConfidenceScore),
			humanCategory,    // HumanCategory
			comment.Category, // FinalClassification
		}
		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError,
				utils.NewErrorResponse(http.StatusInternalServerError, "Failed to write csv data row: "+err.Error()))
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to flush csv data: "+err.Error()))
		return
	}

	// Set response headers and return the CSV file.
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="comments.csv"`)
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}

// ExportTSV generates a TSV file of comments based on provided IDs and sends it as a downloadable response.
func ExportTSV(c *gin.Context) {
	var req utils.ExportCommentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "Invalid request payload"))
		return
	}

	// Convert ObjectID string to primitive.ObjectID
	var objIDs []primitive.ObjectID
	for _, id := range req.IDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				utils.NewErrorResponse(http.StatusBadRequest, "Invalid ObjectID: "+id))
			return
		}
		objIDs = append(objIDs, objID)
	}

	// Query MongoDB for matching comments
	comments, err := db.FindCommentsByIDs(c, objIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to get comments: "+err.Error()))
		return
	}

	// Create CSV file in memory
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = '\t'

	// Write header row
	if err := writer.Write(utils.Headers[:]); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to write TSV header: "+err.Error()))
		return
	}

	// Write data rows
	for _, comment := range comments {
		// Add AI generated category
		aiCategory := comment.Category
		if len(comment.UpdateHistory) > 1 {
			aiCategory = comment.UpdateHistory[0].Category
		}
		humanCategory := ""
		if len(comment.UpdateHistory) > 1 {
			humanCategory = comment.UpdateHistory[len(comment.UpdateHistory)-1].Category
		}
		row := []string{
			comment.IdentificationNo,
			comment.ModeOfAttendance,
			comment.TypeOfAttendance,
			comment.NESBIndicator,
			comment.Citizenship,
			comment.StudyArea,
			comment.RawComment,
			comment.TrainCategory,
			aiCategory, // ClassifiedCategory
			fmt.Sprintf("%.2f", comment.ConfidenceScore),
			humanCategory,    // HumanCategory
			comment.Category, // FinalClassification
		}
		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError,
				utils.NewErrorResponse(http.StatusInternalServerError, "Failed to write TSV data row: "+err.Error()))
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to flush TSV data: "+err.Error()))
		return
	}

	// Set response headers and return the TSV file.
	c.Header("Content-Type", "text/tab-separated-values")
	c.Header("Content-Disposition", `attachment; filename="comments.tsv"`)
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	c.Data(http.StatusOK, "text/tab-separated-values", buf.Bytes())
}
