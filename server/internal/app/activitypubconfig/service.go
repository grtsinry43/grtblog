package activitypubconfig

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strconv"
	"strings"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

type Service struct {
	core *sysconfig.Service
	repo domainconfig.SysConfigRepository
}

func NewService(repo domainconfig.SysConfigRepository) *Service {
	return &Service{
		core: sysconfig.NewService(repo, config.TurnstileConfig{}, nil),
		repo: repo,
	}
}

func (s *Service) ListConfigs(ctx context.Context, keys []string) ([]domainconfig.SysConfig, error) {
	return s.core.ListConfigs(ctx, keys)
}

func (s *Service) UpdateConfigs(ctx context.Context, items []sysconfig.UpdateItem) ([]domainconfig.SysConfig, error) {
	updated, err := s.core.UpdateConfigs(ctx, items)
	if err != nil {
		return nil, err
	}
	settings, err := s.Settings(ctx)
	if err != nil {
		return updated, err
	}
	if !settings.Enabled {
		return updated, nil
	}
	keyUpdates, err := s.ensureKeyPairUpdates(settings)
	if err != nil {
		return updated, err
	}
	if len(keyUpdates) == 0 {
		return updated, nil
	}
	keyUpdated, err := s.core.UpdateConfigs(ctx, keyUpdates)
	if err != nil {
		return updated, err
	}
	updated = append(updated, keyUpdated...)
	return updated, nil
}

type Settings struct {
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

func (s *Service) Settings(ctx context.Context) (Settings, error) {
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
		return Settings{}, err
	}
	lookup := make(map[string]domainconfig.SysConfig, len(items))
	for _, item := range items {
		lookup[item.Key] = item
	}
	actorUsername := parseString(lookup["activitypub.actorUsername"], "blog")
	if strings.TrimSpace(actorUsername) == "" {
		actorUsername = "blog"
	}
	return Settings{
		Enabled:                parseBool(lookup["activitypub.enabled"], false),
		InstanceName:           parseString(lookup["activitypub.instanceName"], ""),
		InstanceURL:            parseString(lookup["activitypub.instanceURL"], ""),
		ActorUsername:          actorUsername,
		PublicKey:              parseString(lookup["activitypub.publicKey"], ""),
		PrivateKey:             parseString(lookup["activitypub.privateKey"], ""),
		SignatureAlg:           parseString(lookup["activitypub.signatureAlg"], "rsa-sha256"),
		RequireHTTPS:           parseBool(lookup["activitypub.requireHTTPS"], true),
		AllowInbound:           parseBool(lookup["activitypub.allowInbound"], true),
		AllowOutbound:          parseBool(lookup["activitypub.allowOutbound"], true),
		AutoAcceptFollow:       parseBool(lookup["activitypub.autoAcceptFollow"], true),
		AcceptInboundComment:   parseBool(lookup["activitypub.acceptInboundComment"], true),
		MentionToAdmin:         parseBool(lookup["activitypub.mentionToAdmin"], true),
		PublishTypes:           parseJSON(lookup["activitypub.publishTypes"], json.RawMessage("[\"article\",\"moment\",\"thinking\"]")),
		FediverseReplyTemplate: parseString(lookup["activitypub.fediverseReplyTemplate"], ""),
	}, nil
}

func parseString(cfg domainconfig.SysConfig, fallback string) string {
	val := valueOrDefault(cfg)
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}

func parseBool(cfg domainconfig.SysConfig, fallback bool) bool {
	val := valueOrDefault(cfg)
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseJSON(cfg domainconfig.SysConfig, fallback json.RawMessage) json.RawMessage {
	val := valueOrDefault(cfg)
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return json.RawMessage(val)
}

func valueOrDefault(cfg domainconfig.SysConfig) string {
	if strings.TrimSpace(cfg.Value) != "" {
		return cfg.Value
	}
	if cfg.DefaultValue != nil {
		return *cfg.DefaultValue
	}
	return ""
}

func (s *Service) ensureKeyPairUpdates(settings Settings) ([]sysconfig.UpdateItem, error) {
	if strings.TrimSpace(settings.PublicKey) != "" && strings.TrimSpace(settings.PrivateKey) != "" {
		return nil, nil
	}
	alg := strings.TrimSpace(settings.SignatureAlg)
	if alg == "" {
		alg = "rsa-sha256"
	}
	pub, priv, err := generateKeyPair(alg)
	if err != nil {
		return nil, err
	}
	pubRaw, err := json.Marshal(pub)
	if err != nil {
		return nil, err
	}
	privRaw, err := json.Marshal(priv)
	if err != nil {
		return nil, err
	}
	return []sysconfig.UpdateItem{
		{Key: "activitypub.publicKey", Value: toRaw(pubRaw)},
		{Key: "activitypub.privateKey", Value: toRaw(privRaw)},
	}, nil
}

func generateRSAKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateBytes})
	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicBytes})
	return string(publicPEM), string(privatePEM), nil
}

func generateEd25519KeyPair() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	privatePKCS8, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privatePKCS8})
	publicPKIX, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", "", err
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicPKIX})
	return string(publicPEM), string(privatePEM), nil
}

func generateKeyPair(algorithm string) (string, string, error) {
	switch strings.ToLower(strings.TrimSpace(algorithm)) {
	case "rsa-sha256", "rsa_sha256":
		return generateRSAKeyPair()
	case "ed25519":
		return generateEd25519KeyPair()
	default:
		return "", "", errors.New("不支持的签名算法")
	}
}

func toRaw(raw []byte) *json.RawMessage {
	msg := json.RawMessage(raw)
	return &msg
}
