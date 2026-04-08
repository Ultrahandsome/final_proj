package db

import (
	"CommentClassifier/internal/data"
	"CommentClassifier/internal/rpcapi"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	DB                    *mongo.Client
	CommentCollection     *mongo.Collection
	FileVersionCollection *mongo.Collection
	CategoryCollection    *mongo.Collection
	UserCollection        *mongo.Collection
)

// InitMongo creates a MongoDB client and save it in the global variable db
func InitMongo(ctx context.Context) error {
	// Load environment variables from init.env file
	err := LoadEnvFromFile()
	if err != nil {
		log.Printf("Error loading init.env: %v", err)
		// Continue even if there's an error loading the file
	}

	// Get MongoDB address for containerisation
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	log.Println("Connecting to MongoDB on: " + mongoURI)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		return err
	}

	// Define the collections: comments collection and categories collection
	DB = client
	CommentCollection = client.Database("comments").Collection("comments")
	CategoryCollection = client.Database("comments").Collection("categories")
	FileVersionCollection = client.Database("comments").Collection("fileversion")
	UserCollection = client.Database("comments").Collection("users")

	log.Println("Connected to MongoDB on: " + mongoURI)

	// Initialise default admin user if needed
	err = InitDefaultAdmin(ctx)
	if err != nil {
		log.Printf("Error initializing admin user: %v", err)
		// Continue even if there's an error creating the admin user
	}

	return nil
}

// LoadEnvFromFile loads environment variables from init.env file
func LoadEnvFromFile() error {
	// Check if init.env file exists
	if _, err := os.Stat("init.env"); os.IsNotExist(err) {
		log.Println("init.env file not found, default admin not created")
		return nil
	}

	// Read the file
	content, err := os.ReadFile("init.env")
	if err != nil {
		return err
	}

	// Parse each line
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		// Skip empty lines or comments
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by '=' and set environment variable
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes if present
			value = strings.Trim(value, "\"'")
			os.Setenv(key, value)
		}
	}

	return nil
}

// InsertComments inserts comments into collection "comments"
// It returns numbers of comments inserted into MongoDB
// As MongoDB requires data in the []interface{} format, we use []interface{} here.
func InsertComments(ctx context.Context, comments []interface{}) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	result, err := CommentCollection.InsertMany(ctx, comments)
	if err != nil {
		return 0, err
	}
	return len(result.InsertedIDs), nil
}

// CountComments filters comments by categories, then count matching documents
func CountComments(ctx context.Context, categories []string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Build the MongoDB filter
	filter := bson.M{}
	if len(categories) > 0 {
		filter["category"] = bson.M{"$in": categories}
	}

	// Count the total matching documents
	total, err := CommentCollection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindComments find matching comments filtered by categories in paging biases
func FindComments(ctx context.Context, skip, limit int64, categories []string) ([]data.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Category filter
	filter := bson.M{}
	if len(categories) > 0 {
		filter["category"] = bson.M{"$in": categories}
	}
	// Paging filter
	opts := options.Find().
		SetSkip(skip).SetLimit(limit).
		SetSort(bson.D{{Key: "lastUpdated", Value: -1}})

	// Find comments with filters
	cursor, err := CommentCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}(cursor, ctx)

	// Collect matching comments
	var comments []data.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// FindCommentsByIDs queries MongoDB for matching comments by given ObjectID(s)
func FindCommentsByIDs(ctx context.Context, ids []primitive.ObjectID) ([]data.Comment, error) {
	// Filter comments by their IDs
	// No filter if ids array is nil
	filter := bson.M{}
	if len(ids) > 0 {
		filter = bson.M{"_id": bson.M{"$in": ids}}
	}

	cursor, err := CommentCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}(cursor, ctx)

	var comments []data.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// FindLatestComments returns all comments with latest UUID
func FindLatestComments(ctx context.Context) ([]data.Comment, error) {
	v, err := GetLatestFileVersion(ctx)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"uuid": v}
	cursor, err := CommentCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}(cursor, ctx)

	var comments []data.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// UpdateClassifiedComment updates a comment by its ObjectID
// It sets ConfidenceScore, Category, and SimilarComments
func UpdateClassifiedComment(ctx context.Context, classified *rpcapi.ClassifiedComment) error {
	// Convert classified.Id to primitive.ObjectID
	objID, err := primitive.ObjectIDFromHex(classified.Id)
	if err != nil {
		log.Printf("Invalid ObjectID %s: %v", classified.Id, err)
		return err
	}

	// Build the update criteria
	now := time.Now().Unix()
	update := bson.M{
		"$set": bson.M{
			"category":        classified.Label,
			"similarComments": classified.SimilarComment,
			"confidenceScore": float64(classified.Score),
			"keywords":        classified.Keywords,
			"updateHistory": []data.UpdateHistory{{
				User:     "AI",
				Category: classified.Label,
				Comment:  "",
				Time:     now,
			}},
			"createAt":    now,
			"lastUpdated": now,
		},
	}

	// Filter document to find the comment by its ObjectID
	filter := bson.M{"_id": objID}

	// Execute the update
	_, err = CommentCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Failed to update comment with ID %s: %v", classified.Id, err)
		return err
	}

	return nil
}

// GetLatestFileVersion returns the most recent file version in MongoDB
func GetLatestFileVersion(ctx context.Context) (string, error) {
	var version data.FileVersion
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	err := FileVersionCollection.FindOne(ctx, bson.M{}, opts).Decode(&version)
	if err != nil {
		return "", err
	}
	return version.UUID, nil
}

// SetFileVersion adds the most recent file version in MongoDB
func SetFileVersion(ctx context.Context, u string) error {
	fileVersion := data.FileVersion{
		UUID:      u,
		CreatedAt: time.Now(),
	}
	_, err := FileVersionCollection.InsertOne(ctx, fileVersion)
	if err != nil {
		return err
	}
	return nil
}

// AddCategory adds a category to DB if it's not present
func AddCategory(ctx context.Context, category string) error {
	filter := bson.M{"name": category}
	count, err := CategoryCollection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}
	// Insert a new category if it's not present
	if count == 0 {
		_, err = CategoryCollection.InsertOne(ctx, data.Category{Name: category})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAllCategories returns all categories in DB
func GetAllCategories(ctx context.Context) ([]data.Category, error) {
	cursor, err := CategoryCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}(cursor, ctx)

	var categories []data.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

// UpdateCommentCategory updates a comment's category and update modification history
func UpdateCommentCategory(ctx context.Context, id primitive.ObjectID, category, user, comment string) error {
	now := time.Now().Unix()
	newHistory := data.UpdateHistory{
		User:     user,
		Category: category,
		Comment:  comment,
		Time:     now,
	}
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set":  bson.M{"category": category, "lastUpdated": now},
		"$push": bson.M{"updateHistory": newHistory},
	}

	_, err := CommentCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}
