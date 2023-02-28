package services_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go-service-template/domain"
	"go-service-template/domain/dto"
	"go-service-template/domain/googlemaps"
	"go-service-template/mocks"
	"go-service-template/services"
	"go-service-template/utils"
	"testing"
)

var (
	testCtx       = domain.CreateAppContext(context.Background(), "")
	createLocData = dto.CreateLocationRequest{
		SupplierID:     1,
		Name:           "New Name",
		Address:        "Address",
		City:           "City",
		State:          "State",
		Zipcode:        "12344",
		LocationTypeID: 2,
		ContactPerson:  utils.ToPointer[string]("contactPerson"),
		PhoneNumber:    utils.ToPointer[string]("phone"),
		Email:          utils.ToPointer[string]("email"),
	}
	updateLocData = dto.UpdateLocationRequest{
		ID:             uuid.New().String(),
		SupplierID:     3,
		Name:           "Updated name",
		Address:        "Updated Address",
		City:           "Updated City",
		State:          "Updated State",
		Zipcode:        "Updated 12344",
		LocationTypeID: 4,
		ContactPerson:  utils.ToPointer[string]("Updated contactPerson"),
		PhoneNumber:    utils.ToPointer[string]("Updated phone"),
		Email:          utils.ToPointer[string]("Updated email"),
		Active:         false,
	}
)

type LocationServiceSuite struct {
	suite.Suite
	dbFactoryMock     *mocks.DatabaseFactory
	googleMapsAPIMock *mocks.GoogleMapsAPI
	locationsDBMock   *mocks.LocationsDB
	locationService   *services.LocationService
}

func (s *LocationServiceSuite) SetupSuite() {
	dbFactoryMock := new(mocks.DatabaseFactory)
	locationsDBMock := new(mocks.LocationsDB)
	googleMapsMock := new(mocks.GoogleMapsAPI)

	s.locationService = services.NewLocationService(dbFactoryMock, googleMapsMock)
	s.dbFactoryMock = dbFactoryMock
	s.locationsDBMock = locationsDBMock
	s.googleMapsAPIMock = googleMapsMock
}

func (s *LocationServiceSuite) SetupTest() {
	s.dbFactoryMock.ExpectedCalls = nil
	s.locationsDBMock.ExpectedCalls = nil
	s.googleMapsAPIMock.ExpectedCalls = nil
}

func (s *LocationServiceSuite) assertAllExpectations() {
	s.dbFactoryMock.AssertExpectations(s.T())
	s.locationsDBMock.AssertExpectations(s.T())
	s.googleMapsAPIMock.AssertExpectations(s.T())
}

func TestLocationServiceSuite(t *testing.T) {
	suite.Run(t, new(LocationServiceSuite))
}

func (s *LocationServiceSuite) Test_CreateLocation_Success() {
	s.googleMapsAPIMock.On("ValidateAddress", mock.Anything, mock.Anything).Return(&googlemaps.AddressValidateMatch{}, nil)
	s.dbFactoryMock.On("GetLocationsDB").Return(s.locationsDBMock, nil)
	s.locationsDBMock.On("CheckLocationNameExistence", testCtx, createLocData.Name).Return(false, nil).Once()

	s.locationsDBMock.On("StartTx").Return(nil).Once()
	s.locationsDBMock.On("CommitTx").Return(nil).Once()

	s.locationsDBMock.On("CreateLocation", testCtx, mock.Anything).Return(nil).Once()
	s.locationsDBMock.On("CreateSubLocation", testCtx, mock.Anything).Run(func(args mock.Arguments) {
		subLocation := args.Get(1).(domain.SubLocation)
		assert.Equal(s.T(), services.DefaultSubLocationName, subLocation.Name)
	}).Return(nil).Once()

	location, err := s.locationService.CreateLocation(testCtx, createLocData)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), location.Name, createLocData.Name)
	s.assertAllExpectations()
}

func (s *LocationServiceSuite) Test_CreateLocation_FailsIfNewLocationNameIsAlreadyInUse() {
	s.googleMapsAPIMock.On("ValidateAddress", mock.Anything, mock.Anything).Return(&googlemaps.AddressValidateMatch{}, nil)
	s.dbFactoryMock.On("GetLocationsDB").Return(s.locationsDBMock, nil)
	s.locationsDBMock.On("CheckLocationNameExistence", testCtx, createLocData.Name).Return(true, nil).Once()

	_, err := s.locationService.CreateLocation(testCtx, createLocData)

	assert.NotNil(s.T(), err)
	assert.IsType(s.T(), domain.NameAlreadyInUseErr{}, err)
	s.assertAllExpectations()
}

func (s *LocationServiceSuite) Test_CreateLocation_FailsIfAddressValidationCannotFindAddress() {
	s.googleMapsAPIMock.On("ValidateAddress", mock.Anything, mock.Anything).Return(nil, nil)

	_, err := s.locationService.CreateLocation(testCtx, createLocData)

	assert.NotNil(s.T(), err)
	assert.IsType(s.T(), domain.AddressNotValidErr{}, err)
	s.assertAllExpectations()
}

func (s *LocationServiceSuite) Test_UpdateLocation_Success() {
	var existingLocation = domain.Location{
		ID:   updateLocData.ID,
		Name: "SomeName",
		Information: domain.LocationInformation{
			ID:        uuid.New().String(),
			Address:   "Address",
			City:      "City",
			State:     "State",
			Zipcode:   "Zipcode",
			Latitude:  123.4,
			Longitude: 567.8,
			ContactInformation: domain.ContactInformation{
				ContactPerson: utils.ToPointer[string]("ContactPerson"),
				PhoneNumber:   utils.ToPointer[string]("PhoneNumber"),
				Email:         utils.ToPointer[string]("Email"),
			},
		},
		LocationType: domain.LocationType{ID: 1, Type: "Type"},
		Supplier:     domain.Supplier{ID: 2, Name: "Supplier"},
		Active:       true,
	}

	s.dbFactoryMock.On("GetLocationsDB").Return(s.locationsDBMock, nil)
	s.locationsDBMock.On("StartTx").Return(nil).Once()
	s.locationsDBMock.On("CommitTx").Return(nil).Once()

	s.locationsDBMock.On("GetLocationByID", testCtx, updateLocData.ID).Return(&existingLocation, nil).Once()
	s.locationsDBMock.On("CheckLocationNameExistence", testCtx, updateLocData.Name).Return(false, nil).Once()

	s.googleMapsAPIMock.On("ValidateAddress", mock.Anything, mock.Anything).Return(&googlemaps.AddressValidateMatch{}, nil)
	s.locationsDBMock.On("UpdateLocation", testCtx, mock.Anything).Run(func(args mock.Arguments) {
		updatedLocation := args.Get(1).(domain.Location)
		assert.Equal(s.T(), updateLocData.Name, updatedLocation.Name)
	}).Return(nil).Once()

	updatedLocation, err := s.locationService.UpdateLocation(testCtx, updateLocData)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), updateLocData.Name, updatedLocation.Name)
	s.assertAllExpectations()
}

func (s *LocationServiceSuite) Test_UpdateLocation_FailsIfAddressValidationCannotFindAddress() {
	s.dbFactoryMock.On("GetLocationsDB").Return(s.locationsDBMock, nil)
	s.locationsDBMock.On("StartTx").Return(nil).Once()
	s.locationsDBMock.On("RollbackTx").Return(nil).Once()

	s.locationsDBMock.On("GetLocationByID", testCtx, updateLocData.ID).Return(&domain.Location{Name: "OldName"}, nil).Once()
	s.locationsDBMock.On("CheckLocationNameExistence", testCtx, updateLocData.Name).Return(false, nil).Once()
	s.googleMapsAPIMock.On("ValidateAddress", mock.Anything, mock.Anything).Return(nil, nil)

	_, err := s.locationService.UpdateLocation(testCtx, updateLocData)

	assert.NotNil(s.T(), err)
	assert.IsType(s.T(), domain.AddressNotValidErr{}, err)
	s.assertAllExpectations()
}

func (s *LocationServiceSuite) Test_UpdateLocation_FailsIfUpdatedLocationNameIsAlreadyInUse() {
	s.dbFactoryMock.On("GetLocationsDB").Return(s.locationsDBMock, nil)
	s.locationsDBMock.On("StartTx").Return(nil).Once()
	s.locationsDBMock.On("RollbackTx").Return(nil).Once()

	s.locationsDBMock.On("GetLocationByID", testCtx, updateLocData.ID).Return(&domain.Location{Name: "OldName"}, nil).Once()
	s.locationsDBMock.On("CheckLocationNameExistence", testCtx, updateLocData.Name).Return(true, nil).Once()

	_, err := s.locationService.UpdateLocation(testCtx, updateLocData)

	assert.NotNil(s.T(), err)
	assert.IsType(s.T(), domain.NameAlreadyInUseErr{}, err)
	s.assertAllExpectations()
}

func (s *LocationServiceSuite) Test_UpdateLocation_FailsIfLocationCannotBeFound() {
	s.dbFactoryMock.On("GetLocationsDB").Return(s.locationsDBMock, nil)
	s.locationsDBMock.On("StartTx").Return(nil).Once()
	s.locationsDBMock.On("RollbackTx").Return(nil).Once()

	s.locationsDBMock.On("GetLocationByID", testCtx, updateLocData.ID).Return(nil, nil).Once()

	_, err := s.locationService.UpdateLocation(testCtx, updateLocData)

	assert.NotNil(s.T(), err)
	assert.IsType(s.T(), domain.BusinessErr{}, err)
	s.assertAllExpectations()
}

func (s *LocationServiceSuite) Test_GetPaginatedLocations_Success() {
	filters := domain.LocationsFilters{}

	s.dbFactoryMock.On("GetLocationsDB").Return(s.locationsDBMock, nil)
	s.locationsDBMock.On("GetPaginatedLocations", testCtx, filters).Return(domain.CursorPage[domain.Location]{}, nil).Once()

	_, err := s.locationService.GetPaginatedLocations(testCtx, domain.LocationsFilters{})

	assert.Nil(s.T(), err)
	s.assertAllExpectations()
}

func (s *LocationServiceSuite) Test_GetLocationByID_Success() {
	locationID := uuid.New().String()

	s.dbFactoryMock.On("GetLocationsDB").Return(s.locationsDBMock, nil)
	s.locationsDBMock.On("GetLocationByID", testCtx, locationID).Return(&domain.Location{ID: locationID}, nil).Once()

	location, err := s.locationService.GetLocationByID(testCtx, locationID)

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), location)
	s.assertAllExpectations()
}
