package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	common_logger "github.com/kittipat1413/go-common/framework/logger"
	middleware "github.com/kittipat1413/go-common/framework/middleware/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger_Default(t *testing.T) {
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a buffer to capture logs.
	var logOutput bytes.Buffer
	// Create a logger that writes to the buffer.
	logger, err := common_logger.NewLogger(common_logger.Config{
		Level:  common_logger.INFO,
		Output: &logOutput,
	})
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Apply the middleware with the custom logger.
	router.Use(middleware.RequestLogger(
		middleware.WithRequestLogger(logger),
	))

	// Add a test route.
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Perform a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent")
	router.ServeHTTP(w, req)

	// Assert the response.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "OK"}`, w.Body.String())

	// Check that logs were written.
	logs := logOutput.String()
	assert.Contains(t, logs, `"method":"GET"`)
	assert.Contains(t, logs, `"path":"/test"`)
	assert.Contains(t, logs, `"status_code":200`)
	assert.Contains(t, logs, `"user_agent":"TestAgent"`)
}

func TestRequestLogger_WithFilter(t *testing.T) {
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a buffer to capture logs.
	var logOutput bytes.Buffer
	// Create a logger that writes to the buffer.
	logger, err := common_logger.NewLogger(common_logger.Config{
		Level:  common_logger.INFO,
		Output: &logOutput,
	})
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Define a filter to skip logging for /skip route.
	skipLoggingFilter := func(req *http.Request) bool {
		return req.URL.Path != "/skip"
	}

	// Apply the middleware with the custom logger and filter.
	router.Use(middleware.RequestLogger(
		middleware.WithRequestLogger(logger),
		middleware.WithRequestLoggerFilter(skipLoggingFilter),
	))

	// Add test routes.
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Test"})
	})
	router.GET("/skip", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Skip"})
	})

	// Perform a request to /test.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.JSONEq(t, `{"message": "Test"}`, w1.Body.String())

	// Perform a request to /skip.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/skip", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.JSONEq(t, `{"message": "Skip"}`, w2.Body.String())

	// Check that logs were written only for the /test route.
	logs := logOutput.String()
	assert.Contains(t, logs, `"path":"/test"`)
	assert.NotContains(t, logs, `"path":"/skip"`)
}

func TestRequestLogger_ContextLogger(t *testing.T) {
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a buffer to capture logs.
	var logOutput bytes.Buffer
	// Create a logger that writes to the buffer.
	logger, err := common_logger.NewLogger(common_logger.Config{
		Level:  common_logger.INFO,
		Output: &logOutput,
	})
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Apply the middleware with the custom logger.
	router.Use(middleware.RequestLogger(
		middleware.WithRequestLogger(logger),
	))

	// Add a test route that uses the logger from the context.
	router.GET("/test", func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := common_logger.FromContext(ctx)
		logger.Info(ctx, "Handler log message", nil)
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Perform a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert the response.
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the handler's log message includes request-specific fields.
	logs := logOutput.String()
	assert.Contains(t, logs, `"message":"Handler log message"`)
	assert.Contains(t, logs, `"path":"/test"`)
}
