package activitypub

import (
	"encoding/json"
	"testing"
)

func TestRemoteActorUnmarshalFlexibleMedia(t *testing.T) {
	raw := []byte(`{
		"id":"https://mastodon.example/users/alice",
		"preferredUsername":"alice",
		"name":"Alice",
		"inbox":"https://mastodon.example/users/alice/inbox",
		"icon":[{"type":"Image","url":[{"type":"Link","href":"https://cdn.example/avatar.png"}]}],
		"image":{"type":"Image","url":{"href":"https://cdn.example/header.jpg"}},
		"endpoints":{"sharedInbox":"https://mastodon.example/inbox"},
		"publicKey":{"id":"https://mastodon.example/users/alice#main-key","publicKeyPem":"-----BEGIN PUBLIC KEY-----\\nabc\\n-----END PUBLIC KEY-----"}
	}`)

	var actor remoteActor
	if err := json.Unmarshal(raw, &actor); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if actor.ID != "https://mastodon.example/users/alice" {
		t.Fatalf("unexpected actor id: %s", actor.ID)
	}
	if actor.PreferredUsername != "alice" {
		t.Fatalf("unexpected preferred username: %s", actor.PreferredUsername)
	}
	if actor.avatarURL() != "https://cdn.example/avatar.png" {
		t.Fatalf("unexpected avatar url: %s", actor.avatarURL())
	}
	if actor.Inbox != "https://mastodon.example/users/alice/inbox" {
		t.Fatalf("unexpected inbox: %s", actor.Inbox)
	}
	if actor.Endpoints.SharedInbox != "https://mastodon.example/inbox" {
		t.Fatalf("unexpected shared inbox: %s", actor.Endpoints.SharedInbox)
	}
	if actor.PublicKey.PublicKeyPEM == "" {
		t.Fatalf("expected public key pem")
	}
}

func TestParseThemeHeroProfile(t *testing.T) {
	raw := json.RawMessage(`{"home":{"hero":{"avatarUrl":"/uploads/avatar.webp","description":"个人简介"}}}`)
	avatar, description := parseThemeHeroProfile(raw)
	if avatar != "/uploads/avatar.webp" {
		t.Fatalf("unexpected avatar: %s", avatar)
	}
	if description != "个人简介" {
		t.Fatalf("unexpected description: %s", description)
	}
}

func TestBuildActorImageRef(t *testing.T) {
	ref := buildActorImageRef("https://example.com/avatar.svg?x=1")
	if ref == nil {
		t.Fatalf("expected image ref")
	}
	if ref.MediaType != "image/svg+xml" {
		t.Fatalf("unexpected media type: %s", ref.MediaType)
	}
}
