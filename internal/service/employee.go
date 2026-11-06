package service

import (
	"context"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type EmployeeRepository interface {
	Create(context.Context, model.CreateEmployeeParams) (model.Employee, error)
	FindByID(context.Context, string) (model.Employee, error)
	List(context.Context, model.EmployeeFilter) (model.EmployeePage, error)
	Update(context.Context, model.UpdateEmployeeParams) (model.Employee, error)
	Deactivate(context.Context, string) (model.Employee, error)
}

type EmployeeService struct {
	repository EmployeeRepository
	now        func() time.Time
}

type CreateEmployeeInput struct {
	DepartmentID   string
	ManagerID      *string
	Name           string
	Email          string
	Password       string
	Role           model.Role
	TotalLeaveDays int
}

func NewEmployeeService(repository EmployeeRepository) *EmployeeService {
	return &EmployeeService{repository: repository, now: time.Now}
}

func (s *EmployeeService) Create(ctx context.Context, input CreateEmployeeInput) (model.Employee, error) {
	fields := make(map[string]string)
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 150 {
		fields["name"] = "name must contain 2 to 150 characters"
	}
	if !validEmail(email) {
		fields["email"] = "email must be a valid address"
	}
	if len(input.Password) < 8 || len(input.Password) > 72 {
		fields["password"] = "password must contain 8 to 72 bytes"
	}
	if !input.Role.Valid() {
		fields["role"] = "role must be admin, hr, manager, or employee"
	}
	if !validID(input.DepartmentID) {
		fields["department_id"] = "department_id must be a UUID"
	}
	if input.ManagerID != nil && !validID(*input.ManagerID) {
		fields["manager_id"] = "manager_id must be a UUID"
	}
	if input.TotalLeaveDays < 1 || input.TotalLeaveDays > 365 {
		fields["total_leave_days"] = "total_leave_days must be between 1 and 365"
	}
	if len(fields) > 0 {
		return model.Employee{}, &apperror.ValidationError{Fields: fields}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.Employee{}, err
	}
	return s.repository.Create(ctx, model.CreateEmployeeParams{
		DepartmentID: input.DepartmentID, ManagerID: input.ManagerID, Name: name, Email: email,
		PasswordHash: string(hash), Role: input.Role, LeaveYear: s.now().Year(), TotalLeaveDays: input.TotalLeaveDays,
	})
}

func (s *EmployeeService) GetByID(ctx context.Context, id string) (model.Employee, error) {
	if !validID(id) {
		return model.Employee{}, &apperror.ValidationError{Fields: map[string]string{"id": "employee ID must be a UUID"}}
	}
	return s.repository.FindByID(ctx, id)
}

func (s *EmployeeService) List(ctx context.Context, filter model.EmployeeFilter) (model.EmployeePage, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.DepartmentID != nil && !validID(*filter.DepartmentID) {
		return model.EmployeePage{}, &apperror.ValidationError{Fields: map[string]string{"department_id": "department_id must be a UUID"}}
	}
	return s.repository.List(ctx, filter)
}

func (s *EmployeeService) Update(ctx context.Context, actorID string, params model.UpdateEmployeeParams) (model.Employee, error) {
	fields := make(map[string]string)
	if !validID(params.ID) {
		fields["id"] = "employee ID must be a UUID"
	}
	if params.DepartmentID != nil && !validID(*params.DepartmentID) {
		fields["department_id"] = "department_id must be a UUID"
	}
	if params.ManagerID != nil {
		if !validID(*params.ManagerID) {
			fields["manager_id"] = "manager_id must be a UUID"
		} else if *params.ManagerID == params.ID {
			fields["manager_id"] = "employee cannot manage themselves"
		}
	}
	if params.Name != nil {
		trimmed := strings.TrimSpace(*params.Name)
		params.Name = &trimmed
		if utf8.RuneCountInString(trimmed) < 2 || utf8.RuneCountInString(trimmed) > 150 {
			fields["name"] = "name must contain 2 to 150 characters"
		}
	}
	if params.Email != nil {
		normalized := strings.ToLower(strings.TrimSpace(*params.Email))
		params.Email = &normalized
		if !validEmail(normalized) {
			fields["email"] = "email must be a valid address"
		}
	}
	if params.Role != nil && !params.Role.Valid() {
		fields["role"] = "role must be admin, hr, manager, or employee"
	}
	if actorID == params.ID && params.Role != nil && *params.Role != model.RoleAdmin {
		fields["role"] = "administrators cannot remove their own administrator role"
	}
	if len(fields) > 0 {
		return model.Employee{}, &apperror.ValidationError{Fields: fields}
	}
	return s.repository.Update(ctx, params)
}

func (s *EmployeeService) Deactivate(ctx context.Context, actorID, employeeID string) (model.Employee, error) {
	if actorID == employeeID {
		return model.Employee{}, &apperror.ValidationError{Fields: map[string]string{"id": "administrators cannot deactivate themselves"}}
	}
	if !validID(employeeID) {
		return model.Employee{}, &apperror.ValidationError{Fields: map[string]string{"id": "employee ID must be a UUID"}}
	}
	return s.repository.Deactivate(ctx, employeeID)
}

func validID(value string) bool { _, err := uuid.Parse(value); return err == nil }

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
