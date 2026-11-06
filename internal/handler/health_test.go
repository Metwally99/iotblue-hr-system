package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type pingerStub struct {
	err error
}

func (p pingerStub) Ping(context.Context) error { return p.err }

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", NewHealthHandler(pingerStub{}).Health)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestReadyWhenDatabaseIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ready", NewHealthHandler(pingerStub{err: errors.New("offline")}).Ready)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
