package core

import (
	"context"
	"errors"
	"testing"

	"github.com/0x3ea/SteamPulse/internal/steam"
)

// fakeSteam is a stand-in SteamClient for testing core without the network.
// It returns the canned players/games and the configured err.
type fakeSteam struct {
	players []steam.PlayerSummary
	games   []steam.Game
	err     error
}

func (f fakeSteam) GetPlayerSummaries(ctx context.Context, ids []string) ([]steam.PlayerSummary, error) {
	return f.players, f.err
}

func (f fakeSteam) GetPlayerOwnedGames(ctx context.Context, steamID string) ([]steam.Game, error) {
	return f.games, f.err
}

func TestGetProfile_HappyPath(t *testing.T) {
	svc := NewService(fakeSteam{players: []steam.PlayerSummary{
		{SteamID: "1", PersonaName: "tester", CommunityVisibilityState: 3,
			PersonaState: 1, TimeCreated: 100, LocCountryCode: "CN"},
	}})
	p, err := svc.GetProfile(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if !p.IsPublic || !p.Online || p.PersonaName != "tester" || p.CountryCode != "CN" {
		t.Errorf("unexpected profile: %+v", p)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	svc := NewService(fakeSteam{}) // empty players → private/nonexistent
	_, err := svc.GetProfile(context.Background(), "1")
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("want ErrProfileNotFound, got %v", err)
	}
}

func TestGetProfile_PropagatesSteamError(t *testing.T) {
	svc := NewService(fakeSteam{err: errors.New("boom")})
	if _, err := svc.GetProfile(context.Background(), "1"); err == nil {
		t.Fatal("want error propagated from steam, got nil")
	}
}

// TestGetProfile_AggregatesGames checks the game-library fields that GetProfile
// derives via summarizeGames: TotalGames, TotalPlaytime (hours), and TopGames
// ordered most-played-first.
func TestGetProfile_AggregatesGames(t *testing.T) {
	svc := NewService(fakeSteam{
		players: []steam.PlayerSummary{{SteamID: "1", CommunityVisibilityState: 3}},
		games: []steam.Game{
			{Name: "C", PlaytimeForever: 60},
			{Name: "A", PlaytimeForever: 240}, // most played
			{Name: "B", PlaytimeForever: 120},
		},
	})
	p, err := svc.GetProfile(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if p.TotalGames != 3 {
		t.Errorf("TotalGames: want 3, got %d", p.TotalGames)
	}
	if p.TotalPlaytime != 7 { // (240+120+60)/60
		t.Errorf("TotalPlaytime: want 7 hours, got %d", p.TotalPlaytime)
	}
	if len(p.TopGames) != 3 || p.TopGames[0].Name != "A" {
		t.Errorf("TopGames not ordered by playtime desc: %+v", p.TopGames)
	}
}

// TestGetProfile_TopGamesCappedAtFive: with more than 5 games, TopGames holds
// only the 5 most-played, while TotalGames still counts all of them.
func TestGetProfile_TopGamesCappedAtFive(t *testing.T) {
	games := []steam.Game{
		{Name: "g1", PlaytimeForever: 600},
		{Name: "g2", PlaytimeForever: 500},
		{Name: "g3", PlaytimeForever: 400},
		{Name: "g4", PlaytimeForever: 300},
		{Name: "g5", PlaytimeForever: 200},
		{Name: "g6", PlaytimeForever: 100},
		{Name: "g7", PlaytimeForever: 50},
	}
	svc := NewService(fakeSteam{
		players: []steam.PlayerSummary{{SteamID: "1", CommunityVisibilityState: 3}},
		games:   games,
	})
	p, err := svc.GetProfile(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if p.TotalGames != 7 {
		t.Errorf("TotalGames: want 7, got %d", p.TotalGames)
	}
	if len(p.TopGames) != 5 {
		t.Fatalf("TopGames len: want 5, got %d", len(p.TopGames))
	}
	want := []string{"g1", "g2", "g3", "g4", "g5"}
	for i, w := range want {
		if p.TopGames[i].Name != w {
			t.Errorf("TopGames[%d]: want %s, got %s", i, w, p.TopGames[i].Name)
		}
	}
}

// TestSummarizeGames covers the pure aggregation directly: minute→hour
// conversion, descending order, the n cap, and the empty/under-n edges.
func TestSummarizeGames(t *testing.T) {
	tests := []struct {
		name         string
		games        []steam.Game
		n            int
		wantHours    int
		wantTopNames []string
	}{
		{
			name: "sum and order",
			games: []steam.Game{
				{Name: "A", PlaytimeForever: 120},
				{Name: "B", PlaytimeForever: 60},
			},
			n: 5, wantHours: 3, wantTopNames: []string{"A", "B"},
		},
		{
			name: "caps at n",
			games: []steam.Game{
				{Name: "A", PlaytimeForever: 300},
				{Name: "B", PlaytimeForever: 200},
				{Name: "C", PlaytimeForever: 100},
			},
			n: 2, wantHours: 10, wantTopNames: []string{"A", "B"},
		},
		{
			name:         "n larger than games",
			games:        []steam.Game{{Name: "A", PlaytimeForever: 60}},
			n:            5,
			wantHours:    1,
			wantTopNames: []string{"A"},
		},
		{name: "empty", games: nil, n: 5, wantHours: 0, wantTopNames: []string{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hours, top := summarizeGames(tc.games, tc.n)
			if hours != tc.wantHours {
				t.Errorf("hours: want %d, got %d", tc.wantHours, hours)
			}
			if len(top) != len(tc.wantTopNames) {
				t.Fatalf("top len: want %d, got %d (%+v)", len(tc.wantTopNames), len(top), top)
			}
			for i, w := range tc.wantTopNames {
				if top[i].Name != w {
					t.Errorf("top[%d]: want %s, got %s", i, w, top[i].Name)
				}
			}
		})
	}
}
