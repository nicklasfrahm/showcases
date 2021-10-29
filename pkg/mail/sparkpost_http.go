package mail

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
)

type SparkpostHTTPMailer struct {
	Config *Config

	mailProvider *MailProvider
	httpClient   *http.Client
	mutex        sync.Mutex
}

type SparkpostRecipient struct {
	Address string `json:"address"`
}

type SparkpostContent struct {
	From    string `json:"from"`
	Subject string `json:"subject"`
	Text    string `json:"text"`
}

type SparkpostMail struct {
	Recipients []SparkpostRecipient `json:"recipients"`
	Content    SparkpostContent     `json:"content"`
}

func (m *SparkpostHTTPMailer) MailProvider() MailProvider {
	// Return a copy to prevent race conditions and state inconsistencies.
	return *m.mailProvider
}

func (m *SparkpostHTTPMailer) SetDisabled(disabled bool) {
	m.mutex.Lock()
	m.mailProvider.Disabled = disabled
	m.mutex.Unlock()
}

func (m *SparkpostHTTPMailer) Send(mail *Mail) error {
	// Encode API message body.
	recipients := make([]SparkpostRecipient, len(mail.Recipients))
	for i, recipient := range mail.Recipients {
		recipients[i] = SparkpostRecipient{
			Address: recipient,
		}
	}
	reqJson, err := json.Marshal(SparkpostMail{
		Content: SparkpostContent{
			From:    m.Config.From,
			Subject: mail.Subject,
			Text:    mail.Message,
		},
		Recipients: recipients,
	})
	if err != nil {
		m.SetDisabled(true)
		return err
	}

	// Create a new HTTP request.
	req, err := http.NewRequest(http.MethodPost, m.Config.URI+"/transmissions", bytes.NewReader(reqJson))
	if err != nil {
		m.SetDisabled(true)
		return err
	}

	// Add headers to HTTP request.
	req.Header.Set("Authorization", m.Config.APIKey)
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

func NewSparkpostHTTP(config *Config) Mailer {
	client := &http.Client{Timeout: config.Timeout}

	return &SparkpostHTTPMailer{
		Config: config,
		mailProvider: &MailProvider{
			Name:      MailerSparkpostHTTP,
			Transport: "HTTP",
			Disabled:  false,
		},
		httpClient: client,
	}
}
