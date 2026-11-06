package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS([]string{"http://localhost:3000"}))
	router.GET("/resource", func(c *gin.Context) { c.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodGet, "/resource", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected configured origin, got %q", got)
	}
}

func TestCORSHandlesPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS([]string{"http://localhost:3000"}))

	request := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected allowed methods header")
	}
}

func TestCORSRejectsUnknownPreflightOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS([]string{"http://localhost:3000"}))

	request := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	request.Header.Set("Origin", "https://untrusted.example")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS grant, got %q", got)
	}
}
