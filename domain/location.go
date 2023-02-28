package domain

const (
	ReconLocationTypeID       = 1
	WholesaleLocationTypeID   = 2
	LastMileLocationTypeID    = 3
	RetailReadyLocationTypeID = 4
	CrossDockLocationTypeID   = 5
	StorageLocationTypeID     = 6
	NotOnSiteLocationTypeID   = 7
)

type Location struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Information  LocationInformation `json:"information"`
	LocationType LocationType        `json:"location_type"`
	Supplier     Supplier            `json:"supplier"`
	Active       bool                `json:"active"`
}

func (l Location) GetUniqueOrderedIdentifier() string {
	return l.Name
}

type LocationInformation struct {
	ID                 string             `json:"-"`
	Address            string             `json:"address"`
	City               string             `json:"city"`
	State              string             `json:"state"`
	Zipcode            string             `json:"zipcode"`
	Latitude           float64            `json:"latitude"`
	Longitude          float64            `json:"longitude"`
	ContactInformation ContactInformation `json:"contact_information"`
}

type LocationType struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type Supplier struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ContactInformation struct {
	ContactPerson *string `json:"contact_person"`
	PhoneNumber   *string `json:"phone_number"`
	Email         *string `json:"email"`
}
