package model

import "time"

type AttendanceRecord struct {
	ID             string
	EmployeeID     string
	EmployeeName   string
	DepartmentID   string
	DepartmentName string
	AttendanceDate time.Time
	CheckIn        time.Time
	CheckOut       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AttendanceFilter struct {
	EmployeeID   *string
	DepartmentID *string
	StartDate    *time.Time
	EndDate      *time.Time
	Search       string
	Page         int
	PageSize     int
}

type AttendancePage struct {
	Records    []AttendanceRecord
	Page       int
	PageSize   int
	TotalItems int
}

type AttendanceReportRow struct {
	EmployeeID     string
	EmployeeName   string
	DepartmentID   string
	DepartmentName string
	PresentDays    int
	CompletedDays  int
	TotalHours     float64
}
