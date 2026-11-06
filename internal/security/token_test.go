package security

import (
	"testing"
	"time"

	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

func TestTokenRoundTrip(t *testing.T) {
	manager, err := NewTokenManager("a-test-secret-that-is-at-least-32-characters", "test-api", 15*time.Minute)
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}
	manager.now = func() time.Time { return time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC) }

	raw, expiresAt, err := manager.Generate(model.Employee{ID: "employee-id", Role: model.RoleAdmin})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	claims, err := manager.Verify(raw)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.Subject != "employee-id" || claims.Role != model.RoleAdmin {
		t.Fatalf("claims = %#v", claims)
	}
	if !expiresAt.Equal(time.Date(2026, 8, 23, 12, 15, 0, 0, time.UTC)) {
		t.Fatalf("expiresAt = %v", expiresAt)
	}
}

func TestTokenRejectsWrongSecret(t *testing.T) {
	issuer := "test-api"
	first, _ := NewTokenManager("first-secret-that-is-at-least-32-characters", issuer, time.Minute)
	second, _ := NewTokenManager("second-secret-that-is-at-least-32-characters", issuer, time.Minute)
	raw, _, _ := first.Generate(model.Employee{ID: "employee-id", Role: model.RoleEmployee})

	if _, err := second.Verify(raw); err == nil {
		t.Fatal("expected token signed with another secret to fail")
	}
}
