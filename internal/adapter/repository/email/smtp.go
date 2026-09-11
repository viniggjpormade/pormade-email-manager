package email

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mime"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type smtpClient struct {
	Host     string
	Password string
	Port     string
	Username string
}
type SmtpRepository struct {
	smtpClient *smtpClient
}

func NewSmtpRepository(connectInfo domain.ConnectInfo) (*SmtpRepository, error) {
	if connectInfo.SMTPHost == nil || strings.TrimSpace(*connectInfo.SMTPHost) == "" {
		return nil, fmt.Errorf("o host smtp não pode ser vazio")
	}
	if connectInfo.SMTPPassword == nil || strings.TrimSpace(*connectInfo.SMTPPassword) == "" {
		return nil, fmt.Errorf("a senha do smtp não pode ser vazia")
	}
	if connectInfo.SMTPPort == nil || strings.TrimSpace(*connectInfo.SMTPPort) == "" {
		return nil, fmt.Errorf("a porta do smtp não pode ser vazia")
	}
	client := smtpClient{
		Host:     *connectInfo.SMTPHost,
		Password: *connectInfo.SMTPPassword,
		Port:     *connectInfo.SMTPPort,
		Username: connectInfo.Username,
	}

	return &SmtpRepository{
		smtpClient: &client,
	}, nil
}

type attachmentData struct {
	Name        string
	ContentType string
	Base64Data  string
	Error       error
}

func processAttachments(attachments []domain.AttachmentsDTO) ([]attachmentData, error) {
	if len(attachments) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	attachmentsData := make([]attachmentData, len(attachments))

	for index, attachment := range attachments {
		wg.Add(1)
		go func(index int, attachment domain.AttachmentsDTO) {
			defer wg.Done()

			contentType := attachment.ContentType
			if contentType == "" {
				contentType = http.DetectContentType(attachment.Data)
			}

			b64 := base64.StdEncoding.EncodeToString(attachment.Data)

			attachmentsData[index] = attachmentData{
				Name:        attachment.Filename,
				ContentType: contentType,
				Base64Data:  b64,
			}
		}(index, attachment)
	}

	wg.Wait()

	for _, attachmentData := range attachmentsData {
		if attachmentData.Error != nil {
			return nil, attachmentData.Error
		}
	}

	return attachmentsData, nil
}

// Monta o email
func buildEmail(params domain.EmailParams, attachments []attachmentData) (string, []byte) {
	buffer := bytes.NewBuffer(nil)
	boundary := fmt.Sprintf("delimitador_%d", time.Now().UnixNano())

	messageId := writeHeader(buffer, &params, boundary)

	writeBody(buffer, &params, boundary)

	writeAttachments(buffer, attachments, boundary)

	buffer.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return messageId, buffer.Bytes()
}

func writeBody(buffer *bytes.Buffer, params *domain.EmailParams, boundary string) {
	var htmlTagRegex = regexp.MustCompile(`(?i)<\/?[a-z][a-z0-9]*(\s+[^>]*)?\/?>`)
	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))

	contentType := "text/plain"
	if htmlTagRegex.MatchString(params.Body) {
		contentType = "text/html"
	}
	buffer.WriteString(fmt.Sprintf("Content-Type: %s; charset=\"utf-8\"\r\n", contentType))

	buffer.WriteString("\r\n")
	buffer.WriteString(params.Body)
	buffer.WriteString("\r\n\r\n")
}

func writeHeader(buffer *bytes.Buffer, params *domain.EmailParams, boundary string) string {
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	messageID := fmt.Sprintf("<%d.%s@pormade.com.br>", time.Now().UnixNano(), hex.EncodeToString(randBytes))
	fmt.Fprintf(buffer, "Message-ID: %s\r\n", messageID)

	fmt.Fprintf(buffer, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))

	fmt.Fprintf(buffer, "From: %s\r\n", params.From)
	fmt.Fprintf(buffer, "To: %s\r\n", strings.Join(params.To, ", "))

	encodedSubject := mime.QEncoding.Encode("utf-8", params.Subject)
	fmt.Fprintf(buffer, "Subject: %s\r\n", encodedSubject)

	if params.InReplyTo != nil {
		fmt.Fprintf(buffer, "In-Reply-To: %s\r\n", params.InReplyTo)

		if params.References != nil {
			fmt.Fprintf(buffer, "References: %s\r\n", params.References)
		} else {
			fmt.Fprintf(buffer, "References: %s\r\n", params.InReplyTo)
		}
	}

	buffer.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(buffer, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary)
	buffer.WriteString("\r\n")
	return messageID
}

func writeAttachments(buffer *bytes.Buffer, attachments []attachmentData, boundary string) {
	for _, attachment := range attachments {
		buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buffer.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", attachment.ContentType, attachment.Name))
		buffer.WriteString("Content-Transfer-Encoding: base64\r\n")
		buffer.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name))
		buffer.WriteString("\r\n")

		writeBase64Lines(buffer, attachment.Base64Data)
		buffer.WriteString("\r\n")
	}
}

func writeBase64Lines(buffer *bytes.Buffer, base64Data string) {
	for i := 0; i < len(base64Data); i += 76 {
		end := i + 76
		if end > len(base64Data) {
			end = len(base64Data)
		}
		buffer.WriteString(base64Data[i:end])
		buffer.WriteString("\r\n")
	}
}

func (repository *SmtpRepository) SendEmail(params domain.EmailParams) (*domain.SendEmailResponse, error) {
	attachments, err := processAttachments(params.Attachments)
	if err != nil {
		return nil, err
	}

	messageId, emailBody := buildEmail(params, attachments)

	auth := smtp.PlainAuth("", repository.smtpClient.Username, repository.smtpClient.Password, repository.smtpClient.Host)
	err = smtp.SendMail(repository.smtpClient.Host+":"+repository.smtpClient.Port, auth, params.From, params.To, emailBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar o e-mail: %w", err)
	}
	response := domain.SendEmailResponse{
		ID:     messageId,
		Status: "sent",
	}

	return &response, nil
}
