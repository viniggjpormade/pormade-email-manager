package router

import (
	"github.com/gin-gonic/gin"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/http/handlers"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/delivery/http/middlewares"
	"github.com/viniggjpormade/pormade-email-manager/internal/adapter/repository/postgres"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/account"
	"gorm.io/gorm"
)

func SetupAccountRoutes(r *gin.Engine, db *gorm.DB) {
	accountRepository := postgres.NewAccountRepository(db)
	validateTokenUseCase := account.NewValidateTokenUseCase(accountRepository)

	createUseCase := account.NewCreateUseCase(accountRepository)

	handler := handlers.NewAccountHandler(createUseCase)

	authMiddleware := middlewares.EnsureAuth(validateTokenUseCase)

	api := r.Group("/accounts")
	{
		api.POST("", authMiddleware, handler.Create)

	}
}
