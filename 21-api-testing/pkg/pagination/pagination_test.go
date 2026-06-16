package pagination_test

import (
	"net/http"
	"testing"
	"workshop/internal/model"
	"workshop/pkg/pagination"

	"github.com/stretchr/testify/assert"
)

func TestExtractPaginationFromURL_Simple(t *testing.T) {
	// Test default values
	req, _ := http.NewRequest("GET", "/users", nil)
	page, limit, order, sort, search := pagination.ExtractPaginationFromURL(req)

	assert.Equal(t, 1, page)
	assert.Equal(t, 10, limit)
	assert.Equal(t, "id", order)
	assert.Equal(t, "asc", sort)
	assert.Equal(t, "", search)
}

func TestExtractPaginationFromURL_WithParams(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users?page=2&limit=20&order=name&sort=desc&search=john", nil)
	page, limit, order, sort, search := pagination.ExtractPaginationFromURL(req)

	assert.Equal(t, 2, page)
	assert.Equal(t, 20, limit)
	assert.Equal(t, "name", order)
	assert.Equal(t, "desc", sort)
	assert.Equal(t, "john", search)
}

func TestExtractPaginationFromURL_InvalidValues(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users?page=invalid&limit=invalid&sort=invalid", nil)
	page, limit, order, sort, search := pagination.ExtractPaginationFromURL(req)

	assert.Equal(t, 1, page)
	assert.Equal(t, 10, limit)
	assert.Equal(t, "id", order)
	assert.Equal(t, "asc", sort)
	assert.Equal(t, "", search)
}

func TestExtractPaginationFromURL_NegativeValues(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users?page=-1&limit=-5", nil)
	page, limit, _, _, _ := pagination.ExtractPaginationFromURL(req)

	assert.Equal(t, 1, page)
	assert.Equal(t, 10, limit)
}

func TestExtractPaginationFromURL_WithDefaultOrder(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users", nil)
	_, _, order, _, _ := pagination.ExtractPaginationFromURL(req, "name")

	assert.Equal(t, "name", order)
}

func TestExtractPaginationFromURL_OrderUppercase(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users?order=NAME&sort=DESC", nil)
	_, _, order, sort, _ := pagination.ExtractPaginationFromURL(req)

	assert.Equal(t, "name", order)
	assert.Equal(t, "desc", sort)
}

func TestGetMeta_Simple(t *testing.T) {
	paginationModel := model.Pagination{
		Page:  1,
		Limit: 10,
		Count: 100,
	}

	result := pagination.GetMeta("", "id", "asc", paginationModel)

	assert.Equal(t, "id", result.Order)
	assert.Equal(t, "asc", result.Sort)
	assert.Equal(t, "", result.Search)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 100, result.Pagination.Total)
	assert.Equal(t, 10, result.Pagination.TotalPages)
	assert.True(t, result.Pagination.HasNext)
	assert.False(t, result.Pagination.HasPrev)
}

func TestGetMeta_WithSearch(t *testing.T) {
	paginationModel := model.Pagination{
		Page:  1,
		Limit: 10,
		Count: 100,
	}

	result := pagination.GetMeta("john", "name", "desc", paginationModel)

	assert.Equal(t, "john", result.Search)
}

func TestGetMeta_EmptyResult(t *testing.T) {
	paginationModel := model.Pagination{
		Page:  1,
		Limit: 10,
		Count: 0,
	}

	result := pagination.GetMeta("", "id", "asc", paginationModel)

	assert.Equal(t, 0, result.Pagination.Total)
	assert.Equal(t, 0, result.Pagination.TotalPages)
	assert.False(t, result.Pagination.HasNext)
	assert.False(t, result.Pagination.HasPrev)
}
