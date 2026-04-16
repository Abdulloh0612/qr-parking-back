package server

import (
	"net/url"
	"strings"

	"qr-parking/docs"
)

// applySwaggerFromBaseURL sets Swagger host/schemes from APP_BASE_URL so Swagger UI
// "Try it out" calls the public API (not localhost / not raw {{.Host}}).
func applySwaggerFromBaseURL(baseURL string) {
	raw := strings.TrimSpace(baseURL)
	if raw == "" {
		return
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return
	}
	docs.SwaggerInfo.Host = u.Host
	switch u.Scheme {
	case "https":
		docs.SwaggerInfo.Schemes = []string{"https"}
	case "http":
		docs.SwaggerInfo.Schemes = []string{"http"}
	default:
		docs.SwaggerInfo.Schemes = nil
	}
}
