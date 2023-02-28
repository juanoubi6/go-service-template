package dto

type CreateLocationRequest struct {
	SupplierID     int     `json:"supplier_id"`
	Name           string  `json:"name"`
	Address        string  `json:"address"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Zipcode        string  `json:"zipcode"`
	LocationTypeID int     `json:"location_type_id"`
	ContactPerson  *string `json:"contact_person"`
	PhoneNumber    *string `json:"phone_number"`
	Email          *string `json:"email"`
}

type UpdateLocationRequest struct {
	ID             string  `json:"id"`
	SupplierID     int     `json:"supplier_id"`
	Name           string  `json:"name"`
	Address        string  `json:"address"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Zipcode        string  `json:"zipcode"`
	LocationTypeID int     `json:"location_type_id"`
	ContactPerson  *string `json:"contact_person"`
	PhoneNumber    *string `json:"phone_number"`
	Email          *string `json:"email"`
	Active         bool    `json:"active"`
}
