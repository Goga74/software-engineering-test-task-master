package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// JSONLogger is a middleware that logs all incoming HTTP requests in JSON format
func JSONLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate request duration in milliseconds
		duration := time.Since(start).Milliseconds()

		// Extract route parameters
		params := make(map[string]interface{})
		for _, param := range c.Params {
			// Map parameter names like "username" to "user_id" format
			key := param.Key
			if key == "username" || key == "id" || key == "uuid" {
				params["user_id"] = param.Value
			} else {
				params[key] = param.Value
			}
		}

		// Build log entry
		logEntry := map[string]interface{}{
			"timestamp":                    time.Now().Format(time.RFC3339Nano),
			"http.server.request.duration": duration,
			"http.log.level":               getLogLevel(c.Writer.Status()),
			"http.request.method":          c.Request.Method,
			"http.response.status_code":    c.Writer.Status(),
			"http.route":                   c.FullPath(),
			"http.request.message":         "Incoming request:",
			"server.address":               c.Request.URL.Path,
			"http.request.host":            c.Request.Host,
		}

		// Add route parameters to the log entry
		for key, value := range params {
			logEntry[key] = value
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(logEntry)
		if err != nil {
			log.Printf("Error marshaling log entry: %v", err)
			return
		}

		// Log the JSON output with standard log prefix
		log.Printf("Incoming request: %s", string(jsonData))
	}
}

// getLogLevel determines the log level based on HTTP status code
func getLogLevel(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "error"
	case statusCode >= 400:
		return "warning"
	default:
		return "info"
	}
}
