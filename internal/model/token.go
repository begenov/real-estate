package model

type TokenDetails struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	AccessUuid   string `json:"accessUuid"`
	RefreshUuid  string `json:"refreshUuid"`
	AtExpires    int64  `json:"atExpires"`
	RtExpires    int64  `json:"rtExpires"`
	SessionUuid  string `json:"sessionUuid"`
	UserId       int64  `json:"userId"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
