package setupstate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
)

const setupMarkerFileName = ".setupdone"

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
	users  identity.Repository
	sysCfg *sysconfig.Service
}

func NewService(users identity.Repository, sysCfg *sysconfig.Service) *Service {
	return &Service{
		users:  users,
		sysCfg: sysCfg,
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

	missingKeys := make([]string, 0, len(requiredWebsiteInfoKeys))
	for _, key := range requiredWebsiteInfoKeys {
		val, err := s.sysCfg.GetWebsiteInfoValue(ctx, key)
		if err != nil {
			if errors.Is(err, domainconfig.ErrSysConfigNotFound) {
				missingKeys = append(missingKeys, key)
				continue
			}
			return nil, err
		}
		if strings.TrimSpace(val) == "" {
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
	path := filepath.Join("storage", setupMarkerFileName)
	if needsSetup {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return
		}
		return
	}
	if _, err := os.Stat(path); err == nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644)
}
