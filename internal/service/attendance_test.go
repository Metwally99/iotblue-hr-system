package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type attendanceRepositoryStub struct {
	employeeID string
	date       time.Time
	instant    time.Time
	filter     model.AttendanceFilter
}

func (s *attendanceRepositoryStub) CheckIn(_ context.Context, employeeID string, date, instant time.Time) (model.AttendanceRecord, error) {
	s.employeeID, s.date, s.instant = employeeID, date, instant
	return model.AttendanceRecord{EmployeeID: employeeID, AttendanceDate: date, CheckIn: instant}, nil
}
func (s *attendanceRepositoryStub) CheckOut(_ context.Context, employeeID string, date, instant time.Time) (model.AttendanceRecord, error) {
	s.employeeID, s.date, s.instant = employeeID, date, instant
	return model.AttendanceRecord{EmployeeID: employeeID, AttendanceDate: date, CheckOut: &instant}, nil
}
func (s *attendanceRepositoryStub) List(_ context.Context, filter model.AttendanceFilter) (model.AttendancePage, error) {
	s.filter = filter
	return model.AttendancePage{Page: filter.Page, PageSize: filter.PageSize}, nil
}
func (s *attendanceRepositoryStub) Report(context.Context, model.AttendanceFilter) ([]model.AttendanceReportRow, error) {
	return nil, nil
}

func TestCheckInUsesCompanyDateAndUTCInstant(t *testing.T) {
	location, err := time.LoadLocation("Africa/Cairo")
	if err != nil {
		t.Fatal(err)
	}
	repository := &attendanceRepositoryStub{}
	service := NewAttendanceService(repository, location)
	service.now = func() time.Time { return time.Date(2026, 8, 23, 22, 30, 0, 0, time.UTC) }

	_, err = service.CheckIn(context.Background(), "employee-id")
	if err != nil {
		t.Fatalf("CheckIn() error = %v", err)
	}
	if repository.date.Format(time.DateOnly) != "2026-08-24" {
		t.Fatalf("attendance date = %s", repository.date.Format(time.DateOnly))
	}
	if !repository.instant.Equal(time.Date(2026, 8, 23, 22, 30, 0, 0, time.UTC)) {
		t.Fatalf("stored instant = %v", repository.instant)
	}
}

func TestListMineOverridesEmployeeIDAndCapsPageSize(t *testing.T) {
	repository := &attendanceRepositoryStub{}
	service := NewAttendanceService(repository, time.UTC)
	actor := model.Identity{EmployeeID: "6085a1c8-842f-427b-b05c-e7f8f4366b6f", Role: model.RoleEmployee}

	_, err := service.List(context.Background(), actor, true, model.AttendanceFilter{PageSize: 1000})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repository.filter.EmployeeID == nil || *repository.filter.EmployeeID != actor.EmployeeID || repository.filter.PageSize != 100 {
		t.Fatalf("filter = %#v", repository.filter)
	}
}

func TestReportRequiresValidDateRange(t *testing.T) {
	service := NewAttendanceService(&attendanceRepositoryStub{}, time.UTC)
	end := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	_, err := service.Report(context.Background(), model.AttendanceFilter{StartDate: &start, EndDate: &end})
	var validation *apperror.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Report() error = %v", err)
	}
}
