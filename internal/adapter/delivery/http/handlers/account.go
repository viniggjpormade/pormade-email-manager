package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/account"
)

type AccountHandler struct {
	createUseCase account.CreateUseCase
}

func NewAccountHandler(
	createUseCase account.CreateUseCase,
) *AccountHandler {
	return &AccountHandler{
		createUseCase: createUseCase,
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
func (h *AccountHandler) Create(c *gin.Context) {
	var input account.CreateAccountDto

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.createUseCase.Execute(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}
