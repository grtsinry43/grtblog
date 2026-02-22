package ws

import "testing"

type testPresenceResolver struct{}

func (testPresenceResolver) Resolve(contentType string, rawURL string) (PresenceResolvedView, bool) {
	if contentType == "" || rawURL == "" {
		return PresenceResolvedView{}, false
	}
	return PresenceResolvedView{
		ContentType: contentType,
		Title:       "Test",
		URL:         rawURL,
	}, true
}

func TestPresenceHubOnlineDedupByVisitorID(t *testing.T) {
	hub := NewPresenceHub(nil, testPresenceResolver{})
	clientA1 := &Client{}
	clientA2 := &Client{}
	clientB := &Client{}
	clientAnon := &Client{}

	hub.Register(clientA1)
	hub.Register(clientA2)
	hub.Register(clientB)
	hub.Register(clientAnon)

	hub.Identify(clientA1, "visitor-a")
	hub.Identify(clientA2, "visitor-a")
	hub.Identify(clientB, "visitor-b")

	hub.Update(clientA1, PresenceClientPayload{ContentType: "article", URL: "/posts/demo"})
	hub.Update(clientA2, PresenceClientPayload{ContentType: "article", URL: "/posts/demo"})
	hub.Update(clientB, PresenceClientPayload{ContentType: "article", URL: "/posts/demo"})
	hub.Update(clientAnon, PresenceClientPayload{ContentType: "article", URL: "/posts/demo"})

	snapshot := snapshotPresenceHub(hub)
	if snapshot.Online != 3 {
		t.Fatalf("expected online=3 after dedup, got %d", snapshot.Online)
	}
	if len(snapshot.Pages) != 1 {
		t.Fatalf("expected 1 page group, got %d", len(snapshot.Pages))
	}
	if snapshot.Pages[0].Connections != 3 {
		t.Fatalf("expected page connections=3 after dedup, got %d", snapshot.Pages[0].Connections)
	}
}

func TestPresenceHubUpdateCanSetVisitorIDWithoutView(t *testing.T) {
	hub := NewPresenceHub(nil, testPresenceResolver{})
	client1 := &Client{}
	client2 := &Client{}

	hub.Register(client1)
	hub.Register(client2)
	hub.Update(client1, PresenceClientPayload{VisitorID: "visitor-a"})
	hub.Update(client2, PresenceClientPayload{VisitorID: "visitor-a"})

	snapshot := snapshotPresenceHub(hub)
	if snapshot.Online != 1 {
		t.Fatalf("expected online=1 when same visitor connects twice, got %d", snapshot.Online)
	}
	if len(snapshot.Pages) != 0 {
		t.Fatalf("expected no page groups before presence report, got %d", len(snapshot.Pages))
	}
}

func snapshotPresenceHub(hub *PresenceHub) PresenceSnapshotPayload {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	return hub.snapshotLocked()
}
