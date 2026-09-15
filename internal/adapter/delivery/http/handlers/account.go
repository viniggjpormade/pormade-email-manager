package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/account"
)

type AccountHandler struct {
	createUseCase account.CreateUseCase
	updateUseCase account.UpdateUseCase
}

func NewAccountHandler(
	createUseCase account.CreateUseCase,
	updateUseCase account.UpdateUseCase,
) *AccountHandler {
	return &AccountHandler{
		createUseCase: createUseCase,
		updateUseCase: updateUseCase,
	}
}

// Create godoc
// @Summary Criar conta
// @Description Cria uma nova conta de email
// @Tags Account
// @Accept json
// @Produce json
// @Param account body account.CreateAccountDto true "Dados da conta"
// @Success 201 {object} domain.Account
// @Failure 400 {object} map[string]string "error message"
// @Router /accounts [post]
func (handler *AccountHandler) Create(context *gin.Context) {
	var input account.CreateAccountDto

	if err := context.ShouldBind(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := handler.createUseCase.Execute(input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, acc)
}

// Update godoc
// @Summary Atualizar conta
// @Description Atualiza uma conta de email existente
// @Tags Account
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param account body account.UpdateAccountDto true "Dados da conta a serem atualizados"
// @Success 200 {object} domain.Account
// @Failure 400 {object} map[string]string "error message"
// @Failure 401 {object} map[string]string "error message"
// @Router /accounts [patch]
func (handler *AccountHandler) Update(context *gin.Context) {
	accContext, exists := context.Get("account")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found in context"})
		return
	}

	accEntity, ok := accContext.(domain.Account)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast account from context"})
		return
	}

	var input account.UpdateAccountDto
	if err := context.ShouldBind(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := handler.updateUseCase.Execute(accEntity, input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, acc)
}
