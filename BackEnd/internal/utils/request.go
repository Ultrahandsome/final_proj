package utils

// GetCommentsRequest represents the query parameters for retrieving comments with pagination and optional category filtering.
// Page specifies the page number to fetch.
// Limit defines the maximum number of users to retrieve per page.
// Categories specifies categories used to filter comments
type GetCommentsRequest struct {
	Page       int64    `json:"page"`
	Limit      int64    `json:"limit"`
	Categories []string `json:"categories"`
}

// ExportCommentRequest requests export matching comments to CSV/xlsx file
// If the IDs filter is nil, it exports all comments in DB
type ExportCommentRequest struct {
	IDs []string `json:"ids"`
}

// OverwriteCategoryRequest add a modification history record
type OverwriteCategoryRequest struct {
	ID       string `json:"id" binding:"required"`
	Category string `json:"category" binding:"required"`
	Comment  string `json:"comment"` // Optional
}

// UserLoginRequest login using an existing user
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateUserRequest represents a request payload to create a new user with mandatory username and password fields.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

// DeleteUserRequest represents a request to delete a user identified by their unique ID.
type DeleteUserRequest struct {
	ID string `json:"id" binding:"required"`
}

// GetUsersRequest represents the request parameters for retrieving a paginated list of users.
// Page specifies the page number to fetch.
// Limit defines the maximum number of users to retrieve per page.
type GetUsersRequest struct {
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
}
