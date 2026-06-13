package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/relay/backend/internal/user"
)

const currentUserKey = "current_user"

func Middleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header must use Bearer token"})
			return
		}

		currentUser, err := service.AuthenticateToken(c.Request.Context(), parts[1])
		if err != nil {
			status := http.StatusUnauthorized
			message := "invalid or expired token"
			if errors.Is(err, ErrInactiveUser) {
				status = http.StatusForbidden
				message = "user account is inactive"
			}

			c.AbortWithStatusJSON(status, gin.H{"error": message})
			return
		}

		c.Set(currentUserKey, currentUser)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*user.User, bool) {
	value, exists := c.Get(currentUserKey)
	if !exists {
		return nil, false
	}

	currentUser, ok := value.(*user.User)
	return currentUser, ok
}
