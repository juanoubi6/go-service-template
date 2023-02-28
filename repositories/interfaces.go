package repositories

import (
	"context"
	"database/sql"
	"go-service-template/domain"
	"go-service-template/domain/googlemaps"
)

type DBReader interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type QueryExecutor interface {
	Ping() error
	StartTx() error
	RollbackTx() error
	CommitTx() error
	Exec(ctx domain.ApplicationContext, stmt string, fields ...interface{}) (sql.Result, error)
	WithTx(ctx domain.ApplicationContext, fn func(fnCtx domain.ApplicationContext) error) error
}

type LocationsDB interface {
	QueryExecutor
	CreateLocation(ctx domain.ApplicationContext, location domain.Location) error
	UpdateLocation(ctx domain.ApplicationContext, location domain.Location) error
	CreateSubLocation(ctx domain.ApplicationContext, subLocation domain.SubLocation) error
	GetLocationByID(ctx domain.ApplicationContext, id string) (*domain.Location, error)
	CheckLocationNameExistence(ctx domain.ApplicationContext, name string) (bool, error)
	GetPaginatedLocations(ctx domain.ApplicationContext, filters domain.LocationsFilters) (domain.CursorPage[domain.Location], error)
}

type DatabaseFactory interface {
	GetLocationsDB() (LocationsDB, error)
}

type GoogleMapsAPI interface {
	ValidateAddress(ctx domain.ApplicationContext, request googlemaps.AddressValidationRequest) (*googlemaps.AddressValidateMatch, error)
}
