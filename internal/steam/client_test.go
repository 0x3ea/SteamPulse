package steam

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_RequiresAPIKey(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("want error for empty API key, got nil")
	}
}

func TestGetPlayerSummaries_HappyPath(t *testing.T) {
	const payload = `{"response":{"players":[
		{"steamid":"76561198000000001","personaname":"tester",
		 "profileurl":"https://steamcommunity.com/id/tester/",
		 "avatarfull":"https://example.com/a.jpg",
		 "communityvisibilitystate":3,"personastate":1,"timecreated":1000000000}
	]}}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("key"); got != "k" {
			t.Errorf("key param: want k, got %q", got)
		}
		if got := r.URL.Query().Get("steamids"); got != "76561198000000001" {
			t.Errorf("steamids param: want 76561198000000001, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	c, err := New("k", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	players, err := c.GetPlayerSummaries(context.Background(), []string{"76561198000000001"})
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	if len(players) != 1 {
		t.Fatalf("want 1 player, got %d", len(players))
	}
	p := players[0]
	if p.PersonaName != "tester" || !p.IsPublic() || p.TimeCreated != 1000000000 {
		t.Errorf("unexpected player: %+v", p)
	}
}

// Private/unknown profile → HTTP 200 with empty players (NOT an error).
func TestGetPlayerSummaries_PrivateProfileEmptyNotError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer srv.Close()

	c, _ := New("k", WithBaseURL(srv.URL))
	players, err := c.GetPlayerSummaries(context.Background(), []string{"76561198000000002"})
	if err != nil {
		t.Fatalf("empty players must not be an error, got %v", err)
	}
	if len(players) != 0 {
		t.Fatalf("want 0 players, got %d", len(players))
	}
}

// Bad key → HTTP 401 → error.
func TestGetPlayerSummaries_BadKeyIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"response":{"error":{"code":401,"message":"Access denied"}}}`))
	}))
	defer srv.Close()

	c, _ := New("bad", WithBaseURL(srv.URL))
	if _, err := c.GetPlayerSummaries(context.Background(), []string{"1"}); err == nil {
		t.Fatal("want error for 401 bad key, got nil")
	}
}

func TestGetPlayerSummaries_EmptyIDsSkipsCall(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	c, _ := New("k", WithBaseURL(srv.URL))
	players, err := c.GetPlayerSummaries(context.Background(), nil)
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	if players != nil {
		t.Errorf("want nil players for empty input, got %v", players)
	}
	if called {
		t.Error("expected no HTTP call for empty input")
	}
}
