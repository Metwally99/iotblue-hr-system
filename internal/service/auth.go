package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
	"github.com/mahmoudmetwally/hr-management-system/internal/security"
)

type AuthEmployeeRepository interface {
	FindByEmail(context.Context, string) (model.Employee, error)
	FindByID(context.Context, string) (model.Employee, error)
}

type AuthTokenManager interface {
	Generate(model.Employee) (string, time.Time, error)
	Verify(string) (security.Claims, error)
}

type AuthService struct {
	employees AuthEmployeeRepository
	tokens    AuthTokenManager
}

type LoginResult struct {
	AccessToken string
	ExpiresAt   time.Time
	Employee    model.Employee
}

func NewAuthService(employees AuthEmployeeRepository, tokens AuthTokenManager) *AuthService {
	return &AuthService{employees: employees, tokens: tokens}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return LoginResult{}, apperror.ErrUnauthorized
	}

	employee, err := s.employees.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			// Comparing a fixed hash makes unknown-email timing closer to a normal failed login.
			_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(password))
			return LoginResult{}, apperror.ErrUnauthorized
		}
		return LoginResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(employee.PasswordHash), []byte(password)) != nil {
		return LoginResult{}, apperror.ErrUnauthorized
	}
	if !employee.IsActive {
		return LoginResult{}, apperror.ErrInactive
	}

	token, expiresAt, err := s.tokens.Generate(employee)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{AccessToken: token, ExpiresAt: expiresAt, Employee: employee}, nil
}

func (s *AuthService) Authenticate(ctx context.Context, rawToken string) (model.Identity, error) {
	claims, err := s.tokens.Verify(rawToken)
	if err != nil {
		return model.Identity{}, apperror.ErrUnauthorized
	}
	employee, err := s.employees.FindByID(ctx, claims.Subject)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return model.Identity{}, apperror.ErrUnauthorized
		}
		return model.Identity{}, err
	}
	if !employee.IsActive {
		return model.Identity{}, apperror.ErrInactive
	}
	return model.Identity{EmployeeID: employee.ID, Name: employee.Name, Email: employee.Email, Role: employee.Role}, nil
}

const dummyPasswordHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoO5y8Cq1cCnZLQJ6wN2OaF9D4E7Q7Q2W2"
