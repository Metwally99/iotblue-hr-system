package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type AttendanceRepository struct{ pool *pgxpool.Pool }

func NewAttendanceRepository(pool *pgxpool.Pool) *AttendanceRepository {
	return &AttendanceRepository{pool: pool}
}

func (r *AttendanceRepository) CheckIn(ctx context.Context, employeeID string, attendanceDate, checkIn time.Time) (model.AttendanceRecord, error) {
	var record model.AttendanceRecord
	err := r.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO attendance_records (employee_id, attendance_date, check_in)
			VALUES ($1, $2, $3)
			RETURNING id, employee_id, attendance_date, check_in, check_out, created_at, updated_at
		)
		SELECT i.id, i.employee_id, e.name, d.id, d.name, i.attendance_date, i.check_in, i.check_out, i.created_at, i.updated_at
		FROM inserted i JOIN employees e ON e.id = i.employee_id JOIN departments d ON d.id = e.department_id
	`, employeeID, attendanceDate, checkIn).Scan(attendanceDestinations(&record)...)
	if err != nil {
		return model.AttendanceRecord{}, fmt.Errorf("check in: %w", translatePostgresError(err))
	}
	return record, nil
}

func (r *AttendanceRepository) CheckOut(ctx context.Context, employeeID string, attendanceDate, checkOut time.Time) (model.AttendanceRecord, error) {
	var record model.AttendanceRecord
	err := r.pool.QueryRow(ctx, `
		WITH updated AS (
			UPDATE attendance_records SET check_out = $3, updated_at = now()
			WHERE employee_id = $1 AND attendance_date = $2 AND check_out IS NULL
			RETURNING id, employee_id, attendance_date, check_in, check_out, created_at, updated_at
		)
		SELECT u.id, u.employee_id, e.name, d.id, d.name, u.attendance_date, u.check_in, u.check_out, u.created_at, u.updated_at
		FROM updated u JOIN employees e ON e.id = u.employee_id JOIN departments d ON d.id = e.department_id
	`, employeeID, attendanceDate, checkOut).Scan(attendanceDestinations(&record)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.AttendanceRecord{}, apperror.ErrInvalidState
	}
	if err != nil {
		return model.AttendanceRecord{}, fmt.Errorf("check out: %w", err)
	}
	return record, nil
}

func (r *AttendanceRepository) List(ctx context.Context, filter model.AttendanceFilter) (model.AttendancePage, error) {
	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.pool.Query(ctx, `
		SELECT ar.id, ar.employee_id, e.name, d.id, d.name, ar.attendance_date, ar.check_in, ar.check_out, ar.created_at, ar.updated_at,
		       count(*) OVER()
		FROM attendance_records ar
		JOIN employees e ON e.id = ar.employee_id
		JOIN departments d ON d.id = e.department_id
		WHERE ($1::uuid IS NULL OR ar.employee_id = $1)
		  AND ($2::uuid IS NULL OR e.department_id = $2)
		  AND ($3::date IS NULL OR ar.attendance_date >= $3)
		  AND ($4::date IS NULL OR ar.attendance_date <= $4)
		  AND ($5 = '' OR e.name ILIKE '%' || $5 || '%' OR e.email ILIKE '%' || $5 || '%')
		ORDER BY ar.attendance_date DESC, ar.check_in DESC, ar.id
		LIMIT $6 OFFSET $7
	`, filter.EmployeeID, filter.DepartmentID, filter.StartDate, filter.EndDate, filter.Search, filter.PageSize, offset)
	if err != nil {
		return model.AttendancePage{}, fmt.Errorf("list attendance: %w", err)
	}
	defer rows.Close()
	page := model.AttendancePage{Records: make([]model.AttendanceRecord, 0), Page: filter.Page, PageSize: filter.PageSize}
	for rows.Next() {
		var record model.AttendanceRecord
		destinations := append(attendanceDestinations(&record), &page.TotalItems)
		if err := rows.Scan(destinations...); err != nil {
			return model.AttendancePage{}, fmt.Errorf("scan attendance: %w", err)
		}
		page.Records = append(page.Records, record)
	}
	if err := rows.Err(); err != nil {
		return model.AttendancePage{}, fmt.Errorf("iterate attendance: %w", err)
	}
	return page, nil
}

func (r *AttendanceRepository) Report(ctx context.Context, filter model.AttendanceFilter) ([]model.AttendanceReportRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.name, d.id, d.name,
		       count(*)::integer AS present_days,
		       count(ar.check_out)::integer AS completed_days,
		       COALESCE(round((sum(extract(epoch FROM (ar.check_out - ar.check_in))) FILTER (WHERE ar.check_out IS NOT NULL) / 3600.0)::numeric, 2), 0)::float8 AS total_hours
		FROM attendance_records ar
		JOIN employees e ON e.id = ar.employee_id
		JOIN departments d ON d.id = e.department_id
		WHERE ar.attendance_date BETWEEN $1 AND $2
		  AND ($3::uuid IS NULL OR e.department_id = $3)
		  AND ($4::uuid IS NULL OR e.id = $4)
		GROUP BY e.id, e.name, d.id, d.name
		ORDER BY d.name, e.name
	`, *filter.StartDate, *filter.EndDate, filter.DepartmentID, filter.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("attendance report: %w", err)
	}
	defer rows.Close()
	result := make([]model.AttendanceReportRow, 0)
	for rows.Next() {
		var row model.AttendanceReportRow
		if err := rows.Scan(&row.EmployeeID, &row.EmployeeName, &row.DepartmentID, &row.DepartmentName, &row.PresentDays, &row.CompletedDays, &row.TotalHours); err != nil {
			return nil, fmt.Errorf("scan attendance report: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attendance report: %w", err)
	}
	return result, nil
}

func attendanceDestinations(record *model.AttendanceRecord) []any {
	return []any{&record.ID, &record.EmployeeID, &record.EmployeeName, &record.DepartmentID, &record.DepartmentName,
		&record.AttendanceDate, &record.CheckIn, &record.CheckOut, &record.CreatedAt, &record.UpdatedAt}
}
