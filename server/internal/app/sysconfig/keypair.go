package sysconfig

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strings"

	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

// ensureKeyPairs checks whether federation/activitypub keys need auto-generation
// after a config update and generates them if needed.
func (s *Service) ensureKeyPairs(ctx context.Context, updatedKeys []string) {
	for _, key := range updatedKeys {
		switch key {
		case "federation.enabled":
			_ = s.ensureKeyPairForPrefix(ctx, "federation")
		case "activitypub.enabled":
			_ = s.ensureKeyPairForPrefix(ctx, "activitypub")
		}
	}
}

func (s *Service) ensureKeyPairForPrefix(ctx context.Context, prefix string) error {
	enabledKey := prefix + ".enabled"
	pubKeyKey := prefix + ".publicKey"
	privKeyKey := prefix + ".privateKey"
	algKey := prefix + ".signatureAlg"

	keys := []string{enabledKey, pubKeyKey, privKeyKey, algKey}
	items, err := s.repo.List(ctx, keys)
	if err != nil {
		return err
	}
	lookup := make(map[string]domainconfig.SysConfig, len(items))
	for _, item := range items {
		lookup[item.Key] = item
	}

	if !cfgParseBool(lookup[enabledKey], false) {
		return nil
	}
	pubKey := cfgParseString(lookup[pubKeyKey], "")
	privKey := cfgParseString(lookup[privKeyKey], "")
	if strings.TrimSpace(pubKey) != "" && strings.TrimSpace(privKey) != "" {
		return nil
	}

	alg := cfgParseString(lookup[algKey], "rsa-sha256")
	pub, priv, err := generateKeyPair(alg)
	if err != nil {
		return err
	}

	pubRaw, err := json.Marshal(pub)
	if err != nil {
		return err
	}
	privRaw, err := json.Marshal(priv)
	if err != nil {
		return err
	}

	_, err = s.UpdateConfigs(ctx, []UpdateItem{
		{Key: pubKeyKey, Value: toRawPtr(pubRaw)},
		{Key: privKeyKey, Value: toRawPtr(privRaw)},
	})
	return err
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

func toRawPtr(raw []byte) *json.RawMessage {
	msg := json.RawMessage(raw)
	return &msg
}
