package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
	"github.com/mahmoudmetwally/hr-management-system/internal/repository"
	"github.com/mahmoudmetwally/hr-management-system/internal/service"
)

func TestEmployeeLeaveAndAttendanceWorkflow(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `TRUNCATE attendance_records, leave_requests, leave_balances, employees, departments RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}

	departments := repository.NewDepartmentRepository(pool)
	department, err := departments.Create(ctx, "Engineering")
	if err != nil {
		t.Fatal(err)
	}
	employees := repository.NewEmployeeRepository(pool)
	employeeService := service.NewEmployeeService(employees)
	admin, err := employeeService.Create(ctx, service.CreateEmployeeInput{
		DepartmentID: department.ID, Name: "Admin User", Email: "admin@example.com",
		Password: "AdminPassword123!", Role: model.RoleAdmin, TotalLeaveDays: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	employee, err := employeeService.Create(ctx, service.CreateEmployeeInput{
		DepartmentID: department.ID, ManagerID: &admin.ID, Name: "Regular Employee", Email: "employee@example.com",
		Password: "EmployeePassword123!", Role: model.RoleEmployee, TotalLeaveDays: 20,
	})
	if err != nil {
		t.Fatal(err)
	}

	leaves := repository.NewLeaveRepository(pool)
	today := time.Now().UTC()
	start := time.Date(today.Year(), time.June, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(today.Year(), time.June, 12, 0, 0, 0, 0, time.UTC)
	request, err := leaves.Create(ctx, employee.ID, start, end, model.LeaveAnnual)
	if err != nil {
		t.Fatal(err)
	}
	approved, balance, err := leaves.Decide(ctx, request.ID, admin.ID, model.RoleAdmin, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if approved.Status != model.LeaveApproved || balance == nil || balance.UsedDays != 3 || balance.AvailableDays() != 17 {
		t.Fatalf("approved = %#v, balance = %#v", approved, balance)
	}

	attendance := repository.NewAttendanceRepository(pool)
	attendanceDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	checkIn := time.Now().UTC()
	record, err := attendance.CheckIn(ctx, employee.ID, attendanceDate, checkIn)
	if err != nil {
		t.Fatal(err)
	}
	record, err = attendance.CheckOut(ctx, employee.ID, attendanceDate, checkIn.Add(8*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if record.CheckOut == nil {
		t.Fatal("expected completed attendance record")
	}
}
