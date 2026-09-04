package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type externalEmailProvider struct {
	apiKey string
}

func NewExternalEmailProvider(apiKey string) domain.OutboundEmailProvider {
	return &externalEmailProvider{
		apiKey: apiKey,
	}
}

type externalEmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type externalEmailPayload struct {
	From           string                    `json:"from"`
	To             []string                  `json:"to"`
	Cc             []string                  `json:"cc"`
	Bcc            []string                  `json:"bcc"`
	Subject        string                    `json:"subject"`
	Text           string                    `json:"text"`
	Html           string                    `json:"html"`
	ReplyTo        string                    `json:"replyTo"`
	InReplyTo      *string                   `json:"inReplyTo,omitempty"`
	References     []string                  `json:"references,omitempty"`
	IdempotencyKey string                    `json:"idempotencyKey"`
	Attachments    []externalEmailAttachment `json:"attachments"`
}

func (provider *externalEmailProvider) SendEmail(from string, to string, subject string, body string, inReplyTo string, attachments []domain.AttachmentDTO) (*domain.SendEmailResponse, error) {
	baseUrl := os.Getenv("EMAIL_PROVIDER_BASE_URL")

	externalAttachments := make([]externalEmailAttachment, 0)
	for _, attachment := range attachments {
		extAttachment := externalEmailAttachment{
			Filename:    attachment.Filename,
			ContentType: attachment.ContentType,
			Content:     base64.StdEncoding.EncodeToString(attachment.Data),
		}
		externalAttachments = append(externalAttachments, extAttachment)
	}

	payload := externalEmailPayload{
		From:           from,
		To:             []string{to},
		Cc:             []string{},
		Bcc:            []string{},
		Subject:        subject,
		Text:           body,
		Html:           body,
		ReplyTo:        from,
		IdempotencyKey: uuid.New().String(),
		Attachments:    externalAttachments,
	}

	if inReplyTo != "" {
		formattedInReplyTo := inReplyTo
		if !strings.HasPrefix(formattedInReplyTo, "<") {
			formattedInReplyTo = "<" + formattedInReplyTo + ">"
		}
		payload.InReplyTo = &formattedInReplyTo
		payload.References = []string{formattedInReplyTo}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar JSON: %w", err)
	}

	baseUrl = strings.TrimSuffix(baseUrl, "/")

	url := fmt.Sprintf("%s/mail/send", baseUrl)
	return provider.executeRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
}

func (provider *externalEmailProvider) VerifyEmailStatus(id string) (*domain.SendEmailResponse, error) {
	baseUrl := os.Getenv("EMAIL_PROVIDER_BASE_URL")
	baseUrl = strings.TrimSuffix(baseUrl, "/")

	url := fmt.Sprintf("%s/mail/%s", baseUrl, id)
	return provider.executeRequest(http.MethodGet, url, nil)
}

func (provider *externalEmailProvider) executeRequest(method, url string, body io.Reader) (*domain.SendEmailResponse, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisicao HTTP: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	if provider.apiKey != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.apiKey))
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("erro de rede na api externa: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("api externa retornou HTTP Status %d: %s", response.StatusCode, string(bodyBytes))
	}

	var emailResponse domain.SendEmailResponse
	if err := json.NewDecoder(response.Body).Decode(&emailResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON da API: %w", err)
	}

	return &emailResponse, nil
}
