package dto_test

import (
	"testing"
	"workshop/internal/dto"
	"workshop/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestPaginationResponse_Transform(t *testing.T) {
	// Test case 1: Page 1 with next page
	var resp1 dto.PaginationResponse
	resp1.Transform(model.Pagination{Page: 1, Limit: 10, Count: 25})

	assert.Equal(t, 1, resp1.Page)
	assert.Equal(t, 10, resp1.Limit)
	assert.Equal(t, 25, resp1.Total)
	assert.Equal(t, 3, resp1.TotalPages) // (25+10-1)/10 = 3
	assert.True(t, resp1.HasNext)
	assert.False(t, resp1.HasPrev)

	// Test case 2: Page 2 with prev and next
	var resp2 dto.PaginationResponse
	resp2.Transform(model.Pagination{Page: 2, Limit: 10, Count: 25})

	assert.Equal(t, 2, resp2.Page)
	assert.Equal(t, 10, resp2.Limit)
	assert.Equal(t, 25, resp2.Total)
	assert.Equal(t, 3, resp2.TotalPages)
	assert.True(t, resp2.HasNext)
	assert.True(t, resp2.HasPrev)

	// Test case 3: Last page
	var resp3 dto.PaginationResponse
	resp3.Transform(model.Pagination{Page: 3, Limit: 10, Count: 25})

	assert.Equal(t, 3, resp3.Page)
	assert.Equal(t, 10, resp3.Limit)
	assert.Equal(t, 25, resp3.Total)
	assert.Equal(t, 3, resp3.TotalPages)
	assert.False(t, resp3.HasNext)
	assert.True(t, resp3.HasPrev)

	// Test case 4: Empty data
	var resp4 dto.PaginationResponse
	resp4.Transform(model.Pagination{Page: 1, Limit: 10, Count: 0})

	assert.Equal(t, 1, resp4.Page)
	assert.Equal(t, 10, resp4.Limit)
	assert.Equal(t, 0, resp4.Total)
	assert.Equal(t, 0, resp4.TotalPages) // (0+10-1)/10 = 0
	assert.False(t, resp4.HasNext)
	assert.False(t, resp4.HasPrev)
}
