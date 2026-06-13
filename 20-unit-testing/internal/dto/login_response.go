package dto

import "workshop/internal/model"

type LoginResponse struct {
	Token    string       `json:"token"`
	User     UserResponse `json:"user"`
	Accesses []string     `json:"permissions"`
}

func (u *LoginResponse) Transform(token string, user model.User, accesses []string) {
	u.Token = token
	u.Accesses = accesses

	userResp := UserResponse{}
	userResp.Transform(user)
	u.User = userResp
}
