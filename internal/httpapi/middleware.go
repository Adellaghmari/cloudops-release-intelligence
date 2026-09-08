package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	headerRequestID     = "X-Request-ID"
	headerCorrelationID = "X-Correlation-ID"
	maxIncomingIDLen    = 128
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := sanitizeIncomingID(c.GetHeader(headerRequestID))
		if reqID == "" {
			reqID = newPrefixedID("req_")
		}
		corrID := sanitizeIncomingID(c.GetHeader(headerCorrelationID))
		if corrID == "" {
			corrID = reqID
		}
		ctx := WithIDs(c.Request.Context(), reqID, corrID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("request_id", reqID)
		c.Set("correlation_id", corrID)
		c.Header(headerRequestID, reqID)
		c.Header(headerCorrelationID, corrID)
		c.Next()
	}
}

func CORSMiddleware(origins []string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, o := range origins {
		allowed[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-Request-ID, X-Correlation-ID")
			c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-Correlation-ID")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

func sanitizeIncomingID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxIncomingIDLen {
		return ""
	}
	for _, r := range raw {
		if !(r == '.' || r == '_' || r == '-' || r == ':' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return ""
		}
	}
	return raw
}

func newPrefixedID(prefix string) string {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return prefix + "fallback"
	}
	return prefix + hex.EncodeToString(b)
}
