package core

import (
	"context"
	"errors"
	"testing"

	"github.com/0x3ea/SteamPulse/internal/steam"
)

type fakeSteam struct {
	players []steam.PlayerSummary
	err     error
}

func (f fakeSteam) GetPlayerSummaries(ctx context.Context, ids []string) ([]steam.PlayerSummary, error) {
	return f.players, f.err
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
