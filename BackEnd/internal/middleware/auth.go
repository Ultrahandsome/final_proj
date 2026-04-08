package middleware

import (
	"CommentClassifier/internal/utils"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// JWTMiddleware is a middleware function that validates the JWT token from the "X-Token" header and sets claims in context.
// If the token is missing, invalid, or expired, it responds with an unauthorized error and aborts the request.
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("X-Token")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized,
				utils.NewErrorResponse(http.StatusUnauthorized, "Authorization header required"))
			c.Abort()
			return
		}

		claims, err := ValidateToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized,
				utils.NewErrorResponse(http.StatusUnauthorized, "Invalid or expired token"))
			c.Abort()
			return
		}

		// Store claims in the context for later use
		c.Set("claims", claims)
		c.Next()
	}
}

// RoleMiddleware creates a middleware to enforce role-based access control by checking user roles in JWT claims.
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized,
				utils.NewErrorResponse(http.StatusUnauthorized, "Authentication required"))
			c.Abort()
			return
		}

		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			log.Printf("Invalid token structure: %v\n", claims)
			c.JSON(http.StatusInternalServerError,
				utils.NewErrorResponse(http.StatusInternalServerError, "Invalid token structure"))
			c.Abort()
			return
		}

		userRole, ok := mapClaims["role"].(string)
		if !ok {
			log.Printf("Invalid token structure: %v\n", claims)
			c.JSON(http.StatusInternalServerError,
				utils.NewErrorResponse(http.StatusInternalServerError, "Invalid token structure"))
			c.Abort()
			return
		}

		// Check if user's role is in the allowed roles
		allowed := false
		for _, role := range roles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden,
				utils.NewErrorResponse(http.StatusForbidden, "Insufficient permissions"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateToken validates a JWT token string and returns its claims if valid or an error if invalid.
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return utils.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
