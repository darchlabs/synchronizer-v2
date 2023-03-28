package util

import (
	"math"

	"github.com/gofiber/fiber/v2"
)

const DefaultPage = 1
const DefaultLimit = 10

type Pagination struct {
	Page   int64 `json:"page" query:"page"`
	Limit  int64 `json:"limit" query:"limit"`
	Offset int64 `json:"offset"`
}

type PaginationMeta struct {
	Page          int64 `json:"page"`
	Limit         int64 `json:"limit"`
	TotalElements int64 `json:"totalElements"`
	TotalPages    int64 `json:"totalPages"`
}

func (p *Pagination) GetPaginationFromFiber(c *fiber.Ctx) error {
	// get pagination from query params
	err := c.QueryParser(p)
	if err != nil {
		return err
	}

	// define default if it's necessary
	if p.Page == 0 {
		p.Page = DefaultPage
	}

	if p.Limit == 0 {
		p.Limit = DefaultLimit
	}

	// calculate offset
	p.Offset = (p.Page - 1) * p.Limit

	return nil
}

func (p *Pagination) GetPaginationMeta(totalElements int64) PaginationMeta {
	// ger total pages value
	totalPages := int64(math.Ceil(float64(totalElements) / float64(p.Limit)))

	return PaginationMeta{
		Page:          p.Page,
		Limit:         p.Limit,
		TotalElements: totalElements,
		TotalPages:    totalPages,
	}
}
