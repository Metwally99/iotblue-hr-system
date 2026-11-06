package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type leaveRepositoryStub struct {
	createdStart time.Time
	createdEnd   time.Time
	createdType  model.LeaveType
	request      model.LeaveRequest
	filter       model.LeaveFilter
}

func (s *leaveRepositoryStub) Create(_ context.Context, _ string, start, end time.Time, leaveType model.LeaveType) (model.LeaveRequest, error) {
	s.createdStart, s.createdEnd, s.createdType = start, end, leaveType
	return model.LeaveRequest{StartDate: start, EndDate: end, LeaveType: leaveType}, nil
}
func (s *leaveRepositoryStub) GetByID(context.Context, string) (model.LeaveRequest, error) {
	return s.request, nil
}
func (s *leaveRepositoryStub) List(_ context.Context, filter model.LeaveFilter) (model.LeavePage, error) {
	s.filter = filter
	return model.LeavePage{Page: filter.Page, PageSize: filter.PageSize}, nil
}
func (s *leaveRepositoryStub) GetBalance(context.Context, string, int, model.LeaveType) (model.LeaveBalance, error) {
	return model.LeaveBalance{TotalDays: 20, UsedDays: 7}, nil
}
func (s *leaveRepositoryStub) Decide(context.Context, string, string, model.Role, bool, *string) (model.LeaveRequest, *model.LeaveBalance, error) {
	return model.LeaveRequest{}, nil, nil
}

func TestSubmitLeaveNormalizesDatesAndType(t *testing.T) {
	repository := &leaveRepositoryStub{}
	service := NewLeaveService(repository)
	service.now = func() time.Time { return time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC) }

	_, err := service.Submit(context.Background(), "employee", time.Date(2026, 8, 24, 15, 0, 0, 0, time.Local), time.Date(2026, 8, 26, 20, 0, 0, 0, time.Local), model.LeaveSick)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if repository.createdStart.Hour() != 0 || repository.createdEnd.Hour() != 0 || repository.createdType != model.LeaveSick {
		t.Fatalf("stored values = %v, %v, %q", repository.createdStart, repository.createdEnd, repository.createdType)
	}
}

func TestSubmitLeaveRejectsPastAndReverseDates(t *testing.T) {
	service := NewLeaveService(&leaveRepositoryStub{})
	service.now = func() time.Time { return time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC) }
	_, err := service.Submit(context.Background(), "employee", time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), model.LeaveAnnual)
	var validation *apperror.ValidationError
	if !errors.As(err, &validation) || len(validation.Fields) < 2 {
		t.Fatalf("Submit() error = %#v", err)
	}
}

func TestEmployeeCannotReadAnotherEmployeesRequest(t *testing.T) {
	repository := &leaveRepositoryStub{request: model.LeaveRequest{ID: "6085a1c8-842f-427b-b05c-e7f8f4366b6f", EmployeeID: "other"}}
	service := NewLeaveService(repository)
	_, err := service.Get(context.Background(), model.Identity{EmployeeID: "self", Role: model.RoleEmployee}, repository.request.ID)
	if !errors.Is(err, apperror.ErrForbidden) {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestListMineOverridesEmployeeFilter(t *testing.T) {
	repository := &leaveRepositoryStub{}
	service := NewLeaveService(repository)
	actor := model.Identity{EmployeeID: "6085a1c8-842f-427b-b05c-e7f8f4366b6f", Role: model.RoleEmployee}
	_, err := service.List(context.Background(), actor, true, model.LeaveFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repository.filter.EmployeeID == nil || *repository.filter.EmployeeID != actor.EmployeeID {
		t.Fatalf("filter = %#v", repository.filter)
	}
}

func TestInclusiveLeaveDays(t *testing.T) {
	start := time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	if days := InclusiveLeaveDays(start, end); days != 3 {
		t.Fatalf("InclusiveLeaveDays() = %d, want 3", days)
	}
}

func TestAvailableLeaveIsDerived(t *testing.T) {
	balance := model.LeaveBalance{TotalDays: 20, UsedDays: 7}
	if balance.AvailableDays() != 13 {
		t.Fatalf("AvailableDays() = %d", balance.AvailableDays())
	}
}
