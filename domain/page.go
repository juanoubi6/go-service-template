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

	// Sort elements ASC
	sort.Slice(data, func(i, j int) bool {
		return data[i].GetUniqueOrderedIdentifier() < data[j].GetUniqueOrderedIdentifier()
	})

	if filters.Direction == NextPage {
		if len(data) > filters.Limit {
			page.NextPage = utils.ToPointer[string]((data[len(data)-2]).GetUniqueOrderedIdentifier())
			page.Data = data[0 : len(data)-1] // Remove last element
		} else {
			page.NextPage = nil
			page.Data = data
		}

		if filters.Cursor == "" {
			page.PreviousPage = nil
		}else{
			page.PreviousPage = utils.ToPointer[string]((data[0]).GetUniqueOrderedIdentifier())
		}
	} else {
		page.NextPage = utils.ToPointer[string]((data[len(data)-1]).GetUniqueOrderedIdentifier())

		if len(data) > filters.Limit {
			page.PreviousPage = utils.ToPointer[string](data[1].GetUniqueOrderedIdentifier())
			page.Data = data[1:] // Remove first element
		} else {
			page.PreviousPage = nil
			page.Data = data
		}
	}

	return page
}
