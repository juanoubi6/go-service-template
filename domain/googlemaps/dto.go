package googlemaps

type AddressValidationRequest struct {
	City         string `json:"city"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	FullAddress  string `json:"full_address"`
	State        string `json:"state"`
	LongForm     bool   `json:"long_form"`
	Zipcode      string `json:"zip_code"`
}

type AddressValidationResponse struct {
	Matches []AddressValidateMatch `json:"matches"`
}

type AddressValidateMatch struct {
	StreetNumber       string   `json:"street_number,omitempty"`
	Route              string   `json:"route,omitempty"`
	City               string   `json:"city,omitempty"`
	State              string   `json:"state,omitempty"`
	ZipCode            string   `json:"zip_code,omitempty"`
	ZipCodeSuffix      string   `json:"zip_code_suffix,omitempty"`
	County             string   `json:"county,omitempty"`
	Country            string   `json:"country,omitempty"`
	FullAddress        string   `json:"full_address,omitempty"`
	Neighborhood       string   `json:"neighborhood,omitempty"`
	Latitude           float64  `json:"latitude"`
	Longitude          float64  `json:"longitude"`
	PartialMatch       bool     `json:"partial_match"`
	MatchType          string   `json:"match_type,omitempty"`
	LocationType       string   `json:"location_type,omitempty"`
	PostCodeLocalities []string `json:"postcode_localities,omitempty"`
}
