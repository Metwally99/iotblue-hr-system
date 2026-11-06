package service

import (
	"context"
	"strings"
	"time"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type AttendanceRepository interface {
	CheckIn(context.Context, string, time.Time, time.Time) (model.AttendanceRecord, error)
	CheckOut(context.Context, string, time.Time, time.Time) (model.AttendanceRecord, error)
	List(context.Context, model.AttendanceFilter) (model.AttendancePage, error)
	Report(context.Context, model.AttendanceFilter) ([]model.AttendanceReportRow, error)
}

type AttendanceService struct {
	repository AttendanceRepository
	location   *time.Location
	now        func() time.Time
}

func NewAttendanceService(repository AttendanceRepository, location *time.Location) *AttendanceService {
	return &AttendanceService{repository: repository, location: location, now: time.Now}
}

func (s *AttendanceService) CheckIn(ctx context.Context, employeeID string) (model.AttendanceRecord, error) {
	now := s.now().In(s.location)
	return s.repository.CheckIn(ctx, employeeID, dateOnly(now), now.UTC())
}

func (s *AttendanceService) CheckOut(ctx context.Context, employeeID string) (model.AttendanceRecord, error) {
	now := s.now().In(s.location)
	return s.repository.CheckOut(ctx, employeeID, dateOnly(now), now.UTC())
}

func (s *AttendanceService) List(ctx context.Context, actor model.Identity, ownOnly bool, filter model.AttendanceFilter) (model.AttendancePage, error) {
	if ownOnly || (actor.Role != model.RoleAdmin && actor.Role != model.RoleHR) {
		filter.EmployeeID = &actor.EmployeeID
	}
	if err := validateAttendanceFilter(filter, false); err != nil {
		return model.AttendancePage{}, err
	}
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
	return s.repository.List(ctx, filter)
}

func (s *AttendanceService) Report(ctx context.Context, filter model.AttendanceFilter) ([]model.AttendanceReportRow, error) {
	if err := validateAttendanceFilter(filter, true); err != nil {
		return nil, err
	}
	return s.repository.Report(ctx, filter)
}

func validateAttendanceFilter(filter model.AttendanceFilter, requireRange bool) error {
	fields := make(map[string]string)
	if filter.EmployeeID != nil && !validID(*filter.EmployeeID) {
		fields["employee_id"] = "employee_id must be a UUID"
	}
	if filter.DepartmentID != nil && !validID(*filter.DepartmentID) {
		fields["department_id"] = "department_id must be a UUID"
	}
	if requireRange && (filter.StartDate == nil || filter.EndDate == nil) {
		fields["date_range"] = "start_date and end_date are required"
	}
	if filter.StartDate != nil && filter.EndDate != nil {
		start, end := dateOnly(*filter.StartDate), dateOnly(*filter.EndDate)
		if end.Before(start) {
			fields["end_date"] = "end_date cannot be before start_date"
		}
		if end.Sub(start) > 366*24*time.Hour {
			fields["date_range"] = "date range cannot exceed 366 days"
		}
	}
	if len(fields) > 0 {
		return &apperror.ValidationError{Fields: fields}
	}
	return nil
}
