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

type LeaveRepository struct{ pool *pgxpool.Pool }

func NewLeaveRepository(pool *pgxpool.Pool) *LeaveRepository { return &LeaveRepository{pool: pool} }

func (r *LeaveRepository) Create(ctx context.Context, employeeID string, start, end time.Time, leaveType model.LeaveType) (model.LeaveRequest, error) {
	var request model.LeaveRequest
	err := r.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO leave_requests (employee_id, start_date, end_date, leave_type)
			VALUES ($1, $2, $3, $4)
			RETURNING id, employee_id, start_date, end_date, leave_type, status, reviewed_by, reviewed_at, review_note, created_at, updated_at
		)
		SELECT i.id, i.employee_id, e.name, i.start_date, i.end_date, i.leave_type, i.status,
		       i.reviewed_by, i.reviewed_at, i.review_note, i.created_at, i.updated_at
		FROM inserted i JOIN employees e ON e.id = i.employee_id
	`, employeeID, start, end, leaveType).Scan(leaveRequestDestinations(&request)...)
	if err != nil {
		return model.LeaveRequest{}, fmt.Errorf("create leave request: %w", translatePostgresError(err))
	}
	return request, nil
}

func (r *LeaveRepository) GetByID(ctx context.Context, id string) (model.LeaveRequest, error) {
	var request model.LeaveRequest
	err := r.pool.QueryRow(ctx, `
		SELECT lr.id, lr.employee_id, e.name, lr.start_date, lr.end_date, lr.leave_type, lr.status,
		       lr.reviewed_by, lr.reviewed_at, lr.review_note, lr.created_at, lr.updated_at
		FROM leave_requests lr JOIN employees e ON e.id = lr.employee_id
		WHERE lr.id = $1
	`, id).Scan(leaveRequestDestinations(&request)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.LeaveRequest{}, apperror.ErrNotFound
	}
	if err != nil {
		return model.LeaveRequest{}, fmt.Errorf("get leave request: %w", err)
	}
	return request, nil
}

func (r *LeaveRepository) List(ctx context.Context, filter model.LeaveFilter) (model.LeavePage, error) {
	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.pool.Query(ctx, `
		SELECT lr.id, lr.employee_id, e.name, lr.start_date, lr.end_date, lr.leave_type, lr.status,
		       lr.reviewed_by, lr.reviewed_at, lr.review_note, lr.created_at, lr.updated_at,
		       count(*) OVER()
		FROM leave_requests lr JOIN employees e ON e.id = lr.employee_id
		WHERE ($1::uuid IS NULL OR lr.employee_id = $1)
		  AND ($2::text IS NULL OR lr.status = $2)
		  AND ($3::integer IS NULL OR extract(year FROM lr.start_date) = $3)
		  AND ($4 = '' OR e.name ILIKE '%' || $4 || '%')
		  AND ($5::uuid IS NULL OR e.manager_id = $5)
		ORDER BY lr.created_at DESC, lr.id
		LIMIT $5 OFFSET $6
	`, filter.EmployeeID, filter.Status, filter.Year, filter.Search, filter.ManagerID, filter.PageSize, offset)
	if err != nil {
		return model.LeavePage{}, fmt.Errorf("list leave requests: %w", err)
	}
	defer rows.Close()
	page := model.LeavePage{Requests: make([]model.LeaveRequest, 0), Page: filter.Page, PageSize: filter.PageSize}
	for rows.Next() {
		var request model.LeaveRequest
		destinations := append(leaveRequestDestinations(&request), &page.TotalItems)
		if err := rows.Scan(destinations...); err != nil {
			return model.LeavePage{}, fmt.Errorf("scan leave request: %w", err)
		}
		page.Requests = append(page.Requests, request)
	}
	if err := rows.Err(); err != nil {
		return model.LeavePage{}, fmt.Errorf("iterate leave requests: %w", err)
	}
	return page, nil
}

func (r *LeaveRepository) GetBalance(ctx context.Context, employeeID string, year int, leaveType model.LeaveType) (model.LeaveBalance, error) {
	return getBalance(ctx, r.pool, employeeID, year, leaveType, false)
}

func (r *LeaveRepository) Decide(ctx context.Context, requestID, reviewerID string, reviewerRole model.Role, approve bool, note *string) (model.LeaveRequest, *model.LeaveBalance, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.LeaveRequest{}, nil, fmt.Errorf("begin leave decision: %w", err)
	}
	defer tx.Rollback(ctx)

	var request model.LeaveRequest
	err = tx.QueryRow(ctx, `
		SELECT lr.id, lr.employee_id, e.name, lr.start_date, lr.end_date, lr.leave_type, lr.status,
		       lr.reviewed_by, lr.reviewed_at, lr.review_note, lr.created_at, lr.updated_at
		FROM leave_requests lr JOIN employees e ON e.id = lr.employee_id
		WHERE lr.id = $1
		  AND ($3 IN ('admin', 'hr') OR ($3 = 'manager' AND e.manager_id = $2))
		FOR UPDATE OF lr
	`, requestID, reviewerID, reviewerRole).Scan(leaveRequestDestinations(&request)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.LeaveRequest{}, nil, apperror.ErrNotFound
	}
	if err != nil {
		return model.LeaveRequest{}, nil, fmt.Errorf("lock leave request: %w", err)
	}
	if request.Status != model.LeavePending {
		return model.LeaveRequest{}, nil, apperror.ErrInvalidState
	}

	status := model.LeaveRejected
	var balance *model.LeaveBalance
	if approve {
		status = model.LeaveApproved
		locked, err := getBalance(ctx, tx, request.EmployeeID, request.StartDate.Year(), request.LeaveType, true)
		if err != nil {
			return model.LeaveRequest{}, nil, err
		}
		requestedDays := int(request.EndDate.Sub(request.StartDate).Hours()/24) + 1
		if locked.AvailableDays() < requestedDays {
			return model.LeaveRequest{}, nil, apperror.ErrInsufficientLeave
		}
		if _, err := tx.Exec(ctx, `
			UPDATE leave_balances SET used_days = used_days + $2, updated_at = now() WHERE id = $1
		`, locked.ID, requestedDays); err != nil {
			return model.LeaveRequest{}, nil, fmt.Errorf("consume leave balance: %w", err)
		}
		locked.UsedDays += requestedDays
		balance = &locked
	}

	err = tx.QueryRow(ctx, `
		UPDATE leave_requests SET status = $2, reviewed_by = $3, reviewed_at = now(), review_note = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, employee_id, $5, start_date, end_date, leave_type, status, reviewed_by, reviewed_at, review_note, created_at, updated_at
	`, request.ID, status, reviewerID, note, request.EmployeeName).Scan(leaveRequestDestinations(&request)...)
	if err != nil {
		return model.LeaveRequest{}, nil, fmt.Errorf("record leave decision: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.LeaveRequest{}, nil, fmt.Errorf("commit leave decision: %w", err)
	}
	return request, balance, nil
}

type balanceQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func getBalance(ctx context.Context, querier balanceQuerier, employeeID string, year int, leaveType model.LeaveType, lock bool) (model.LeaveBalance, error) {
	query := `SELECT id, employee_id, leave_year, leave_type, total_days, used_days, created_at, updated_at FROM leave_balances WHERE employee_id = $1 AND leave_year = $2 AND leave_type = $3`
	if lock {
		query += ` FOR UPDATE`
	}
	var balance model.LeaveBalance
	err := querier.QueryRow(ctx, query, employeeID, year, leaveType).Scan(
		&balance.ID, &balance.EmployeeID, &balance.LeaveYear, &balance.LeaveType, &balance.TotalDays, &balance.UsedDays, &balance.CreatedAt, &balance.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.LeaveBalance{}, apperror.ErrNotFound
	}
	if err != nil {
		return model.LeaveBalance{}, fmt.Errorf("get leave balance: %w", err)
	}
	return balance, nil
}

func leaveRequestDestinations(request *model.LeaveRequest) []any {
	return []any{&request.ID, &request.EmployeeID, &request.EmployeeName, &request.StartDate, &request.EndDate,
		&request.LeaveType, &request.Status, &request.ReviewedBy, &request.ReviewedAt, &request.ReviewNote,
		&request.CreatedAt, &request.UpdatedAt}
}
