package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	middleware "github.com/kittipat1413/go-common/framework/middleware/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_Default(t *testing.T) {
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the RequestID middleware with default options.
	router.Use(middleware.RequestID())

	// Handler to test the request ID.
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := middleware.GetRequestIDFromContext(c.Request.Context())
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No Request ID"})
			return
		}
		c.String(http.StatusOK, requestID)
	})

	// Perform a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert the response.
	require.Equal(t, http.StatusOK, w.Code)
	requestID := w.Body.String()
	require.NotEmpty(t, requestID)

	// Check that the request ID is set in the response header.
	assert.Equal(t, requestID, w.Header().Get(middleware.DefaultRequestIDHeader))
}

func TestRequestID_WithExistingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestID())

	router.GET("/test", func(c *gin.Context) {
		requestID, _ := middleware.GetRequestIDFromContext(c.Request.Context())
		c.String(http.StatusOK, requestID)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	existingID := "existing-request-id-123"
	req.Header.Set(middleware.DefaultRequestIDHeader, existingID)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, existingID, w.Body.String())
	assert.Equal(t, existingID, w.Header().Get(middleware.DefaultRequestIDHeader))
}

func TestRequestID_WithCustomHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	customHeader := "X-Custom-Request-ID"
	router.Use(middleware.RequestID(
		middleware.WithRequestIDHeader(customHeader),
	))

	router.GET("/test", func(c *gin.Context) {
		requestID, _ := middleware.GetRequestIDFromContext(c.Request.Context())
		c.String(http.StatusOK, requestID)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	requestID := w.Body.String()
	require.NotEmpty(t, requestID)

	assert.Equal(t, requestID, w.Header().Get(customHeader))
}

func TestRequestID_WithCustomGenerator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	customGenerator := func() string {
		return "custom-generated-id"
	}
	router.Use(middleware.RequestID(
		middleware.WithRequestIDGenerator(customGenerator),
	))

	router.GET("/test", func(c *gin.Context) {
		requestID, _ := middleware.GetRequestIDFromContext(c.Request.Context())
		c.String(http.StatusOK, requestID)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	requestID := w.Body.String()
	assert.Equal(t, "custom-generated-id", requestID)
}

func TestRequestID_LimitExistingIDLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestID())

	router.GET("/test", func(c *gin.Context) {
		requestID, _ := middleware.GetRequestIDFromContext(c.Request.Context())
		c.String(http.StatusOK, requestID)
	})

	// Create an excessively long request ID.
	existingID := "long-request-ID-0123456789-0123456789-0123456789-0123456789-0123456789-0123456789"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(middleware.DefaultRequestIDHeader, existingID)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	requestID := w.Body.String()
	assert.NotEqual(t, existingID, requestID)
	assert.Equal(t, requestID, w.Header().Get(middleware.DefaultRequestIDHeader))
}
