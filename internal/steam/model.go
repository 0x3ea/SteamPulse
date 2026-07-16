package steam

// PlayerSummary Maps to the JSON field in the GetPlayerSummaries response.
// unnecessary field will be ignored.
// api document: https://developer.valvesoftware.com/wiki/Steam_Web_API#GetPlayerSummaries_(v0002)
type PlayerSummary struct {
	SteamID     string `json:"steamid"`
	PersonaName string `json:"personaname"`
	ProfileURL  string `json:"profileurl"`

	Avatar       string `json:"avatar"`       // 32px
	AvatarMedium string `json:"avatarmedium"` // 64px
	AvatarFull   string `json:"avatarfull"`   // 184px

	// 1 = private, 3 = public.
	CommunityVisibilityState int `json:"communityvisibilitystate"`
	// 0 - Offline, 1 - Online, 2 - Busy, 3 - Away, 4 - Snooze, 5 - looking to trade, 6 - looking to play.
	PersonaState   int    `json:"personastate"`
	LastLogOff     int64  `json:"lastlogoff"`  // unix seconds
	TimeCreated    int64  `json:"timecreated"` // unix seconds, account creation
	RealName       string `json:"realname"`
	LocCountryCode string `json:"loccountrycode"`
	PrimaryClanID  string `json:"primaryclanid"`
}

// IsPublic reports whether the profile is visible to this API
func (p PlayerSummary) IsPublic() bool {
	return p.CommunityVisibilityState == 3
}

// Game Maps to the JSON field in the GetPlayerOwnedGames response.
type Game struct {
	AppID           int64  `json:"appid"`
	Name            string `json:"name"`
	PlaytimeForever int    `json:"playtime_forever"` // minutes
	RtimeLastPlayed int64  `json:"rtime_last_played"`
	ImgIconURL      string `json:"img_icon_url"`
}
