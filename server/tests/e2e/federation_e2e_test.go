//go:build federation_e2e

package e2e_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/database"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
	"github.com/grtsinry43/grtblog-v2/server/internal/security/jwt"
	appserver "github.com/grtsinry43/grtblog-v2/server/internal/server"
)

const (
	adminPrefix = "/api/v2/admin"
	siteAURL    = "https://192.0.2.10"
	siteBURL    = "https://198.51.100.20"
)

type testSite struct {
	name      string
	baseURL   string
	token     string
	username  string
	articleID int64
	db        *gorm.DB
	server    *appserver.Server
	app       *fiber.App
}

type appTransport struct {
	mu   sync.RWMutex
	apps map[string]*fiber.App
}

func newAppTransport() *appTransport {
	return &appTransport{apps: make(map[string]*fiber.App)}
}

func (r *appTransport) register(rawURL string, app *fiber.App) {
	parsed, _ := url.Parse(rawURL)
	r.mu.Lock()
	r.apps[parsed.Host] = app
	r.mu.Unlock()
}

func (r *appTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		return nil, fmt.Errorf("test federation transport only permits HTTPS")
	}
	r.mu.RLock()
	app := r.apps[req.URL.Host]
	r.mu.RUnlock()
	if app == nil {
		return nil, fmt.Errorf("test federation transport rejected unknown host %q", req.URL.Host)
	}
	cloned := req.Clone(req.Context())
	cloned.RequestURI = ""
	return app.Test(cloned, -1)
}

type envelope[T any] struct {
	Code   int    `json:"code"`
	BizErr string `json:"bizErr"`
	Msg    string `json:"msg"`
	Data   T      `json:"data"`
}

type proxyResponse struct {
	RequestID  string `json:"request_id"`
	DeliveryID int64  `json:"delivery_id"`
	StatusCode int    `json:"status_code"`
}

type manifest struct {
	ProtocolVersion string `json:"protocol_version"`
	Instance        struct {
		URL string `json:"url"`
	} `json:"instance"`
	Features []string `json:"features"`
}

type publicKeyDoc struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
}

type endpointsDoc struct {
	BaseURL   string            `json:"base_url"`
	Endpoints map[string]string `json:"endpoints"`
}

func TestFederationBetweenTwoIsolatedInstances(t *testing.T) {
	dsnA := strings.TrimSpace(os.Getenv("FEDERATION_E2E_DB_DSN_A"))
	dsnB := strings.TrimSpace(os.Getenv("FEDERATION_E2E_DB_DSN_B"))
	if dsnA == "" || dsnB == "" {
		t.Skip("set FEDERATION_E2E_DB_DSN_A and FEDERATION_E2E_DB_DSN_B to migrated temporary PostgreSQL databases")
	}

	transport := newAppTransport()
	federationClient := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	a := newTestSite(t, "A", siteAURL, dsnA, federationClient)
	b := newTestSite(t, "B", siteBURL, dsnB, federationClient)
	transport.register(a.baseURL, a.app)
	transport.register(b.baseURL, b.app)

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = a.server.Shutdown(shutdownCtx)
		_ = b.server.Shutdown(shutdownCtx)
		closeDB(a.db)
		closeDB(b.db)
	})

	t.Run("bilateral discovery and security", func(t *testing.T) {
		assertWellKnown(t, a)
		assertWellKnown(t, b)
		assertAdminDiscovery(t, a, b)
		assertAdminDiscovery(t, b, a)
		assertUnsignedCitationRejected(t, a)
		assertUnsignedCitationRejected(t, b)
	})

	t.Run("A to B", func(t *testing.T) { runDirection(t, a, b) })
	t.Run("B to A", func(t *testing.T) { runDirection(t, b, a) })
}

func newTestSite(t *testing.T, name, baseURL, dsn string, federationClient *http.Client) *testSite {
	t.Helper()
	cfg := config.Config{
		App: config.AppConfig{
			Name:                "grtblog-e2e-" + strings.ToLower(name),
			Env:                 "test",
			Port:                "0",
			HTMLSnapshotBaseURL: "http://127.0.0.1:9",
		},
		Database: config.DatabaseConfig{Driver: "postgres", DSN: dsn},
		Auth: config.AuthConfig{
			Secret:    "federation-e2e-secret-" + name,
			Issuer:    "federation-e2e-" + name,
			AccessTTL: time.Hour,
		},
		Turnstile: config.TurnstileConfig{Enabled: false},
		Redis:     config.RedisConfig{Prefix: "e2e:" + strings.ToLower(name) + ":"},
	}
	db, err := database.New(cfg.Database)
	if err != nil {
		t.Fatalf("open %s database: %v", name, err)
	}

	ctx := context.Background()
	publicKey, privateKey := generateRSAKeyPair(t)
	setFederationConfig(t, db, map[string]string{
		"federation.enabled":         "true",
		"federation.instanceName":    "E2E Site " + name,
		"federation.instanceURL":     baseURL,
		"federation.publicKey":       publicKey,
		"federation.privateKey":      privateKey,
		"federation.signatureAlg":    "rsa-sha256",
		"federation.requireHTTPS":    "true",
		"federation.allowInbound":    "true",
		"federation.allowOutbound":   "true",
		"federation.defaultPolicies": `{"allow_citation":true,"allow_mention":true,"auto_approve_friendlink":false,"auto_approve_friendlink_citation":false}`,
		"federation.rateLimits":      `{}`,
	})

	username := "admin_" + strings.ToLower(name)
	user := &identity.User{
		Username: username,
		Nickname: "Admin " + name,
		Email:    strings.ToLower(name) + "@e2e.invalid",
		IsActive: true,
		IsAdmin:  true,
	}
	if err := persistence.NewIdentityRepository(db).Create(ctx, user); err != nil {
		t.Fatalf("create %s admin: %v", name, err)
	}
	article := &content.Article{
		Title:            "Federation E2E Article " + name,
		Summary:          "isolated federation test article",
		Content:          "# Federation E2E\n\nSite " + name,
		ContentHash:      "e2e-content-" + strings.ToLower(name),
		AuthorID:         user.ID,
		ShortURL:         "federation-e2e-" + strings.ToLower(name),
		IsPublished:      true,
		IsOriginal:       true,
		ExtInfo:          []byte(`{}`),
		ContentUpdatedAt: time.Now().UTC(),
		CreatedAt:        time.Now().UTC(),
	}
	if err := persistence.NewContentRepository(db).CreateArticle(ctx, article); err != nil {
		t.Fatalf("create %s article: %v", name, err)
	}

	manager := jwt.NewManager(cfg.Auth)
	token, _, err := manager.Generate(user.ID, true)
	if err != nil {
		t.Fatalf("generate %s token: %v", name, err)
	}
	srv := appserver.NewWithOptions(cfg, db, appserver.Options{
		FederationHTTPClient: federationClient,
		DisableRedis:         true,
	})
	return &testSite{
		name: name, baseURL: baseURL, token: token, username: username,
		articleID: article.ID, db: db, server: srv, app: srv.App(),
	}
}

func runDirection(t *testing.T, source, target *testSite) {
	t.Helper()
	t.Run("citation review and callback", func(t *testing.T) {
		var sent proxyResponse
		adminJSON(t, source, http.MethodPost, "/federation/citations/request", map[string]any{
			"target_instance_url": target.baseURL,
			"target_post_id":      strconv.FormatInt(target.articleID, 10),
			"source_article_id":   source.articleID,
			"citation_context":    "isolated e2e citation",
			"citation_type":       "reference",
		}, &sent)
		assertDispatchAccepted(t, sent)

		var inbound model.FederatedCitation
		mustFind(t, target.db, &inbound, "source_request_id = ?", sent.RequestID)
		if inbound.Status != "pending" || inbound.TargetArticleID != target.articleID {
			t.Fatalf("invalid inbound citation: %+v", inbound)
		}
		adminJSON(t, target, http.MethodPut, fmt.Sprintf("/federation/citations/%d/review", inbound.ID), map[string]string{"status": "approved"}, nil)
		out := mustOutbound(t, source.db, sent.RequestID)
		if out.Status != "approved" || out.LastCallbackAt == nil {
			t.Fatalf("citation callback did not approve source delivery: %+v", out)
		}
		callbackAt := *out.LastCallbackAt
		adminJSON(t, target, http.MethodPut, fmt.Sprintf("/federation/citations/%d/review", inbound.ID), map[string]string{"status": "approved"}, nil)
		out = mustOutbound(t, source.db, sent.RequestID)
		if out.LastCallbackAt == nil || !out.LastCallbackAt.Equal(callbackAt) {
			t.Fatalf("duplicate citation review emitted another callback")
		}
	})

	t.Run("mention review and callback", func(t *testing.T) {
		var sent proxyResponse
		adminJSON(t, source, http.MethodPost, "/federation/mentions/notify", map[string]any{
			"target_instance_url": target.baseURL,
			"mentioned_user":      target.username,
			"source_article_id":   source.articleID,
			"mention_context":     "isolated e2e mention",
			"mention_type":        "discussion",
		}, &sent)
		assertDispatchAccepted(t, sent)

		var inbound model.FederatedMention
		mustFind(t, target.db, &inbound, "source_request_id = ?", sent.RequestID)
		if inbound.Status != "pending" || inbound.MentionContext == "" {
			t.Fatalf("invalid inbound mention: %+v", inbound)
		}
		adminJSON(t, target, http.MethodPut, fmt.Sprintf("/federation/mentions/%d/review", inbound.ID), map[string]string{"status": "approved"}, nil)
		out := mustOutbound(t, source.db, sent.RequestID)
		if out.Status != "approved" || out.LastCallbackAt == nil {
			t.Fatalf("mention callback did not approve source delivery: %+v", out)
		}
	})

	t.Run("friend link handshake and initial timeline sync", func(t *testing.T) {
		var sent proxyResponse
		adminJSON(t, source, http.MethodPost, "/friend-links/federation/request", map[string]string{
			"target_url": target.baseURL,
			"message":    "isolated e2e friend link",
		}, &sent)
		assertDispatchAccepted(t, sent)

		var application model.FriendLinkApplication
		mustFind(t, target.db, &application, "source_request_id = ?", sent.RequestID)
		if application.Status != "pending" || application.ApplyChannel != "federation" || !application.SignatureVerified {
			t.Fatalf("invalid inbound friend-link application: %+v", application)
		}
		adminJSON(t, target, http.MethodPut, fmt.Sprintf("/friend-links/applications/%d/approve", application.ID), nil, nil)
		out := mustOutbound(t, source.db, sent.RequestID)
		if out.Status != "approved" || out.LastCallbackAt == nil {
			t.Fatalf("friend-link callback did not approve source delivery: %+v", out)
		}

		waitFor(t, 5*time.Second, func() (bool, error) {
			var link model.FriendLink
			if err := source.db.Where("url = ? AND type = ? AND is_active = true", target.baseURL, "federation").First(&link).Error; err != nil {
				return false, nil
			}
			if link.InstanceID == nil {
				return false, nil
			}
			var cached int64
			err := source.db.Model(&model.FederatedPostCache{}).
				Where("friend_link_id = ? AND instance_id = ?", link.ID, *link.InstanceID).
				Count(&cached).Error
			return cached > 0, err
		})
	})
}

func assertWellKnown(t *testing.T, site *testSite) {
	t.Helper()
	var m manifest
	publicJSON(t, site, "/.well-known/blog-federation/manifest.json", &m)
	if m.ProtocolVersion == "" || m.Instance.URL != site.baseURL || !contains(m.Features, "cross-citation") || !contains(m.Features, "cross-mention") {
		t.Fatalf("%s invalid federation manifest: %+v", site.name, m)
	}
	var key publicKeyDoc
	publicJSON(t, site, "/.well-known/blog-federation/public-key.json", &key)
	if key.Algorithm != "rsa-sha256" || !strings.Contains(key.PublicKey, "PUBLIC KEY") || !strings.HasPrefix(key.KeyID, site.baseURL) {
		t.Fatalf("%s invalid public key document", site.name)
	}
	var endpoints endpointsDoc
	publicJSON(t, site, "/.well-known/blog-federation/endpoints.json", &endpoints)
	for _, name := range []string{"friendlink_request", "timeline", "post_detail", "citation_request", "mention_notify", "outbound_result"} {
		if endpoints.Endpoints[name] == "" {
			t.Fatalf("%s endpoints document is missing %q", site.name, name)
		}
	}
}

func assertAdminDiscovery(t *testing.T, source, target *testSite) {
	t.Helper()
	var discovered struct {
		Manifest  map[string]any `json:"manifest"`
		PublicKey map[string]any `json:"public_key"`
		Endpoints map[string]any `json:"endpoints"`
	}
	adminJSON(t, source, http.MethodGet, "/federation/remote/check?target_url="+url.QueryEscape(target.baseURL), nil, &discovered)
	if len(discovered.Manifest) == 0 || len(discovered.PublicKey) == 0 || len(discovered.Endpoints) == 0 {
		t.Fatalf("%s did not fully discover %s", source.name, target.name)
	}
}

func assertUnsignedCitationRejected(t *testing.T, target *testSite) {
	t.Helper()
	req := newRequest(t, http.MethodPost, target.baseURL+"/api/federation/citations/request", map[string]any{
		"source_instance_url": siteAURL,
		"source_post":         map[string]string{"url": siteAURL + "/posts/forged"},
		"target_post_id":      strconv.FormatInt(target.articleID, 10),
		"citation_context":    "forged",
	})
	resp := appResponse(t, target.app, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s accepted unsigned citation: HTTP %d %s", target.name, resp.StatusCode, raw)
	}
	var count int64
	if err := target.db.Model(&model.FederatedCitation{}).Where("citation_context = ?", "forged").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("unsigned citation changed database: count=%d err=%v", count, err)
	}
}

func adminJSON(t *testing.T, site *testSite, method, path string, body, dst any) {
	t.Helper()
	req := newRequest(t, method, site.baseURL+adminPrefix+path, body)
	req.Header.Set("Authorization", "Bearer "+site.token)
	decodeEnvelope(t, appResponse(t, site.app, req), dst)
}

func publicJSON(t *testing.T, site *testSite, path string, dst any) {
	t.Helper()
	req := newRequest(t, http.MethodGet, site.baseURL+path, nil)
	resp := appResponse(t, site.app, req)
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET %s returned HTTP %d: %s", path, resp.StatusCode, raw)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode GET %s: %v", path, err)
	}
}

func decodeEnvelope(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("request returned HTTP %d: %s", resp.StatusCode, raw)
	}
	var wrapped envelope[json.RawMessage]
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		t.Fatalf("decode response envelope: %v; body=%s", err, raw)
	}
	if wrapped.Code != 0 || wrapped.BizErr != "OK" {
		t.Fatalf("business error code=%d bizErr=%s msg=%s", wrapped.Code, wrapped.BizErr, wrapped.Msg)
	}
	if dst != nil && len(wrapped.Data) > 0 && string(wrapped.Data) != "null" {
		if err := json.Unmarshal(wrapped.Data, dst); err != nil {
			t.Fatalf("decode response data: %v; data=%s", err, wrapped.Data)
		}
	}
}

func newRequest(t *testing.T, method, endpoint string, body any) *http.Request {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, endpoint, reader)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return req
}

func appResponse(t *testing.T, app *fiber.App, req *http.Request) *http.Response {
	t.Helper()
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL, err)
	}
	return resp
}

func setFederationConfig(t *testing.T, db *gorm.DB, values map[string]string) {
	t.Helper()
	for key, value := range values {
		result := db.Model(&model.SysConfig{}).Where("config_key = ?", key).Update("value", value)
		if result.Error != nil || result.RowsAffected != 1 {
			t.Fatalf("update %s: rows=%d err=%v; did all migrations run?", key, result.RowsAffected, result.Error)
		}
	}
}

func generateRSAKeyPair(t *testing.T) (string, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal RSA public key: %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	return string(publicPEM), string(privatePEM)
}

func assertDispatchAccepted(t *testing.T, sent proxyResponse) {
	t.Helper()
	if sent.RequestID == "" || sent.DeliveryID <= 0 || sent.StatusCode < 200 || sent.StatusCode >= 300 {
		t.Fatalf("dispatch was not accepted: %+v", sent)
	}
}

func mustFind(t *testing.T, db *gorm.DB, dst any, query string, args ...any) {
	t.Helper()
	if err := db.Where(query, args...).First(dst).Error; err != nil {
		t.Fatalf("find database record: %v", err)
	}
}

func mustOutbound(t *testing.T, db *gorm.DB, requestID string) model.OutboundDelivery {
	t.Helper()
	var result model.OutboundDelivery
	mustFind(t, db, &result, "request_id = ?", requestID)
	return result
}

func waitFor(t *testing.T, timeout time.Duration, check func() (bool, error)) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ok, err := check()
		if err != nil {
			t.Fatalf("poll federation state: %v", err)
		}
		if ok {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("timed out after %s waiting for federation state", timeout)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func closeDB(db *gorm.DB) {
	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
