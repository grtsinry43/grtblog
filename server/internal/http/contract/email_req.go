package contract

import "encoding/json"

type CreateEmailTemplateReq struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	EventName       string   `json:"eventName"`
	SubjectTemplate string   `json:"subjectTemplate"`
	HTMLTemplate    string   `json:"htmlTemplate"`
	TextTemplate    string   `json:"textTemplate"`
	ToEmails        []string `json:"toEmails"`
	IsEnabled       bool     `json:"isEnabled"`
}

type UpdateEmailTemplateReq struct {
	Name            string   `json:"name"`
	EventName       string   `json:"eventName"`
	SubjectTemplate string   `json:"subjectTemplate"`
	HTMLTemplate    string   `json:"htmlTemplate"`
	TextTemplate    string   `json:"textTemplate"`
	ToEmails        []string `json:"toEmails"`
	IsEnabled       bool     `json:"isEnabled"`
}

type EmailTemplatePreviewReq struct {
	Variables json.RawMessage `json:"variables"`
}

type EmailTemplateTestReq struct {
	ToEmails  []string        `json:"toEmails"`
	Variables json.RawMessage `json:"variables"`
}

type EmailSubscribeReq struct {
	Email      string   `json:"email"`
	EventNames []string `json:"eventNames"`
}

type EmailUnsubscribeReq struct {
	Token     string `json:"token,omitempty"`
	Email     string `json:"email,omitempty"`
	EventName string `json:"eventName,omitempty"`
}

type BatchUpdateEmailSubscriptionStatusReq struct {
	IDs    []int64 `json:"ids"`
	Status string  `json:"status"`
}
