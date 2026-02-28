package contract

import "time"

type EmailTemplateResp struct {
	ID              int64     `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	EventName       string    `json:"eventName"`
	SubjectTemplate string    `json:"subjectTemplate"`
	HTMLTemplate    string    `json:"htmlTemplate"`
	TextTemplate    string    `json:"textTemplate"`
	ToEmails        []string  `json:"toEmails"`
	IsEnabled       bool      `json:"isEnabled"`
	IsInternal      bool      `json:"isInternal"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type EmailTemplatePreviewResp struct {
	Subject  string `json:"subject"`
	HTMLBody string `json:"htmlBody"`
	TextBody string `json:"textBody"`
}

type EmailEventListResp struct {
	Events []string `json:"events"`
}

type EmailEventFieldResp struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type EmailEventDescriptorResp struct {
	Name        string                `json:"name"`
	Title       string                `json:"title"`
	Category    string                `json:"category"`
	Public      bool                  `json:"public"`
	Description string                `json:"description"`
	Fields      []EmailEventFieldResp `json:"fields"`
}

type EmailEventCatalogResp struct {
	Items []EmailEventDescriptorResp `json:"items"`
}

type EmailSubscriptionResp struct {
	ID             int64      `json:"id"`
	Email          string     `json:"email"`
	EventName      string     `json:"eventName"`
	Status         string     `json:"status"`
	Token          string     `json:"token,omitempty"`
	SourceIP       string     `json:"sourceIp,omitempty"`
	UnsubscribedAt *time.Time `json:"unsubscribedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type EmailPublicSubscriptionResp struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	EventName string    `json:"eventName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type EmailSubscriptionListResp struct {
	Items []EmailSubscriptionResp `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}

type EmailSubscribeBatchResp struct {
	Items []EmailPublicSubscriptionResp `json:"items"`
}

type EmailOutboxResp struct {
	ID           int64      `json:"id"`
	TemplateCode string     `json:"templateCode"`
	EventName    string     `json:"eventName"`
	ToEmails     []string   `json:"toEmails"`
	Subject      string     `json:"subject"`
	HTMLBody     string     `json:"htmlBody,omitempty"`
	TextBody     string     `json:"textBody,omitempty"`
	Status       string     `json:"status"`
	RetryCount   int        `json:"retryCount"`
	NextRetryAt  time.Time  `json:"nextRetryAt"`
	LastError    string     `json:"lastError,omitempty"`
	SentAt       *time.Time `json:"sentAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type EmailOutboxListResp struct {
	Items []EmailOutboxResp `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}
