package middleware_test

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	middleware "github.com/kittipat1413/go-common/framework/middleware/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMiddleware_BasicMetricsExposure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(middleware.Prometheus(""))

	// Define a test route
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "world"})
	})

	// Define metrics endpoint
	router.GET("/metrics", middleware.MetricsHandler())

	// Make a test request to /hello
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hello", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "world"}`, w.Body.String())

	// Make a request to /metrics
	metricsResp := httptest.NewRecorder()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(metricsResp, metricsReq)
	require.Equal(t, http.StatusOK, metricsResp.Code)

	// Check that expected metrics exist
	body := metricsResp.Body.String()
	scanner := bufio.NewScanner(strings.NewReader(body))

	var foundCountMetric, foundDurationMetric bool
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "requests_total") {
			foundCountMetric = true
		}
		if strings.Contains(line, "request_duration_seconds") {
			foundDurationMetric = true
		}
	}
	assert.True(t, foundCountMetric, "Expected metric requests_total not found")
	assert.True(t, foundDurationMetric, "Expected metric request_duration_seconds not found")
}

func TestPrometheusMiddleware_MetricsForNotFoundRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Prometheus(""))

	// Intentionally not defining route to simulate 404
	router.GET("/metrics", middleware.MetricsHandler())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Make a request to metrics
	metricsResp := httptest.NewRecorder()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(metricsResp, metricsReq)

	assert.Contains(t, metricsResp.Body.String(), "requests_total")
	assert.Contains(t, metricsResp.Body.String(), "path=\"/notfound\"")
	assert.Contains(t, metricsResp.Body.String(), "status=\"404\"")
}

func TestPrometheusMiddleware_RequestSizeMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(middleware.Prometheus(""))
	router.POST("/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "received"})
	})
	router.GET("/metrics", middleware.MetricsHandler())

	// Define a POST request with body and headers
	body := strings.NewReader(`{"foo":"bar"}`)
	req := httptest.NewRequest(http.MethodPost, "/echo", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom", "abc")
	req.Host = "example.com"
	req.ContentLength = int64(body.Len())

	// Execute the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Query metrics
	metricsResp := httptest.NewRecorder()
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(metricsResp, metricsReq)

	// Look for request_size_bytes_sum with /echo label
	bodyLines := strings.Split(metricsResp.Body.String(), "\n")
	var matchedLine string
	for _, line := range bodyLines {
		if strings.HasPrefix(line, "request_size_bytes_sum") &&
			strings.Contains(line, `path="/echo"`) {
			matchedLine = line
			break
		}
	}

	require.NotEmpty(t, matchedLine, "Expected request_size_bytes_sum for /echo not found")
	parts := strings.Fields(matchedLine)
	require.Len(t, parts, 2)

	sizeValue, err := strconv.ParseFloat(parts[1], 64)
	require.NoError(t, err)

	// Assert a non-zero request size was recorded
	assert.Greater(t, sizeValue, 0.0)
}
