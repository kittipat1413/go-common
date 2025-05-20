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

var (
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestSize     *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec

	// for thread-safe initialization of Prometheus metrics
	prometheusInitOnce sync.Once
)

// initPrometheusMetrics sets up and registers Prometheus metrics with the provided namespace.
// This function initializes the following metrics:
//
//   - requests_total: Counter for total HTTP requests.
//   - request_duration_seconds: Histogram of request latencies.
//   - request_size_bytes: Histogram of request sizes.
//   - response_size_bytes: Histogram of response sizes.
//
// These metrics are labeled with HTTP status, method, and path.
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

// Prometheus returns a Gin middleware that exports HTTP metrics to Prometheus.
//
// It captures the following metrics per request:
//   - requests_total
//   - request_duration_seconds
//   - request_size_bytes
//   - response_size_bytes
//
// Parameters:
//   - serviceName: used as the namespace prefix for the exported metrics.
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
//
// Example:
//
//	router.GET("/metrics", middleware.MetricsHandler())
func MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// approximateRequestSize estimates the size in bytes of an incoming HTTP request.
// It includes the request line, headers, and content length (if present).
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
