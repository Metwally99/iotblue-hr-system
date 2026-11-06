package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type DepartmentService interface {
	Create(context.Context, string) (model.Department, error)
	GetByID(context.Context, string) (model.Department, error)
	List(context.Context) ([]model.Department, error)
	Update(context.Context, string, string) (model.Department, error)
}

type DepartmentHandler struct{ service DepartmentService }

func NewDepartmentHandler(service DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

type departmentRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *DepartmentHandler) Create(c *gin.Context) {
	var request departmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, validationError(map[string]string{"name": "name is required"}))
		return
	}
	department, err := h.service.Create(c.Request.Context(), request.Name)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": departmentJSON(department)})
}

func (h *DepartmentHandler) Get(c *gin.Context) {
	department, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": departmentJSON(department)})
}

func (h *DepartmentHandler) List(c *gin.Context) {
	departments, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	data := make([]gin.H, 0, len(departments))
	for _, department := range departments {
		data = append(data, departmentJSON(department))
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *DepartmentHandler) Update(c *gin.Context) {
	var request departmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, validationError(map[string]string{"name": "name is required"}))
		return
	}
	department, err := h.service.Update(c.Request.Context(), c.Param("id"), request.Name)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": departmentJSON(department)})
}

func departmentJSON(department model.Department) gin.H {
	return gin.H{"id": department.ID, "name": department.Name, "created_at": department.CreatedAt, "updated_at": department.UpdatedAt}
}
