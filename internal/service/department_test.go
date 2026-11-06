package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type departmentRepositoryStub struct {
	createdName string
}

func (s *departmentRepositoryStub) Create(_ context.Context, name string) (model.Department, error) {
	s.createdName = name
	return model.Department{ID: "department-id", Name: name}, nil
}
func (s *departmentRepositoryStub) GetByID(context.Context, string) (model.Department, error) {
	return model.Department{}, nil
}
func (s *departmentRepositoryStub) List(context.Context) ([]model.Department, error) {
	return nil, nil
}
func (s *departmentRepositoryStub) Update(_ context.Context, id, name string) (model.Department, error) {
	return model.Department{ID: id, Name: name}, nil
}

func TestCreateDepartmentTrimsName(t *testing.T) {
	repository := &departmentRepositoryStub{}
	service := NewDepartmentService(repository)

	department, err := service.Create(context.Background(), "  Engineering  ")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.createdName != "Engineering" || department.Name != "Engineering" {
		t.Fatalf("department = %#v, repository name = %q", department, repository.createdName)
	}
}

func TestCreateDepartmentRejectsShortName(t *testing.T) {
	service := NewDepartmentService(&departmentRepositoryStub{})

	_, err := service.Create(context.Background(), "A")
	var validation *apperror.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Create() error = %v", err)
	}
}
