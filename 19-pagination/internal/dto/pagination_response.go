package dto

import "workshop/internal/model"

type MetaResponse struct {
	Order      string             `json:"order,omitempty"`
	Sort       string             `json:"sort,omitempty"`
	Search     string             `json:"search,omitempty"`
	Filter     any                `json:"filter,omitempty"`
	Pagination PaginationResponse `json:"pagination,omitempty"`
}

type PaginationResponse struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

func (u *PaginationResponse) Transform(pagination model.Pagination) {
	u.Page = pagination.Page
	u.Limit = pagination.Limit
	u.Total = pagination.Count
	u.TotalPages = (u.Total + u.Limit - 1) / u.Limit
	u.HasNext = u.Page < u.TotalPages
	u.HasPrev = u.Page > 1
}
