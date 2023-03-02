package services

import (
	"fmt"
	"github.com/google/uuid"
	"go-service-template/domain"
	"go-service-template/domain/dto"
	"go-service-template/domain/googlemaps"
	"go-service-template/monitor"
	"go-service-template/repositories"
	"go-service-template/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

const (
	DefaultSubLocationName = "DEFAULT"
)

type LocationService struct {
	logger        monitor.StdLogger
	dbFactory     repositories.DatabaseFactory
	googleMapsAPI repositories.GoogleMapsAPI
}

func NewLocationService(dbFactory repositories.DatabaseFactory, googleMapsAPI repositories.GoogleMapsAPI) *LocationService {
	return &LocationService{
		logger:        monitor.GetStdLogger("LocationService"),
		dbFactory:     dbFactory,
		googleMapsAPI: googleMapsAPI,
	}
}

func (s *LocationService) CreateLocation(ctx monitor.ApplicationContext, newLocationData dto.CreateLocationRequest) (location domain.Location, err error) {
	fnName := "LocationService.CreateLocation"

	ctx, span := ctx.StartSpan(fnName, trace.WithAttributes(attribute.String("new_location_data", utils.ToJSON(newLocationData))))
	defer span.End()

	newLocation, err := s.buildNewLocation(ctx, newLocationData)
	if err != nil {
		return location, err
	}

	newDefaultSubLocation := s.buildDefaultSubLocationForLocation(newLocation)

	db, err := s.dbFactory.GetLocationsDB()
	if err != nil {
		return location, err
	}

	// Check if location name is already in use
	nameInUse, err := db.CheckLocationNameExistence(ctx, newLocationData.Name)
	if err != nil {
		s.logger.Error(ctx, fnName, "failed to check location name existence", err)
		return location, err
	}
	if nameInUse {
		errVal := domain.NameAlreadyInUseErr{Msg: fmt.Sprintf("location name '%v' is already in use", newLocationData.Name)}
		s.logger.Warn(ctx, fnName, errVal.Msg)
		return location, errVal
	}

	if err = db.WithTx(ctx, func(ctx monitor.ApplicationContext) error {
		var txErr error

		// Create the location
		if txErr = db.CreateLocation(ctx, newLocation); txErr != nil {
			return fmt.Errorf("error creating new location: %w", txErr)
		}

		// Create it's default sub location
		if txErr = db.CreateSubLocation(ctx, newDefaultSubLocation); txErr != nil {
			return fmt.Errorf("error creating new sub location: %w", txErr)
		}

		return nil
	}); err != nil {
		s.logger.Error(ctx, fnName, "tx failed", err)
		return location, err
	}

	return newLocation, nil
}

func (s *LocationService) UpdateLocation(ctx monitor.ApplicationContext, updatedLocationData dto.UpdateLocationRequest) (location domain.Location, err error) {
	fnName := "LocationService.UpdateLocation"

	ctx, span := ctx.StartSpan(fnName, trace.WithAttributes(attribute.String("updated_location_data", utils.ToJSON(updatedLocationData))))
	defer span.End()

	db, err := s.dbFactory.GetLocationsDB()
	if err != nil {
		return location, err
	}

	var existingLocation *domain.Location

	if err = db.WithTx(ctx, func(ctx monitor.ApplicationContext) error {
		var txErr error

		// Retrieve existing location
		existingLocation, txErr = db.GetLocationByID(ctx, updatedLocationData.ID)
		if txErr != nil {
			return fmt.Errorf("error finding location with ID %v: %w", updatedLocationData.ID, txErr)
		}
		if existingLocation == nil {
			return domain.BusinessErr{Msg: fmt.Sprintf("location with ID %v does not exist", updatedLocationData.ID)}
		}

		// Check if location name changed. If so, validate the new name is not in use
		if !strings.EqualFold(existingLocation.Name, updatedLocationData.Name) {
			var nameInUse bool
			nameInUse, txErr = db.CheckLocationNameExistence(ctx, updatedLocationData.Name)
			if txErr != nil {
				return txErr
			}
			if nameInUse {
				return domain.NameAlreadyInUseErr{Msg: fmt.Sprintf("location name '%v' is already in use", updatedLocationData.Name)}
			}
		}

		// Update location fields
		if txErr = s.updateLocation(ctx, db, existingLocation, updatedLocationData); txErr != nil {
			return txErr
		}

		return nil
	}); err != nil {
		s.logger.Error(ctx, fnName, "tx failed", err)
		return location, err
	}

	return *existingLocation, nil
}

func (s *LocationService) GetLocationByID(ctx monitor.ApplicationContext, id string) (*domain.Location, error) {
	fnName := "LocationService.GetLocationByID"

	ctx, span := ctx.StartSpan(fnName, trace.WithAttributes(attribute.String("location_id", id)))
	defer span.End()

	db, err := s.dbFactory.GetLocationsDB()
	if err != nil {
		return nil, err
	}

	return db.GetLocationByID(ctx, id)
}

func (s *LocationService) GetPaginatedLocations(ctx monitor.ApplicationContext, filters domain.LocationsFilters) (page domain.CursorPage[domain.Location], err error) {
	fnName := "LocationService.GetPaginatedLocations"

	ctx, span := ctx.StartSpan(fnName, trace.WithAttributes(attribute.String("filters", utils.ToJSON(filters))))
	defer span.End()

	db, err := s.dbFactory.GetLocationsDB()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return page, err
	}

	page, err = db.GetPaginatedLocations(ctx, filters)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		s.logger.Error(ctx, fnName, "failed to retrieve paginated locations", err)
		return page, err
	}

	return page, nil
}

func (s *LocationService) buildNewLocation(ctx monitor.ApplicationContext, data dto.CreateLocationRequest) (domain.Location, error) {
	supplierName := SupplierMap[data.SupplierID]
	locationTypeName := LocationTypeMap[data.LocationTypeID]

	// Use Google Maps API to retrieve the address and fill latitude and longitude values
	validatedAddress, err := s.googleMapsAPI.ValidateAddress(ctx, googlemaps.AddressValidationRequest{
		City:         data.City,
		AddressLine1: data.Address,
		State:        data.State,
		LongForm:     true,
		Zipcode:      data.Zipcode,
	})
	if err != nil {
		return domain.Location{}, err
	}

	if validatedAddress == nil {
		s.logger.Warn(ctx, "buildNewLocation", "failed to validate address", monitor.LoggingParam{Name: "address_data", Value: data})
		return domain.Location{}, domain.AddressNotValidErr{Msg: "the address information does not correspond to a valid address"}
	}

	return domain.Location{
		ID:   uuid.New().String(),
		Name: data.Name,
		Information: domain.LocationInformation{
			ID:        uuid.New().String(),
			Address:   data.Address,
			City:      data.City,
			State:     data.State,
			Zipcode:   data.Zipcode,
			Latitude:  validatedAddress.Latitude,
			Longitude: validatedAddress.Longitude,
			ContactInformation: domain.ContactInformation{
				ContactPerson: data.ContactPerson,
				PhoneNumber:   data.PhoneNumber,
				Email:         data.Email,
			},
		},
		LocationType: domain.LocationType{ID: data.LocationTypeID, Type: locationTypeName},
		Supplier:     domain.Supplier{ID: data.SupplierID, Name: supplierName},
		Active:       true,
	}, nil
}

func (s *LocationService) buildDefaultSubLocationForLocation(location domain.Location) domain.SubLocation {
	return domain.SubLocation{
		ID:              uuid.New().String(),
		Name:            DefaultSubLocationName,
		SubLocationType: domain.SubLocationType{ID: domain.DefaultSubLocationTypeID},
		Active:          true,
		LocationID:      location.ID,
	}
}

func (s *LocationService) updateLocation(
	ctx monitor.ApplicationContext,
	db repositories.LocationsDB,
	location *domain.Location,
	updateData dto.UpdateLocationRequest,
) error {
	// Use Google Maps API to retrieve the address and fill latitude and longitude values
	validatedAddress, err := s.googleMapsAPI.ValidateAddress(ctx, googlemaps.AddressValidationRequest{
		City:         updateData.City,
		AddressLine1: updateData.Address,
		State:        updateData.State,
		LongForm:     true,
		Zipcode:      updateData.Zipcode,
	})
	if err != nil {
		return err
	}

	if validatedAddress == nil {
		s.logger.Warn(ctx, "updateLocation", "failed to validate address", monitor.LoggingParam{Name: "address_data", Value: updateData})
		return domain.AddressNotValidErr{Msg: "the address information does not correspond to a valid address"}
	}

	location.Name = updateData.Name
	location.Supplier.ID = updateData.SupplierID
	location.Supplier.Name = SupplierMap[updateData.SupplierID]
	location.LocationType.ID = updateData.LocationTypeID
	location.LocationType.Type = LocationTypeMap[updateData.LocationTypeID]
	location.Active = updateData.Active

	location.Information.Address = updateData.Address
	location.Information.City = updateData.City
	location.Information.State = updateData.State
	location.Information.Zipcode = updateData.Zipcode
	location.Information.Latitude = validatedAddress.Latitude
	location.Information.Longitude = validatedAddress.Longitude

	location.Information.ContactInformation.ContactPerson = updateData.ContactPerson
	location.Information.ContactInformation.PhoneNumber = updateData.PhoneNumber
	location.Information.ContactInformation.Email = updateData.Email

	return db.UpdateLocation(ctx, *location)
}
