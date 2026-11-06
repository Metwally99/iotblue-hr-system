package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type LeaveRepository interface {
	Create(context.Context, string, time.Time, time.Time, model.LeaveType) (model.LeaveRequest, error)
	GetByID(context.Context, string) (model.LeaveRequest, error)
	List(context.Context, model.LeaveFilter) (model.LeavePage, error)
	GetBalance(context.Context, string, int, model.LeaveType) (model.LeaveBalance, error)
	Decide(context.Context, string, string, model.Role, bool, *string) (model.LeaveRequest, *model.LeaveBalance, error)
}

type LeaveService struct {
	repository LeaveRepository
	now        func() time.Time
}

func NewLeaveService(repository LeaveRepository) *LeaveService {
	return &LeaveService{repository: repository, now: time.Now}
}

func (s *LeaveService) Submit(ctx context.Context, employeeID string, start, end time.Time, leaveType model.LeaveType) (model.LeaveRequest, error) {
	fields := make(map[string]string)
	today := dateOnly(s.now())
	start = dateOnly(start)
	end = dateOnly(end)
	if start.Before(today) {
		fields["start_date"] = "start_date cannot be in the past"
	}
	if end.Before(start) {
		fields["end_date"] = "end_date cannot be before start_date"
	}
	if start.Year() != end.Year() {
		fields["date_range"] = "a leave request must stay within one calendar year"
	}
	if start.Year() != today.Year() {
		fields["start_date"] = "version 1 accepts leave requests for the current calendar year"
	}
	if !leaveType.Valid() {
		fields["leave_type"] = "leave_type must be annual, sick, or holiday"
	}
	if len(fields) > 0 {
		return model.LeaveRequest{}, &apperror.ValidationError{Fields: fields}
	}
	return s.repository.Create(ctx, employeeID, start, end, leaveType)
}

func (s *LeaveService) Get(ctx context.Context, actor model.Identity, id string) (model.LeaveRequest, error) {
	if !validID(id) {
		return model.LeaveRequest{}, &apperror.ValidationError{Fields: map[string]string{"id": "leave request ID must be a UUID"}}
	}
	request, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return model.LeaveRequest{}, err
	}
	if !isHRRole(actor.Role) && request.EmployeeID != actor.EmployeeID {
		return model.LeaveRequest{}, apperror.ErrForbidden
	}
	return request, nil
}

func (s *LeaveService) List(ctx context.Context, actor model.Identity, ownOnly bool, filter model.LeaveFilter) (model.LeavePage, error) {
	if ownOnly || !isHRRole(actor.Role) {
		filter.EmployeeID = &actor.EmployeeID
		if actor.Role == model.RoleManager && !ownOnly {
			filter.EmployeeID = nil
			filter.ManagerID = &actor.EmployeeID
		}
	}
	if filter.EmployeeID != nil && !validID(*filter.EmployeeID) {
		return model.LeavePage{}, &apperror.ValidationError{Fields: map[string]string{"employee_id": "employee_id must be a UUID"}}
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

func (s *LeaveService) Balance(ctx context.Context, actor model.Identity, employeeID string, year int) (model.LeaveBalance, error) {
	if !isHRRole(actor.Role) && actor.EmployeeID != employeeID {
		return model.LeaveBalance{}, apperror.ErrForbidden
	}
	if !validID(employeeID) {
		return model.LeaveBalance{}, &apperror.ValidationError{Fields: map[string]string{"employee_id": "employee_id must be a UUID"}}
	}
	if year < 2000 || year > 2200 {
		return model.LeaveBalance{}, &apperror.ValidationError{Fields: map[string]string{"year": "year must be between 2000 and 2200"}}
	}
	return s.repository.GetBalance(ctx, employeeID, year, model.LeaveAnnual)
}

func (s *LeaveService) Decide(ctx context.Context, reviewer model.Identity, requestID string, approve bool, note *string) (model.LeaveRequest, *model.LeaveBalance, error) {
	if !isHRRole(reviewer.Role) && reviewer.Role != model.RoleManager {
		return model.LeaveRequest{}, nil, apperror.ErrForbidden
	}
	if !validID(requestID) {
		return model.LeaveRequest{}, nil, &apperror.ValidationError{Fields: map[string]string{"id": "leave request ID must be a UUID"}}
	}
	if note != nil {
		trimmed := strings.TrimSpace(*note)
		note = &trimmed
		if utf8.RuneCountInString(trimmed) > 500 {
			return model.LeaveRequest{}, nil, &apperror.ValidationError{Fields: map[string]string{"review_note": "review_note cannot exceed 500 characters"}}
		}
	}
	return s.repository.Decide(ctx, requestID, reviewer.EmployeeID, reviewer.Role, approve, note)
}

func isHRRole(role model.Role) bool { return role == model.RoleAdmin || role == model.RoleHR }

func InclusiveLeaveDays(start, end time.Time) int {
	start, end = dateOnly(start), dateOnly(end)
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
