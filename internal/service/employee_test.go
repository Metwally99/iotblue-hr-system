package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type employeeRepositoryStub struct {
	createParams model.CreateEmployeeParams
	filter       model.EmployeeFilter
}

func (s *employeeRepositoryStub) Create(_ context.Context, params model.CreateEmployeeParams) (model.Employee, error) {
	s.createParams = params
	return model.Employee{ID: "94b61561-f428-40f5-a7d8-56ef5bd67a09", Email: params.Email, PasswordHash: params.PasswordHash}, nil
}
func (s *employeeRepositoryStub) FindByID(context.Context, string) (model.Employee, error) {
	return model.Employee{}, nil
}
func (s *employeeRepositoryStub) List(_ context.Context, filter model.EmployeeFilter) (model.EmployeePage, error) {
	s.filter = filter
	return model.EmployeePage{Page: filter.Page, PageSize: filter.PageSize}, nil
}
func (s *employeeRepositoryStub) Update(context.Context, model.UpdateEmployeeParams) (model.Employee, error) {
	return model.Employee{}, nil
}
func (s *employeeRepositoryStub) Deactivate(context.Context, string) (model.Employee, error) {
	return model.Employee{}, nil
}

func TestCreateEmployeeNormalizesAndHashesCredentials(t *testing.T) {
	repository := &employeeRepositoryStub{}
	service := NewEmployeeService(repository)
	service.now = func() time.Time { return time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC) }

	_, err := service.Create(context.Background(), CreateEmployeeInput{
		DepartmentID: "9e2f2d18-251e-4b70-98b4-d4ce61806210",
		Name:         "  Mahmoud Metwally  ", Email: " Mahmoud@Example.com ", Password: "strong-password",
		Role: model.RoleEmployee, TotalLeaveDays: 21,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.createParams.Name != "Mahmoud Metwally" || repository.createParams.Email != "mahmoud@example.com" {
		t.Fatalf("params = %#v", repository.createParams)
	}
	if repository.createParams.LeaveYear != 2026 || repository.createParams.TotalLeaveDays != 21 {
		t.Fatalf("leave params = %#v", repository.createParams)
	}
	if bcrypt.CompareHashAndPassword([]byte(repository.createParams.PasswordHash), []byte("strong-password")) != nil {
		t.Fatal("stored password hash does not verify")
	}
}

func TestCreateEmployeeCollectsValidationErrors(t *testing.T) {
	service := NewEmployeeService(&employeeRepositoryStub{})
	_, err := service.Create(context.Background(), CreateEmployeeInput{Name: "A", Email: "bad", Password: "short"})
	var validation *apperror.ValidationError
	if !errors.As(err, &validation) || len(validation.Fields) < 4 {
		t.Fatalf("Create() error = %#v", err)
	}
}

func TestListEmployeesAppliesPaginationDefaultsAndMaximum(t *testing.T) {
	repository := &employeeRepositoryStub{}
	service := NewEmployeeService(repository)
	_, err := service.List(context.Background(), model.EmployeeFilter{Page: 0, PageSize: 1000})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repository.filter.Page != 1 || repository.filter.PageSize != 100 {
		t.Fatalf("filter = %#v", repository.filter)
	}
}

func TestAdminCannotDeactivateSelf(t *testing.T) {
	service := NewEmployeeService(&employeeRepositoryStub{})
	id := "94b61561-f428-40f5-a7d8-56ef5bd67a09"
	_, err := service.Deactivate(context.Background(), id, id)
	var validation *apperror.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Deactivate() error = %v", err)
	}
}
