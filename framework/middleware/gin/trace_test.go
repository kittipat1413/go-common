package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	middleware "github.com/kittipat1413/go-common/framework/middleware/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otelpropagation "go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestTraceMiddleware_Default(t *testing.T) {
	// Set Gin to Test Mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the SpanRecorder and TracerProvider.
	sr := tracetest.NewSpanRecorder()
	tp := tracesdk.NewTracerProvider(tracesdk.WithSpanProcessor(sr))

	// Apply the middleware with the test TracerProvider.
	router.Use(middleware.Trace(
		middleware.WithTracerProvider(tp),
	))

	// Add a test route.
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Perform a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Host = "example.com"
	router.ServeHTTP(w, req)

	// Assert the response.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "OK"}`, w.Body.String())

	// Retrieve the spans.
	spans := sr.Ended()
	require.Len(t, spans, 1)

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
	assert.Equal(t, "example.com", attrMap[semconv.ServerAddressKey].AsString())
}

func TestTraceMiddleware_CustomSpanNameFormatter(t *testing.T) {
	// Set Gin to Test Mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the SpanRecorder and TracerProvider.
	sr := tracetest.NewSpanRecorder()
	tp := tracesdk.NewTracerProvider(tracesdk.WithSpanProcessor(sr))

	// Custom SpanNameFormatter.
	spanNameFormatter := func(req *http.Request) string {
		return fmt.Sprintf("custom %s %s", req.Method, req.URL.Path)
	}

	// Apply the middleware with the custom SpanNameFormatter.
	router.Use(middleware.Trace(
		middleware.WithTracerProvider(tp),
		middleware.WithSpanNameFormatter(spanNameFormatter),
	))

	// Add a test route.
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Perform a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert the response.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "OK"}`, w.Body.String())

	// Retrieve the spans.
	spans := sr.Ended()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "custom GET /test", span.Name())
}

func TestTraceMiddleware_WithFilter(t *testing.T) {
	// Set Gin to Test Mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the SpanRecorder and TracerProvider.
	sr := tracetest.NewSpanRecorder()
	tp := tracesdk.NewTracerProvider(tracesdk.WithSpanProcessor(sr))

	// Filter to skip tracing requests to /notrace.
	skipTracingFilter := func(req *http.Request) bool {
		return req.URL.Path != "/notrace"
	}

	// Apply the middleware with the filter.
	router.Use(middleware.Trace(
		middleware.WithTracerProvider(tp),
		middleware.WithTraceFilter(skipTracingFilter),
	))

	// Add test routes.
	router.GET("/trace", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Trace"})
	})
	router.GET("/notrace", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "No Trace"})
	})

	// Perform a request to /trace.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/trace", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.JSONEq(t, `{"message": "Trace"}`, w1.Body.String())

	// Perform a request to /notrace.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/notrace", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.JSONEq(t, `{"message": "No Trace"}`, w2.Body.String())

	// Retrieve the spans.
	spans := sr.Ended()
	require.Len(t, spans, 1) // Only one span should be recorded for /trace.

	span := spans[0]
	assert.Equal(t, "GET /trace", span.Name())
}

func TestTraceMiddleware_WithPropagators(t *testing.T) {
	// Set Gin to Test Mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the SpanRecorder and TracerProvider.
	sr := tracetest.NewSpanRecorder()
	tp := tracesdk.NewTracerProvider(tracesdk.WithSpanProcessor(sr))

	// Create a custom propagator (no-op in this case).
	customPropagator := otelpropagation.NewCompositeTextMapPropagator()

	// Apply the middleware with the custom propagator.
	router.Use(middleware.Trace(
		middleware.WithTracerProvider(tp),
		middleware.WithTracePropagators(customPropagator),
	))

	// Add a test route.
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Perform a test request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert the response.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "OK"}`, w.Body.String())

	// Retrieve the spans.
	spans := sr.Ended()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "GET /test", span.Name())
}

func TestTraceMiddleware_ErrorHandling(t *testing.T) {
	// Set Gin to Test Mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the SpanRecorder and TracerProvider.
	sr := tracetest.NewSpanRecorder()
	tp := tracesdk.NewTracerProvider(tracesdk.WithSpanProcessor(sr))

	// Apply the middleware.
	router.Use(middleware.Trace(
		middleware.WithTracerProvider(tp),
	))

	// Add a route that returns an error.
	router.GET("/error", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
	})

	// Perform the request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	// Assert the response.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error": "Something went wrong"}`, w.Body.String())

	// Retrieve the spans.
	spans := sr.Ended()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "GET /error", span.Name())

	// Check that the span status is codes.Error.
	assert.Equal(t, otelcodes.Error, span.Status().Code)
	assert.Equal(t, "Internal Server Error", span.Status().Description)
}

func TestTraceMiddleware_GinErrors(t *testing.T) {
	// Set Gin to Test Mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the SpanRecorder and TracerProvider.
	sr := tracetest.NewSpanRecorder()
	tp := tracesdk.NewTracerProvider(tracesdk.WithSpanProcessor(sr))

	// Apply the middleware.
	router.Use(middleware.Trace(
		middleware.WithTracerProvider(tp),
	))

	// Add a route that adds an error to c.Errors.
	router.GET("/error", func(c *gin.Context) {
		_ = c.Error(fmt.Errorf("test error"))
		c.JSON(http.StatusOK, gin.H{"message": "OK with error"})
	})

	// Perform the request.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	// Assert the response.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "OK with error"}`, w.Body.String())

	// Retrieve the spans.
	spans := sr.Ended()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "GET /error", span.Name())

	// Check that the span has the gin.errors attribute.
	attrs := span.Attributes()
	var ginErrorsAttr attribute.KeyValue
	for _, attr := range attrs {
		if attr.Key == attribute.Key("gin.errors") {
			ginErrorsAttr = attr
			break
		}
	}
	require.NotNil(t, ginErrorsAttr)
	assert.Contains(t, ginErrorsAttr.Value.AsString(), "test error")
}

func TestTraceMiddleware_BuildRequestAttributes(t *testing.T) {
	// Set Gin to Test Mode.
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the SpanRecorder and TracerProvider.
	sr := tracetest.NewSpanRecorder()
	tp := tracesdk.NewTracerProvider(tracesdk.WithSpanProcessor(sr))

	// Apply the middleware.
	router.Use(middleware.Trace(
		middleware.WithTracerProvider(tp),
	))

	// Add a test route.
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Create a test request with headers.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent")
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	// Perform the request.
	router.ServeHTTP(w, req)

	// Assert the response.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "OK"}`, w.Body.String())

	// Retrieve the spans.
	spans := sr.Ended()
	require.Len(t, spans, 1)

	span := spans[0]

	// Check the attributes.
	attrs := span.Attributes()
	attrMap := make(map[attribute.Key]attribute.Value)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	// Assert that specific attributes are set correctly.
	assert.Equal(t, "http", attrMap[semconv.URLSchemeKey].AsString())
	assert.Equal(t, "POST", attrMap[semconv.HTTPRequestMethodKey].AsString())
	assert.Equal(t, "/test", attrMap[semconv.HTTPRouteKey].AsString())
	assert.Equal(t, "TestAgent", attrMap[semconv.UserAgentOriginalKey].AsString())
	assert.Equal(t, "1.2.3.4", attrMap[semconv.ClientAddressKey].AsString())
}
