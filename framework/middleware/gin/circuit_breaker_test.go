package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	middleware "github.com/kittipat1413/go-common/framework/middleware/gin"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_DefaultBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the CircuitBreaker middleware with default options.
	router.Use(middleware.CircuitBreaker())

	router.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Perform a successful request.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/success", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "success"}`, w.Body.String())
}

func TestCircuitBreaker_DefaultTripCondition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the CircuitBreaker middleware with default trip condition (5 consecutive failures).
	router.Use(middleware.CircuitBreaker())

	router.GET("/fail", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusInternalServerError)
	})

	// Cause 5 consecutive failures.
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/fail", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}

	// After the circuit breaker trips.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fail", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"error": "Service is temporarily unavailable due to high failure rate. Please try again later."}`, w.Body.String())
}

func TestCircuitBreaker_CustomTripCondition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the CircuitBreaker middleware with a custom trip condition.
	router.Use(middleware.CircuitBreaker(
		middleware.WithCircuitBreakerSettings(gobreaker.Settings{
			Name: "CustomBreaker",
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 3
			},
		}),
	))

	router.GET("/fail", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusInternalServerError)
	})

	// Cause 3 consecutive failures.
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/fail", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}

	// After the circuit breaker trips.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fail", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"error": "Service is temporarily unavailable due to high failure rate. Please try again later."}`, w.Body.String())
}

func TestCircuitBreaker_CustomStatusThreshold(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the CircuitBreaker middleware with a custom status threshold.
	router.Use(middleware.CircuitBreaker(
		middleware.WithCircuitBreakerStatusThreshold(http.StatusBadRequest), // Treat >= 400 as errors.
		middleware.WithCircuitBreakerSettings(gobreaker.Settings{
			Name: "CustomThresholdBreaker",
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 2
			},
		}),
	))

	router.GET("/fail", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusBadRequest)
	})

	// Cause 2 consecutive failures.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/fail", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	// After the circuit breaker trips.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fail", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"error": "Service is temporarily unavailable due to high failure rate. Please try again later."}`, w.Body.String())
}

func TestCircuitBreaker_CustomErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the CircuitBreaker middleware with a custom error handler.
	router.Use(middleware.CircuitBreaker(
		middleware.WithCircuitBreakerSettings(gobreaker.Settings{
			Name: "CustomHandlerBreaker",
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 2
			},
		}),
		middleware.WithCircuitBreakerErrorHandler(func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Custom error handler response",
			})
		}),
	))

	router.GET("/fail", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusInternalServerError)
	})

	// Cause 2 consecutive failures.
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/fail", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}

	// After the circuit breaker trips.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fail", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"error": "Custom error handler response"}`, w.Body.String())
}

func TestCircuitBreaker_SkipFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use the CircuitBreaker middleware with a filter to skip health check routes.
	router.Use(middleware.CircuitBreaker(
		middleware.WithCircuitBreakerFilter(func(req *http.Request) bool {
			return req.URL.Path != "/fail" // Skip circuit breaker for the "/fail" route.
		}),
	))

	router.GET("/fail", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed"})
	})

	// Cause 5 consecutive failures.
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/fail", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}

	// The circuit breaker should not trip because the route is skipped.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fail", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error": "Failed"}`, w.Body.String())
}
