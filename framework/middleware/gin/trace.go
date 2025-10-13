package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// traceOptions holds configuration options for the tracing middleware.
type traceOptions struct {
	tracerProvider    oteltrace.TracerProvider      // tracerProvider is the OpenTelemetry tracer provider to use.
	propagators       propagation.TextMapPropagator // propagators are used to extract and inject context information.
	filters           []TraceFilter                 // filters is a list of functions to determine whether a request should be traced.
	spanNameFormatter SpanNameFormatter             // spanNameFormatter is a function to generate span names based on the request.
}

// TraceOption specifies instrumentation configuration options.
type TraceOption func(*traceOptions)

// TraceFilter determines whether a request should be traced.
// Returns true to create a span for the request, false to skip tracing entirely.
type TraceFilter func(*http.Request) bool

// SpanNameFormatter generates custom span names based on HTTP request information.
// Enables context-aware span naming for better trace organization and filtering.
type SpanNameFormatter func(*http.Request) string

// WithTracePropagators specifies propagators to use for extracting information from the HTTP requests.
func WithTracePropagators(propagators propagation.TextMapPropagator) TraceOption {
	return func(opts *traceOptions) {
		if propagators != nil {
			opts.propagators = propagators
		}
	}
}

// WithTracerProvider sets a custom tracer provider for span creation.
func WithTracerProvider(provider oteltrace.TracerProvider) TraceOption {
	return func(opts *traceOptions) {
		if provider != nil {
			opts.tracerProvider = provider
		}
	}
}

// WithTraceFilter adds request filters for selective tracing.
// All filters must return true for a request to be traced.
func WithTraceFilter(filters ...TraceFilter) TraceOption {
	return func(opts *traceOptions) {
		opts.filters = append(opts.filters, filters...)
	}
}

// WithSpanNameFormatter sets a custom function to format the span name for each request.
func WithSpanNameFormatter(formatter SpanNameFormatter) TraceOption {
	return func(opts *traceOptions) {
		opts.spanNameFormatter = formatter
	}
}

// Trace returns Gin middleware that integrates OpenTelemetry distributed tracing.
// Creates spans for HTTP requests, propagates trace context, and captures comprehensive
// request attributes for distributed system observability and performance monitoring.
//
// The middleware performs the following actions:
//  1. Initializes tracing options using the provided TraceOption functions or falls back to defaults.
//     - If no tracer provider is provided, it uses the global tracer provider from OpenTelemetry.
//     - If no propagators are specified, it uses the global text map propagators.
//  2. Applies user-defined filters to determine whether a request should be traced.
//  3. Extracts tracing context from the incoming request headers using the specified propagators.
//  4. Determines the span name using a custom formatter or defaults to "<METHOD> <PATH>".
//  5. Creates a new span and adds common HTTP attributes to the span (e.g., method, path, client IP).
//  6. Passes the span context through the request for use by downstream handlers and middlewares.
//  7. Records errors and sets the span status based on the HTTP response status code.
//  8. Ends the span once the request processing is complete.
//
// Example Usage:
//
//	// Basic usage with defaults
//	router.Use(Trace())
//
//	// Custom configuration
//	router.Use(Trace(
//	    WithTracerProvider(myTracerProvider),
//	    WithTraceFilter(func(req *http.Request) bool {
//	        return req.URL.Path != "/health" // Skip health checks
//	    }),
//	    WithSpanNameFormatter(func(req *http.Request) string {
//	        return fmt.Sprintf("api-service %s %s", req.Method, req.URL.Path)
//	    }),
//	))
func Trace(options ...TraceOption) gin.HandlerFunc {
	// Initialize default configuration.
	opts := &traceOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// Use the global tracer provider if none is specified.
	if opts.tracerProvider == nil {
		opts.tracerProvider = otel.GetTracerProvider()
	}
	tracer := opts.tracerProvider.Tracer(
		"github.com/kittipat1413/go-common/framework/middleware/gin",
	)

	// Use the global propagators if none are specified.
	if opts.propagators == nil {
		opts.propagators = otel.GetTextMapPropagator()
	}

	// Return the middleware handler.
	return func(c *gin.Context) {
		// Apply filters to determine if the request should be traced.
		for _, filter := range opts.filters {
			if !filter(c.Request) {
				// If a filter rejects the request, proceed without tracing.
				c.Next()
				return
			}
		}

		// Extract the context from the incoming request.
		ctx := opts.propagators.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Determine the span name.
		var spanName string
		if opts.spanNameFormatter != nil {
			spanName = opts.spanNameFormatter(c.Request)
		} else if fullPath := c.FullPath(); fullPath != "" {
			spanName = fmt.Sprintf("%s %s", c.Request.Method, fullPath)
		}
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		}

		// Start a new span with the extracted context.
		ctx, span := tracer.Start(ctx, spanName,
			oteltrace.WithAttributes(buildRequestAttributes(c)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		)
		defer span.End()

		// Pass the span context through the request.
		c.Request = c.Request.WithContext(ctx)

		// Process the request.
		c.Next()

		// Set the span status based on the response status code.
		statusCode := c.Writer.Status()
		code, description := convertHTTPStatusToOtelCode(statusCode)
		span.SetStatus(code, description)
		if statusCode > 0 {
			span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))
		}

		// Record any errors from the Gin context.
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
		}
	}
}

// buildRequestAttributes constructs comprehensive HTTP span attributes following OpenTelemetry conventions.
// Captures request metadata, network information, and client details for rich trace context.
func buildRequestAttributes(c *gin.Context) []attribute.KeyValue {
	// Determine the scheme (http or https).
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	// Initialize the attributes with common HTTP request attributes.
	attributes := []attribute.KeyValue{
		semconv.URLSchemeKey.String(scheme),
		semconv.HTTPRequestMethodKey.String(c.Request.Method),
		semconv.HTTPRouteKey.String(c.FullPath()),
		semconv.URLFullKey.String(c.Request.URL.String()),
		semconv.URLPathKey.String(c.Request.URL.Path),
		semconv.ServerAddressKey.String(c.Request.Host),
	}

	// Add request content length if available.
	if c.Request.ContentLength > 0 {
		attributes = append(attributes, semconv.HTTPRequestBodySizeKey.Int64(c.Request.ContentLength))
	}

	// Parse the Host header to get host and port.
	if host, port, err := net.SplitHostPort(c.Request.Host); err == nil {
		attributes = append(attributes, semconv.ServerAddressKey.String(host))
		if portNum, err := strconv.Atoi(port); err == nil {
			attributes = append(attributes, semconv.ServerPortKey.Int(portNum))
		}
	} else {
		// If unable to split, use the entire Host.
		attributes = append(attributes, semconv.ServerAddressKey.String(c.Request.Host))
	}

	// Parse RemoteAddr to get client IP and port.
	if ip, portStr, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		attributes = append(attributes, semconv.NetworkPeerAddressKey.String(ip))
		if portNum, err := strconv.Atoi(portStr); err == nil {
			attributes = append(attributes, semconv.NetworkPeerPortKey.Int(portNum))
		}
	} else {
		// If unable to split, use the entire RemoteAddr.
		attributes = append(attributes, semconv.NetworkPeerAddressKey.String(c.Request.RemoteAddr))
	}

	// Add client IP address from X-Forwarded-For header if available.
	if xForwardedFor := c.Request.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// Use the first IP in the list.
		if idx := strings.Index(xForwardedFor, ","); idx >= 0 {
			xForwardedFor = xForwardedFor[:idx]
		}
		attributes = append(attributes, semconv.ClientAddressKey.String(xForwardedFor))
	}

	// Add User-Agent header if available.
	if userAgent := c.Request.UserAgent(); userAgent != "" {
		attributes = append(attributes, semconv.UserAgentOriginalKey.String(userAgent))
	}

	return attributes
}

// convertHTTPStatusToOtelCode maps HTTP status codes to OpenTelemetry status codes.
// Follows OpenTelemetry HTTP semantic conventions for consistent error classification.
func convertHTTPStatusToOtelCode(statusCode int) (codes.Code, string) {
	if statusCode < 100 || statusCode >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", statusCode)
	}
	if statusCode >= 400 {
		return codes.Error, http.StatusText(statusCode)
	}
	return codes.Unset, ""
}
