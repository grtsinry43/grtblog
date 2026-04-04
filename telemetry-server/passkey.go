package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v2"
)

const (
	sessionTTL         = 7 * 24 * time.Hour
	sessionCookieName  = "telemetry_session"
	ceremonyTTL        = 5 * time.Minute
)

// adminUser implements webauthn.User for the single admin account.
type adminUser struct {
	credentials []webauthn.Credential
}

func (u *adminUser) WebAuthnID() []byte                         { return []byte("admin") }
func (u *adminUser) WebAuthnName() string                       { return "admin" }
func (u *adminUser) WebAuthnDisplayName() string                { return "Admin" }
func (u *adminUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

// PasskeyService manages WebAuthn registration/login and app sessions.
type PasskeyService struct {
	store      *Store
	cfg        Config
	wa         *webauthn.WebAuthn
	hasCredCached atomic.Bool // cache to avoid DB hit on every request
}

func NewPasskeyService(store *Store, cfg Config) *PasskeyService {
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          cfg.WebAuthnRPID,
		RPDisplayName: "GrtBlog Telemetry",
		RPOrigins:     []string{cfg.WebAuthnOrigin},
	})
	if err != nil {
		log.Fatalf("failed to init webauthn: %v", err)
	}

	ps := &PasskeyService{store: store, cfg: cfg, wa: wa}

	// Warm the credential cache.
	if has, _ := store.HasAnyCredential(context.Background()); has {
		ps.hasCredCached.Store(true)
	}

	return ps
}

// RequireSession enforces Passkey session auth.
// If no credentials exist, admin routes return 404 (hidden).
func (ps *PasskeyService) RequireSession() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !ps.hasCredCached.Load() {
			// Re-check DB in case credential was just registered.
			if has, _ := ps.store.HasAnyCredential(c.UserContext()); has {
				ps.hasCredCached.Store(true)
			} else {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
			}
		}

		token := c.Cookies(sessionCookieName)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		valid, err := ps.store.ValidateAdminSession(c.UserContext(), token)
		if err != nil || !valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired session"})
		}

		return c.Next()
	}
}

// --- Registration ---

func (ps *PasskeyService) RegisterBeginHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		if ps.hasCredCached.Load() {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "credential already registered"})
		}
		hasAny, _ := ps.store.HasAnyCredential(ctx)
		if hasAny {
			ps.hasCredCached.Store(true)
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "credential already registered"})
		}

		// Require setup token.
		if ps.cfg.SetupToken == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "SETUP_TOKEN not configured"})
		}
		if c.Get("X-Setup-Token") != ps.cfg.SetupToken {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
		}

		user := &adminUser{}
		creation, sessionData, err := ps.wa.BeginRegistration(user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("begin registration: %v", err)})
		}

		// Store session data for FinishRegistration.
		sessionJSON, _ := json.Marshal(sessionData)
		sessionID := mustRandomHex(16)
		if err := ps.store.SaveWebAuthnSession(ctx, sessionID, sessionJSON, "register", time.Now().Add(ceremonyTTL)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "save session failed"})
		}

		return c.JSON(fiber.Map{
			"publicKey": creation,
			"sessionId": sessionID,
		})
	}
}

func (ps *PasskeyService) RegisterFinishHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		sessionID := c.Get("X-Session-Id")
		if sessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing X-Session-Id header"})
		}

		// Retrieve and delete ceremony session (use-once, type-checked).
		sessionJSON, err := ps.store.GetWebAuthnSession(ctx, sessionID, "register")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired session"})
		}
		_ = ps.store.DeleteWebAuthnSession(ctx, sessionID)

		var sessionData webauthn.SessionData
		if err := json.Unmarshal(sessionJSON, &sessionData); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "corrupt session"})
		}

		fakeReq, _ := http.NewRequest("POST", "/", bytes.NewReader(c.Body()))
		fakeReq.Header.Set("Content-Type", "application/json")

		user := &adminUser{}
		credential, err := ps.wa.FinishRegistration(user, sessionData, fakeReq)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("registration failed: %v", err)})
		}

		// Persist credential.
		credJSON, _ := json.Marshal(credential)
		credID := base64.RawURLEncoding.EncodeToString(credential.ID)
		if err := ps.store.SaveCredentialJSON(ctx, credID, credJSON, "admin"); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "save credential failed"})
		}
		ps.hasCredCached.Store(true)

		// Issue app session.
		token := mustRandomHex(32)
		expiresAt := time.Now().Add(sessionTTL)
		if err := ps.store.SaveAdminSession(ctx, token, expiresAt); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session creation failed"})
	}

		c.Cookie(&fiber.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			Expires:  expiresAt,
			HTTPOnly: true,
			Secure:   ps.isSecureOrigin(),
			SameSite: "Lax",
		})

		return c.JSON(fiber.Map{"status": "registered"})
	}
}

// --- Login ---

func (ps *PasskeyService) LoginBeginHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		if !ps.hasCredCached.Load() {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
		}

		user, err := ps.loadAdminUser(ctx)
		if err != nil || len(user.credentials) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
		}

		assertion, sessionData, err := ps.wa.BeginLogin(user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("begin login: %v", err)})
		}

		sessionJSON, _ := json.Marshal(sessionData)
		sessionID := mustRandomHex(16)
		if err := ps.store.SaveWebAuthnSession(ctx, sessionID, sessionJSON, "login", time.Now().Add(ceremonyTTL)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "save session failed"})
		}

		return c.JSON(fiber.Map{
			"publicKey": assertion,
			"sessionId": sessionID,
		})
	}
}

func (ps *PasskeyService) LoginFinishHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		sessionID := c.Get("X-Session-Id")
		if sessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing X-Session-Id header"})
		}

		// Retrieve and delete ceremony session (use-once, type-checked).
		sessionJSON, err := ps.store.GetWebAuthnSession(ctx, sessionID, "login")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired session"})
		}
		_ = ps.store.DeleteWebAuthnSession(ctx, sessionID)

		var sessionData webauthn.SessionData
		if err := json.Unmarshal(sessionJSON, &sessionData); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "corrupt session"})
		}

		user, err := ps.loadAdminUser(ctx)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "no credentials found"})
		}

		// Build standard http.Request for the library (cryptographic verification happens here).
		fakeReq, _ := http.NewRequest("POST", "/", bytes.NewReader(c.Body()))
		fakeReq.Header.Set("Content-Type", "application/json")

		credential, err := ps.wa.FinishLogin(user, sessionData, fakeReq)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": fmt.Sprintf("login failed: %v", err)})
		}

		// Update credential (sign count etc).
		credJSON, _ := json.Marshal(credential)
		credID := base64.RawURLEncoding.EncodeToString(credential.ID)
		if err := ps.store.SaveCredentialJSON(ctx, credID, credJSON, "admin"); err != nil {
			log.Printf("[passkey] failed to update credential sign count: %v", err)
		}

		// Issue app session.
		token := mustRandomHex(32)
		expiresAt := time.Now().Add(sessionTTL)
		if err := ps.store.SaveAdminSession(ctx, token, expiresAt); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session creation failed"})
	}

		c.Cookie(&fiber.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			Expires:  expiresAt,
			HTTPOnly: true,
			Secure:   ps.isSecureOrigin(),
			SameSite: "Lax",
		})

		return c.JSON(fiber.Map{"status": "authenticated"})
	}
}

// --- Helpers ---

// loadAdminUser reconstructs the admin user with all stored credentials.
func (ps *PasskeyService) loadAdminUser(ctx context.Context) (*adminUser, error) {
	jsons, err := ps.store.ListCredentialJSONs(ctx)
	if err != nil {
		return nil, err
	}
	user := &adminUser{}
	for _, data := range jsons {
		var cred webauthn.Credential
		if err := json.Unmarshal(data, &cred); err != nil {
			continue
		}
		user.credentials = append(user.credentials, cred)
	}
	return user, nil
}

// isSecureOrigin returns true if the configured RP origin uses HTTPS.
func (ps *PasskeyService) isSecureOrigin() bool {
	return strings.HasPrefix(ps.cfg.WebAuthnOrigin, "https://")
}

// mustRandomHex generates n random bytes and returns them as hex string.
// Panics on entropy failure (C3 fix: don't silently ignore rand errors).
func mustRandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return fmt.Sprintf("%x", b)
}
