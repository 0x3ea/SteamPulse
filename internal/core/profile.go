package core

import (
	"context"
	"errors"

	"github.com/0x3ea/SteamPulse/internal/steam"
)

// SteamClient is the subset of the Steam API that core depends on.
type SteamClient interface {
	GetPlayerSummaries(ctx context.Context, steamIDs []string) ([]steam.PlayerSummary, error)
	GetPlayerOwnedGames(ctx context.Context, steamID string) ([]steam.Game, error)
}

// Profile is the player profile card returned by GetProfile.
type Profile struct {
	SteamID       string       `json:"steam_id"`
	PersonaName   string       `json:"persona_name"`
	ProfileURL    string       `json:"profile_url"`
	AvatarFull    string       `json:"avatar_full"`
	IsPublic      bool         `json:"is_public"`
	Online        bool         `json:"online"`
	TimeCreated   int64        `json:"time_created"` // unix seconds
	CountryCode   string       `json:"country_code"`
	TotalPlaytime int          `json:"total_playtime"`
	TotalGames    int          `json:"total_games"`
	TopGames      []steam.Game `json:"top_games"`
}

// ErrProfileNotFound means Steam ID does not exist or is private.
var ErrProfileNotFound = errors.New("core: profile not found or private")

// Service is the entry point for all core operations.
type Service struct {
	steam SteamClient
}

// NewService return a Service backed by SteamClient.
func NewService(s SteamClient) *Service {
	return &Service{steam: s}
}

// GetProfile assembles player's profile card from Steam.
func (s *Service) GetProfile(ctx context.Context, steamID string) (*Profile, error) {
	players, err := s.steam.GetPlayerSummaries(ctx, []string{steamID})
	if err != nil {
		return nil, err
	}
	if len(players) == 0 {
		// private or nonexistent (Steam returns 200, not an error).
		return nil, ErrProfileNotFound
	}

	games, err := s.steam.GetPlayerOwnedGames(ctx, steamID)
	if err != nil {
		return nil, err
	}

	totalHours, top5 := summarizeGames(games, 5)

	p := players[0]
	return &Profile{
		SteamID:       p.SteamID,
		PersonaName:   p.PersonaName,
		ProfileURL:    p.ProfileURL,
		AvatarFull:    p.AvatarFull,
		IsPublic:      p.IsPublic(),
		Online:        p.PersonaState != 0,
		TimeCreated:   p.TimeCreated,
		CountryCode:   p.LocCountryCode,
		TotalPlaytime: totalHours,
		TotalGames:    len(games),
		TopGames:      top5,
	}, nil
}
