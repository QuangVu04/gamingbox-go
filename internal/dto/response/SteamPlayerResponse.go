package response

type SteamPlayerResponse struct {
	Response struct {
		Players []struct {
			SteamID      string `json:"steamid"`
			PersonaName  string `json:"personaname"` 
			ProfileUrl   string `json:"profileurl"`
			AvatarMedium   string `json:"avatarmedium"`   
			RealName     string `json:"realname"`
			LocCountry   string `json:"loccountrycode"`
		} `json:"players"`
	} `json:"response"`
}