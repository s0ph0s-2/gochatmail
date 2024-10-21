package config

import (
	"encoding/json"
	"os"
)

type ChatmailConfig struct {
	MailFullyQualifiedDomainName    string
	MaxEmailsPerMinutePerUser       int
	MaxMailboxSizeMB                int
	MaxMessageSizeB                 int
	DeleteMailsAfterDays            int
	DeleteInactiveUsersAfterDays    int
	UsernameMinLength               int
	UsernameMaxLength               int
	PasswordMinLength               int
	PassthroughSendersList          []string
	PassthroughRecipientsList       []string
	PrivacyContactPostalAddress     string
	PrivacyContactEmailAddress      string
	PrivacyDataOfficerPostalAddress string
	PrivacySupervisorPostalAddress  string
}

func NewChatmailConfig(fqdn string) ChatmailConfig {
	return ChatmailConfig{
		fqdn,
		30,
		100,
		31457280,
		20,
		90,
		9,
		9,
		9,
		[]string{},
		[]string{"xstore@testrun.org"},
		"",
		"",
		"",
		"",
	}
}

func (config ChatmailConfig) Save(filename string) error {
	output_txt, m_err := json.MarshalIndent(config, "", "  ")
	if m_err != nil {
		return m_err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(output_txt)
	return err
}

func LoadChatmailConfigFromFile(filename string, config *ChatmailConfig) error {
	data, r_err := os.ReadFile(filename)
	if r_err != nil {
		return r_err
	}
	j_err := json.Unmarshal(data, config)
	if j_err != nil {
		return j_err
	}
	return nil
}
