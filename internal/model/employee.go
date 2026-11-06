package model

import "time"

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleHR       Role = "hr"
	RoleManager  Role = "manager"
	RoleEmployee Role = "employee"
)

func (r Role) Valid() bool {
	return r == RoleAdmin || r == RoleHR || r == RoleManager || r == RoleEmployee
}

type Employee struct {
	ID           string
	DepartmentID string
	ManagerID    *string
	Name         string
	Email        string
	PasswordHash string
	Role         Role
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Identity struct {
	EmployeeID string
	Name       string
	Email      string
	Role       Role
}

type CreateEmployeeParams struct {
	DepartmentID   string
	ManagerID      *string
	Name           string
	Email          string
	PasswordHash   string
	Role           Role
	LeaveYear      int
	TotalLeaveDays int
}

type UpdateEmployeeParams struct {
	ID           string
	DepartmentID *string
	ManagerID    *string
	Name         *string
	Email        *string
	Role         *Role
}

type EmployeeFilter struct {
	Page         int
	PageSize     int
	Search       string
	DepartmentID *string
	Role         *Role
	IsActive     *bool
}

type EmployeePage struct {
	Employees  []Employee
	Page       int
	PageSize   int
	TotalItems int
}

type BootstrapAdminParams struct {
	DepartmentName string
	Name           string
	Email          string
	PasswordHash   string
	LeaveYear      int
	TotalLeaveDays int
}
