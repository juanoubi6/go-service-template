package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go-service-template/domain"
	"go-service-template/domain/dto"
	customHTTP "go-service-template/http"
	"go-service-template/http/controllers"
	"go-service-template/mocks"
	"go-service-template/utils"

	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	LocName   = "Some Name"
	NextPage  = "NextPage"
	PrevPage  = "PrevPage"
	CursorVal = "Value"
)

var (
	mockCreateLocationRequest = dto.CreateLocationRequest{
		SupplierID:     1,
		Name:           "someName",
		Address:        "Address",
		City:           "City",
		State:          "State",
		Zipcode:        "Zipcode",
		LocationTypeID: 1,
		ContactPerson:  nil,
		PhoneNumber:    nil,
		Email:          nil,
	}
	mockUpdateLocationRequest = dto.UpdateLocationRequest{
		ID:             uuid.New().String(),
		SupplierID:     1,
		Name:           "newName",
		Address:        "newAddress",
		City:           "newCity",
		State:          "newState",
		Zipcode:        "newZipcode",
		LocationTypeID: 1,
		ContactPerson:  nil,
		PhoneNumber:    nil,
		Email:          nil,
		Active:         false,
	}
)

type LocationControllerSuite struct {
	suite.Suite
	locationServiceMock     *mocks.ILocationService
	createLocationEP        customHTTP.Endpoint
	updateLocationEP        customHTTP.Endpoint
	getPaginatedLocationsEP customHTTP.Endpoint
	getLocationDetailsEP    customHTTP.Endpoint
	echoRouter              *echo.Echo
}

func (s *LocationControllerSuite) SetupTest() {
	locationServiceMock := new(mocks.ILocationService)
	controller := controllers.NewLocationController(locationServiceMock, validator.New())

	s.createLocationEP = controller.CreateLocationEndpoint()
	s.updateLocationEP = controller.UpdateLocationEndpoint()
	s.getPaginatedLocationsEP = controller.PaginatedLocationsEndpoint()
	s.getLocationDetailsEP = controller.LocationDetailsEndpoint()
	s.locationServiceMock = locationServiceMock

	s.echoRouter = echo.New()
}

func (s *LocationControllerSuite) assertMockExpectations() {
	s.locationServiceMock.AssertExpectations(s.T())
}

func TestLocationControllerSuite(t *testing.T) {
	suite.Run(t, new(LocationControllerSuite))
}

func (s *LocationControllerSuite) Test_createLocation_Success() {
	w := httptest.NewRecorder()

	bodyBytes, _ := json.Marshal(mockCreateLocationRequest)

	req, _ := http.NewRequest(http.MethodPost, "/v1/locations", bytes.NewBuffer(bodyBytes))

	s.locationServiceMock.On("CreateLocation", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		request := args.Get(1).(dto.CreateLocationRequest)
		assert.Equal(s.T(), "someName", request.Name)
	}).Return(domain.Location{ID: "1"}, nil)

	assert.Nil(s.T(), s.createLocationEP.Handler(s.echoRouter.NewContext(req, w)))

	var response struct {
		Data domain.Location `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		s.FailNow("could not unmarshal response body", err.Error())
	}

	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Equal(s.T(), "1", response.Data.ID)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_createLocation_Returns400OnInvalidBody() {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, "/v1/locations", bytes.NewBuffer([]byte("invalid body")))

	assert.Nil(s.T(), s.createLocationEP.Handler(s.echoRouter.NewContext(req, w)))
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_updateLocation_Success() {
	w := httptest.NewRecorder()

	bodyBytes, _ := json.Marshal(mockUpdateLocationRequest)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/locations/%v", mockUpdateLocationRequest.ID), bytes.NewBuffer(bodyBytes))

	s.locationServiceMock.On("UpdateLocation", mock.Anything, mock.Anything).Return(domain.Location{ID: "1"}, nil)

	echoCtx := s.echoRouter.NewContext(req, w)
	echoCtx.SetPath(s.updateLocationEP.Path)
	echoCtx.SetParamNames("locationID")
	echoCtx.SetParamValues(mockUpdateLocationRequest.ID)

	assert.Nil(s.T(), s.updateLocationEP.Handler(echoCtx))

	var response struct {
		Data domain.Location `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		s.FailNow("could not unmarshal response body", err.Error())
	}

	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Equal(s.T(), "1", response.Data.ID)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_updateLocation_Returns400OnInvalidBody() {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPut, "/v1/locations/uuid", bytes.NewBuffer([]byte("invalid body")))

	echoCtx := s.echoRouter.NewContext(req, w)
	echoCtx.SetPath(s.updateLocationEP.Path)
	echoCtx.SetParamNames("locationID")
	echoCtx.SetParamValues("uuid")

	assert.Nil(s.T(), s.updateLocationEP.Handler(echoCtx))
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_updateLocation_Returns400OnLocationIDMismatchBetweenUrlAndPayload() {
	w := httptest.NewRecorder()

	bodyBytes, _ := json.Marshal(dto.UpdateLocationRequest{ID: "someUUID"})
	req, _ := http.NewRequest(http.MethodPut, "/v1/locations/uuid", bytes.NewBuffer(bodyBytes))

	echoCtx := s.echoRouter.NewContext(req, w)
	echoCtx.SetPath(s.updateLocationEP.Path)
	echoCtx.SetParamNames("locationID")
	echoCtx.SetParamValues("differentUUID")

	assert.Nil(s.T(), s.updateLocationEP.Handler(echoCtx))
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_getPaginatedLocations_Success() {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/locations?limit=%v&direction=%v&name=%v&cursor=%v", controllers.DefaultLimit, domain.NextPage, LocName, CursorVal),
		http.NoBody,
	)

	s.locationServiceMock.On("GetPaginatedLocations", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		filters := args.Get(1).(domain.LocationsFilters)
		assert.Equal(s.T(), LocName, *filters.Name)
		assert.Equal(s.T(), CursorVal, filters.CursorPaginationFilters.Cursor)
		assert.Equal(s.T(), domain.NextPage, filters.CursorPaginationFilters.Direction)
		assert.Equal(s.T(), controllers.DefaultLimit, filters.CursorPaginationFilters.Limit)
	}).Return(
		domain.CursorPage[domain.Location]{
			Limit:        controllers.DefaultLimit,
			Data:         []domain.Location{{}},
			NextPage:     utils.ToPointer(NextPage),
			PreviousPage: utils.ToPointer(PrevPage),
		}, nil,
	)

	assert.Nil(s.T(), s.getPaginatedLocationsEP.Handler(s.echoRouter.NewContext(req, w)))

	var response struct {
		Data domain.CursorPage[domain.Location] `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		s.FailNow("could not unmarshal response body", err.Error())
	}

	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Equal(s.T(), NextPage, *response.Data.NextPage)
	assert.Equal(s.T(), PrevPage, *response.Data.PreviousPage)
	assert.Equal(s.T(), controllers.DefaultLimit, response.Data.Limit)
	assert.Len(s.T(), response.Data.Data, 1)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_getPaginatedLocations_Returns400OnEmptyCursorAndPrevDirection() {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/locations?limit=%v&direction=%v", controllers.DefaultLimit, domain.PreviousPage),
		http.NoBody,
	)

	assert.Nil(s.T(), s.getPaginatedLocationsEP.Handler(s.echoRouter.NewContext(req, w)))
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_getPaginatedLocations_Returns400OnInvalidDirection() {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/locations?limit=%v&direction=%v", controllers.DefaultLimit, "invalidDirection"),
		http.NoBody,
	)

	assert.Nil(s.T(), s.getPaginatedLocationsEP.Handler(s.echoRouter.NewContext(req, w)))
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_getPaginatedLocations_Returns400OnInvalidLimitValue() {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/locations?limit=%v&direction=%v", "nonNumeric", domain.NextPage),
		http.NoBody,
	)

	assert.Nil(s.T(), s.getPaginatedLocationsEP.Handler(s.echoRouter.NewContext(req, w)))
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_getLocationDetails_Success() {
	w := httptest.NewRecorder()

	locationID := uuid.New().String()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/locations/%v", locationID), http.NoBody)

	s.locationServiceMock.On("GetLocationByID", mock.Anything, locationID).Return(&domain.Location{ID: locationID}, nil)

	echoCtx := s.echoRouter.NewContext(req, w)
	echoCtx.SetPath(s.getLocationDetailsEP.Path)
	echoCtx.SetParamNames("locationID")
	echoCtx.SetParamValues(locationID)

	assert.Nil(s.T(), s.getLocationDetailsEP.Handler(echoCtx))

	var response struct {
		Data domain.Location `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		s.FailNow("could not unmarshal response body", err.Error())
	}

	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Equal(s.T(), locationID, response.Data.ID)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_getLocationDetails_Returns404WhenLocationCannotBeFound() {
	w := httptest.NewRecorder()

	locationID := uuid.New().String()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/locations/%v", locationID), http.NoBody)

	s.locationServiceMock.On("GetLocationByID", mock.Anything, locationID).Return(nil, nil)

	echoCtx := s.echoRouter.NewContext(req, w)
	echoCtx.SetPath(s.getLocationDetailsEP.Path)
	echoCtx.SetParamNames("locationID")
	echoCtx.SetParamValues(locationID)

	assert.Nil(s.T(), s.getLocationDetailsEP.Handler(echoCtx))
	assert.Equal(s.T(), http.StatusNotFound, w.Code)
	s.assertMockExpectations()
}
