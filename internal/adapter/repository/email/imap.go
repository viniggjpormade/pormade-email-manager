package email

import (
	"bytes"
	"fmt"
	"io"
	netmail "net/mail"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type ImapRepository struct {
	imapClient *client.Client
}

func NewImapRepository(connectInfo domain.ConnectInfo) (*ImapRepository, error) {
	imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", connectInfo.IMAPHost, connectInfo.IMAPPort), nil)
	if err != nil {
		return nil, fmt.Errorf("erro na conexão: %v", err)
	}

	if err := imapClient.Login(connectInfo.Username, connectInfo.IMAPPassword); err != nil {
		return nil, fmt.Errorf("erro no login: %v", err)
	}

	return &ImapRepository{
		imapClient: imapClient,
	}, nil
}

func (repository *ImapRepository) GetAllUnreadEmails() ([]domain.EmailDTO, error) {
	_, err := repository.imapClient.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("erro ao selecionar INBOX: %v", err)
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	indexes, err := repository.imapClient.UidSearch(criteria)
	if err != nil {
		return nil, fmt.Errorf("erro na busca: %v", err)
	}

	if len(indexes) == 0 {
		return []domain.EmailDTO{}, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(indexes...)

	section := &imap.BodySectionName{BodyPartName: imap.BodyPartName{Specifier: imap.HeaderSpecifier, Fields: []string{"REFERENCES"}}, Peek: true}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, imap.FetchFlags, section.FetchItem()}
	messages := make(chan *imap.Message, len(indexes))
	done := make(chan error, 1)

	go func() {
		done <- repository.imapClient.UidFetch(seqset, items, messages)
	}()

	var result []domain.EmailDTO
	for message := range messages {
		if message.Envelope != nil {
			result = append(result, ParseMessageToDTO(message))
		}
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("erro ao baixar mensagens: %v", err)
	}

	return result, nil
}

func (repository *ImapRepository) GetEmailByUid(uid uint32) (*domain.EmailBodyAndAttachments, error) {
	seqset := new(imap.SeqSet)
	seqset.AddNum(uid)
	messages := make(chan *imap.Message, 1)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	done := make(chan error, 1)
	go func() {
		done <- repository.imapClient.UidFetch(seqset, items, messages)
	}()

	var fullMessage *imap.Message
	for message := range messages {
		fullMessage = message
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("erro ao buscar mensagem pelo UID: %w", err)
	}
	if fullMessage == nil {
		return nil, fmt.Errorf("mensagem não encontrada")
	}

	read := fullMessage.GetBody(section)
	if read == nil {
		return nil, fmt.Errorf("corpo da mensagem não encontrado")
	}

	mailReader, err := mail.CreateReader(read)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar leitor MIME: %w", err)
	}

	var htmlBody string
	var textBody string
	var attachments []domain.AttachmentsDTO

	for {
		part, err := mailReader.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("erro ao ler parte da mensagem: %w", err)
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			contentType, _, _ := header.ContentType()
			body, err := io.ReadAll(part.Body)
			if err != nil {
				return nil, fmt.Errorf("erro ao ler corpo da parte inline: %w", err)
			}
			if strings.EqualFold(contentType, "text/html") {
				htmlBody = string(body)
			} else if strings.EqualFold(contentType, "text/plain") && textBody == "" {
				textBody = string(body)
			}
		case *mail.AttachmentHeader:
			filename, _ := header.Filename()
			contentType, _, _ := header.ContentType()
			body, err := io.ReadAll(part.Body)
			if err != nil {
				return nil, fmt.Errorf("erro ao ler corpo do anexo: %w", err)
			}
			attachments = append(attachments, domain.AttachmentsDTO{
				Filename:    filename,
				ContentType: contentType,
				Data:        body,
			})
		}
	}

	finalBody := textBody
	if htmlBody != "" {
		finalBody = htmlBody
	}

	return &domain.EmailBodyAndAttachments{
		Body:        finalBody,
		Attachments: attachments,
	}, nil
}

func (repository *ImapRepository) Disconnect() error {
	return repository.imapClient.Logout()
}

func mapAddresses(addresses []*imap.Address) []domain.AddressDTO {
	var result []domain.AddressDTO
	for _, address := range addresses {
		if address != nil {
			email := fmt.Sprintf("%s@%s", address.MailboxName, address.HostName)
			result = append(result, domain.AddressDTO{
				Name:  address.PersonalName,
				Email: email,
			})
		}
	}
	return result
}

func ParseMessageToDTO(message *imap.Message) domain.EmailDTO {
	dto := domain.EmailDTO{
		SeqNum: message.SeqNum,
		Uid:    message.Uid,
		Size:   message.Size,
		Flags:  message.Flags,
	}

	if message.Envelope != nil {
		dto.Subject = message.Envelope.Subject
		dto.Date = message.Envelope.Date
		dto.MessageId = message.Envelope.MessageId
		dto.InReplyTo = message.Envelope.InReplyTo
		dto.From = mapAddresses(message.Envelope.From)
		dto.To = mapAddresses(message.Envelope.To)
		dto.ReplyTo = mapAddresses(message.Envelope.ReplyTo)
	}

	for sectionName, body := range message.Body {
		if sectionName.Specifier == imap.HeaderSpecifier {
			b, err := io.ReadAll(body)
			if err == nil {
				msg, err := netmail.ReadMessage(bytes.NewReader(b))
				if err == nil && msg != nil {
					dto.References = msg.Header.Get("References")
				}
			}
		}
	}

	return dto
}
