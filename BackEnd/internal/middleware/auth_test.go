package middleware

import (
	"CommentClassifier/internal/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	// Set a test JWT secret for testing purposes
	utils.JWTSecret = []byte("test-secret-key")
}

// generateTestToken generates a valid JWT token for testing
func generateTestToken(role string, expiry time.Time) (string, error) {
	claims := jwt.MapClaims{
		"role":    role,
		"user_id": "test-user",
		"exp":     expiry.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(utils.JWTSecret)
}

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		checkContext   bool
	}{
		{
			name:           "Missing token",
			setupRequest:   func(req *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "Invalid token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-Token", "invalid-token")
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "Expired token",
			setupRequest: func(req *http.Request) {
				token, _ := generateTestToken("admin", time.Now().Add(-time.Hour))
				req.Header.Set("X-Token", token)
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "Valid token",
			setupRequest: func(req *http.Request) {
				token, _ := generateTestToken("admin", time.Now().Add(time.Hour))
				req.Header.Set("X-Token", token)
			},
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(JWTMiddleware())

			var contextClaims jwt.MapClaims

			router.GET("/test", func(c *gin.Context) {
				if tt.checkContext {
					claims, exists := c.Get("claims")
					assert.True(t, exists)
					contextClaims = claims.(jwt.MapClaims)
				}
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tt.setupRequest(req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkContext {
				assert.NotNil(t, contextClaims)
				assert.Equal(t, "admin", contextClaims["role"])
				assert.Equal(t, "test-user", contextClaims["user_id"])
			}
		})
	}
}

func TestRoleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		roles          []string
		setupContext   func(*gin.Context)
		expectedStatus int
	}{
		{
			name:  "No claims in context",
			roles: []string{"admin"},
			setupContext: func(c *gin.Context) {
				// Don't set any claims
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Invalid claims type",
			roles: []string{"admin"},
			setupContext: func(c *gin.Context) {
				c.Set("claims", "not-a-map-claims")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "Missing role in claims",
			roles: []string{"admin"},
			setupContext: func(c *gin.Context) {
				c.Set("claims", jwt.MapClaims{
					"user_id": "test-user",
					// No role field
				})
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "Insufficient permissions",
			roles: []string{"admin"},
			setupContext: func(c *gin.Context) {
				c.Set("claims", jwt.MapClaims{
					"user_id": "test-user",
					"role":    "user",
				})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:  "Sufficient permissions",
			roles: []string{"admin", "superuser"},
			setupContext: func(c *gin.Context) {
				c.Set("claims", jwt.MapClaims{
					"user_id": "test-user",
					"role":    "admin",
				})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.GET("/test", func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			}, RoleMiddleware(tt.roles...), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		tokenGen    func() string
		expectError bool
		checkClaims func(claims jwt.MapClaims)
	}{
		{
			name: "Valid token",
			tokenGen: func() string {
				token, _ := generateTestToken("admin", time.Now().Add(time.Hour))
				return token
			},
			expectError: false,
			checkClaims: func(claims jwt.MapClaims) {
				assert.Equal(t, "admin", claims["role"])
				assert.Equal(t, "test-user", claims["user_id"])
			},
		},
		{
			name: "Expired token",
			tokenGen: func() string {
				token, _ := generateTestToken("admin", time.Now().Add(-time.Hour))
				return token
			},
			expectError: true,
			checkClaims: nil,
		},
		{
			name: "Invalid token format",
			tokenGen: func() string {
				return "invalid-token-format"
			},
			expectError: true,
			checkClaims: nil,
		},
		{
			name: "Token with wrong signing method",
			tokenGen: func() string {
				claims := jwt.MapClaims{
					"role":    "admin",
					"user_id": "test-user",
					"exp":     time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
				tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
				return tokenString
			},
			expectError: true,
			checkClaims: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.tokenGen()
			claims, err := ValidateToken(tokenString)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				tt.checkClaims(claims)
			}
		})
	}
}
