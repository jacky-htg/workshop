package dto

type LoginResponse struct {
	Token string `json:"token"`
}

func (u *LoginResponse) Transform(token string) {
	u.Token = token
}
