package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
	"github.com/viniggjpormade/pormade-email-manager/internal/usecase/email"
)

type EmailHandler struct {
	sendEmailUseCase email.SendEmailUseCase
}

func NewEmailHandler(
	sendEmailUseCase email.SendEmailUseCase,
) *EmailHandler {
	return &EmailHandler{
		sendEmailUseCase: sendEmailUseCase,
	}
}

// SendEmail godoc
// @Summary Enviar email
// @Description Envia um novo email com suporte a anexos
// @Tags Email
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param to formData string true "Email do destinatário"
// @Param subject formData string true "Assunto do e-mail(Mensagens de retorno devem ter Re: subject do email a ser respondido)"
// @Param body formData string true "Corpo do e-mail (Texto ou HTML)"
// @Param inReplyTo formData string false "ID da mensagem que este e-mail responde (opcional)"
// @Param references formData string false "Cadeia de referências (opcional)"
// @Param attachments formData []file false "Arquivos em anexo"
// @Success 201 {object} map[string]string "mensagem de sucesso"
// @Failure 400 {object} map[string]string "error message"
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /emails [post]
func (handler *EmailHandler) SendEmail(context *gin.Context) {
	var input email.SendEmailDTO

	if err := context.ShouldBind(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, err := context.MultipartForm()
	if err == nil && form != nil {
		files := form.File["attachments"]
		for _, file := range files {
			openedFile, err := file.Open()
			if err != nil {
				continue
			}

			fileBytes := make([]byte, file.Size)
			_, err = openedFile.Read(fileBytes)
			openedFile.Close()
			if err != nil {
				continue
			}

			input.Attachments = append(input.Attachments, domain.AttachmentsDTO{
				Filename:    file.Filename,
				ContentType: file.Header.Get("Content-Type"),
				Data:        fileBytes,
			})
		}
	}

	accountInterface, exists := context.Get("account")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "conta não autenticada"})
		return
	}
	account := accountInterface.(domain.Account)

	err = handler.sendEmailUseCase.Execute(account, input)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "email enviado"})
}
