package domain

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
