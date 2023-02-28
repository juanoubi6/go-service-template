package domain

const (
	DefaultSubLocationTypeID = 1
)

type SubLocation struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	SubLocationType SubLocationType `json:"sub_location_type"`
	Active          bool            `json:"active"`
	LocationID      string          `json:"location_id"`
}

type SubLocationType struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}
