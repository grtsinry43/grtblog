package setupstate

import (
	"context"
	"encoding/json"
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

// Each completed guide is stored as an independent bool key:
//   system.upgrade_guide.2.1.completed = "true"
// This avoids read-modify-write races on a shared JSON array.
const (
	upgradeGuideKeyPrefix = "system.upgrade_guide."
	upgradeGuideKeySuffix = ".completed"
)

// allUpgradeGuideVersions lists every version that ships an upgrade guide.
// To add a guide for a new release, simply append the version string here.
var allUpgradeGuideVersions = []string{"2.1"}

var requiredWebsiteInfoKeys = []string{
	"website_name",
	"public_url",
	"description",
	"keywords",
}

type State struct {
	HasUser                bool
	HasAdmin               bool
	WebsiteInfoReady       bool
	MissingWebsiteInfoKeys []string
	NeedsSetup             bool
	PendingUpgradeGuides   []string
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

	// Only check upgrade guides when initial setup is complete.
	if !state.NeedsSetup {
		state.PendingUpgradeGuides = s.pendingGuides(ctx)
	}

	s.syncMarker(state.NeedsSetup)
	return state, nil
}

func upgradeGuideKey(version string) string {
	return upgradeGuideKeyPrefix + version + upgradeGuideKeySuffix
}

// pendingGuides returns the list of guide versions that have not been completed yet.
func (s *Service) pendingGuides(ctx context.Context) []string {
	keys := make([]string, len(allUpgradeGuideVersions))
	for i, v := range allUpgradeGuideVersions {
		keys[i] = upgradeGuideKey(v)
	}
	items, err := s.sysCfg.ListConfigs(ctx, keys)
	if err != nil {
		return nil // transient error → don't show guide
	}
	completed := make(map[string]bool, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Value) == "true" {
			completed[item.Key] = true
		}
	}
	var pending []string
	for _, v := range allUpgradeGuideVersions {
		if !completed[upgradeGuideKey(v)] {
			pending = append(pending, v)
		}
	}
	return pending
}

// CompleteUpgradeGuide marks a single guide version as completed.
// Each version is an independent key, so no read-modify-write is needed.
func (s *Service) CompleteUpgradeGuide(ctx context.Context, version string) error {
	valRaw := json.RawMessage(`true`)
	vt := "bool"
	_, err := s.sysCfg.UpdateConfigs(ctx, []sysconfig.UpdateItem{
		{Key: upgradeGuideKey(version), Value: &valRaw, ValueType: &vt},
	})
	return err
}

// CompleteAllUpgradeGuides marks every known guide version as completed.
// Used after fresh installation so the admin is not shown guides for features
// they just configured during init.
func (s *Service) CompleteAllUpgradeGuides(ctx context.Context) error {
	valRaw := json.RawMessage(`true`)
	vt := "bool"
	items := make([]sysconfig.UpdateItem, len(allUpgradeGuideVersions))
	for i, v := range allUpgradeGuideVersions {
		items[i] = sysconfig.UpdateItem{
			Key: upgradeGuideKey(v), Value: &valRaw, ValueType: &vt,
		}
	}
	_, err := s.sysCfg.UpdateConfigs(ctx, items)
	return err
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
