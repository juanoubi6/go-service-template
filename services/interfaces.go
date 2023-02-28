package services

import (
	"go-service-template/domain"
	"go-service-template/domain/dto"
)

type ILocationService interface {
	GetLocationByID(ctx domain.ApplicationContext, id string) (*domain.Location, error)
	CreateLocation(ctx domain.ApplicationContext, newLocationData dto.CreateLocationRequest) (domain.Location, error)
	UpdateLocation(ctx domain.ApplicationContext, updatedLocationData dto.UpdateLocationRequest) (domain.Location, error)
	GetPaginatedLocations(ctx domain.ApplicationContext, filters domain.LocationsFilters) (domain.CursorPage[domain.Location], error)
}
