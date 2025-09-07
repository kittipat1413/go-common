package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Global Prometheus metric collectors for HTTP request monitoring.
// These metrics provide comprehensive observability into HTTP traffic patterns.
var (
	requestCount    *prometheus.CounterVec   // Total number of HTTP requests
	requestDuration *prometheus.HistogramVec // Request latency distribution in seconds
	requestSize     *prometheus.HistogramVec // Request size distribution in bytes
	responseSize    *prometheus.HistogramVec // Response size distribution in bytes

	// prometheusInitOnce ensures thread-safe one-time initialization of metrics
	prometheusInitOnce sync.Once
)

// initPrometheusMetrics sets up and registers Prometheus metrics with the provided namespace.
// Creates four key HTTP metrics with consistent labeling for comprehensive request monitoring.
//
// Metrics Created:
//   - {namespace}_requests_total: Counter for total HTTP requests
//   - {namespace}_request_duration_seconds: Histogram of request latencies
//   - {namespace}_request_size_bytes: Histogram of request sizes
//   - {namespace}_response_size_bytes: Histogram of response sizes
//
// All metrics are labeled with: status, method, path for detailed filtering and aggregation.
//
// Bucket Configuration:
//   - Duration: Prometheus defaults (5ms to 10s) for typical web request latencies
//   - Size: Exponential buckets (1KB to 16MB) for HTTP payload sizes
//
// Parameters:
//   - namespace: Metric namespace prefix (typically service name)
func initPrometheusMetrics(namespace string) {
	labels := []string{"status", "method", "path"}

	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed, partitioned by status, method, and path.",
		},
		labels,
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "Histogram of HTTP request latencies in seconds.",
			// Here, we use the prometheus defaults which are for ~10s request length max: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
			Buckets: prometheus.DefBuckets,
		},
		labels,
	)

	requestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_size_bytes",
			Help:      "Histogram of HTTP request sizes in bytes.",
			// Buckets from 1KiB doubling up to ~16MiB
			Buckets: prometheus.ExponentialBuckets(1024, 2, 15),
		},
		labels,
	)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "response_size_bytes",
			Help:      "Histogram of HTTP response sizes in bytes.",
			// Buckets from 1KiB doubling up to ~16MiB
			Buckets: prometheus.ExponentialBuckets(1024, 2, 15),
		},
		labels,
	)

	prometheus.MustRegister(requestCount, requestDuration, requestSize, responseSize)
}

// Prometheus returns Gin middleware that collects HTTP metrics for Prometheus monitoring.
// Captures request counts, latencies, and payload sizes with detailed labeling for
// comprehensive HTTP traffic observability and performance monitoring.
//
// Collected Metrics:
//   - Request count: Total requests segmented by status/method/path
//   - Request duration: Latency histogram for performance monitoring
//   - Request size: Payload size distribution for bandwidth analysis
//   - Response size: Response payload size distribution
//
// Parameters:
//   - serviceName: Namespace prefix for metrics (empty string for no prefix)
//
// Returns:
//   - gin.HandlerFunc: Middleware that captures metrics for each request
//
// Example:
//
//	router := gin.Default()
//	router.Use(middleware.Prometheus("service_name"))
//	router.GET("/metrics", middleware.MetricsHandler())
func Prometheus(serviceName string) gin.HandlerFunc {
	// Initialize Prometheus metrics with the provided service name.
	prometheusInitOnce.Do(func() {
		initPrometheusMetrics(serviceName)
	})
	// Return a Gin middleware function that captures metrics for each request.
	return func(c *gin.Context) {
		start := time.Now()
		reqSz := approximateRequestSize(c.Request)
		c.Next()
		latency := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		labels := prometheus.Labels{"status": status, "method": method, "path": path}

		requestCount.With(labels).Inc()
		requestDuration.With(labels).Observe(latency)
		requestSize.With(labels).Observe(float64(reqSz))
		responseSize.With(labels).Observe(float64(c.Writer.Size()))
	}
}

// MetricsHandler exposes the Prometheus `/metrics` endpoint as a Gin handler.
// Returns all registered metrics in Prometheus exposition format for scraping
// by Prometheus servers or compatible monitoring systems.
//
// Returns:
//   - gin.HandlerFunc: Handler that serves Prometheus metrics endpoint
//
// Example:
//
//	router.GET("/metrics", middleware.MetricsHandler())
func MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// approximateRequestSize estimates the total size of an HTTP request in bytes.
// Includes request line (method + path + protocol), all headers, and content length.
// Provides approximate sizing for bandwidth and payload monitoring.
func approximateRequestSize(r *http.Request) int {
	sz := len(r.Method) + len(r.URL.Path) + len(r.Proto) + len(r.Host)
	for name, vals := range r.Header {
		sz += len(name)
		for _, v := range vals {
			sz += len(v)
		}
	}
	if r.ContentLength > 0 {
		sz += int(r.ContentLength)
	}
	return sz
}
