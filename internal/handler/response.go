package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
)

func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "an unexpected error occurred"
	var details any

	var validation *apperror.ValidationError
	switch {
	case errors.As(err, &validation):
		status, code, message, details = http.StatusUnprocessableEntity, "VALIDATION_ERROR", validation.Error(), validation.Fields
	case errors.Is(err, apperror.ErrUnauthorized):
		status, code, message = http.StatusUnauthorized, "UNAUTHORIZED", "invalid or missing authentication"
	case errors.Is(err, apperror.ErrInactive):
		status, code, message = http.StatusForbidden, "ACCOUNT_INACTIVE", "employee account is inactive"
	case errors.Is(err, apperror.ErrForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", "you do not have permission to perform this action"
	case errors.Is(err, apperror.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "the requested resource was not found"
	case errors.Is(err, apperror.ErrConflict):
		status, code, message = http.StatusConflict, "CONFLICT", "the request conflicts with existing data"
	case errors.Is(err, apperror.ErrInvalidState):
		status, code, message = http.StatusConflict, "INVALID_STATE", "the resource is not in the required state"
	case errors.Is(err, apperror.ErrInsufficientLeave):
		status, code, message = http.StatusConflict, "INSUFFICIENT_LEAVE_BALANCE", "the employee does not have enough available leave"
	}

	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message, "details": details}})
}

func validationError(fields map[string]string) error {
	return &apperror.ValidationError{Fields: fields}
}
