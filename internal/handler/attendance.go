package handler

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type AttendanceService interface {
	CheckIn(context.Context, string) (model.AttendanceRecord, error)
	CheckOut(context.Context, string) (model.AttendanceRecord, error)
	List(context.Context, model.Identity, bool, model.AttendanceFilter) (model.AttendancePage, error)
	Report(context.Context, model.AttendanceFilter) ([]model.AttendanceReportRow, error)
}

type AttendanceHandler struct{ service AttendanceService }

func NewAttendanceHandler(service AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{service: service}
}

func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	identity, _ := CurrentIdentity(c)
	record, err := h.service.CheckIn(c.Request.Context(), identity.EmployeeID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": attendanceJSON(record)})
}

func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	identity, _ := CurrentIdentity(c)
	record, err := h.service.CheckOut(c.Request.Context(), identity.EmployeeID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": attendanceJSON(record)})
}

func (h *AttendanceHandler) ListMine(c *gin.Context) { h.list(c, true) }
func (h *AttendanceHandler) ListAll(c *gin.Context)  { h.list(c, false) }

func (h *AttendanceHandler) list(c *gin.Context, ownOnly bool) {
	identity, _ := CurrentIdentity(c)
	filter, err := attendanceFilter(c, false)
	if err != nil {
		respondError(c, err)
		return
	}
	page, err := h.service.List(c.Request.Context(), identity, ownOnly, filter)
	if err != nil {
		respondError(c, err)
		return
	}
	data := make([]gin.H, 0, len(page.Records))
	for _, record := range page.Records {
		data = append(data, attendanceJSON(record))
	}
	totalPages := 0
	if page.TotalItems > 0 {
		totalPages = int(math.Ceil(float64(page.TotalItems) / float64(page.PageSize)))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": gin.H{
		"page": page.Page, "page_size": page.PageSize, "total_items": page.TotalItems, "total_pages": totalPages,
	}})
}

func (h *AttendanceHandler) Report(c *gin.Context) {
	filter, err := attendanceFilter(c, true)
	if err != nil {
		respondError(c, err)
		return
	}
	rows, err := h.service.Report(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	data := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		data = append(data, gin.H{
			"employee_id": row.EmployeeID, "employee_name": row.EmployeeName,
			"department_id": row.DepartmentID, "department_name": row.DepartmentName,
			"present_days": row.PresentDays, "completed_days": row.CompletedDays, "total_hours": row.TotalHours,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": gin.H{
		"start_date": filter.StartDate.Format(time.DateOnly), "end_date": filter.EndDate.Format(time.DateOnly),
	}})
}

func attendanceFilter(c *gin.Context, requireRange bool) (model.AttendanceFilter, error) {
	page, err := positiveQueryInt(c, "page", 1)
	if err != nil {
		return model.AttendanceFilter{}, err
	}
	pageSize, err := positiveQueryInt(c, "page_size", 20)
	if err != nil {
		return model.AttendanceFilter{}, err
	}
	filter := model.AttendanceFilter{Page: page, PageSize: pageSize, Search: c.Query("search")}
	if value := c.Query("employee_id"); value != "" {
		filter.EmployeeID = &value
	}
	if value := c.Query("department_id"); value != "" {
		filter.DepartmentID = &value
	}
	start, err := optionalDateQuery(c, "start_date")
	if err != nil {
		return filter, err
	}
	end, err := optionalDateQuery(c, "end_date")
	if err != nil {
		return filter, err
	}
	filter.StartDate, filter.EndDate = start, end
	if requireRange && (start == nil || end == nil) {
		return filter, validationError(map[string]string{"date_range": "start_date and end_date are required"})
	}
	return filter, nil
}

func optionalDateQuery(c *gin.Context, name string) (*time.Time, error) {
	value := c.Query(name)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, validationError(map[string]string{name: name + " must use YYYY-MM-DD"})
	}
	return &parsed, nil
}

func attendanceJSON(record model.AttendanceRecord) gin.H {
	return gin.H{
		"id": record.ID, "employee_id": record.EmployeeID, "employee_name": record.EmployeeName,
		"department_id": record.DepartmentID, "department_name": record.DepartmentName,
		"attendance_date": record.AttendanceDate.Format(time.DateOnly), "check_in": record.CheckIn,
		"check_out": record.CheckOut, "created_at": record.CreatedAt, "updated_at": record.UpdatedAt,
	}
}
