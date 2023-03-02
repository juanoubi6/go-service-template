package repositories

import (
	"context"
	"database/sql"
	"go-service-template/domain"
	"go-service-template/domain/googlemaps"
	"go-service-template/monitor"
)

type DBReader interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type QueryExecutor interface {
	Ping() error
	StartTx(ctx monitor.ApplicationContext) error
	RollbackTx() error
	CommitTx() error
	Exec(ctx monitor.ApplicationContext, stmt string, fields ...interface{}) (sql.Result, error)
	WithTx(ctx monitor.ApplicationContext, fn func(fnCtx monitor.ApplicationContext) error) error
}

type LocationsDB interface {
	QueryExecutor
	CreateLocation(ctx monitor.ApplicationContext, location domain.Location) error
	UpdateLocation(ctx monitor.ApplicationContext, location domain.Location) error
	CreateSubLocation(ctx monitor.ApplicationContext, subLocation domain.SubLocation) error
	GetLocationByID(ctx monitor.ApplicationContext, id string) (*domain.Location, error)
	CheckLocationNameExistence(ctx monitor.ApplicationContext, name string) (bool, error)
	GetPaginatedLocations(ctx monitor.ApplicationContext, filters domain.LocationsFilters) (domain.CursorPage[domain.Location], error)
}

type DatabaseFactory interface {
	GetLocationsDB() (LocationsDB, error)
}

type GoogleMapsAPI interface {
	ValidateAddress(ctx monitor.ApplicationContext, request googlemaps.AddressValidationRequest) (*googlemaps.AddressValidateMatch, error)
}
