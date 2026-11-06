package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	appmiddleware "github.com/mahmoudmetwally/hr-management-system/internal/middleware"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
	"github.com/mahmoudmetwally/hr-management-system/internal/service"
)

type AuthService interface {
	Login(context.Context, string, string) (service.LoginResult, error)
	Authenticate(context.Context, string) (model.Identity, error)
}

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler { return &AuthHandler{service: service} }

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type employeeResponse struct {
	ID           string     `json:"id"`
	DepartmentID string     `json:"department_id"`
	ManagerID    *string    `json:"manager_id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Role         model.Role `json:"role"`
	IsActive     bool       `json:"is_active"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, validationError(map[string]string{"body": "email and password are required"}))
		return
	}
	result, err := h.service.Login(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"access_token": result.AccessToken,
		"token_type":   "Bearer",
		"expires_at":   result.ExpiresAt.Format(time.RFC3339),
		"employee":     publicEmployee(result.Employee),
	}})
}

func CurrentIdentity(c *gin.Context) (model.Identity, bool) {
	value, ok := c.Get(appmiddleware.IdentityKey)
	if !ok {
		return model.Identity{}, false
	}
	identity, ok := value.(model.Identity)
	return identity, ok
}

func publicEmployee(employee model.Employee) employeeResponse {
	return employeeResponse{
		ID: employee.ID, DepartmentID: employee.DepartmentID, ManagerID: employee.ManagerID,
		Name: employee.Name, Email: employee.Email, Role: employee.Role, IsActive: employee.IsActive,
	}
}
