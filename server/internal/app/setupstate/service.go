package setupstate

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
)

const setupMarkerFile = ".setupdone"

var requiredWebsiteInfoKeys = []string{
	"website_name",
	"public_url",
}

type State struct {
	HasUser                bool
	HasAdmin               bool
	WebsiteInfoReady       bool
	MissingWebsiteInfoKeys []string
	NeedsSetup             bool
}

type Service struct {
	users identity.Repository
	site  config.WebsiteInfoRepository
}

func NewService(users identity.Repository, site config.WebsiteInfoRepository) *Service {
	return &Service{
		users: users,
		site:  site,
	}
}

func (s *Service) Evaluate(ctx context.Context) (*State, error) {
	userCount, err := s.users.CountUsers(ctx)
	if err != nil {
		return nil, err
	}
	admins, err := s.users.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.site.List(ctx)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string, len(items))
	for _, item := range items {
		if item.Value == nil {
			continue
		}
		values[item.Key] = strings.TrimSpace(*item.Value)
	}

	missingKeys := make([]string, 0, len(requiredWebsiteInfoKeys))
	for _, key := range requiredWebsiteInfoKeys {
		if strings.TrimSpace(values[key]) == "" {
			missingKeys = append(missingKeys, key)
		}
	}

	state := &State{
		HasUser:                userCount > 0,
		HasAdmin:               len(admins) > 0,
		WebsiteInfoReady:       len(missingKeys) == 0,
		MissingWebsiteInfoKeys: missingKeys,
	}
	state.NeedsSetup = !state.HasUser || !state.HasAdmin || !state.WebsiteInfoReady
	s.syncMarker(state.NeedsSetup)
	return state, nil
}

func (s *Service) syncMarker(needsSetup bool) {
	if needsSetup {
		if err := os.Remove(setupMarkerFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			return
		}
		return
	}
	if _, err := os.Stat(setupMarkerFile); err == nil {
		return
	}
	_ = os.WriteFile(setupMarkerFile, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644)
}
