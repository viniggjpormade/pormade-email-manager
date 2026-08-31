package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/account"
)

func EnsureAuth(tokenUsecase account.ValidateTokenUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Não autorizado!"})
			c.Abort()
			return
		}

		accessToken := authHeader[7:]

		if !tokenUsecase.Execute(accessToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Não autorizado!"})
			c.Abort()
			return
		}

		c.Next()
	}
}
