package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
	"github.com/mahmoudmetwally/hr-management-system/internal/security"
)

type authEmployeeRepositoryStub struct {
	employee model.Employee
	err      error
}

func (s authEmployeeRepositoryStub) FindByEmail(context.Context, string) (model.Employee, error) {
	return s.employee, s.err
}
func (s authEmployeeRepositoryStub) FindByID(context.Context, string) (model.Employee, error) {
	return s.employee, s.err
}

type authTokenStub struct {
	claims security.Claims
	err    error
}

func (s authTokenStub) Generate(model.Employee) (string, time.Time, error) {
	return "access-token", time.Date(2026, 8, 23, 12, 15, 0, 0, time.UTC), s.err
}
func (s authTokenStub) Verify(string) (security.Claims, error) { return s.claims, s.err }

func TestLoginSucceedsForActiveEmployee(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("strong-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	employee := model.Employee{ID: "employee-id", Email: "person@example.com", PasswordHash: string(hash), Role: model.RoleEmployee, IsActive: true}
	service := NewAuthService(authEmployeeRepositoryStub{employee: employee}, authTokenStub{})

	result, err := service.Login(context.Background(), " Person@Example.com ", "strong-password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.AccessToken != "access-token" || result.Employee.ID != employee.ID {
		t.Fatalf("result = %#v", result)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	employee := model.Employee{PasswordHash: string(hash), IsActive: true}
	service := NewAuthService(authEmployeeRepositoryStub{employee: employee}, authTokenStub{})

	_, err := service.Login(context.Background(), "person@example.com", "wrong-password")
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Fatalf("Login() error = %v", err)
	}
}

func TestAuthenticateRejectsDeactivatedEmployee(t *testing.T) {
	employee := model.Employee{ID: "employee-id", IsActive: false}
	tokens := authTokenStub{}
	tokens.claims.Subject = employee.ID
	service := NewAuthService(authEmployeeRepositoryStub{employee: employee}, tokens)

	_, err := service.Authenticate(context.Background(), "token")
	if !errors.Is(err, apperror.ErrInactive) {
		t.Fatalf("Authenticate() error = %v", err)
	}
}

func TestLoginHidesUnknownEmployee(t *testing.T) {
	service := NewAuthService(authEmployeeRepositoryStub{err: apperror.ErrNotFound}, authTokenStub{})

	_, err := service.Login(context.Background(), "missing@example.com", "some-password")
	if !errors.Is(err, apperror.ErrUnauthorized) {
		t.Fatalf("Login() error = %v", err)
	}
}
