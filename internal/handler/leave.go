package handler

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type LeaveService interface {
	Submit(context.Context, string, time.Time, time.Time, model.LeaveType) (model.LeaveRequest, error)
	Get(context.Context, model.Identity, string) (model.LeaveRequest, error)
	List(context.Context, model.Identity, bool, model.LeaveFilter) (model.LeavePage, error)
	Balance(context.Context, model.Identity, string, int) (model.LeaveBalance, error)
	Decide(context.Context, model.Identity, string, bool, *string) (model.LeaveRequest, *model.LeaveBalance, error)
}

type LeaveHandler struct{ service LeaveService }

func NewLeaveHandler(service LeaveService) *LeaveHandler { return &LeaveHandler{service: service} }

type submitLeaveRequest struct {
	StartDate string          `json:"start_date" binding:"required"`
	EndDate   string          `json:"end_date" binding:"required"`
	LeaveType model.LeaveType `json:"leave_type" binding:"required"`
}

type decisionRequest struct {
	ReviewNote *string `json:"review_note"`
}

func (h *LeaveHandler) Submit(c *gin.Context) {
	identity, ok := CurrentIdentity(c)
	if !ok {
		respondError(c, validationError(map[string]string{"identity": "missing authentication identity"}))
		return
	}
	var request submitLeaveRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, validationError(map[string]string{"body": "start_date, end_date, and leave_type are required"}))
		return
	}
	start, err := time.Parse(time.DateOnly, request.StartDate)
	if err != nil {
		respondError(c, validationError(map[string]string{"start_date": "start_date must use YYYY-MM-DD"}))
		return
	}
	end, err := time.Parse(time.DateOnly, request.EndDate)
	if err != nil {
		respondError(c, validationError(map[string]string{"end_date": "end_date must use YYYY-MM-DD"}))
		return
	}
	leave, err := h.service.Submit(c.Request.Context(), identity.EmployeeID, start, end, request.LeaveType)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": leaveJSON(leave)})
}

func (h *LeaveHandler) Get(c *gin.Context) {
	identity, _ := CurrentIdentity(c)
	leave, err := h.service.Get(c.Request.Context(), identity, c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": leaveJSON(leave)})
}

func (h *LeaveHandler) ListMine(c *gin.Context) { h.list(c, true) }
func (h *LeaveHandler) ListAll(c *gin.Context)  { h.list(c, false) }

func (h *LeaveHandler) list(c *gin.Context, ownOnly bool) {
	identity, _ := CurrentIdentity(c)
	filter, err := leaveFilter(c)
	if err != nil {
		respondError(c, err)
		return
	}
	page, err := h.service.List(c.Request.Context(), identity, ownOnly, filter)
	if err != nil {
		respondError(c, err)
		return
	}
	data := make([]gin.H, 0, len(page.Requests))
	for _, leave := range page.Requests {
		data = append(data, leaveJSON(leave))
	}
	totalPages := 0
	if page.TotalItems > 0 {
		totalPages = int(math.Ceil(float64(page.TotalItems) / float64(page.PageSize)))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": gin.H{"page": page.Page, "page_size": page.PageSize, "total_items": page.TotalItems, "total_pages": totalPages}})
}

func (h *LeaveHandler) BalanceMine(c *gin.Context) {
	identity, _ := CurrentIdentity(c)
	year, err := queryYear(c)
	if err != nil {
		respondError(c, err)
		return
	}
	h.respondBalance(c, identity, identity.EmployeeID, year)
}

func (h *LeaveHandler) BalanceForEmployee(c *gin.Context) {
	identity, _ := CurrentIdentity(c)
	year, err := queryYear(c)
	if err != nil {
		respondError(c, err)
		return
	}
	h.respondBalance(c, identity, c.Param("id"), year)
}

func (h *LeaveHandler) respondBalance(c *gin.Context, identity model.Identity, employeeID string, year int) {
	balance, err := h.service.Balance(c.Request.Context(), identity, employeeID, year)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": balanceJSON(balance)})
}

func (h *LeaveHandler) Approve(c *gin.Context) { h.decide(c, true) }
func (h *LeaveHandler) Reject(c *gin.Context)  { h.decide(c, false) }

func (h *LeaveHandler) decide(c *gin.Context, approve bool) {
	identity, _ := CurrentIdentity(c)
	var request decisionRequest
	if err := c.ShouldBindJSON(&request); err != nil && c.Request.ContentLength > 0 {
		respondError(c, validationError(map[string]string{"body": "body must be valid JSON"}))
		return
	}
	leave, balance, err := h.service.Decide(c.Request.Context(), identity, c.Param("id"), approve, request.ReviewNote)
	if err != nil {
		respondError(c, err)
		return
	}
	data := gin.H{"leave_request": leaveJSON(leave)}
	if balance != nil {
		data["leave_balance"] = balanceJSON(*balance)
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func leaveFilter(c *gin.Context) (model.LeaveFilter, error) {
	page, err := positiveQueryInt(c, "page", 1)
	if err != nil {
		return model.LeaveFilter{}, err
	}
	pageSize, err := positiveQueryInt(c, "page_size", 20)
	if err != nil {
		return model.LeaveFilter{}, err
	}
	filter := model.LeaveFilter{Page: page, PageSize: pageSize, Search: c.Query("search")}
	if value := c.Query("employee_id"); value != "" {
		filter.EmployeeID = &value
	}
	if value := c.Query("status"); value != "" {
		status := model.LeaveStatus(value)
		if !status.Valid() {
			return filter, validationError(map[string]string{"status": "status must be pending, approved, or rejected"})
		}
		filter.Status = &status
	}
	if value := c.Query("year"); value != "" {
		year, err := strconv.Atoi(value)
		if err != nil {
			return filter, validationError(map[string]string{"year": "year must be an integer"})
		}
		filter.Year = &year
	}
	return filter, nil
}

func queryYear(c *gin.Context) (int, error) {
	value := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, err := strconv.Atoi(value)
	if err != nil {
		return 0, validationError(map[string]string{"year": "year must be an integer"})
	}
	return year, nil
}

func leaveJSON(leave model.LeaveRequest) gin.H {
	return gin.H{"id": leave.ID, "employee_id": leave.EmployeeID, "employee_name": leave.EmployeeName,
		"start_date": leave.StartDate.Format(time.DateOnly), "end_date": leave.EndDate.Format(time.DateOnly),
		"leave_type": leave.LeaveType, "status": leave.Status, "reviewed_by": leave.ReviewedBy,
		"reviewed_at": leave.ReviewedAt, "review_note": leave.ReviewNote, "created_at": leave.CreatedAt, "updated_at": leave.UpdatedAt}
}

func balanceJSON(balance model.LeaveBalance) gin.H {
	return gin.H{"id": balance.ID, "employee_id": balance.EmployeeID, "leave_year": balance.LeaveYear, "leave_type": balance.LeaveType,
		"total_days": balance.TotalDays, "used_days": balance.UsedDays, "available_days": balance.AvailableDays()}
}
