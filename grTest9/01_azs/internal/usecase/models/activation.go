package models

import "time"

type ActivationData struct {
	LastModification time.Time
	KazsAPIKey       string
	ResetPassword    string
	KazsID           string
	URL              string
	ConfigHash       string
	KazsNumber       string
	KazsTimezone     string
	NTPServer        string
	SupportNumber    string
	Logo             string
}

type UpdateActivationData struct {
	KazsID        string  `json:"kazs_id"`
	ConfigHash    *string `json:"config_hash,omitempty"`
	KazsNumber    *string `json:"kazs_number,omitempty"`
	KazsTimezone  *string `json:"kazs_timezone,omitempty"`
	NtpServer     *string `json:"ntp_server,omitempty"`
	SupportNumber *string `json:"support_number,omitempty"`
	Logo          *string `json:"logo,omitempty"`
}
