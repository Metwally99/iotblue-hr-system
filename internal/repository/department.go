package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type DepartmentRepository struct{ pool *pgxpool.Pool }

func NewDepartmentRepository(pool *pgxpool.Pool) *DepartmentRepository {
	return &DepartmentRepository{pool: pool}
}

func (r *DepartmentRepository) Create(ctx context.Context, name string) (model.Department, error) {
	var result model.Department
	err := r.pool.QueryRow(ctx, `
		INSERT INTO departments (name) VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`, name).Scan(&result.ID, &result.Name, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		return model.Department{}, fmt.Errorf("create department: %w", translatePostgresError(err))
	}
	return result, nil
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id string) (model.Department, error) {
	var result model.Department
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, created_at, updated_at FROM departments WHERE id = $1
	`, id).Scan(&result.ID, &result.Name, &result.CreatedAt, &result.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Department{}, apperror.ErrNotFound
	}
	if err != nil {
		return model.Department{}, fmt.Errorf("get department: %w", err)
	}
	return result, nil
}

func (r *DepartmentRepository) List(ctx context.Context) ([]model.Department, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, created_at, updated_at FROM departments ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	defer rows.Close()

	departments := make([]model.Department, 0)
	for rows.Next() {
		var department model.Department
		if err := rows.Scan(&department.ID, &department.Name, &department.CreatedAt, &department.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		departments = append(departments, department)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate departments: %w", err)
	}
	return departments, nil
}

func (r *DepartmentRepository) Update(ctx context.Context, id, name string) (model.Department, error) {
	var result model.Department
	err := r.pool.QueryRow(ctx, `
		UPDATE departments SET name = $2, updated_at = now() WHERE id = $1
		RETURNING id, name, created_at, updated_at
	`, id, name).Scan(&result.ID, &result.Name, &result.CreatedAt, &result.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Department{}, apperror.ErrNotFound
	}
	if err != nil {
		return model.Department{}, fmt.Errorf("update department: %w", translatePostgresError(err))
	}
	return result, nil
}
