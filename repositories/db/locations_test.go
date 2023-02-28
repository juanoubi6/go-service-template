package db

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-service-template/domain"
	"go-service-template/monitor"
	"go-service-template/utils"
	"log"
	"testing"
)

var (
	mockCtx  = monitor.CreateAppContext(context.Background(), "")
	location = domain.Location{
		ID:   uuid.New().String(),
		Name: "New test location deactivated",
		Information: domain.LocationInformation{
			ID:        uuid.New().String(),
			Address:   "StreetName",
			City:      "City",
			State:     "ST",
			Zipcode:   "1234",
			Latitude:  90.0,
			Longitude: -90,
			ContactInformation: domain.ContactInformation{
				ContactPerson: utils.ToPointer[string]("Name"),
				PhoneNumber:   utils.ToPointer[string]("Phone"),
				Email:         utils.ToPointer[string]("Email"),
			},
		},
		LocationType: domain.LocationType{
			ID:   domain.LastMileLocationTypeID,
			Type: "Last Mile",
		},
		Supplier: domain.Supplier{
			ID:   1,
			Name: "SomeSupplier",
		},
		Active: true,
	}
	subLocation = domain.SubLocation{
		ID:   uuid.New().String(),
		Name: "SubLocation",
		SubLocationType: domain.SubLocationType{
			ID:   1,
			Type: "Default",
		},
		Active:     false,
		LocationID: uuid.New().String(),
	}
	existingName    = "existingName"
	notExistingName = "notExistingName"
)

type LocationsDALSuite struct {
	suite.Suite
	dal     *LocationsDAL
	db      *sql.DB
	sqlMock sqlmock.Sqlmock
}

func (s *LocationsDALSuite) SetupTest() {
	db, sqmock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	s.db = db
	s.sqlMock = sqmock
	s.dal = &LocationsDAL{
		TxDBContext:  CreateTxDBContext(db),
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func TestLocationsDALSuiteSuite(t *testing.T) {
	suite.Run(t, new(LocationsDALSuite))
}

func (s *LocationsDALSuite) Test_CreateLocation_Success() {
	s.sqlMock.ExpectPrepare(InsertLocation).ExpectExec().WithArgs(
		location.ID,
		location.Name,
		location.LocationType.ID,
		location.Supplier.ID,
		location.Active,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	s.sqlMock.ExpectPrepare(InsertLocationInformation).ExpectExec().WithArgs(
		location.Information.ID,
		location.ID,
		location.Information.Address,
		location.Information.City,
		location.Information.State,
		location.Information.Zipcode,
		location.Information.ContactInformation.ContactPerson,
		location.Information.ContactInformation.PhoneNumber,
		location.Information.ContactInformation.Email,
		location.Information.Latitude,
		location.Information.Longitude,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.dal.CreateLocation(mockCtx, location)

	assert.Nil(s.T(), err)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *LocationsDALSuite) Test_CreateSubLocation_Success() {
	s.sqlMock.ExpectPrepare(InsertSubLocation).ExpectExec().WithArgs(
		subLocation.ID,
		subLocation.LocationID,
		subLocation.SubLocationType.ID,
		subLocation.Name,
		subLocation.Active,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.dal.CreateSubLocation(mockCtx, subLocation)

	assert.Nil(s.T(), err)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *LocationsDALSuite) Test_UpdateLocation_Success() {
	s.sqlMock.ExpectPrepare(UpdateLocation).ExpectExec().WithArgs(
		location.Name,
		location.LocationType.ID,
		location.Supplier.ID,
		location.Active,
		location.ID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	s.sqlMock.ExpectPrepare(UpdateLocationInformation).ExpectExec().WithArgs(
		location.Information.Address,
		location.Information.City,
		location.Information.State,
		location.Information.Zipcode,
		location.Information.ContactInformation.ContactPerson,
		location.Information.ContactInformation.PhoneNumber,
		location.Information.ContactInformation.Email,
		location.Information.Latitude,
		location.Information.Longitude,
		location.Information.ID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.dal.UpdateLocation(mockCtx, location)

	assert.Nil(s.T(), err)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *LocationsDALSuite) Test_GetLocationByID_Success() {
	locationID := uuid.New().String()

	s.sqlMock.ExpectQuery(GetLocationByID).WithArgs(locationID).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"l.id", "l.name", "l.active",
				"s.id", "s.name",
				"lt.id", "lt.type",
				"li.id", "li.address", "li.city", "li.state", "li.zipcode", "li.contact_person", "li.phone_number", "li.email", "li.latitude", "li.longitude",
			},
		).AddRow(
			locationID, "locName", true,
			1, "supplierName",
			2, "locationType",
			"locInfID", "address", "city", "state", "zipcode", "contactPerson", "phone", "email", 90.0, -90.0,
		),
	)

	location, err := s.dal.GetLocationByID(mockCtx, locationID)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), locationID, location.ID)
	assert.Equal(s.T(), 1, location.Supplier.ID)
	assert.Equal(s.T(), 2, location.LocationType.ID)
	assert.Equal(s.T(), "locInfID", location.Information.ID)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *LocationsDALSuite) Test_CheckLocationNameExistence_ReturnsTrueIfNameExists() {
	s.sqlMock.ExpectQuery(CheckLocationNameExistence).WithArgs(existingName).WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()),
	)

	exists, err := s.dal.CheckLocationNameExistence(mockCtx, existingName)

	assert.Nil(s.T(), err)
	assert.True(s.T(), exists)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *LocationsDALSuite) Test_CheckLocationNameExistence_ReturnsFalseIfNameDoesNotExists() {
	s.sqlMock.ExpectQuery(CheckLocationNameExistence).WithArgs(notExistingName).WillReturnRows(
		sqlmock.NewRows([]string{}),
	)

	exists, err := s.dal.CheckLocationNameExistence(mockCtx, notExistingName)

	assert.Nil(s.T(), err)
	assert.False(s.T(), exists)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *LocationsDALSuite) Test_GetPaginatedLocations_SuccessOnNextDirection() {
	filters := domain.LocationsFilters{
		CursorPaginationFilters: domain.CursorPaginationFilters{
			Cursor:    "val",
			Direction: domain.NextPage,
			Limit:     10,
		},
		Name: utils.ToPointer[string]("name"),
	}

	expectedQuery := `SELECT 
    	l.id, 
    	l.name, 
    	l.active, 
    	s.id, 
    	s.name, 
    	lt.id, 
    	lt.type, 
    	li.id, 
    	li.address, 
    	li.city, 
    	li.state, 
    	li.zipcode, 
    	li.contact_person, 
    	li.phone_number, 
    	li.email, 
    	li.latitude, 
    	li.longitude
	FROM location.locations l 
	    INNER JOIN location.location_information li on l.id = li.location_id 
	    INNER JOIN location.location_types lt on l.location_type_id = lt.id 
	    INNER JOIN location.suppliers s on s.id = l.supplier_id 
	  WHERE l.name LIKE CONCAT ('%',$1::text,'%') 
	  AND l.name > $2 ORDER BY l.name ASC LIMIT 11`

	s.sqlMock.ExpectQuery(expectedQuery).WithArgs(*filters.Name, filters.Cursor).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"l.id", "l.name", "l.active",
				"s.id", "s.name",
				"lt.id", "lt.type",
				"li.id", "li.address", "li.city", "li.state", "li.zipcode", "li.contact_person", "li.phone_number", "li.email", "li.latitude", "li.longitude",
			},
		).AddRow(
			"uuid", "locName", true,
			1, "supplierName",
			2, "locationType",
			"locInfID", "address", "city", "state", "zipcode", "contactPerson", "phone", "email", 90.0, -90.0,
		),
	)

	resp, err := s.dal.GetPaginatedLocations(mockCtx, filters)

	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resp.NextPage)
	assert.Equal(s.T(), "locName", *resp.PreviousPage)
	assert.Len(s.T(), resp.Data, 1)
	assert.Equal(s.T(), filters.Limit, resp.Limit)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *LocationsDALSuite) Test_GetPaginatedLocations_SuccessOnPrevDirection() {
	filters := domain.LocationsFilters{
		CursorPaginationFilters: domain.CursorPaginationFilters{
			Cursor:    "val",
			Direction: domain.PreviousPage,
			Limit:     10,
		},
		Name: utils.ToPointer[string]("name"),
	}

	expectedQuery := `SELECT 
    	l.id, 
    	l.name, 
    	l.active, 
    	s.id, 
    	s.name, 
    	lt.id, 
    	lt.type, 
    	li.id, 
    	li.address, 
    	li.city, 
    	li.state, 
    	li.zipcode, 
    	li.contact_person, 
    	li.phone_number, 
    	li.email, 
    	li.latitude, 
    	li.longitude
	FROM location.locations l 
	    INNER JOIN location.location_information li on l.id = li.location_id 
	    INNER JOIN location.location_types lt on l.location_type_id = lt.id 
	    INNER JOIN location.suppliers s on s.id = l.supplier_id 
	  WHERE l.name LIKE CONCAT ('%',$1::text,'%') 
	  AND l.name < $2 ORDER BY l.name DESC LIMIT 11`

	s.sqlMock.ExpectQuery(expectedQuery).WithArgs(*filters.Name, filters.CursorPaginationFilters.Cursor).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"l.id", "l.name", "l.active",
				"s.id", "s.name",
				"lt.id", "lt.type",
				"li.id", "li.address", "li.city", "li.state", "li.zipcode", "li.contact_person", "li.phone_number", "li.email", "li.latitude", "li.longitude",
			},
		).AddRow(
			"uuid", "locName", true,
			1, "supplierName",
			2, "locationType",
			"locInfID", "address", "city", "state", "zipcode", "contactPerson", "phone", "email", 90.0, -90.0,
		),
	)

	resp, err := s.dal.GetPaginatedLocations(mockCtx, filters)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "locName", *resp.NextPage)
	assert.Nil(s.T(), resp.PreviousPage)
	assert.Len(s.T(), resp.Data, 1)
	assert.Equal(s.T(), filters.Limit, resp.Limit)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *LocationsDALSuite) Test_GetPaginatedLocations_SuccessOnEmptyCursor() {
	filters := domain.LocationsFilters{
		CursorPaginationFilters: domain.CursorPaginationFilters{
			Cursor:    "",
			Direction: domain.NextPage,
			Limit:     10,
		},
		Name: utils.ToPointer[string]("name"),
	}

	expectedQuery := `SELECT 
    	l.id, 
    	l.name, 
    	l.active, 
    	s.id, 
    	s.name, 
    	lt.id, 
    	lt.type, 
    	li.id, 
    	li.address, 
    	li.city, 
    	li.state, 
    	li.zipcode, 
    	li.contact_person, 
    	li.phone_number, 
    	li.email, 
    	li.latitude, 
    	li.longitude
	FROM location.locations l 
	    INNER JOIN location.location_information li on l.id = li.location_id 
	    INNER JOIN location.location_types lt on l.location_type_id = lt.id 
	    INNER JOIN location.suppliers s on s.id = l.supplier_id 
	  WHERE l.name LIKE CONCAT ('%',$1::text,'%') 
		ORDER BY l.name ASC LIMIT 11`

	s.sqlMock.ExpectQuery(expectedQuery).WithArgs(*filters.Name).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"l.id", "l.name", "l.active",
				"s.id", "s.name",
				"lt.id", "lt.type",
				"li.id", "li.address", "li.city", "li.state", "li.zipcode", "li.contact_person", "li.phone_number", "li.email", "li.latitude", "li.longitude",
			},
		).AddRow(
			"uuid", "locName", true,
			1, "supplierName",
			2, "locationType",
			"locInfID", "address", "city", "state", "zipcode", "contactPerson", "phone", "email", 90.0, -90.0,
		),
	)

	resp, err := s.dal.GetPaginatedLocations(mockCtx, filters)

	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resp.NextPage)
	assert.Nil(s.T(), resp.PreviousPage)
	assert.Len(s.T(), resp.Data, 1)
	assert.Equal(s.T(), filters.Limit, resp.Limit)
	if err = s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}
