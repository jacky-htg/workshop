package pagination

import (
	"net/http"
	"strconv"
	"strings"
	"workshop/internal/dto"
	"workshop/internal/model"
)

func ExtractPaginationFromURL(r *http.Request, defaultOrder ...string) (page, limit int, order, sort, search string) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	order = strings.ToLower(r.URL.Query().Get("order"))
	sort = strings.ToLower(r.URL.Query().Get("sort"))
	search = strings.ToLower(r.URL.Query().Get("search"))

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if len(order) == 0 {
		if len(defaultOrder) > 0 {
			order = defaultOrder[0]
		} else {
			order = "id"
		}
	}

	if len(sort) == 0 || !(sort == "asc" || sort == "desc") {
		sort = "asc"
	}

	return
}

func GetMeta(search, order, sort string, pagination model.Pagination) dto.MetaResponse {
	meta := dto.MetaResponse{
		Order: order,
		Sort:  sort,
	}

	if len(search) > 0 {
		meta.Search = search
	}

	var paginationResp dto.PaginationResponse
	paginationResp.Transform(pagination)
	meta.Pagination = paginationResp
	return meta
}
