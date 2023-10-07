package services

import (
	"go-service-template/domain"
	"go-service-template/domain/dto"
	"go-service-template/monitor"
)

type ILocationService interface {
	CreateLocationMock(ctx monitor.ApplicationContext) error
	GetLocationByID(ctx monitor.ApplicationContext, id string) (*domain.Location, error)
	CreateLocation(ctx monitor.ApplicationContext, newLocationData dto.CreateLocationRequest) (domain.Location, error)
	UpdateLocation(ctx monitor.ApplicationContext, updatedLocationData dto.UpdateLocationRequest) (domain.Location, error)
	GetPaginatedLocations(ctx monitor.ApplicationContext, filters domain.LocationsFilters) (domain.CursorPage[domain.Location], error)
}
