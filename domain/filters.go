package domain

import "encoding/json"

const (
	NextPage     = "next"
	PreviousPage = "prev"
)

type CursorPaginationFilters struct {
	Cursor    string `json:"cursor"`
	Direction string `json:"direction"` // next or prev
	Limit     int    `json:"limit"`
}

type LocationsFilters struct {
	CursorPaginationFilters
	Name *string `json:"name"`
}

func (f LocationsFilters) ToJSON() string {
	jsonBytes, err := json.Marshal(f)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}
