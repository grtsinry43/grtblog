package sysconfig

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

// FederationSettings aggregates federation configuration values.
type FederationSettings struct {
	Enabled         bool
	InstanceName    string
	InstanceURL     string
	PublicKey       string
	PrivateKey      string
	SignatureAlg    string
	RequireHTTPS    bool
	AllowInbound    bool
	AllowOutbound   bool
	DefaultPolicies json.RawMessage
	RateLimits      json.RawMessage
}

// FederationSettings returns aggregated federation config from sys_config.
func (s *Service) FederationSettings(ctx context.Context) (FederationSettings, error) {
	keys := []string{
		"federation.enabled",
		"federation.instanceName",
		"federation.instanceURL",
		"federation.publicKey",
		"federation.privateKey",
		"federation.signatureAlg",
		"federation.requireHTTPS",
		"federation.allowInbound",
		"federation.allowOutbound",
		"federation.defaultPolicies",
		"federation.rateLimits",
	}
	items, err := s.repo.List(ctx, keys)
	if err != nil {
		return FederationSettings{}, err
	}
	lookup := makeLookup(items)

	return FederationSettings{
		Enabled:         cfgParseBool(lookup["federation.enabled"], false),
		InstanceName:    cfgParseString(lookup["federation.instanceName"], ""),
		InstanceURL:     cfgParseString(lookup["federation.instanceURL"], ""),
		PublicKey:       cfgParseString(lookup["federation.publicKey"], ""),
		PrivateKey:      cfgParseString(lookup["federation.privateKey"], ""),
		SignatureAlg:    cfgParseString(lookup["federation.signatureAlg"], "rsa-sha256"),
		RequireHTTPS:    cfgParseBool(lookup["federation.requireHTTPS"], true),
		AllowInbound:    cfgParseBool(lookup["federation.allowInbound"], true),
		AllowOutbound:   cfgParseBool(lookup["federation.allowOutbound"], true),
		DefaultPolicies: cfgParseJSON(lookup["federation.defaultPolicies"], json.RawMessage("{}")),
		RateLimits:      cfgParseJSON(lookup["federation.rateLimits"], json.RawMessage("{}")),
	}, nil
}

// ActivityPubSettings aggregates ActivityPub configuration values.
type ActivityPubSettings struct {
	Enabled                bool
	InstanceName           string
	InstanceURL            string
	ActorUsername          string
	PublicKey              string
	PrivateKey             string
	SignatureAlg           string
	RequireHTTPS           bool
	AllowInbound           bool
	AllowOutbound          bool
	AutoAcceptFollow       bool
	AcceptInboundComment   bool
	MentionToAdmin         bool
	PublishTypes           json.RawMessage
	FediverseReplyTemplate string
}

// ActivityPubSettings returns aggregated ActivityPub config from sys_config.
func (s *Service) ActivityPubSettings(ctx context.Context) (ActivityPubSettings, error) {
	keys := []string{
		"activitypub.enabled",
		"activitypub.instanceName",
		"activitypub.instanceURL",
		"activitypub.actorUsername",
		"activitypub.publicKey",
		"activitypub.privateKey",
		"activitypub.signatureAlg",
		"activitypub.requireHTTPS",
		"activitypub.allowInbound",
		"activitypub.allowOutbound",
		"activitypub.autoAcceptFollow",
		"activitypub.acceptInboundComment",
		"activitypub.mentionToAdmin",
		"activitypub.publishTypes",
		"activitypub.fediverseReplyTemplate",
	}
	items, err := s.repo.List(ctx, keys)
	if err != nil {
		return ActivityPubSettings{}, err
	}
	lookup := makeLookup(items)

	actorUsername := cfgParseString(lookup["activitypub.actorUsername"], "blog")
	if strings.TrimSpace(actorUsername) == "" {
		actorUsername = "blog"
	}

	return ActivityPubSettings{
		Enabled:                cfgParseBool(lookup["activitypub.enabled"], false),
		InstanceName:           cfgParseString(lookup["activitypub.instanceName"], ""),
		InstanceURL:            cfgParseString(lookup["activitypub.instanceURL"], ""),
		ActorUsername:          actorUsername,
		PublicKey:              cfgParseString(lookup["activitypub.publicKey"], ""),
		PrivateKey:             cfgParseString(lookup["activitypub.privateKey"], ""),
		SignatureAlg:           cfgParseString(lookup["activitypub.signatureAlg"], "rsa-sha256"),
		RequireHTTPS:           cfgParseBool(lookup["activitypub.requireHTTPS"], true),
		AllowInbound:           cfgParseBool(lookup["activitypub.allowInbound"], true),
		AllowOutbound:          cfgParseBool(lookup["activitypub.allowOutbound"], true),
		AutoAcceptFollow:       cfgParseBool(lookup["activitypub.autoAcceptFollow"], true),
		AcceptInboundComment:   cfgParseBool(lookup["activitypub.acceptInboundComment"], true),
		MentionToAdmin:         cfgParseBool(lookup["activitypub.mentionToAdmin"], true),
		PublishTypes:           cfgParseJSON(lookup["activitypub.publishTypes"], json.RawMessage(`["article","moment","thinking"]`)),
		FediverseReplyTemplate: cfgParseString(lookup["activitypub.fediverseReplyTemplate"], ""),
	}, nil
}

// --- shared parse helpers ---

func makeLookup(items []domainconfig.SysConfig) map[string]domainconfig.SysConfig {
	m := make(map[string]domainconfig.SysConfig, len(items))
	for _, item := range items {
		m[item.Key] = item
	}
	return m
}

func cfgValueOrDefault(cfg domainconfig.SysConfig) string {
	if strings.TrimSpace(cfg.Value) != "" {
		return cfg.Value
	}
	if cfg.DefaultValue != nil {
		return *cfg.DefaultValue
	}
	return ""
}

func cfgParseString(cfg domainconfig.SysConfig, fallback string) string {
	val := cfgValueOrDefault(cfg)
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}

func cfgParseBool(cfg domainconfig.SysConfig, fallback bool) bool {
	val := cfgValueOrDefault(cfg)
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func cfgParseJSON(cfg domainconfig.SysConfig, fallback json.RawMessage) json.RawMessage {
	val := cfgValueOrDefault(cfg)
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return json.RawMessage(val)
}
