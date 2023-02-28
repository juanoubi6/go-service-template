package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

type LocationControllerSuite struct {
	suite.Suite
	locationServiceMock     *mocks.ILocationService
	createLocationEP        customHTTP.Endpoint
	updateLocationEP        customHTTP.Endpoint
	getPaginatedLocationsEP customHTTP.Endpoint
	getLocationDetailsEP    customHTTP.Endpoint
}

func (s *LocationControllerSuite) SetupTest() {
	locationServiceMock := new(mocks.ILocationService)
	controller := controllers.NewLocationController(locationServiceMock)

	s.createLocationEP = controller.CreateLocationEndpoint()
	s.updateLocationEP = controller.UpdateLocationEndpoint()
	s.getPaginatedLocationsEP = controller.PaginatedLocationsEndpoint()
	s.getLocationDetailsEP = controller.LocationDetailsEndpoint()
	s.locationServiceMock = locationServiceMock
}

func (s *LocationControllerSuite) assertMockExpectations() {
	s.locationServiceMock.AssertExpectations(s.T())
}

func TestLocationControllerSuite(t *testing.T) {
	suite.Run(t, new(LocationControllerSuite))
}

func (s *LocationControllerSuite) Test_createLocation_Success() {
	w := httptest.NewRecorder()

	bodyBytes, _ := json.Marshal(dto.CreateLocationRequest{Name: "someName"})

	req, _ := http.NewRequest(http.MethodPost, "/v1/locations", bytes.NewBuffer(bodyBytes))

	s.locationServiceMock.On("CreateLocation", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		request := args.Get(1).(dto.CreateLocationRequest)
		assert.Equal(s.T(), "someName", request.Name)
	}).Return(domain.Location{ID: "1"}, nil)

	s.createLocationEP.Handler(w, req)

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

	s.createLocationEP.Handler(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_updateLocation_Success() {
	w := httptest.NewRecorder()

	bodyBytes, _ := json.Marshal(dto.UpdateLocationRequest{ID: "uuid1"})
	req, _ := http.NewRequest(http.MethodPut, "/v1/locations/uuid1", bytes.NewBuffer(bodyBytes))

	s.locationServiceMock.On("UpdateLocation", mock.Anything, mock.Anything).Return(domain.Location{ID: "1"}, nil)

	router := chi.NewRouter()
	router.MethodFunc(s.updateLocationEP.Method, s.updateLocationEP.Path, s.updateLocationEP.Handler)
	router.ServeHTTP(w, req)

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

	router := chi.NewRouter()
	router.Put(s.updateLocationEP.Path, s.updateLocationEP.Handler)
	router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_updateLocation_Returns400OnLocationIDMismatchBetweenUrlAndPayload() {
	w := httptest.NewRecorder()

	bodyBytes, _ := json.Marshal(dto.UpdateLocationRequest{ID: "someUUID"})
	req, _ := http.NewRequest(http.MethodPut, "/v1/locations/differentUUID", bytes.NewBuffer(bodyBytes))

	router := chi.NewRouter()
	router.Put(s.updateLocationEP.Path, s.updateLocationEP.Handler)
	router.ServeHTTP(w, req)

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

	s.getPaginatedLocationsEP.Handler(w, req)

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

	s.getPaginatedLocationsEP.Handler(w, req)

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

	s.getPaginatedLocationsEP.Handler(w, req)

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

	s.getPaginatedLocationsEP.Handler(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	s.assertMockExpectations()
}

func (s *LocationControllerSuite) Test_getLocationDetails_Success() {
	w := httptest.NewRecorder()

	locationID := uuid.New().String()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/locations/%v", locationID), http.NoBody)

	s.locationServiceMock.On("GetLocationByID", mock.Anything, locationID).Return(&domain.Location{ID: locationID}, nil)

	router := chi.NewRouter()
	router.MethodFunc(s.getLocationDetailsEP.Method, s.getLocationDetailsEP.Path, s.getLocationDetailsEP.Handler)
	router.ServeHTTP(w, req)

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

	router := chi.NewRouter()
	router.MethodFunc(s.getLocationDetailsEP.Method, s.getLocationDetailsEP.Path, s.getLocationDetailsEP.Handler)
	router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
	s.assertMockExpectations()
}
