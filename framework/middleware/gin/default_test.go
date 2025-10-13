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
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestConfigureDefaultMiddlewares(t *testing.T) {
	var logOutput bytes.Buffer
	testLogger, err := common_logger.NewLogger(common_logger.Config{
		Level:  common_logger.INFO,
		Output: &logOutput,
	})
	require.NoError(t, err)

	sr := tracetest.NewSpanRecorder()
	testTracerProvider := tracesdk.NewTracerProvider(tracesdk.WithSpanProcessor(sr))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ConfigureDefaultMiddlewares(middleware.DefaultMiddlewareConfig{
		Logger:         testLogger,
		TracerProvider: testTracerProvider,
	})...)

	router.GET("/test", func(c *gin.Context) {
		requestID, _ := middleware.GetRequestIDFromContext(c.Request.Context())
		c.String(http.StatusOK, requestID)
	})
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	router.GET("/fail", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusInternalServerError)
	})

	t.Run("A span should be recorded with correct attributes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		spans := sr.Ended()
		require.Len(t, spans, 1) // This is first sub-test, so we expect only 1 span

		span := spans[0]
		assert.Equal(t, "GET /test", span.Name())
		assert.Equal(t, oteltrace.SpanKindServer, span.SpanKind())

		// Check that the span status is Unset (i.e., no error).
		assert.Equal(t, otelcodes.Unset, span.Status().Code)

		attrs := span.Attributes()
		attrMap := make(map[attribute.Key]attribute.Value)
		for _, attr := range attrs {
			attrMap[attr.Key] = attr.Value
		}

		// Assert that specific attributes are set correctly.
		assert.Equal(t, "http", attrMap[semconv.URLSchemeKey].AsString())
		assert.Equal(t, "GET", attrMap[semconv.HTTPRequestMethodKey].AsString())
		assert.Equal(t, "/test", attrMap[semconv.HTTPRouteKey].AsString())
		assert.Equal(t, "/test", attrMap[semconv.URLFullKey].AsString())
		assert.Equal(t, "/test", attrMap[semconv.URLPathKey].AsString())
	})

	t.Run("Response Header should contain request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		requestID := w.Body.String()
		require.NotEmpty(t, requestID)
		assert.Equal(t, requestID, w.Header().Get(middleware.DefaultRequestIDHeader))
	})

	t.Run("Logs should contain method, path, and status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		logs := logOutput.String()
		assert.Contains(t, logs, `"method":"GET"`)
		assert.Contains(t, logs, `"path":"/test"`)
		assert.Contains(t, logs, `"status_code":200`)
	})

	t.Run("Panic should be recovered and return 502 Bad Gateway", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Internal Server Error"}`, w.Body.String())
	})

	t.Run("Circuit breaker should trip after 5 consecutive failures", func(t *testing.T) {
		// Reset the circuit breaker with the success handler.
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Cause 5 consecutive failures.
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/fail", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		}

		// After the circuit breaker trips.
		w = httptest.NewRecorder()
		req = httptest.NewRequest("GET", "/fail", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.JSONEq(t, `{"error": "Service is temporarily unavailable due to high failure rate. Please try again later."}`, w.Body.String())
	})
}
