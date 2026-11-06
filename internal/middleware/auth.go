package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mahmoudmetwally/hr-management-system/internal/apperror"
	"github.com/mahmoudmetwally/hr-management-system/internal/model"
)

const IdentityKey = "authenticated_identity"

type Authenticator interface {
	Authenticate(context.Context, string) (model.Identity, error)
}

func Authenticate(authenticator Authenticator, onError func(*gin.Context, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			onError(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}
		identity, err := authenticator.Authenticate(c.Request.Context(), parts[1])
		if err != nil {
			onError(c, err)
			c.Abort()
			return
		}
		c.Set(IdentityKey, identity)
		c.Next()
	}
}

func RequireRole(role model.Role, onError func(*gin.Context, error)) gin.HandlerFunc {
	return RequireRoles(onError, role)
}

func RequireRoles(onError func(*gin.Context, error), roles ...model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, ok := c.Get(IdentityKey)
		identity, valid := value.(model.Identity)
		allowed := false
		if valid {
			for _, role := range roles {
				if identity.Role == role {
					allowed = true
					break
				}
			}
		}
		if !ok || !valid || !allowed {
			onError(c, apperror.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
