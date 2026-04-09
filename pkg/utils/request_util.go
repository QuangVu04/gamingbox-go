package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetQueryInt extracts an integer query parameter from the request
// Returns the parsed value, or a default value if parsing fails
func GetQueryInt(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.DefaultQuery(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetQueryIntWithRange extracts an integer query parameter with min/max bounds
// If value is outside bounds, returns the default value
func GetQueryIntWithRange(c *gin.Context, key string, defaultValue, minValue, maxValue int) int {
	value := GetQueryInt(c, key, defaultValue)

	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}

	return value
}

// GetQueryString extracts a string query parameter from the request
// Returns the value, or a default value if not provided
func GetQueryString(c *gin.Context, key string, defaultValue string) string {
	value := c.DefaultQuery(key, "")
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryBool extracts a boolean query parameter from the request
// Accepts: "true", "1", "yes" as true; anything else is false
func GetQueryBool(c *gin.Context, key string, defaultValue bool) bool {
	valueStr := c.DefaultQuery(key, "")
	if valueStr == "" {
		return defaultValue
	}

	return valueStr == "true" || valueStr == "1" || valueStr == "yes"
}
