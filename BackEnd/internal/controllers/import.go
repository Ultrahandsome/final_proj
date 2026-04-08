package controllers

import (
	"CommentClassifier/internal/data"
	"CommentClassifier/internal/db"
	"CommentClassifier/internal/rpcapi"
	"CommentClassifier/internal/utils"
	"context"
	"encoding/csv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

// UploadExcel handles the Excel file upload
func UploadExcel(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "No file received"))
		return
	}

	// Open the multipart file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Unable to open the uploaded file: "+err.Error()))
		return
	}
	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
			log.Println("Failed to close multipart file: " + err.Error())
		}
	}(f)

	// Parse the Excel file
	excel, err := excelize.OpenReader(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to parse the Excel file: "+err.Error()))
		return
	}
	defer func(excel *excelize.File) {
		err := excel.Close()
		if err != nil {
			log.Println("Failed to close excel file: " + err.Error())
		}
	}(excel)

	// Get the first sheet
	sheet := excel.GetSheetName(0)
	if sheet == "" {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "No sheet found in the Excel file"))
		return
	}

	// Get rows from the first sheet
	rows, err := excel.GetRows(sheet)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "Failed to read rows from the Excel file: "+err.Error()))
		return
	}
	if len(rows) < 1 {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "No rows found in the Excel file"))
		return
	}

	// Generate a new UUID for this batch of comments
	u := uuid.New().String()

	// We assume that the first row is header, subsequent rows contain data
	// As MongoDB requires data in the []interface{} type, we use []interface{} here
	var comments []interface{}

	for i, row := range rows {
		// Skip table header
		if i == 0 {
			continue
		}

		var comment data.Comment

		// Map row values to Comment struct fields based on order:
		// 0: Identification Number
		// 1: Mode of attendance code
		// 2: Type of attendance code
		// 3: NESB indicator
		// 4: Citizenship indicator
		// 5: Study area
		// 6: Course level
		// 7: RawComment
		// 8: TrainCategory
		if len(row) > 8 {
			comment.IdentificationNo = row[0]
			comment.ModeOfAttendance = row[1]
			comment.TypeOfAttendance = row[2]
			comment.NESBIndicator = row[3]
			comment.Citizenship = row[4]
			comment.StudyArea = row[5]
			comment.CourseLevel = row[6]
			comment.RawComment = row[7]
			comment.TrainCategory = row[8]
			comment.Category = utils.DefaultCategory
			comment.UUID = u

			// Ignore empty comments
			if strings.Trim(comment.RawComment, " ") == "" {
				continue
			}
		}
		comments = append(comments, comment)
	}

	if len(comments) == 0 {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "No comments found in the Excel file"))
		return
	}

	numberOfComments, err := db.InsertComments(c, comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to insert comments: "+err.Error()))
		return
	}

	// Update latest file version with this UUID
	if err := db.SetFileVersion(c, u); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to save new version: "+err.Error()))
		return
	}

	// Start a Goroutine to cooperate with AI module
	ClassifyComments()

	c.JSON(http.StatusCreated,
		utils.NewSuccessResponse(http.StatusCreated,
			utils.UploadSuccessfulResponse{NumberOfComments: numberOfComments}))
	return
}

// UploadCSV handles CSV file upload
// This is quite similar to UploadExcel
func UploadCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "No file received"))
		return
	}

	// Open the multipart file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Unable to open the uploaded file: "+err.Error()))
		return
	}
	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
			log.Println("Failed to close multipart file: " + err.Error())
		}
	}(f)

	// Parse and read the CSV file
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to read CSV file: "+err.Error()))
		return
	}

	if len(records) < 1 {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "CSV file contains no data"))
	}

	// Generate a new UUID for this batch of comments
	u := uuid.New().String()

	var comments []interface{}
	for i, row := range records {
		// Skip table header
		if i == 0 {
			continue
		}

		var comment data.Comment
		if len(row) > 8 {
			comment.IdentificationNo = row[0]
			comment.ModeOfAttendance = row[1]
			comment.TypeOfAttendance = row[2]
			comment.NESBIndicator = row[3]
			comment.Citizenship = row[4]
			comment.StudyArea = row[5]
			comment.CourseLevel = row[6]
			comment.RawComment = row[7]
			comment.TrainCategory = row[8]
			comment.Category = utils.DefaultCategory
			comment.UUID = u

			// Ignore empty comments
			if strings.Trim(comment.RawComment, " ") == "" {
				continue
			}
		}
		comments = append(comments, comment)
	}

	if len(comments) == 0 {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "No comments found in the CSV file"))
	}

	numberOfComments, err := db.InsertComments(c, comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to insert comments: "+err.Error()))
		return
	}

	// Update latest file version with this UUID
	if err := db.SetFileVersion(c, u); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to save new version: "+err.Error()))
		return
	}

	// Start a Goroutine to cooperate with AI module
	ClassifyComments()

	c.JSON(http.StatusCreated,
		utils.NewSuccessResponse(http.StatusCreated,
			utils.UploadSuccessfulResponse{NumberOfComments: numberOfComments}))
	return
}

// UploadTSV handles the upload of a TSV file, processes its content, and stores parsed comments in a database.
func UploadTSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "No file received"))
		return
	}

	// Open the multipart file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Unable to open the uploaded file: "+err.Error()))
		return
	}
	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
			log.Println("Failed to close multipart file: " + err.Error())
		}
	}(f)

	// Parse and read the TSV file
	reader := csv.NewReader(f)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to read TSV file: "+err.Error()))
		return
	}

	if len(records) < 1 {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "CSV file contains no data"))
	}

	// Generate a new UUID for this batch of comments
	u := uuid.New().String()

	var comments []interface{}
	for i, row := range records {
		// Skip table header
		if i == 0 {
			continue
		}

		var comment data.Comment
		if len(row) > 8 {
			comment.IdentificationNo = row[0]
			comment.ModeOfAttendance = row[1]
			comment.TypeOfAttendance = row[2]
			comment.NESBIndicator = row[3]
			comment.Citizenship = row[4]
			comment.StudyArea = row[5]
			comment.CourseLevel = row[6]
			comment.RawComment = row[7]
			comment.TrainCategory = row[8]
			comment.Category = utils.DefaultCategory
			comment.UUID = u

			// Ignore empty comments
			if strings.Trim(comment.RawComment, " ") == "" {
				continue
			}
		}
		comments = append(comments, comment)
	}

	if len(comments) == 0 {
		c.JSON(http.StatusBadRequest,
			utils.NewErrorResponse(http.StatusBadRequest, "No comments found in the TSV file"))
	}

	numberOfComments, err := db.InsertComments(c, comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to insert comments: "+err.Error()))
		return
	}

	// Update latest file version with this UUID
	if err := db.SetFileVersion(c, u); err != nil {
		c.JSON(http.StatusInternalServerError,
			utils.NewErrorResponse(http.StatusInternalServerError, "Failed to save new version: "+err.Error()))
		return
	}

	// Start a Goroutine to cooperate with AI module
	ClassifyComments()

	c.JSON(http.StatusCreated,
		utils.NewSuccessResponse(http.StatusCreated,
			utils.UploadSuccessfulResponse{NumberOfComments: numberOfComments}))
	return
}

func ClassifyComments() {
	go func() {
		// Find latest comments
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comments, err := db.FindLatestComments(ctx)
		if err != nil {
			log.Println("Failed to find latest comments: ", err.Error())
			return
		}

		if len(comments) == 0 {
			// Unlikely to happen
			log.Println("No comments found in MongoDB")
			return
		}
		log.Printf("Start processing batch: %s\n", comments[0].UUID)

		// Convert comments to gRPC messages
		var rawComments []*rpcapi.RawComment
		for _, comment := range comments {
			rawComments = append(rawComments, &rpcapi.RawComment{
				Id:         comment.ID.Hex(),
				RawComment: comment.RawComment,
			})
		}

		// Build gRPC request
		request := &rpcapi.RawComments{Comments: rawComments}

		// Send to server via server.GrpcClientConnection
		client := rpcapi.NewCommentClassifierClient(rpcapi.GrpcClientConnection)
		stream, err := client.ClassifyComments(context.Background(), request)
		if err != nil {
			log.Printf("Failed to classify comments: %s", err.Error())
			return
		}

		// Process the streamed responses
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				// End of stream
				log.Printf("Finished processing batch: %s\n", comments[0].UUID)
				return
			}
			if err != nil {
				log.Printf("Error when receivig streamed response: %s", err.Error())
				return
			}

			for _, classified := range resp.Comments {
				err := db.UpdateClassifiedComment(ctx, classified)
				if err != nil {
					log.Printf("Failed to update classified comment: %s", err.Error())
				}
				err = db.AddCategory(ctx, classified.Label)
				if err != nil {
					log.Printf("Failed to update category: %s", err.Error())
				}
			}
		}
	}()
}
