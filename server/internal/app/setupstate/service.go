package setupstate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"gorm.io/gorm"
)

const setupMarkerFileName = ".setupdone"

// UpgradeGuideTask is a code-owned task definition. Completion state lives in
// upgrade_guide_state; presentation remains owned by the admin frontend.
type UpgradeGuideTask struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Type     string `json:"type"`
	Title    string `json:"title,omitempty"`
	Required bool   `json:"required,omitempty"`
	Revision int    `json:"revision,omitempty"`
}

var upgradeGuideRegistry = []UpgradeGuideTask{{ID: "2.1-overview-and-features", Version: "2.1", Type: "release-guide", Title: "2.1 版本介绍与功能设置", Revision: 1}}

var ErrUnknownUpgradeGuide = errors.New("unknown upgrade guide")

var requiredWebsiteInfoKeys = []string{
	"website_name",
	"public_url",
	"description",
	"keywords",
}

type State struct {
	HasUser                  bool
	HasAdmin                 bool
	WebsiteInfoReady         bool
	MissingWebsiteInfoKeys   []string
	NeedsSetup               bool
	PendingUpgradeGuideTasks []UpgradeGuideTask
}

type Service struct {
	users  identity.Repository
	sysCfg *sysconfig.Service
	db     *gorm.DB
}

func NewService(users identity.Repository, sysCfg *sysconfig.Service, db ...*gorm.DB) *Service {
	var database *gorm.DB
	if len(db) > 0 {
		database = db[0]
	}
	return &Service{
		users:  users,
		sysCfg: sysCfg,
		db:     database,
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
		state.PendingUpgradeGuideTasks, err = s.pendingGuideTasks(ctx)
		if err != nil {
			return nil, err
		}
	}

	s.syncMarker(state.NeedsSetup)
	return state, nil
}

// pendingGuideTasks derives pending state from the registry revision and persisted decisions.
func (s *Service) pendingGuideTasks(ctx context.Context) ([]UpgradeGuideTask, error) {
	if s.db == nil {
		return append([]UpgradeGuideTask(nil), upgradeGuideRegistry...), nil
	}
	type row struct {
		TaskID   string
		Revision int
		Status   string
	}
	var rows []row
	if err := s.db.WithContext(ctx).Table("upgrade_guide_state").Find(&rows).Error; err != nil {
		return nil, err
	}
	states := make(map[string]row, len(rows))
	for _, r := range rows {
		states[r.TaskID] = r
	}
	pending := make([]UpgradeGuideTask, 0)
	for _, task := range upgradeGuideRegistry {
		state, ok := states[task.ID]
		if !ok || state.Revision < task.Revision || state.Status != "completed" {
			pending = append(pending, task)
		}
	}
	return pending, nil
}

// CompleteUpgradeGuide marks a registered task revision as completed.
func (s *Service) CompleteUpgradeGuide(ctx context.Context, taskID string) error {
	var task UpgradeGuideTask
	known := false
	for _, candidate := range upgradeGuideRegistry {
		if taskID == candidate.ID {
			task = candidate
			known = true
			break
		}
	}
	if !known {
		return ErrUnknownUpgradeGuide
	}
	if s.db == nil {
		return errors.New("upgrade guide state database unavailable")
	}
	return s.db.WithContext(ctx).Exec(`INSERT INTO upgrade_guide_state (task_id, revision, status, selection, decided_at, updated_at) VALUES (?, ?, 'completed', '{}'::jsonb, NOW(), NOW()) ON CONFLICT (task_id) DO UPDATE SET revision=EXCLUDED.revision, status='completed', decided_at=NOW(), updated_at=NOW()`, task.ID, task.Revision).Error
}

// CompleteAllUpgradeGuides marks every guide currently covered by fresh-install setup as completed.
// Used after fresh installation so the admin is not shown guides for features
// they just configured during init.
func (s *Service) CompleteAllUpgradeGuides(ctx context.Context) error {
	for _, task := range upgradeGuideRegistry {
		if err := s.CompleteUpgradeGuide(ctx, task.ID); err != nil {
			return fmt.Errorf("complete %s: %w", task.ID, err)
		}
	}
	return nil
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
