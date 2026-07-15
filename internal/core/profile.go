package core

import (
	"context"
	"errors"

	"github.com/0x3ea/SteamPulse/internal/steam"
)

// only have GetPlayerSummaries method could be SteamClient
type SteamClient interface {
	GetPlayerSummaries(ctx context.Context, steamIDs []string) ([]steam.PlayerSummary, error)
}

type Profile struct {
	SteamID     string `json:"steam_id"`
	PersonaName string `json:"persona_name"`
	ProfileURL  string `json:"profile_url"`
	AvatarFull  string `json:"avatar_full"`
	IsPublic    bool   `json:"is_public"`
	Online      bool   `json:"online"`
	TimeCreated int64  `json:"time_created"` // unix seconds
	CountryCode string `json:"country_code"`

	// TODO(Phase 1): AccountValue, TopGames, TotalPlaytime — added as features land.
}

// Steam ID does not exist or is private.
var ErrProfileNotFound = errors.New("profile not found or private")

// the entry point for all core operations.
type Service struct {
	steam SteamClient
}

func NewService(s SteamClient) *Service {
	return &Service{steam: s}
}

// assembles player's profile card from Steam.
func (s *Service) GetProfile(ctx context.Context, steamID string) (*Profile, error) {
	players, err := s.steam.GetPlayerSummaries(ctx, []string{steamID})
	if err != nil {
		return nil, err
	}
	if len(players) == 0 {
		// private or nonexistent (Steam returns 200, not an error).
		return nil, ErrProfileNotFound
	}
	p := players[0]
	return &Profile{
		SteamID:     p.SteamID,
		PersonaName: p.PersonaName,
		ProfileURL:  p.ProfileURL,
		AvatarFull:  p.AvatarFull,
		IsPublic:    p.IsPublic(),
		Online:      p.PersonaState != 0,
		TimeCreated: p.TimeCreated,
		CountryCode: p.LocCountryCode,
	}, nil
}
