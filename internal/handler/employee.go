package handler

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
	"github.com/mahmoudmetwally/hr-management-system/internal/service"
)

type EmployeeService interface {
	Create(context.Context, service.CreateEmployeeInput) (model.Employee, error)
	GetByID(context.Context, string) (model.Employee, error)
	List(context.Context, model.EmployeeFilter) (model.EmployeePage, error)
	Update(context.Context, string, model.UpdateEmployeeParams) (model.Employee, error)
	Deactivate(context.Context, string, string) (model.Employee, error)
}

type EmployeeHandler struct{ service EmployeeService }

func NewEmployeeHandler(service EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

type createEmployeeRequest struct {
	DepartmentID   string     `json:"department_id" binding:"required"`
	ManagerID      *string    `json:"manager_id"`
	Name           string     `json:"name" binding:"required"`
	Email          string     `json:"email" binding:"required"`
	Password       string     `json:"password" binding:"required"`
	Role           model.Role `json:"role" binding:"required"`
	TotalLeaveDays int        `json:"total_leave_days" binding:"required"`
}

type updateEmployeeRequest struct {
	DepartmentID *string     `json:"department_id"`
	ManagerID    *string     `json:"manager_id"`
	Name         *string     `json:"name"`
	Email        *string     `json:"email"`
	Role         *model.Role `json:"role"`
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var request createEmployeeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, validationError(map[string]string{"body": "required employee fields are missing"}))
		return
	}
	employee, err := h.service.Create(c.Request.Context(), service.CreateEmployeeInput{
		DepartmentID: request.DepartmentID, ManagerID: request.ManagerID, Name: request.Name,
		Email: request.Email, Password: request.Password, Role: request.Role, TotalLeaveDays: request.TotalLeaveDays,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": publicEmployee(employee)})
}

func (h *EmployeeHandler) Get(c *gin.Context) {
	employee, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": publicEmployee(employee)})
}

func (h *EmployeeHandler) List(c *gin.Context) {
	filter, err := employeeFilter(c)
	if err != nil {
		respondError(c, err)
		return
	}
	page, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	data := make([]employeeResponse, 0, len(page.Employees))
	for _, employee := range page.Employees {
		data = append(data, publicEmployee(employee))
	}
	totalPages := 0
	if page.TotalItems > 0 {
		totalPages = int(math.Ceil(float64(page.TotalItems) / float64(page.PageSize)))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": gin.H{
		"page": page.Page, "page_size": page.PageSize, "total_items": page.TotalItems, "total_pages": totalPages,
	}})
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	var request updateEmployeeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, validationError(map[string]string{"body": "body must be valid JSON"}))
		return
	}
	identity, ok := CurrentIdentity(c)
	if !ok {
		respondError(c, validationError(map[string]string{"identity": "missing authentication identity"}))
		return
	}
	employee, err := h.service.Update(c.Request.Context(), identity.EmployeeID, model.UpdateEmployeeParams{
		ID: c.Param("id"), DepartmentID: request.DepartmentID, ManagerID: request.ManagerID,
		Name: request.Name, Email: request.Email, Role: request.Role,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": publicEmployee(employee)})
}

func (h *EmployeeHandler) Deactivate(c *gin.Context) {
	identity, ok := CurrentIdentity(c)
	if !ok {
		respondError(c, validationError(map[string]string{"identity": "missing authentication identity"}))
		return
	}
	employee, err := h.service.Deactivate(c.Request.Context(), identity.EmployeeID, c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": publicEmployee(employee)})
}

func employeeFilter(c *gin.Context) (model.EmployeeFilter, error) {
	page, err := positiveQueryInt(c, "page", 1)
	if err != nil {
		return model.EmployeeFilter{}, err
	}
	pageSize, err := positiveQueryInt(c, "page_size", 20)
	if err != nil {
		return model.EmployeeFilter{}, err
	}
	filter := model.EmployeeFilter{Page: page, PageSize: pageSize, Search: c.Query("search")}
	if value := c.Query("department_id"); value != "" {
		filter.DepartmentID = &value
	}
	if value := c.Query("role"); value != "" {
		role := model.Role(value)
		if !role.Valid() {
			return filter, validationError(map[string]string{"role": "role must be admin, hr, manager, or employee"})
		}
		filter.Role = &role
	}
	if value := c.Query("is_active"); value != "" {
		active, err := strconv.ParseBool(value)
		if err != nil {
			return filter, validationError(map[string]string{"is_active": "is_active must be true or false"})
		}
		filter.IsActive = &active
	}
	return filter, nil
}

func positiveQueryInt(c *gin.Context, name string, fallback int) (int, error) {
	value := c.Query(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, validationError(map[string]string{name: name + " must be a positive integer"})
	}
	return parsed, nil
}
