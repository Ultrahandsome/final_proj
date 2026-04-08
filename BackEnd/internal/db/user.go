package db

import (
	"CommentClassifier/internal/data"
	"CommentClassifier/internal/utils"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

// RegisterUser add a new user to MongoDB
func RegisterUser(ctx context.Context, username, password, role string) error {
	// Username should be unique
	user := data.User{}
	err := UserCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == nil {
		return errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert user into MongoDB
	user.Role = role
	user.Username = username
	user.Password = string(hashedPassword)

	if _, err := UserCollection.InsertOne(ctx, user); err != nil {
		return err
	}
	return nil
}

// FindUserByUsername find a user by its username
func FindUserByUsername(ctx context.Context, username string) (*data.User, error) {
	user := data.User{}
	err := UserCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	return &user, nil
}

// DeleteUserByID delete a user from MongoDB by its ObjectID
func DeleteUserByID(ctx context.Context, ID primitive.ObjectID) error {
	_, err := UserCollection.DeleteOne(ctx, bson.M{"_id": ID})
	return err
}

// CountUsers counts the total number of user documents in the UserCollection in the database.
func CountUsers(ctx context.Context) (int64, error) {
	return UserCollection.CountDocuments(ctx, bson.M{})
}

// GetAllUsers retrieves a paginated list of users from the database based on the specified skip and limit values.
// Returns a slice of data.User and an error if any issues occur during the query.
func GetAllUsers(ctx context.Context, skip, limit int64) ([]data.User, error) {
	opt := options.Find().SetLimit(limit).SetSkip(skip)
	cursor, err := UserCollection.Find(ctx, bson.M{}, opt)
	if err != nil {
		return make([]data.User, 0), err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}(cursor, ctx)

	var users []data.User
	if err := cursor.All(ctx, &users); err != nil {
		return make([]data.User, 0), err
	}
	return users, nil
}

// CheckAdminUserExists checks if an admin user already exists in the database
func CheckAdminUserExists(ctx context.Context) (bool, error) {
	filter := bson.M{"username": "admin", "role": "admin"}

	count, err := UserCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// InitDefaultAdmin checks for admin user and creates one if needed
func InitDefaultAdmin(ctx context.Context) error {
	// Check if an admin user already exists
	exists, err := CheckAdminUserExists(ctx)
	if err != nil {
		return err
	}

	// If admin already exists, do nothing
	if exists {
		log.Println("Admin user already exists")
		return nil
	}

	// Get admin password from environment variable
	adminPassword := os.Getenv("ADMIN_PASSWD")
	if adminPassword == "" {
		log.Println("No ADMIN_PASSWORD environment variable found, admin user not created")
		return nil
	}

	// Create the admin user
	err = RegisterUser(ctx, "admin", adminPassword, utils.RoleAdmin)
	return err
}
