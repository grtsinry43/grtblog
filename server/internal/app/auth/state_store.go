package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// OAuthState 保存 state 与 PKCE 信息，存储在 Redis。
type OAuthState struct {
	Provider         string    `json:"provider"`
	Redirect         string    `json:"redirect,omitempty"`
	CodeVerifier     string    `json:"code_verifier,omitempty"`
	ContextNonceHash string    `json:"context_nonce_hash,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type StateStore interface {
	Save(ctx context.Context, state string, data OAuthState, ttl time.Duration) error
	Load(ctx context.Context, state string) (*OAuthState, error)
	Delete(ctx context.Context, state string) error
}

const OAuthStateNonceCookieName = "oauth_state_nonce"

type redisStateStore struct {
	client *redis.Client
	prefix string
}

func NewRedisStateStore(client *redis.Client, prefix string) StateStore {
	return &redisStateStore{
		client: client,
		prefix: prefix,
	}
}

func (s *redisStateStore) Save(ctx context.Context, state string, data OAuthState, ttl time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return s.client.Set(ctx, s.key(state), b, ttl).Err()
}

func (s *redisStateStore) Load(ctx context.Context, state string) (*OAuthState, error) {
	val, err := s.client.Get(ctx, s.key(state)).Bytes()
	if err != nil {
		return nil, err
	}
	var data OAuthState
	if err := json.Unmarshal(val, &data); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}
	return &data, nil
}

func (s *redisStateStore) Delete(ctx context.Context, state string) error {
	return s.client.Del(ctx, s.key(state)).Err()
}

func (s *redisStateStore) key(state string) string {
	return s.prefix + "oauth_state:" + state
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateState 返回随机 state。
func GenerateState() (string, error) {
	return randomString(24)
}

// GenerateContextNonce 生成用于绑定 OAuth state 与发起端上下文的一次性随机值。
func GenerateContextNonce() (string, error) {
	return randomString(24)
}

// HashContextNonce 对 context nonce 进行摘要，用于持久化存储。
func HashContextNonce(nonce string) string {
	nonce = strings.TrimSpace(nonce)
	if nonce == "" {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(hashSHA256([]byte(nonce)))
}

// VerifyContextNonce 校验 context nonce 是否与预期摘要一致。
func VerifyContextNonce(nonce, expectedHash string) bool {
	expectedHash = strings.TrimSpace(expectedHash)
	if expectedHash == "" {
		return strings.TrimSpace(nonce) == ""
	}
	actual := HashContextNonce(nonce)
	if actual == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expectedHash)) == 1
}

// GenerateCodeVerifier 生成 PKCE code verifier。
func GenerateCodeVerifier() (string, string, error) {
	verifier, err := randomString(32)
	if err != nil {
		return "", "", err
	}
	challenge := base64.RawURLEncoding.EncodeToString(hashSHA256([]byte(verifier)))
	return verifier, challenge, nil
}

func hashSHA256(data []byte) []byte {
	h := sha256Sum(data)
	return h[:]
}

func sha256Sum(data []byte) [32]byte {
	return sha256.Sum256(data)
}
