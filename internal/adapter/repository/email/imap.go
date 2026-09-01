package email

import (
	"fmt"
	"io"
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
	imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", connectInfo.Address, connectInfo.Port), nil)
	if err != nil {
		return nil, fmt.Errorf("erro na conexão: %v", err)
	}

	if err := imapClient.Login(connectInfo.Username, connectInfo.Password); err != nil {
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
	ids, err := repository.imapClient.UidSearch(criteria)
	if err != nil {
		return nil, fmt.Errorf("erro na busca: %v", err)
	}

	if len(ids) == 0 {
		return []domain.EmailDTO{}, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, imap.FetchFlags}
	messages := make(chan *imap.Message, len(ids))
	done := make(chan error, 1)

	go func() {
		done <- repository.imapClient.UidFetch(seqset, items, messages)
	}()

	var result []domain.EmailDTO
	for msg := range messages {
		if msg.Envelope != nil {
			result = append(result, ParseMessageToDTO(msg))
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

	var fullMsg *imap.Message
	for msg := range messages {
		fullMsg = msg
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("erro ao buscar mensagem pelo UID: %w", err)
	}
	if fullMsg == nil {
		return nil, fmt.Errorf("mensagem não encontrada")
	}

	r := fullMsg.GetBody(section)
	if r == nil {
		return nil, fmt.Errorf("corpo da mensagem não encontrado")
	}

	mr, err := mail.CreateReader(r)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar leitor MIME: %w", err)
	}

	var htmlBody string
	var textBody string
	var attachments []domain.AttachmentDTO

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("erro ao ler parte da mensagem: %w", err)
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			contentType, _, _ := h.ContentType()
			b, err := io.ReadAll(p.Body)
			if err != nil {
				return nil, fmt.Errorf("erro ao ler corpo da parte inline: %w", err)
			}
			if strings.EqualFold(contentType, "text/html") {
				htmlBody = string(b)
			} else if strings.EqualFold(contentType, "text/plain") && textBody == "" {
				textBody = string(b)
			}
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			contentType, _, _ := h.ContentType()
			b, err := io.ReadAll(p.Body)
			if err != nil {
				return nil, fmt.Errorf("erro ao ler corpo do anexo: %w", err)
			}
			attachments = append(attachments, domain.AttachmentDTO{
				Filename:    filename,
				ContentType: contentType,
				Data:        b,
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

func mapAddresses(addrs []*imap.Address) []domain.AddressDTO {
	var result []domain.AddressDTO
	for _, addr := range addrs {
		if addr != nil {
			email := fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName)
			result = append(result, domain.AddressDTO{
				Name:  addr.PersonalName,
				Email: email,
			})
		}
	}
	return result
}

func ParseMessageToDTO(msg *imap.Message) domain.EmailDTO {
	dto := domain.EmailDTO{
		SeqNum: msg.SeqNum,
		Uid:    msg.Uid,
		Size:   msg.Size,
		Flags:  msg.Flags,
	}

	if msg.Envelope != nil {
		dto.Subject = msg.Envelope.Subject
		dto.Date = msg.Envelope.Date
		dto.MessageId = msg.Envelope.MessageId
		dto.InReplyTo = msg.Envelope.InReplyTo
		dto.From = mapAddresses(msg.Envelope.From)
		dto.To = mapAddresses(msg.Envelope.To)
		dto.ReplyTo = mapAddresses(msg.Envelope.ReplyTo)
	}

	return dto
}
