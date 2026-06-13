package dto

import "workshop/internal/model"

type AccessTreeResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`

	Childrens []AccessResponse `json:"childrens"`
}

func (u *AccessTreeResponse) Transform(access model.AccessTree) {
	u.ID = access.ID
	u.Name = access.Name
	u.Alias = access.Alias

	for _, val := range access.Childrens {
		var child AccessResponse
		child.Transform(val)
		u.Childrens = append(u.Childrens, child)
	}
}

type AccessResponse struct {
	ID       int    `json:"id"`
	ParentID *int   `json:"parent_id"`
	Name     string `json:"name"`
	Alias    string `json:"alias"`
}

func (u *AccessResponse) Transform(access model.Access) {
	u.ID = access.ID
	u.ParentID = access.ParentID
	u.Name = access.Name
	u.Alias = access.Alias
}
