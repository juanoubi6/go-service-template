package domain

import (
	"go-service-template/utils"
	"sort"
)

type Pageable interface {
	GetUniqueOrderedIdentifier() string
}

// ExampleCursorPage is only for swagger usages (swaggo/swag still does not support generics)
type ExampleCursorPage struct {
	Data         []any   `json:"data"`
	Limit        int     `json:"limit"`
	NextPage     *string `json:"next_page"`
	PreviousPage *string `json:"previous_page"`
}

type CursorPage[T Pageable] struct {
	Data         []T     `json:"data"`
	Limit        int     `json:"limit"`
	NextPage     *string `json:"next_page"`
	PreviousPage *string `json:"previous_page"`
}

func BuildCursorPage[T Pageable](data []T, filters CursorPaginationFilters) CursorPage[T] {
	if len(data) == 0 {
		return CursorPage[T]{Data: []T{}, Limit: 0, NextPage: nil, PreviousPage: nil}
	}

	page := CursorPage[T]{Limit: filters.Limit}

	if filters.Direction == NextPage {
		// Data elements should be in ASC order here
		if len(data) > filters.Limit {
			page.NextPage = utils.ToPointer[string]((data[len(data)-2]).GetUniqueOrderedIdentifier())
			page.PreviousPage = utils.ToPointer[string]((data[0]).GetUniqueOrderedIdentifier())
			page.Data = data[0 : len(data)-1] // Remove last element
		} else {
			page.NextPage = nil
			page.PreviousPage = utils.ToPointer[string]((data[0]).GetUniqueOrderedIdentifier())
			page.Data = data
		}

		if filters.Cursor == "" {
			page.PreviousPage = nil
		}
	} else {
		// Data elements should be in DESC order here. We need to order them in ASC
		sort.Slice(data, func(i, j int) bool {
			return data[i].GetUniqueOrderedIdentifier() < data[j].GetUniqueOrderedIdentifier()
		})
		if len(data) > filters.Limit {
			page.PreviousPage = utils.ToPointer[string](data[1].GetUniqueOrderedIdentifier())
			page.NextPage = utils.ToPointer[string]((data[len(data)-1]).GetUniqueOrderedIdentifier())
			page.Data = data[1:] // Remove first element
		} else {
			page.PreviousPage = nil
			page.NextPage = utils.ToPointer[string]((data[len(data)-1]).GetUniqueOrderedIdentifier())
			page.Data = data
		}
	}

	return page
}
