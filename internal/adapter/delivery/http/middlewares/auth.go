package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/account"
)

func EnsureAuth(tokenUsecase account.ValidateTokenUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Não autorizado! Header vazio."})
			c.Abort()
			return
		}

		accessToken := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			accessToken = authHeader[7:]
		}

		acc, err := tokenUsecase.Execute(accessToken)
		if err != nil || acc == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Não autorizado!"})
			c.Abort()
			return
		}

		c.Set("account", *acc)
		c.Next()
	}
}
