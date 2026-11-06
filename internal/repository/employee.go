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

type EmployeeRepository struct {
	pool *pgxpool.Pool
}

func NewEmployeeRepository(pool *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{pool: pool}
}

func (r *EmployeeRepository) FindByEmail(ctx context.Context, email string) (model.Employee, error) {
	return r.findOne(ctx, `
		SELECT id, department_id, manager_id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM employees
		WHERE email = $1
	`, email)
}

func (r *EmployeeRepository) FindByID(ctx context.Context, id string) (model.Employee, error) {
	return r.findOne(ctx, `
		SELECT id, department_id, manager_id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM employees
		WHERE id = $1
	`, id)
}

func (r *EmployeeRepository) Create(ctx context.Context, params model.CreateEmployeeParams) (model.Employee, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Employee{}, fmt.Errorf("begin employee creation: %w", err)
	}
	defer tx.Rollback(ctx)

	var employee model.Employee
	err = tx.QueryRow(ctx, `
		INSERT INTO employees (department_id, manager_id, name, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, department_id, manager_id, name, email, password_hash, role, is_active, created_at, updated_at
	`, params.DepartmentID, params.ManagerID, params.Name, params.Email, params.PasswordHash, params.Role).Scan(
		&employee.ID, &employee.DepartmentID, &employee.ManagerID, &employee.Name, &employee.Email,
		&employee.PasswordHash, &employee.Role, &employee.IsActive, &employee.CreatedAt, &employee.UpdatedAt,
	)
	if err != nil {
		return model.Employee{}, fmt.Errorf("insert employee: %w", translatePostgresError(err))
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO leave_balances (employee_id, leave_year, total_days)
		VALUES ($1, $2, $3)
	`, employee.ID, params.LeaveYear, params.TotalLeaveDays)
	if err != nil {
		return model.Employee{}, fmt.Errorf("create initial leave balance: %w", translatePostgresError(err))
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Employee{}, fmt.Errorf("commit employee creation: %w", err)
	}
	return employee, nil
}

func (r *EmployeeRepository) List(ctx context.Context, filter model.EmployeeFilter) (model.EmployeePage, error) {
	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, department_id, manager_id, name, email, password_hash, role, is_active, created_at, updated_at,
		       count(*) OVER()
		FROM employees
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
		  AND ($2::uuid IS NULL OR department_id = $2)
		  AND ($3::text IS NULL OR role = $3)
		  AND ($4::boolean IS NULL OR is_active = $4)
		ORDER BY name, id
		LIMIT $5 OFFSET $6
	`, filter.Search, filter.DepartmentID, filter.Role, filter.IsActive, filter.PageSize, offset)
	if err != nil {
		return model.EmployeePage{}, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	page := model.EmployeePage{Employees: make([]model.Employee, 0), Page: filter.Page, PageSize: filter.PageSize}
	for rows.Next() {
		var employee model.Employee
		if err := rows.Scan(
			&employee.ID, &employee.DepartmentID, &employee.ManagerID, &employee.Name, &employee.Email,
			&employee.PasswordHash, &employee.Role, &employee.IsActive, &employee.CreatedAt, &employee.UpdatedAt,
			&page.TotalItems,
		); err != nil {
			return model.EmployeePage{}, fmt.Errorf("scan employee list: %w", err)
		}
		page.Employees = append(page.Employees, employee)
	}
	if err := rows.Err(); err != nil {
		return model.EmployeePage{}, fmt.Errorf("iterate employees: %w", err)
	}
	return page, nil
}

func (r *EmployeeRepository) Update(ctx context.Context, params model.UpdateEmployeeParams) (model.Employee, error) {
	var employee model.Employee
	err := r.pool.QueryRow(ctx, `
		UPDATE employees SET
			department_id = COALESCE($2, department_id),
			manager_id = COALESCE($3, manager_id),
			name = COALESCE($4, name),
			email = COALESCE($5, email),
			role = COALESCE($6, role),
			updated_at = now()
		WHERE id = $1
		RETURNING id, department_id, manager_id, name, email, password_hash, role, is_active, created_at, updated_at
	`, params.ID, params.DepartmentID, params.ManagerID, params.Name, params.Email, params.Role).Scan(
		&employee.ID, &employee.DepartmentID, &employee.ManagerID, &employee.Name, &employee.Email,
		&employee.PasswordHash, &employee.Role, &employee.IsActive, &employee.CreatedAt, &employee.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Employee{}, apperror.ErrNotFound
	}
	if err != nil {
		return model.Employee{}, fmt.Errorf("update employee: %w", translatePostgresError(err))
	}
	return employee, nil
}

func (r *EmployeeRepository) Deactivate(ctx context.Context, id string) (model.Employee, error) {
	var employee model.Employee
	err := r.pool.QueryRow(ctx, `
		UPDATE employees SET is_active = false, updated_at = now()
		WHERE id = $1 AND is_active = true
		RETURNING id, department_id, manager_id, name, email, password_hash, role, is_active, created_at, updated_at
	`, id).Scan(
		&employee.ID, &employee.DepartmentID, &employee.ManagerID, &employee.Name, &employee.Email,
		&employee.PasswordHash, &employee.Role, &employee.IsActive, &employee.CreatedAt, &employee.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Employee{}, apperror.ErrNotFound
	}
	if err != nil {
		return model.Employee{}, fmt.Errorf("deactivate employee: %w", err)
	}
	return employee, nil
}

func (r *EmployeeRepository) BootstrapAdmin(ctx context.Context, params model.BootstrapAdminParams) (bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return false, fmt.Errorf("begin admin bootstrap: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('hr_admin_bootstrap'))`); err != nil {
		return false, fmt.Errorf("lock admin bootstrap: %w", err)
	}
	var employeeCount int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM employees`).Scan(&employeeCount); err != nil {
		return false, fmt.Errorf("count employees: %w", err)
	}
	if employeeCount > 0 {
		return false, tx.Commit(ctx)
	}
	var departmentID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO departments (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET updated_at = now()
		RETURNING id
	`, params.DepartmentName).Scan(&departmentID); err != nil {
		return false, fmt.Errorf("create bootstrap department: %w", err)
	}
	var employeeID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO employees (department_id, name, email, password_hash, role)
		VALUES ($1, $2, $3, $4, 'admin') RETURNING id
	`, departmentID, params.Name, params.Email, params.PasswordHash).Scan(&employeeID); err != nil {
		return false, fmt.Errorf("create bootstrap admin: %w", translatePostgresError(err))
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO leave_balances (employee_id, leave_year, total_days) VALUES ($1, $2, $3)
	`, employeeID, params.LeaveYear, params.TotalLeaveDays); err != nil {
		return false, fmt.Errorf("create bootstrap leave balance: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit admin bootstrap: %w", err)
	}
	return true, nil
}

func (r *EmployeeRepository) findOne(ctx context.Context, query string, value string) (model.Employee, error) {
	var employee model.Employee
	err := r.pool.QueryRow(ctx, query, value).Scan(
		&employee.ID,
		&employee.DepartmentID,
		&employee.ManagerID,
		&employee.Name,
		&employee.Email,
		&employee.PasswordHash,
		&employee.Role,
		&employee.IsActive,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Employee{}, apperror.ErrNotFound
	}
	if err != nil {
		return model.Employee{}, fmt.Errorf("find employee: %w", err)
	}
	return employee, nil
}
