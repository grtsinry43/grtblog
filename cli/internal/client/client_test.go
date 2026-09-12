package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, New(srv.URL, "gt_testtoken", "test")
}

func writeEnvelope(t *testing.T, w http.ResponseWriter, code int, bizErr, msg string, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":   code,
		"bizErr": bizErr,
		"msg":    msg,
		"data":   json.RawMessage(raw),
	})
}

func TestGetSuccess(t *testing.T) {
	srv, cli := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/auth/profile" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "gt_testtoken" {
			t.Errorf("unexpected auth header: %s", got)
		}
		writeEnvelope(t, w, 0, "OK", "success", map[string]any{
			"id": 1, "username": "grt", "isAdmin": true,
		})
	})
	_ = srv

	var user User
	if err := cli.Get("/auth/profile", nil, &user); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if user.Username != "grt" || !user.IsAdmin {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestBearerTokenForJWT(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer jwttoken" {
			t.Errorf("unexpected auth header: %s", got)
		}
		writeEnvelope(t, w, 0, "OK", "success", nil)
	}))
	defer srv.Close()
	cli := New(srv.URL, "jwttoken", "test")
	if err := cli.Get("/x", nil, nil); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
}

func TestBizError(t *testing.T) {
	_, cli := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeEnvelope(t, w, 401, "NOT_LOGIN", "token 无效", nil)
	})

	err := cli.Get("/auth/profile", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if !apiErr.IsAuth() {
		t.Errorf("expected auth error, got %+v", apiErr)
	}
	if apiErr.Msg != "token 无效" {
		t.Errorf("unexpected msg: %s", apiErr.Msg)
	}
}

func TestPostBody(t *testing.T) {
	_, cli := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["title"] != "hello" {
			t.Errorf("unexpected body: %v", body)
		}
		writeEnvelope(t, w, 0, "OK", "success", map[string]any{"id": 42})
	})

	var out struct {
		ID int `json:"id"`
	}
	if err := cli.Post("/articles", map[string]any{"title": "hello"}, &out); err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if out.ID != 42 {
		t.Errorf("unexpected id: %d", out.ID)
	}
}

func TestQueryParams(t *testing.T) {
	_, cli := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("published") != "true" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		writeEnvelope(t, w, 0, "OK", "success", map[string]any{"items": []any{}, "total": 0})
	})
	var list ArticleList
	if err := cli.Get("/admin/articles", map[string]string{"page": "2", "published": "true"}, &list); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
}

func TestNonJSONError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>bad gateway</html>"))
	}))
	defer srv.Close()
	cli := New(srv.URL, "gt_x", "test")
	err := cli.Get("/x", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.HTTPStatus != http.StatusBadGateway {
		t.Errorf("unexpected status: %d", apiErr.HTTPStatus)
	}
}
