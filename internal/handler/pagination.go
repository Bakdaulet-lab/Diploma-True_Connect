package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

// parsePagination extracts and clamps page/per_page query parameters.
func parsePagination(c *gin.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return page, perPage
}

// parseCursorPagination extracts and clamps limit and reads cursor query parameters.
func parseCursorPagination(c *gin.Context) (cursor string, limit int) {
	cursor = c.Query("cursor")
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit < 1 {
		limit = defaultPerPage
	}
	if limit > maxPerPage {
		limit = maxPerPage
	}

	return cursor, limit

}
