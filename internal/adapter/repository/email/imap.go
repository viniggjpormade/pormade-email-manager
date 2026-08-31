package email

import (
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
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
		dto.From = mapAddresses(msg.Envelope.From)
		dto.To = mapAddresses(msg.Envelope.To)
	}

	return dto
}
