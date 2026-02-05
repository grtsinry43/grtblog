package friendlink

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
)

type AdminService struct {
	appRepo      social.FriendLinkApplicationRepository
	linkRepo     social.FriendLinkRepository
	instanceRepo federation.FederationInstanceRepository
	userRepo     identity.Repository
	events       appEvent.Bus
}

func NewAdminService(appRepo social.FriendLinkApplicationRepository, linkRepo social.FriendLinkRepository, instanceRepo federation.FederationInstanceRepository, userRepo identity.Repository, events appEvent.Bus) *AdminService {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &AdminService{
		appRepo:      appRepo,
		linkRepo:     linkRepo,
		instanceRepo: instanceRepo,
		userRepo:     userRepo,
		events:       events,
	}
}

type ApplicationListOptions struct {
	Status       string
	ApplyChannel string
	Keyword      string
	Page         int
	PageSize     int
}

type FriendLinkListOptions struct {
	IsActive *bool
	Kind     string
	SyncMode string
	Keyword  string
	Page     int
	PageSize int
}

type CreateFriendLinkCmd struct {
	Name         string
	URL          string
	Logo         *string
	Description  *string
	RSSURL       *string
	Kind         string
	SyncMode     string
	InstanceID   *int64
	SyncInterval *int
	IsActive     bool
	UserID       *int64
}

type UpdateFriendLinkCmd struct {
	ID           int64
	Name         string
	URL          string
	Logo         *string
	Description  *string
	RSSURL       *string
	Kind         string
	SyncMode     string
	InstanceID   *int64
	SyncInterval *int
	IsActive     bool
	UserID       *int64
}

func (s *AdminService) ListApplications(ctx context.Context, options ApplicationListOptions) ([]social.FriendLinkApplication, int64, error) {
	return s.appRepo.List(ctx, social.FriendLinkApplicationListOptions{
		Status:       strings.TrimSpace(options.Status),
		ApplyChannel: strings.TrimSpace(options.ApplyChannel),
		Keyword:      strings.TrimSpace(options.Keyword),
		Page:         options.Page,
		PageSize:     options.PageSize,
	})
}

func (s *AdminService) ApproveApplication(ctx context.Context, id int64) (*social.FriendLinkApplication, error) {
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app.Status == social.FriendLinkAppStatusBlocked {
		return nil, social.ErrFriendLinkApplicationBlocked
	}
	app.Status = social.FriendLinkAppStatusApproved
	if err := s.ensureFriendLink(ctx, app); err != nil {
		return nil, err
	}
	if app.ApplyChannel == social.FriendLinkApplyChannelFederation {
		_ = s.activateFederationInstance(ctx, app.URL)
	}
	if err := s.appRepo.UpdateByID(ctx, app); err != nil {
		return nil, err
	}
	s.publishApplicationStatusEvent(ctx, "friendlink.application.approved", app)
	return app, nil
}

func (s *AdminService) RejectApplication(ctx context.Context, id int64) (*social.FriendLinkApplication, error) {
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app.Status == social.FriendLinkAppStatusBlocked {
		return nil, social.ErrFriendLinkApplicationBlocked
	}
	app.Status = social.FriendLinkAppStatusRejected
	_ = s.deactivateFriendLink(ctx, app.URL)
	if err := s.appRepo.UpdateByID(ctx, app); err != nil {
		return nil, err
	}
	s.publishApplicationStatusEvent(ctx, "friendlink.application.rejected", app)
	return app, nil
}

func (s *AdminService) BlockApplication(ctx context.Context, id int64) (*social.FriendLinkApplication, error) {
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	app.Status = social.FriendLinkAppStatusBlocked
	if err := s.appRepo.UpdateByID(ctx, app); err != nil {
		return nil, err
	}
	_ = s.deactivateFriendLink(ctx, app.URL)
	if app.ApplyChannel == social.FriendLinkApplyChannelFederation {
		_ = s.blockFederationInstance(ctx, app.URL)
	}
	s.publishApplicationStatusEvent(ctx, "friendlink.application.blocked", app)
	return app, nil
}

func (s *AdminService) UpdateApplicationStatus(ctx context.Context, id int64, status string) (*social.FriendLinkApplication, error) {
	status = strings.TrimSpace(status)
	switch status {
	case social.FriendLinkAppStatusApproved:
		return s.ApproveApplication(ctx, id)
	case social.FriendLinkAppStatusRejected:
		return s.RejectApplication(ctx, id)
	case social.FriendLinkAppStatusBlocked:
		return s.BlockApplication(ctx, id)
	case social.FriendLinkAppStatusPending:
		app, err := s.appRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		app.Status = social.FriendLinkAppStatusPending
		_ = s.deactivateFriendLink(ctx, app.URL)
		if err := s.appRepo.UpdateByID(ctx, app); err != nil {
			return nil, err
		}
		return app, nil
	default:
		return nil, errors.New("不支持的状态")
	}
}

func (s *AdminService) ListFriendLinks(ctx context.Context, options FriendLinkListOptions) ([]social.FriendLink, int64, error) {
	return s.linkRepo.List(ctx, social.FriendLinkListOptions{
		IsActive: options.IsActive,
		Kind:     strings.TrimSpace(options.Kind),
		SyncMode: strings.TrimSpace(options.SyncMode),
		Keyword:  strings.TrimSpace(options.Keyword),
		Page:     options.Page,
		PageSize: options.PageSize,
	})
}

func (s *AdminService) CreateFriendLink(ctx context.Context, cmd CreateFriendLinkCmd) (*social.FriendLink, error) {
	name := strings.TrimSpace(cmd.Name)
	url := strings.TrimSpace(cmd.URL)
	if name == "" || url == "" {
		return nil, errors.New("友链名称和URL不能为空")
	}
	if _, err := s.linkRepo.FindByURL(ctx, url); err == nil {
		return nil, errors.New("友链已存在")
	} else if !errors.Is(err, social.ErrFriendLinkNotFound) {
		return nil, err
	}
	kind := strings.TrimSpace(cmd.Kind)
	if kind == "" {
		kind = social.FriendLinkKindManual
	}
	syncMode := strings.TrimSpace(cmd.SyncMode)
	if syncMode == "" {
		syncMode = social.FriendLinkSyncModeNone
	}
	link := &social.FriendLink{
		Name:             name,
		URL:              url,
		Logo:             cmd.Logo,
		Description:      cmd.Description,
		RSSURL:           cmd.RSSURL,
		Kind:             kind,
		SyncMode:         syncMode,
		InstanceID:       cmd.InstanceID,
		SyncInterval:     cmd.SyncInterval,
		TotalPostsCached: 0,
		UserID:           cmd.UserID,
		IsActive:         cmd.IsActive,
	}
	if err := s.linkRepo.Create(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *AdminService) UpdateFriendLink(ctx context.Context, cmd UpdateFriendLinkCmd) (*social.FriendLink, error) {
	name := strings.TrimSpace(cmd.Name)
	url := strings.TrimSpace(cmd.URL)
	if name == "" || url == "" {
		return nil, errors.New("友链名称和URL不能为空")
	}
	link, err := s.linkRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	link.Name = name
	link.URL = url
	link.Logo = cmd.Logo
	link.Description = cmd.Description
	link.RSSURL = cmd.RSSURL
	link.Kind = strings.TrimSpace(cmd.Kind)
	if link.Kind == "" {
		link.Kind = social.FriendLinkKindManual
	}
	link.SyncMode = strings.TrimSpace(cmd.SyncMode)
	if link.SyncMode == "" {
		link.SyncMode = social.FriendLinkSyncModeNone
	}
	link.InstanceID = cmd.InstanceID
	link.SyncInterval = cmd.SyncInterval
	link.IsActive = cmd.IsActive
	link.UserID = cmd.UserID
	if err := s.linkRepo.Update(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *AdminService) DeleteFriendLink(ctx context.Context, id int64) error {
	return s.linkRepo.Delete(ctx, id)
}

func (s *AdminService) BlockFriendLink(ctx context.Context, id int64) (*social.FriendLink, error) {
	link, err := s.linkRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	link.IsActive = false
	if err := s.linkRepo.Update(ctx, link); err != nil {
		return nil, err
	}
	app, err := s.appRepo.FindByURL(ctx, link.URL)
	if err != nil && !errors.Is(err, social.ErrFriendLinkApplicationNotFound) {
		return nil, err
	}
	if app == nil {
		status := social.FriendLinkAppStatusBlocked
		applyChannel := social.FriendLinkApplyChannelAdmin
		requestedSync := link.SyncMode
		if strings.TrimSpace(requestedSync) == "" {
			requestedSync = social.FriendLinkSyncModeNone
		}
		if link.Kind == social.FriendLinkKindFederation {
			applyChannel = social.FriendLinkApplyChannelFederation
			requestedSync = social.FriendLinkSyncModeFederation
		}
		app = &social.FriendLinkApplication{
			Name:              toOptionalString(link.Name),
			URL:               link.URL,
			Logo:              link.Logo,
			Description:       link.Description,
			ApplyChannel:      applyChannel,
			RequestedSyncMode: requestedSync,
			RSSURL:            link.RSSURL,
			Manifest:          json.RawMessage("{}"),
			SignatureVerified: false,
			UserID:            link.UserID,
			Status:            status,
		}
		if err := s.appRepo.Create(ctx, app); err != nil {
			return nil, err
		}
	} else if app.Status != social.FriendLinkAppStatusBlocked {
		app.Status = social.FriendLinkAppStatusBlocked
		if err := s.appRepo.UpdateByID(ctx, app); err != nil {
			return nil, err
		}
	}
	if link.Kind == social.FriendLinkKindFederation {
		_ = s.blockFederationInstance(ctx, link.URL)
	}
	return link, nil
}

func (s *AdminService) ensureFriendLink(ctx context.Context, app *social.FriendLinkApplication) error {
	url := strings.TrimSpace(app.URL)
	if url == "" {
		return errors.New("友链URL不能为空")
	}
	link, err := s.linkRepo.FindByURL(ctx, url)
	if err != nil && !errors.Is(err, social.ErrFriendLinkNotFound) {
		return err
	}
	kind := social.FriendLinkKindManual
	syncMode := app.RequestedSyncMode
	var instanceID *int64
	if app.ApplyChannel == social.FriendLinkApplyChannelFederation {
		kind = social.FriendLinkKindFederation
		syncMode = social.FriendLinkSyncModeFederation
		instance, err := s.instanceRepo.GetByBaseURL(ctx, url)
		if err != nil {
			if errors.Is(err, federation.ErrFederationInstanceNotFound) {
				return errors.New("联邦实例不存在")
			}
			return err
		}
		instanceID = &instance.ID
	}
	if link == nil {
		link = &social.FriendLink{
			Name:             fallbackName(app.Name, url),
			URL:              url,
			Logo:             app.Logo,
			Description:      app.Description,
			RSSURL:           app.RSSURL,
			Kind:             kind,
			SyncMode:         syncMode,
			InstanceID:       instanceID,
			TotalPostsCached: 0,
			UserID:           app.UserID,
			IsActive:         true,
		}
		return s.linkRepo.Create(ctx, link)
	}
	link.Name = fallbackName(app.Name, link.URL)
	link.Logo = app.Logo
	link.Description = app.Description
	link.RSSURL = app.RSSURL
	link.Kind = kind
	link.SyncMode = syncMode
	link.InstanceID = instanceID
	link.UserID = app.UserID
	link.IsActive = true
	return s.linkRepo.Update(ctx, link)
}

func (s *AdminService) publishApplicationStatusEvent(ctx context.Context, name string, app *social.FriendLinkApplication) {
	if app == nil {
		return
	}
	recipientEmail := s.resolveApplicantEmail(ctx, app.UserID)
	_ = s.events.Publish(ctx, appEvent.Generic{
		EventName: name,
		At:        time.Now(),
		Payload: map[string]any{
			"ID":             app.ID,
			"URL":            app.URL,
			"Status":         app.Status,
			"Name":           toValue(app.Name),
			"recipientEmail": recipientEmail,
		},
	})
}

func (s *AdminService) resolveApplicantEmail(ctx context.Context, userID *int64) string {
	if userID == nil || *userID <= 0 || s.userRepo == nil {
		return ""
	}
	user, err := s.userRepo.FindByID(ctx, *userID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(user.Email)
}

func toValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (s *AdminService) deactivateFriendLink(ctx context.Context, url string) error {
	link, err := s.linkRepo.FindByURL(ctx, url)
	if err != nil {
		return err
	}
	link.IsActive = false
	return s.linkRepo.Update(ctx, link)
}

func (s *AdminService) blockFederationInstance(ctx context.Context, baseURL string) error {
	instance, err := s.instanceRepo.GetByBaseURL(ctx, baseURL)
	if err != nil {
		return err
	}
	instance.Status = "blocked"
	return s.instanceRepo.Update(ctx, instance)
}

func (s *AdminService) activateFederationInstance(ctx context.Context, baseURL string) error {
	instance, err := s.instanceRepo.GetByBaseURL(ctx, baseURL)
	if err != nil {
		return err
	}
	instance.Status = "active"
	return s.instanceRepo.Update(ctx, instance)
}

func fallbackName(name *string, fallback string) string {
	if name == nil {
		return fallback
	}
	value := strings.TrimSpace(*name)
	if value == "" {
		return fallback
	}
	return value
}
