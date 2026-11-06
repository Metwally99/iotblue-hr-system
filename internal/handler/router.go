package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	appmiddleware "github.com/mahmoudmetwally/hr-management-system/internal/middleware"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

type RouterDependencies struct {
	Logger             *slog.Logger
	Database           DatabasePinger
	Auth               AuthService
	Departments        DepartmentService
	Employees          EmployeeService
	Leave              LeaveService
	Attendance         AttendanceService
	CORSAllowedOrigins []string
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(appmiddleware.CORS(deps.CORSAllowedOrigins))
	router.Use(appmiddleware.RequestID(), appmiddleware.RequestLogger(deps.Logger))

	health := NewHealthHandler(deps.Database)
	router.GET("/health", health.Health)
	router.GET("/ready", health.Ready)
	router.StaticFile("/openapi.yaml", "api/openapi.yaml")

	api := router.Group("/api/v1")
	api.GET("/", func(c *gin.Context) {
		deps.Logger.Debug("api root requested")
		c.JSON(200, gin.H{"data": gin.H{"name": "HR Management API", "version": "v1"}})
	})

	auth := NewAuthHandler(deps.Auth)
	api.POST("/auth/login", auth.Login)

	protected := api.Group("")
	protected.Use(appmiddleware.Authenticate(deps.Auth, respondError))

	departments := NewDepartmentHandler(deps.Departments)
	protected.GET("/departments", departments.List)
	protected.GET("/departments/:id", departments.Get)
	admin := protected.Group("")
	admin.Use(appmiddleware.RequireRoles(respondError, model.RoleAdmin, model.RoleHR))
	admin.POST("/departments", departments.Create)
	admin.PATCH("/departments/:id", departments.Update)

	employees := NewEmployeeHandler(deps.Employees)
	admin.POST("/employees", employees.Create)
	admin.GET("/employees", employees.List)
	admin.GET("/employees/:id", employees.Get)
	admin.PATCH("/employees/:id", employees.Update)
	admin.PATCH("/employees/:id/deactivate", employees.Deactivate)

	leave := NewLeaveHandler(deps.Leave)
	protected.POST("/leave-requests", leave.Submit)
	protected.GET("/leave-requests", leave.ListAll)
	protected.GET("/leave-requests/:id", leave.Get)
	reviewers := protected.Group("")
	reviewers.Use(appmiddleware.RequireRoles(respondError, model.RoleAdmin, model.RoleHR, model.RoleManager))
	reviewers.POST("/leave-requests/:id/approve", leave.Approve)
	reviewers.POST("/leave-requests/:id/reject", leave.Reject)
	protected.GET("/employees/:id/leave-balances", leave.BalanceForEmployee)

	attendance := NewAttendanceHandler(deps.Attendance)
	protected.POST("/attendance/check-in", attendance.CheckIn)
	protected.POST("/attendance/check-out", attendance.CheckOut)
	protected.GET("/attendance", attendance.ListAll)
	admin.GET("/attendance/report", attendance.Report)

	return router
}
