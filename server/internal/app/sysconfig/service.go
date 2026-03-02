package sysconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/security/turnstile"
)

// Service 负责从数据库读取系统配置并做类型转换。
type Service struct {
	repo             domainconfig.SysConfigRepository
	defaultTurnstile config.TurnstileConfig
	events           appEvent.Bus
}

func NewService(repo domainconfig.SysConfigRepository, defaults config.TurnstileConfig, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{
		repo:             repo,
		defaultTurnstile: defaults,
		events:           events,
	}
}

// Turnstile 返回实时的 Turnstile 配置，优先读取 sys_config，未配置时回退到 env 默认值。
// 约定 key：
// - turnstile.enabled: bool 字符串
// - turnstile.secret: Turnstile Secret
// - turnstile.siteKey: Turnstile Site Key（给前端）
// - turnstile.verifyURL: 覆盖校验端点
// - turnstile.timeoutSeconds: 请求超时秒数
func (s *Service) Turnstile(ctx context.Context) (turnstile.Settings, error) {
	settings := turnstile.Settings{
		Enabled:   s.defaultTurnstile.Enabled,
		Secret:    strings.TrimSpace(s.defaultTurnstile.Secret),
		SiteKey:   "",
		VerifyURL: strings.TrimSpace(s.defaultTurnstile.VerifyURL),
		Timeout:   s.defaultTurnstile.Timeout,
	}

	applyString := func(key string, apply func(string) error) error {
		cfg, err := s.repo.GetByKey(ctx, key)
		if err != nil {
			if err == domainconfig.ErrSysConfigNotFound {
				return nil
			}
			return fmt.Errorf("load %s: %w", key, err)
		}
		val := strings.TrimSpace(cfg.Value)
		if val == "" {
			return nil
		}
		return apply(val)
	}

	if err := applyString("turnstile.enabled", func(val string) error {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("parse bool: %w", err)
		}
		settings.Enabled = b
		return nil
	}); err != nil {
		return settings, err
	}

	_ = applyString("turnstile.secret", func(val string) error {
		settings.Secret = val
		return nil
	})
	_ = applyString("turnstile.siteKey", func(val string) error {
		settings.SiteKey = val
		return nil
	})
	_ = applyString("turnstile.verifyURL", func(val string) error {
		settings.VerifyURL = val
		return nil
	})
	if err := applyString("turnstile.timeoutSeconds", func(val string) error {
		sec, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("parse timeoutSeconds: %w", err)
		}
		if sec > 0 {
			settings.Timeout = time.Duration(sec) * time.Second
		}
		return nil
	}); err != nil {
		return settings, err
	}

	// 如果启用但缺失 Secret，直接报错（fail-closed），避免误配置导致防护失效。
	if settings.Enabled && strings.TrimSpace(settings.Secret) == "" {
		return settings, turnstile.ErrMissingSecret
	}
	return settings, nil
}

type HotArticleThresholds struct {
	Views    int64
	Likes    int64
	Comments int64
}

type CommentSettings struct {
	Disabled          bool
	RequireModeration bool
}

// CommentSettings 返回评论开关配置。
// 约定 key：
// - comment.disabled: 全站禁评
// - comment.requireModeration: 全站评论需要审核
func (s *Service) CommentSettings(ctx context.Context) CommentSettings {
	settings := CommentSettings{
		Disabled:          false,
		RequireModeration: false,
	}
	if cfg, err := s.repo.GetByKey(ctx, "comment.disabled"); err == nil {
		if val, parseErr := strconv.ParseBool(strings.TrimSpace(cfg.Value)); parseErr == nil {
			settings.Disabled = val
		}
	}
	if cfg, err := s.repo.GetByKey(ctx, "comment.requireModeration"); err == nil {
		if val, parseErr := strconv.ParseBool(strings.TrimSpace(cfg.Value)); parseErr == nil {
			settings.RequireModeration = val
		}
	}
	return settings
}

// HotArticleThresholds 返回热门文章判定阈值
func (s *Service) HotArticleThresholds(ctx context.Context) HotArticleThresholds {
	const (
		viewsKey    = "article.hot.views"
		likesKey    = "article.hot.likes"
		commentsKey = "article.hot.comments"
		defaultV    = 100
		defaultL    = 10
		defaultC    = 5
	)

	t := HotArticleThresholds{
		Views:    defaultV,
		Likes:    defaultL,
		Comments: defaultC,
	}

	if cfg, err := s.repo.GetByKey(ctx, viewsKey); err == nil {
		if v, err := strconv.ParseInt(cfg.Value, 10, 64); err == nil {
			t.Views = v
		}
	}
	if cfg, err := s.repo.GetByKey(ctx, likesKey); err == nil {
		if v, err := strconv.ParseInt(cfg.Value, 10, 64); err == nil {
			t.Likes = v
		}
	}
	if cfg, err := s.repo.GetByKey(ctx, commentsKey); err == nil {
		if v, err := strconv.ParseInt(cfg.Value, 10, 64); err == nil {
			t.Comments = v
		}
	}
	return t
}

// UploadMaxSizeBytes 返回上传文件的最大大小（字节），范围 1MB~50MB，默认 50MB。
func (s *Service) UploadMaxSizeBytes(ctx context.Context) int {
	const (
		uploadKey     = "upload.maxSizeMB"
		defaultSizeMB = 50
		minSizeMB     = 1
		maxSizeMB     = 50
	)

	sizeMB := defaultSizeMB
	cfg, err := s.repo.GetByKey(ctx, uploadKey)
	if err == nil {
		val := strings.TrimSpace(cfg.Value)
		if parsed, parseErr := strconv.Atoi(val); parseErr == nil {
			sizeMB = parsed
		}
	}

	if sizeMB < minSizeMB {
		sizeMB = minSizeMB
	}
	if sizeMB > maxSizeMB {
		sizeMB = maxSizeMB
	}
	return sizeMB * 1024 * 1024
}

type WebhookSettings struct {
	Timeout   time.Duration
	Workers   int
	QueueSize int
}

// WebhookSettings 返回 Webhook 发送配置，优先读取 sys_config，未配置时回退默认值。
// 约定 key：
// - webhook.timeoutSeconds: 请求超时秒数
// - webhook.workers: 并发 worker 数
// - webhook.queueSize: 队列长度
func (s *Service) WebhookSettings(ctx context.Context) (WebhookSettings, error) {
	const (
		timeoutKey  = "webhook.timeoutSeconds"
		workersKey  = "webhook.workers"
		queueKey    = "webhook.queueSize"
		defaultSec  = 30
		defaultWork = 4
		defaultQ    = 200
	)

	settings := WebhookSettings{
		Timeout:   time.Duration(defaultSec) * time.Second,
		Workers:   defaultWork,
		QueueSize: defaultQ,
	}

	applyInt := func(key string, apply func(int) error) error {
		cfg, err := s.repo.GetByKey(ctx, key)
		if err != nil {
			if err == domainconfig.ErrSysConfigNotFound {
				return nil
			}
			return fmt.Errorf("load %s: %w", key, err)
		}
		val := strings.TrimSpace(cfg.Value)
		if val == "" {
			return nil
		}
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("parse %s: %w", key, err)
		}
		return apply(parsed)
	}

	if err := applyInt(timeoutKey, func(val int) error {
		if val > 0 {
			settings.Timeout = time.Duration(val) * time.Second
		}
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyInt(workersKey, func(val int) error {
		if val > 0 {
			settings.Workers = val
		}
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyInt(queueKey, func(val int) error {
		if val > 0 {
			settings.QueueSize = val
		}
		return nil
	}); err != nil {
		return settings, err
	}

	return settings, nil
}

const (
	EmailTLSModeNone     = "none"
	EmailTLSModeStartTLS = "starttls"
	EmailTLSModeTLS      = "tls"
)

type EmailSettings struct {
	Enabled      bool
	FromAddress  string
	FromName     string
	DefaultTo    []string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	TLSMode      string
	Timeout      time.Duration
	Workers      int
	QueueSize    int
	MaxRetries   int
}

// EmailSettings 返回邮件发送配置，优先读取 sys_config，未配置时回退默认值。
// 约定 key：
// - email.enabled
// - email.from.address
// - email.from.name
// - email.defaultTo
// - email.smtp.host
// - email.smtp.port
// - email.smtp.username
// - email.smtp.password
// - email.smtp.tlsMode
// - email.send.timeoutSeconds
// - email.send.workers
// - email.send.queueSize
// - email.send.maxRetries
func (s *Service) EmailSettings(ctx context.Context) (EmailSettings, error) {
	settings := EmailSettings{
		Enabled:      false,
		FromAddress:  "",
		FromName:     "",
		DefaultTo:    []string{},
		SMTPHost:     "",
		SMTPPort:     587,
		SMTPUsername: "",
		SMTPPassword: "",
		TLSMode:      EmailTLSModeStartTLS,
		Timeout:      10 * time.Second,
		Workers:      2,
		QueueSize:    200,
		MaxRetries:   3,
	}

	applyString := func(key string, apply func(string) error) error {
		cfg, err := s.repo.GetByKey(ctx, key)
		if err != nil {
			if err == domainconfig.ErrSysConfigNotFound {
				return nil
			}
			return fmt.Errorf("load %s: %w", key, err)
		}
		value := strings.TrimSpace(cfg.Value)
		if value == "" {
			return nil
		}
		return apply(value)
	}
	applyInt := func(key string, apply func(int) error) error {
		return applyString(key, func(value string) error {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("parse %s: %w", key, err)
			}
			return apply(parsed)
		})
	}

	if err := applyString("email.enabled", func(value string) error {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse email.enabled: %w", err)
		}
		settings.Enabled = parsed
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyString("email.from.address", func(value string) error {
		settings.FromAddress = value
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyString("email.from.name", func(value string) error {
		settings.FromName = value
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyString("email.defaultTo", func(value string) error {
		settings.DefaultTo = parseCSVValues(value)
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyString("email.smtp.host", func(value string) error {
		settings.SMTPHost = value
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyInt("email.smtp.port", func(value int) error {
		if value > 0 {
			settings.SMTPPort = value
		}
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyString("email.smtp.username", func(value string) error {
		settings.SMTPUsername = value
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyString("email.smtp.password", func(value string) error {
		settings.SMTPPassword = value
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyString("email.smtp.tlsMode", func(value string) error {
		switch strings.ToLower(value) {
		case EmailTLSModeNone, EmailTLSModeStartTLS, EmailTLSModeTLS:
			settings.TLSMode = strings.ToLower(value)
			return nil
		default:
			return fmt.Errorf("unsupported tls mode: %s", value)
		}
	}); err != nil {
		return settings, err
	}
	if err := applyInt("email.send.timeoutSeconds", func(value int) error {
		if value > 0 {
			settings.Timeout = time.Duration(value) * time.Second
		}
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyInt("email.send.workers", func(value int) error {
		if value > 0 {
			settings.Workers = value
		}
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyInt("email.send.queueSize", func(value int) error {
		if value > 0 {
			settings.QueueSize = value
		}
		return nil
	}); err != nil {
		return settings, err
	}
	if err := applyInt("email.send.maxRetries", func(value int) error {
		if value > 0 {
			settings.MaxRetries = value
		}
		return nil
	}); err != nil {
		return settings, err
	}
	if settings.Enabled && len(settings.DefaultTo) == 0 {
		return settings, fmt.Errorf("email.defaultTo is required when email is enabled")
	}

	return settings, nil
}

func parseCSVValues(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

// EmailSubscriptionBlockedIPs 返回邮件订阅接口的 IP 黑名单。
// 约定 key：
// - email.subscription.blockedIPs: 逗号分隔 IP 列表
func (s *Service) EmailSubscriptionBlockedIPs(ctx context.Context) ([]string, error) {
	cfg, err := s.repo.GetByKey(ctx, "email.subscription.blockedIPs")
	if err != nil {
		if err == domainconfig.ErrSysConfigNotFound {
			return []string{}, nil
		}
		return nil, err
	}
	return parseCSVValues(cfg.Value), nil
}

type UpdateItem struct {
	Key          string
	Value        *json.RawMessage
	IsSensitive  *bool
	GroupPath    *string
	Label        *string
	Description  *string
	ValueType    *string
	EnumOptions  *json.RawMessage
	DefaultValue *json.RawMessage
	VisibleWhen  *json.RawMessage
	Sort         *int
	Meta         *json.RawMessage
}

const (
	valueTypeString = "string"
	valueTypeNumber = "number"
	valueTypeBool   = "bool"
	valueTypeEnum   = "enum"
	valueTypeJSON   = "json"
)

type UpdateValidationError struct {
	Key     string
	Message string
	Cause   error
}

func (e *UpdateValidationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (s *Service) ListConfigs(ctx context.Context, keys []string) ([]domainconfig.SysConfig, error) {
	return s.repo.List(ctx, keys)
}

func (s *Service) UpdateConfigs(ctx context.Context, items []UpdateItem) ([]domainconfig.SysConfig, error) {
	if len(items) == 0 {
		return nil, nil
	}
	uniqueKeys := make(map[string]struct{}, len(items))
	keys := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		if _, ok := uniqueKeys[key]; ok {
			continue
		}
		uniqueKeys[key] = struct{}{}
		keys = append(keys, key)
	}
	existingList, err := s.repo.List(ctx, keys)
	if err != nil {
		return nil, err
	}
	existingMap := make(map[string]domainconfig.SysConfig, len(existingList))
	for _, cfg := range existingList {
		existingMap[cfg.Key] = cfg
	}

	toUpsert := make([]domainconfig.SysConfig, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		current, exists := existingMap[key]
		next := current
		if !exists {
			next = domainconfig.SysConfig{Key: key}
		}
		changed := false

		targetValueType := normalizeValueType(current.ValueType)
		if targetValueType == "" {
			targetValueType = valueTypeString
		}
		if item.ValueType != nil {
			targetValueType = normalizeValueType(*item.ValueType)
			if err := validateValueType(targetValueType); err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "valueType 无效",
					Cause:   err,
				}
			}
			if !exists || targetValueType != current.ValueType {
				next.ValueType = targetValueType
				changed = true
			}
		} else if err := validateValueType(targetValueType); err != nil {
			return nil, &UpdateValidationError{
				Key:     key,
				Message: "valueType 无效",
				Cause:   err,
			}
		} else if !exists {
			next.ValueType = targetValueType
		}

		targetSensitive := current.IsSensitive
		if item.IsSensitive != nil {
			targetSensitive = *item.IsSensitive
			if !exists || targetSensitive != current.IsSensitive {
				changed = true
			}
		} else if !exists {
			targetSensitive = false
		}
		next.IsSensitive = targetSensitive

		if item.GroupPath != nil {
			groupPath := normalizeGroupPath(*item.GroupPath)
			if !exists || groupPath != current.GroupPath {
				next.GroupPath = groupPath
				changed = true
			}
		} else if !exists {
			next.GroupPath = ""
		}

		if item.Label != nil {
			if !exists || *item.Label != current.Label {
				next.Label = *item.Label
				changed = true
			}
		} else if !exists {
			next.Label = ""
		}

		if item.Description != nil {
			if !exists || *item.Description != current.Description {
				next.Description = *item.Description
				changed = true
			}
		} else if !exists {
			next.Description = ""
		}

		enumOptions := current.EnumOptions
		enumValues := []string(nil)
		if item.EnumOptions != nil {
			if targetValueType != valueTypeEnum {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "enumOptions 仅适用于 enum 类型",
				}
			}
			normalized, values, err := normalizeEnumOptions(*item.EnumOptions)
			if err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "enumOptions 无效",
					Cause:   err,
				}
			}
			enumOptions = normalized
			enumValues = values
			next.EnumOptions = enumOptions
			changed = true
		} else if !exists {
			enumOptions = emptyJSONArray
			next.EnumOptions = enumOptions
		} else {
			if len(enumOptions) == 0 {
				enumOptions = emptyJSONArray
			}
			next.EnumOptions = enumOptions
		}

		if targetValueType == valueTypeEnum && len(enumValues) == 0 {
			values, err := extractEnumOptionValues(enumOptions)
			if err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "enumOptions 无效",
					Cause:   err,
				}
			}
			enumValues = values
		}

		if item.VisibleWhen != nil {
			normalized, err := normalizeJSONArray(*item.VisibleWhen)
			if err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "visibleWhen 无效",
					Cause:   err,
				}
			}
			next.VisibleWhen = normalized
			changed = true
		} else if !exists {
			next.VisibleWhen = emptyJSONArray
		} else if len(next.VisibleWhen) == 0 {
			next.VisibleWhen = emptyJSONArray
		}

		if item.Meta != nil {
			normalized, err := normalizeJSONObject(*item.Meta)
			if err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "meta 无效",
					Cause:   err,
				}
			}
			next.Meta = normalized
			changed = true
		} else if !exists {
			next.Meta = emptyJSONObject
		} else if len(next.Meta) == 0 {
			next.Meta = emptyJSONObject
		}

		if item.DefaultValue != nil {
			parsed, err := parseDefaultValueByType(targetValueType, *item.DefaultValue)
			if err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "defaultValue 无效",
					Cause:   err,
				}
			}
			next.DefaultValue = parsed
			changed = true
		} else if !exists {
			next.DefaultValue = nil
		}

		if item.Sort != nil {
			if !exists || *item.Sort != current.Sort {
				next.Sort = *item.Sort
				changed = true
			}
		} else if !exists {
			next.Sort = 0
		}

		valueSet := false
		if item.Value != nil {
			parsed, isEmpty, err := parseValueByType(targetValueType, *item.Value)
			if err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "value 无效",
					Cause:   err,
				}
			}
			if !(targetSensitive && isEmpty) {
				next.Value = parsed
				valueSet = true
				if !exists || parsed != current.Value {
					changed = true
				}
			}
		} else if !exists {
			next.Value = ""
		}

		if item.ValueType != nil && item.Value == nil && exists {
			if err := validateStoredValue(targetValueType, current.Value); err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "value 与 valueType 不匹配",
					Cause:   err,
				}
			}
		}

		if targetValueType == valueTypeEnum && len(enumValues) > 0 {
			checkValue := next.Value
			if !valueSet {
				checkValue = current.Value
			}
			if err := validateEnumValue(enumValues, checkValue); err != nil {
				return nil, &UpdateValidationError{
					Key:     key,
					Message: "value 不在 enumOptions 中",
					Cause:   err,
				}
			}
			if next.DefaultValue != nil {
				if err := validateEnumValue(enumValues, *next.DefaultValue); err != nil {
					return nil, &UpdateValidationError{
						Key:     key,
						Message: "defaultValue 不在 enumOptions 中",
						Cause:   err,
					}
				}
			}
		}

		if !exists && !valueSet {
			continue
		}
		if !changed {
			continue
		}
		next.ValueType = targetValueType
		if len(next.EnumOptions) == 0 {
			next.EnumOptions = emptyJSONArray
		}
		if len(next.VisibleWhen) == 0 {
			next.VisibleWhen = emptyJSONArray
		}
		if len(next.Meta) == 0 {
			next.Meta = emptyJSONObject
		}
		toUpsert = append(toUpsert, next)
	}

	if err := s.repo.Upsert(ctx, toUpsert); err != nil {
		return nil, err
	}
	if len(toUpsert) > 0 {
		updatedKeys := make([]string, 0, len(toUpsert))
		for _, item := range toUpsert {
			updatedKeys = append(updatedKeys, item.Key)
		}
		_ = s.events.Publish(ctx, appEvent.Generic{
			EventName: "sysconfig.updated",
			At:        time.Now(),
			Payload: map[string]any{
				"Keys":  updatedKeys,
				"Count": len(updatedKeys),
			},
		})
		// Auto-generate keypairs when federation/activitypub is enabled
		s.ensureKeyPairs(ctx, updatedKeys)
	}
	return s.repo.List(ctx, nil)
}

var (
	emptyJSONArray  = json.RawMessage("[]")
	emptyJSONObject = json.RawMessage("{}")
)

func normalizeValueType(valueType string) string {
	valueType = strings.TrimSpace(strings.ToLower(valueType))
	if valueType == "" {
		return valueTypeString
	}
	return valueType
}

func validateValueType(valueType string) error {
	switch valueType {
	case valueTypeString, valueTypeNumber, valueTypeBool, valueTypeEnum, valueTypeJSON:
		return nil
	default:
		return fmt.Errorf("unsupported valueType: %s", valueType)
	}
}

func normalizeGroupPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	return path
}

func normalizeJSONArray(raw json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return emptyJSONArray, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(trimmed, &items); err != nil {
		return nil, err
	}
	return append(json.RawMessage(nil), trimmed...), nil
}

func normalizeJSONObject(raw json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return emptyJSONObject, nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &obj); err != nil {
		return nil, err
	}
	return append(json.RawMessage(nil), trimmed...), nil
}

func parseStringValue(raw json.RawMessage) (string, bool, error) {
	var val *string
	if err := json.Unmarshal(raw, &val); err != nil {
		return "", false, err
	}
	if val == nil {
		return "", true, nil
	}
	return *val, *val == "", nil
}

func parseNumberValue(raw json.RawMessage) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var num json.Number
	if err := decoder.Decode(&num); err != nil {
		return "", err
	}
	if _, err := num.Float64(); err != nil {
		return "", err
	}
	return num.String(), nil
}

func parseBoolValue(raw json.RawMessage) (string, error) {
	var val bool
	if err := json.Unmarshal(raw, &val); err != nil {
		return "", err
	}
	return strconv.FormatBool(val), nil
}

func parseValueByType(valueType string, raw json.RawMessage) (string, bool, error) {
	switch valueType {
	case valueTypeString, valueTypeEnum:
		val, isEmpty, err := parseStringValue(raw)
		return val, isEmpty, err
	case valueTypeNumber:
		val, err := parseNumberValue(raw)
		return val, false, err
	case valueTypeBool:
		val, err := parseBoolValue(raw)
		return val, false, err
	case valueTypeJSON:
		trimmed := bytes.TrimSpace(raw)
		if len(trimmed) == 0 {
			return "", false, fmt.Errorf("empty json")
		}
		if !json.Valid(trimmed) {
			return "", false, fmt.Errorf("invalid json")
		}
		return string(trimmed), false, nil
	default:
		return "", false, fmt.Errorf("unsupported valueType: %s", valueType)
	}
}

func parseDefaultValueByType(valueType string, raw json.RawMessage) (*string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}
	switch valueType {
	case valueTypeString, valueTypeEnum:
		val, _, err := parseStringValue(trimmed)
		if err != nil {
			return nil, err
		}
		return &val, nil
	case valueTypeNumber:
		val, err := parseNumberValue(trimmed)
		if err != nil {
			return nil, err
		}
		return &val, nil
	case valueTypeBool:
		val, err := parseBoolValue(trimmed)
		if err != nil {
			return nil, err
		}
		return &val, nil
	case valueTypeJSON:
		if !json.Valid(trimmed) {
			return nil, fmt.Errorf("invalid json")
		}
		val := string(trimmed)
		return &val, nil
	default:
		return nil, fmt.Errorf("unsupported valueType: %s", valueType)
	}
}

func validateStoredValue(valueType string, value string) error {
	switch valueType {
	case valueTypeString, valueTypeEnum:
		return nil
	case valueTypeNumber:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return err
		}
		return nil
	case valueTypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return err
		}
		return nil
	case valueTypeJSON:
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("invalid json")
		}
		return nil
	default:
		return fmt.Errorf("unsupported valueType: %s", valueType)
	}
}

func extractEnumOptionValues(raw json.RawMessage) ([]string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(trimmed, &items); err != nil {
		return nil, err
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		var strValue string
		if err := json.Unmarshal(item, &strValue); err == nil {
			values = append(values, strValue)
			continue
		}
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(item, &obj); err != nil {
			return nil, err
		}
		valueRaw, ok := obj["value"]
		if !ok {
			return nil, fmt.Errorf("enum option missing value")
		}
		if err := json.Unmarshal(valueRaw, &strValue); err != nil {
			return nil, err
		}
		values = append(values, strValue)
	}
	return values, nil
}

func normalizeEnumOptions(raw json.RawMessage) (json.RawMessage, []string, error) {
	normalized, err := normalizeJSONArray(raw)
	if err != nil {
		return nil, nil, err
	}
	values, err := extractEnumOptionValues(normalized)
	if err != nil {
		return nil, nil, err
	}
	return normalized, values, nil
}

func validateEnumValue(values []string, value string) error {
	if len(values) == 0 {
		return nil
	}
	for _, candidate := range values {
		if candidate == value {
			return nil
		}
	}
	return fmt.Errorf("invalid enum value")
}

// GetConfigValue 返回指定 key 的 sys_config 值（字符串）。
// 如果 key 不存在返回 domainconfig.ErrSysConfigNotFound。
func (s *Service) GetConfigValue(ctx context.Context, key string) (string, error) {
	cfg, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return "", err
	}
	return cfg.Value, nil
}
