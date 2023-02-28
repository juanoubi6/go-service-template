package dto

type CreateLocationRequest struct {
	SupplierID     int     `json:"supplier_id" validate:"required"`
	Name           string  `json:"name" validate:"required"`
	Address        string  `json:"address" validate:"required"`
	City           string  `json:"city" validate:"required"`
	State          string  `json:"state" validate:"required"`
	Zipcode        string  `json:"zipcode" validate:"required"`
	LocationTypeID int     `json:"location_type_id" validate:"required"`
	ContactPerson  *string `json:"contact_person"`
	PhoneNumber    *string `json:"phone_number"`
	Email          *string `json:"email"`
}

type UpdateLocationRequest struct {
	ID             string  `json:"id" validate:"required"`
	SupplierID     int     `json:"supplier_id" validate:"required"`
	Name           string  `json:"name" validate:"required"`
	Address        string  `json:"address" validate:"required"`
	City           string  `json:"city" validate:"required"`
	State          string  `json:"state" validate:"required"`
	Zipcode        string  `json:"zipcode" validate:"required"`
	LocationTypeID int     `json:"location_type_id" validate:"required"`
	ContactPerson  *string `json:"contact_person"`
	PhoneNumber    *string `json:"phone_number"`
	Email          *string `json:"email"`
	Active         bool    `json:"active"`
}
