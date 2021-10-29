package mail

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
)

type SendgridHTTPMailer struct {
	Config *Config

	mailProvider *MailProvider
	httpClient   *http.Client
	mutex        sync.Mutex
}

type SendgridMIMETypedContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type SendgridPersonalization struct {
	To []SendgridAccount `json:"to"`
}

type SendgridAccount struct {
	Email string `json:"email"`
}

type SendgridMail struct {
	Personalizations []SendgridPersonalization  `json:"personalizations"`
	From             []SendgridAccount          `json:"from"`
	Subject          string                     `json:"subject"`
	Content          []SendgridMIMETypedContent `json:"content"`
}

func (m *SendgridHTTPMailer) MailProvider() MailProvider {
	// Return a copy to prevent race conditions and state inconsistencies.
	return *m.mailProvider
}

func (m *SendgridHTTPMailer) SetDisabled(disabled bool) {
	m.mutex.Lock()
	m.mailProvider.Disabled = disabled
	m.mutex.Unlock()
}

func (m *SendgridHTTPMailer) Send(mail *Mail) error {
	// Encode API message body.
	recipients := make([]SendgridAccount, len(mail.Recipients))
	for i, recipient := range mail.Recipients {
		recipients[i] = SendgridAccount{
			Email: recipient,
		}
	}
	reqJson, err := json.Marshal(SendgridMail{
		Personalizations: []SendgridPersonalization{
			{
				To: recipients,
			},
		},
		From: []SendgridAccount{
			{
				Email: m.Config.From,
			},
		},
		Subject: mail.Subject,
		Content: []SendgridMIMETypedContent{
			{
				Type:  "text/plain",
				Value: mail.Message,
			},
		},
	})
	if err != nil {
		m.SetDisabled(true)
		return err
	}

	// Create a new HTTP request.
	req, err := http.NewRequest(http.MethodPost, m.Config.URI+"/mail/send", bytes.NewReader(reqJson))
	if err != nil {
		m.SetDisabled(true)
		return err
	}

	// Add headers to HTTP request.
	req.Header.Set("Authorization", "Bearer "+m.Config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := m.httpClient.Do(req)
	if err != nil {
		m.SetDisabled(true)
		return err
	}

	// Display response.
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		m.SetDisabled(true)
		return err
	}

	if res.StatusCode != 200 {
		m.SetDisabled(true)
		return errors.New(string(data))
	}

	// Add information about the use mail provider.
	mail.MailProvider = new(MailProvider)
	*mail.MailProvider = *m.mailProvider

	return nil
}

func NewSendgridHTTP(config *Config) Mailer {
	client := &http.Client{Timeout: config.Timeout}

	return &SendgridHTTPMailer{
		Config: config,
		mailProvider: &MailProvider{
			Name:      MailerSendgridHTTP,
			Transport: "HTTP",
			Disabled:  false,
		},
		httpClient: client,
	}
}
