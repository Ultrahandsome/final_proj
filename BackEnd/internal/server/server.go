package server

import (
	"CommentClassifier/internal/controllers"
	"CommentClassifier/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

var (
	Server *gin.Engine
)

func InitServer() {
	Server = gin.Default()
	Server.Use(cors.New(cors.Config{
		// Allow all origins for front-end development
		// AllowOrigins:     []string{"*"},
		AllowOrigins:     []string{"http://localhost:8000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	registerRouters(Server)
}

func registerRouters(s *gin.Engine) {
	// For front-end usage
	apiGroup := s.Group("/api")
	apiGroup.GET("/dashboard", controllers.Dashboard)
	apiGroup.POST("/login", controllers.Login)
	apiGroup.POST("/comments", controllers.GetComments)
	apiGroup.GET("/categories", controllers.GetAllCategories)

	ig := apiGroup.Group("/upload")
	ig.POST("/csv", controllers.UploadCSV)
	ig.POST("/excel", controllers.UploadExcel)
	ig.POST("/tsv", controllers.UploadTSV)

	eg := apiGroup.Group("/export")
	eg.POST("/excel", controllers.ExportExcel)
	eg.POST("/csv", controllers.ExportCSV)
	eg.POST("/tsv", controllers.ExportTSV)

	ug := apiGroup.Group("/user")
	ug.POST("", controllers.GetUsers)
	ug.POST("/add", controllers.CreateUser)
	ug.POST("/delete", controllers.DeleteUser)

	// Comment moderation (requires login)
	cg := apiGroup.Group("/comment")
	cg.Use(middleware.JWTMiddleware())
	cg.POST("/category", controllers.OverwriteCategory)

	afg := apiGroup.Group("/info")
	afg.Use(middleware.JWTMiddleware())
	afg.GET("", controllers.GetUserFromToken)

	// ========== Old APIs ==========
	// Register upload controllers
	uploadGroup := s.Group("/upload")
	uploadGroup.POST("/excel", controllers.UploadExcel)
	uploadGroup.POST("/csv", controllers.UploadCSV)
	uploadGroup.POST("/tsv", controllers.UploadTSV)

	s.GET("/dashboard", controllers.Dashboard)

	// Register comment retrieval controllers
	s.POST("/comments", controllers.GetComments)

	// Comment moderation (requires login)
	commentGroup := s.Group("/comment")
	commentGroup.Use(middleware.JWTMiddleware())
	commentGroup.POST("/category", controllers.OverwriteCategory)

	// Register export controllers
	exportGroup := s.Group("/export")
	exportGroup.POST("/excel", controllers.ExportExcel)
	exportGroup.POST("/csv", controllers.ExportCSV)
	exportGroup.POST("/tsv", controllers.ExportTSV)

	// Register categories controllers
	s.GET("/categories", controllers.GetAllCategories)

	// Register user management controllers
	s.POST("/login", controllers.Login)
	userGroup := s.Group("/user")
	userGroup.POST("", controllers.GetUsers)
	userGroup.POST("/add", controllers.CreateUser)
	userGroup.POST("/delete", controllers.DeleteUser)

	fg := s.Group("/info")
	fg.Use(middleware.JWTMiddleware())
	fg.GET("", controllers.GetUserFromToken)
}
