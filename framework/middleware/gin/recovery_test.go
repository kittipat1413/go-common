package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kittipat1413/go-common/framework/logger"
	logger_mocks "github.com/kittipat1413/go-common/framework/logger/mocks"
	middleware "github.com/kittipat1413/go-common/framework/middleware/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware_Default(t *testing.T) {
	// Setup Gin in test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the Recovery middleware with default settings.
	router.Use(middleware.Recovery())

	// Add a route that will panic.
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Create a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	// Perform the request.
	router.ServeHTTP(w, req)

	// Assert that the status code is 500 Internal Server Error.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Assert that the response body contains the default error message.
	assert.JSONEq(t, `{"error": "Internal Server Error"}`, w.Body.String())
}

func TestRecoveryMiddleware_WithLogger(t *testing.T) {
	// Setup the gomock controller.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup Gin in test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a mock logger.
	mockLogger := logger_mocks.NewMockLogger(ctrl)

	// Define what we expect the logger to be called with.
	mockLogger.EXPECT().
		Error(gomock.Any(), "Panic recovered", nil, gomock.Any()).
		Times(1)

	// Use the Recovery middleware with the custom logger.
	router.Use(middleware.Recovery(
		middleware.WithRecoveryLogger(mockLogger),
	))

	// Add a route that will panic.
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Create a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	// Perform the request.
	router.ServeHTTP(w, req)

	// Assert that the status code is 500 Internal Server Error.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Assert that the response body contains the default error message.
	assert.JSONEq(t, `{"error": "Internal Server Error"}`, w.Body.String())
}

func TestRecoveryMiddleware_WithCustomHandler(t *testing.T) {
	// Setup Gin in test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Define a custom recovery handler.
	customHandler := func(c *gin.Context, err interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "from custom handler",
		})
	}

	// Use the Recovery middleware with the custom handler.
	router.Use(middleware.Recovery(
		middleware.WithRecoveryHandler(customHandler),
	))

	// Add a route that will panic.
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Create a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	// Perform the request.
	router.ServeHTTP(w, req)

	// Assert that the status code is 500 Internal Server Error.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Assert that the response body contains the custom error message.
	assert.JSONEq(t, `{"error": "from custom handler"}`, w.Body.String())
}

func TestRecoveryMiddleware_WithLoggerAndCustomHandler(t *testing.T) {
	// Setup the gomock controller.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup Gin in test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a mock logger.
	mockLogger := logger_mocks.NewMockLogger(ctrl)

	// Define what we expect the logger to be called with.
	mockLogger.EXPECT().
		Error(gomock.Any(), "Panic recovered", nil, gomock.Any()).
		Times(1)

	// Define a custom recovery handler.
	customHandler := func(c *gin.Context, err interface{}) {
		c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"error": "Bad Gateway",
		})
	}

	// Use the Recovery middleware with both custom logger and handler.
	router.Use(middleware.Recovery(
		middleware.WithRecoveryLogger(mockLogger),
		middleware.WithRecoveryHandler(customHandler),
	))

	// Add a route that will panic.
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Create a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	// Perform the request.
	router.ServeHTTP(w, req)

	// Assert that the status code is 502 Bad Gateway.
	assert.Equal(t, http.StatusBadGateway, w.Code)

	// Assert that the response body contains the custom error message.
	assert.JSONEq(t, `{"error": "Bad Gateway"}`, w.Body.String())
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	// Setup Gin in test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the Recovery middleware.
	router.Use(middleware.Recovery())

	// Add a normal route.
	router.GET("/no_panic", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "All good!"})
	})

	// Create a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/no_panic", nil)

	// Perform the request.
	router.ServeHTTP(w, req)

	// Assert that the status code is 200 OK.
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert that the response body contains the expected message.
	assert.JSONEq(t, `{"message": "All good!"}`, w.Body.String())
}

func TestRecoveryMiddleware_NilLogger(t *testing.T) {
	// Setup Gin in test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the Recovery middleware with a nil logger.
	router.Use(middleware.Recovery(
		middleware.WithRecoveryLogger(nil),
	))

	// Add a route that will panic.
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic with nil logger")
	})

	// Create a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	// Perform the request.
	router.ServeHTTP(w, req)

	// Assert that the status code is 500 Internal Server Error.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Assert that the response body contains the default error message.
	assert.JSONEq(t, `{"error": "Internal Server Error"}`, w.Body.String())
}

func TestRecoveryMiddleware_ContextLogger(t *testing.T) {
	// Setup the gomock controller.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup Gin in test mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a mock logger.
	mockLogger := logger_mocks.NewMockLogger(ctrl)

	// Define what we expect the logger to be called with.
	mockLogger.EXPECT().
		Error(gomock.Any(), "Panic recovered", nil, gomock.Any()).
		Times(1)

	// Create a middleware to inject the logger into the context.
	router.Use(func(c *gin.Context) {
		ctx := logger.NewContext(c.Request.Context(), mockLogger)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Use the Recovery middleware without providing a logger.
	router.Use(middleware.Recovery())

	// Add a route that will panic.
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic with context logger")
	})

	// Create a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	// Perform the request.
	router.ServeHTTP(w, req)

	// Assert that the status code is 500 Internal Server Error.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Assert that the response body contains the default error message.
	assert.JSONEq(t, `{"error": "Internal Server Error"}`, w.Body.String())
}
