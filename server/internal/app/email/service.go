package email

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	domainemail "github.com/grtsinry43/grtblog-v2/server/internal/domain/email"
)

type Service struct {
	repo        domainemail.Repository
	sender      *Sender
	websiteInfo domainconfig.WebsiteInfoRepository
}

func NewService(repo domainemail.Repository, sender *Sender, websiteInfo domainconfig.WebsiteInfoRepository) *Service {
	return &Service{repo: repo, sender: sender, websiteInfo: websiteInfo}
}

func (s *Service) ListEvents() []string {
	return append([]string(nil), AvailableEventNames...)
}

func (s *Service) ListPublicEvents() []string {
	return append([]string(nil), PublicSubscribableEventNames...)
}

func (s *Service) ListEventCatalog() []EventDescriptor {
	return EventCatalog()
}

func (s *Service) ListTemplates(ctx context.Context) ([]*domainemail.Template, error) {
	return s.repo.ListTemplates(ctx)
}

func (s *Service) GetTemplateByCode(ctx context.Context, code string) (*domainemail.Template, error) {
	return s.repo.GetTemplateByCode(ctx, strings.TrimSpace(code))
}

func (s *Service) CreateTemplate(ctx context.Context, tpl *domainemail.Template) error {
	if tpl == nil {
		return domainemail.ErrEmailTemplateRenderFailed
	}
	normalizeTemplate(tpl)
	if err := validateTemplate(tpl); err != nil {
		return err
	}
	return s.repo.CreateTemplate(ctx, tpl)
}

func (s *Service) UpdateTemplate(ctx context.Context, code string, tpl *domainemail.Template) error {
	if tpl == nil {
		return domainemail.ErrEmailTemplateRenderFailed
	}
	tpl.Code = strings.TrimSpace(code)
	normalizeTemplate(tpl)
	if err := validateTemplate(tpl); err != nil {
		return err
	}
	return s.repo.UpdateTemplate(ctx, tpl)
}

func (s *Service) DeleteTemplate(ctx context.Context, code string) error {
	return s.repo.DeleteTemplateByCode(ctx, strings.TrimSpace(code))
}

func (s *Service) PreviewTemplate(ctx context.Context, code string, variables map[string]any) (RenderedTemplate, error) {
	tpl, err := s.repo.GetTemplateByCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return RenderedTemplate{}, err
	}
	rendered, err := RenderTemplate(tpl, s.mergeTemplateVariables(ctx, tpl.EventName, variables))
	if err != nil {
		return RenderedTemplate{}, fmt.Errorf("%w: %v", domainemail.ErrEmailTemplateRenderFailed, err)
	}
	return rendered, nil
}

func (s *Service) TestSend(ctx context.Context, code string, to []string, variables map[string]any) error {
	rendered, err := s.PreviewTemplate(ctx, code, variables)
	if err != nil {
		return err
	}
	if err := s.sender.Send(ctx, Message{
		To:       to,
		Subject:  rendered.Subject,
		HTMLBody: rendered.HTMLBody,
		TextBody: rendered.TextBody,
	}); err != nil {
		return err
	}
	return nil
}

func (s *Service) mergeTemplateVariables(ctx context.Context, eventName string, input map[string]any) map[string]any {
	merged := appEvent.BuildGlobalTemplateVariables(ctx, s.websiteInfo)
	if strings.TrimSpace(eventName) != "" {
		merged["eventName"] = eventName
	}
	merged["occurredAt"] = time.Now().Format(time.RFC3339)
	for key, value := range input {
		merged[key] = value
	}
	return merged
}

func (s *Service) Subscribe(ctx context.Context, email string, eventName string, sourceIP string) (*domainemail.Subscription, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	eventName = strings.TrimSpace(eventName)
	if email == "" || eventName == "" {
		return nil, domainemail.ErrEmailSubscriptionInvalid
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, domainemail.ErrEmailSubscriptionInvalid
	}
	if !IsPublicSubscribableEventName(eventName) {
		return nil, domainemail.ErrEmailSubscriptionEventInvalid
	}
	if existing, err := s.repo.GetSubscriptionByEmailEvent(ctx, email, eventName); err == nil {
		if existing.Status == domainemail.SubscriptionStatusBlocked {
			return nil, domainemail.ErrEmailSubscriptionStatusInvalid
		}
	} else if !errors.Is(err, domainemail.ErrEmailSubscriptionNotFound) {
		return nil, err
	}
	token, err := generateSubscriptionToken()
	if err != nil {
		return nil, err
	}
	sub := &domainemail.Subscription{
		Email:     email,
		EventName: eventName,
		Status:    domainemail.SubscriptionStatusActive,
		Token:     token,
		SourceIP:  strings.TrimSpace(sourceIP),
	}
	if err := s.repo.CreateOrUpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Service) Unsubscribe(ctx context.Context, token string, email string, eventName string) error {
	token = strings.TrimSpace(token)
	if token != "" {
		return s.repo.UnsubscribeByToken(ctx, token)
	}
	email = strings.ToLower(strings.TrimSpace(email))
	eventName = strings.TrimSpace(eventName)
	if email == "" || eventName == "" {
		return domainemail.ErrEmailSubscriptionInvalid
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return domainemail.ErrEmailSubscriptionInvalid
	}
	if !IsPublicSubscribableEventName(eventName) {
		return domainemail.ErrEmailSubscriptionEventInvalid
	}
	return s.repo.UnsubscribeByEmailEvent(ctx, email, eventName)
}

func (s *Service) ListSubscriptions(ctx context.Context, options domainemail.SubscriptionListOptions) ([]*domainemail.Subscription, int64, error) {
	if options.Status != nil {
		status := strings.TrimSpace(*options.Status)
		if status != "" && !isValidSubscriptionStatus(status) {
			return nil, 0, domainemail.ErrEmailSubscriptionStatusInvalid
		}
	}
	return s.repo.ListSubscriptions(ctx, options)
}

func (s *Service) BatchUpdateSubscriptionStatus(ctx context.Context, ids []int64, status string) error {
	status = strings.TrimSpace(status)
	if !isValidSubscriptionStatus(status) {
		return domainemail.ErrEmailSubscriptionStatusInvalid
	}
	uniq := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		return domainemail.ErrEmailSubscriptionInvalid
	}
	return s.repo.BatchUpdateSubscriptionStatus(ctx, uniq, status)
}

func validateTemplate(tpl *domainemail.Template) error {
	if strings.TrimSpace(tpl.Code) == "" || strings.TrimSpace(tpl.Name) == "" {
		return domainemail.ErrEmailTemplateRenderFailed
	}
	if !IsValidEventName(tpl.EventName) {
		return domainemail.ErrEmailTemplateEventInvalid
	}
	if strings.TrimSpace(tpl.SubjectTemplate) == "" {
		return domainemail.ErrEmailTemplateRenderFailed
	}
	if strings.TrimSpace(tpl.HTMLTemplate) == "" && strings.TrimSpace(tpl.TextTemplate) == "" {
		return domainemail.ErrEmailTemplateRenderFailed
	}
	if _, err := RenderTemplate(tpl, map[string]any{}); err != nil {
		// empty vars preview failure is acceptable for missingkey, but syntax errors should fail
		if !isMissingKeyErr(err) {
			return fmt.Errorf("%w: %v", domainemail.ErrEmailTemplateRenderFailed, err)
		}
	}
	return nil
}

func normalizeTemplate(tpl *domainemail.Template) {
	tpl.Code = strings.TrimSpace(tpl.Code)
	tpl.Name = strings.TrimSpace(tpl.Name)
	tpl.EventName = strings.TrimSpace(tpl.EventName)
	tpl.SubjectTemplate = strings.TrimSpace(tpl.SubjectTemplate)
	tpl.HTMLTemplate = strings.TrimSpace(tpl.HTMLTemplate)
	tpl.TextTemplate = strings.TrimSpace(tpl.TextTemplate)
	tpl.ToEmails = normalizeRecipients(tpl.ToEmails)
}

func isMissingKeyErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "map has no entry for key")
}

func generateSubscriptionToken() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func isValidSubscriptionStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case domainemail.SubscriptionStatusActive, domainemail.SubscriptionStatusUnsubscribed, domainemail.SubscriptionStatusBlocked:
		return true
	default:
		return false
	}
}
