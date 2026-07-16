package core

import (
	"slices"
	"sort"

	"github.com/0x3ea/SteamPulse/internal/steam"
)

// summarizeGames Calculate Total playtime in hours and the n most-played games
func summarizeGames(games []steam.Game, n int) (totalHours int, topN []steam.Game) {
	var minutes int
	for _, g := range games {
		minutes += g.PlaytimeForever
	}
	totalHours = minutes / 60

	sort.Slice(games, func(i, j int) bool {
		return games[i].PlaytimeForever > games[j].PlaytimeForever
	})
	m := min(n, len(games))
	topN = slices.Clone(games[:m])

	return
}
