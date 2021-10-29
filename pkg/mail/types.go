package mail

import (
	"time"

	"github.com/rs/zerolog"
)

const (
	MailerSendgridHTTP  = "sendgrid-http"
	MailerSparkpostHTTP = "sparkpost-http"
)

type MailProvider struct {
	Name      string `json:"name"`
	Transport string `json:"transport"`
	Disabled  bool   `json:"disabled"`
}

type Mail struct {
	Recipients []string `json:"recipients"`
	Subject    string   `json:"subject"`
	Message    string   `json:"message"`

	MailProvider *MailProvider `json:"mail_provider"`
}

type Mailer interface {
	MailProvider() MailProvider
	SetDisabled(bool)

	Send(*Mail) error
}

type Config struct {
	APIKey  string          `json:"-"`
	URI     string          `json:"-"`
	From    string          `json:"-"`
	Logger  *zerolog.Logger `json:"-"`
	Timeout time.Duration   `json:"-"`
}
