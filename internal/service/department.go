package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type DepartmentRepository interface {
	Create(context.Context, string) (model.Department, error)
	GetByID(context.Context, string) (model.Department, error)
	List(context.Context) ([]model.Department, error)
	Update(context.Context, string, string) (model.Department, error)
}

type DepartmentService struct{ repository DepartmentRepository }

func NewDepartmentService(repository DepartmentRepository) *DepartmentService {
	return &DepartmentService{repository: repository}
}

func (s *DepartmentService) Create(ctx context.Context, name string) (model.Department, error) {
	name, err := validDepartmentName(name)
	if err != nil {
		return model.Department{}, err
	}
	return s.repository.Create(ctx, name)
}

func (s *DepartmentService) GetByID(ctx context.Context, id string) (model.Department, error) {
	if strings.TrimSpace(id) == "" {
		return model.Department{}, &apperror.ValidationError{Fields: map[string]string{"id": "department ID is required"}}
	}
	return s.repository.GetByID(ctx, id)
}

func (s *DepartmentService) List(ctx context.Context) ([]model.Department, error) {
	return s.repository.List(ctx)
}

func (s *DepartmentService) Update(ctx context.Context, id, name string) (model.Department, error) {
	name, err := validDepartmentName(name)
	if err != nil {
		return model.Department{}, err
	}
	return s.repository.Update(ctx, id, name)
}

func validDepartmentName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 100 {
		return "", &apperror.ValidationError{Fields: map[string]string{"name": "name must contain 2 to 100 characters"}}
	}
	return name, nil
}
