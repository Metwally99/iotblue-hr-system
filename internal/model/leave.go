package model

import "time"

type LeaveStatus string

const (
	LeavePending  LeaveStatus = "pending"
	LeaveApproved LeaveStatus = "approved"
	LeaveRejected LeaveStatus = "rejected"
)

func (s LeaveStatus) Valid() bool {
	return s == LeavePending || s == LeaveApproved || s == LeaveRejected
}

type LeaveType string

const (
	LeaveAnnual  LeaveType = "annual"
	LeaveSick    LeaveType = "sick"
	LeaveHoliday LeaveType = "holiday"
)

func (t LeaveType) Valid() bool {
	return t == LeaveAnnual || t == LeaveSick || t == LeaveHoliday
}

type LeaveRequest struct {
	ID           string
	EmployeeID   string
	EmployeeName string
	StartDate    time.Time
	EndDate      time.Time
	LeaveType    LeaveType
	Status       LeaveStatus
	ReviewedBy   *string
	ReviewedAt   *time.Time
	ReviewNote   *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type LeaveBalance struct {
	ID         string
	EmployeeID string
	LeaveYear  int
	LeaveType  LeaveType
	TotalDays  int
	UsedDays   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (b LeaveBalance) AvailableDays() int { return b.TotalDays - b.UsedDays }

type LeaveFilter struct {
	EmployeeID *string
	ManagerID  *string
	Status     *LeaveStatus
	Year       *int
	Search     string
	Page       int
	PageSize   int
}

type LeavePage struct {
	Requests   []LeaveRequest
	Page       int
	PageSize   int
	TotalItems int
}
