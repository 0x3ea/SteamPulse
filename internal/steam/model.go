package steam

// Maps to the JSON field in the GetPlayerSummaries response
// unnecessary field will be ignored
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

// whether the profile is visible to this API
func (p PlayerSummary) IsPublic() bool {
	return p.CommunityVisibilityState == 3
}
