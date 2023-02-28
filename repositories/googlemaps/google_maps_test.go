package googlemaps_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go-service-template/domain/googlemaps"
	customHTTP "go-service-template/http"
	"go-service-template/mocks"
	"go-service-template/monitor"
	googleMapsRepo "go-service-template/repositories/googlemaps"
	"go-service-template/utils"
	"net/http"
	"testing"
	"time"
)

var (
	mockCtx     = monitor.CreateAppContext(context.Background(), "")
	mockRequest = googlemaps.AddressValidationRequest{}
)

type GoogleMapsRepositorySuite struct {
	suite.Suite
	httpClientMock       *mocks.CustomHTTPClient
	googleMapsRepository *googleMapsRepo.Repository
}

func (s *GoogleMapsRepositorySuite) SetupTest() {
	httpClientMock := new(mocks.CustomHTTPClient)

	googleMapsRepository := googleMapsRepo.NewGoogleMapsRepository(httpClientMock)

	s.googleMapsRepository = googleMapsRepository
	s.httpClientMock = httpClientMock
}

func (s *GoogleMapsRepositorySuite) assertMockExpectations() {
	s.httpClientMock.AssertExpectations(s.T())
}

func TestInventoryMgmtRepositorySuite(t *testing.T) {
	suite.Run(t, new(GoogleMapsRepositorySuite))
}

func (s *GoogleMapsRepositorySuite) Test_ValidateAddress_Success() {
	validateAddressResponse := utils.GetJSONFileContent("validate-address-many-matches")

	s.httpClientMock.On(
		"DoWithRetry",
		mockCtx,
		mock.Anything,
		time.Duration(5)*time.Second,
		3,
		time.Second,
		[]int{},
	).Return(
		customHTTP.CustomHTTPResponse{
			StatusCode:   http.StatusOK,
			BodyPayload:  validateAddressResponse,
			Headers:      http.Header{},
			BaseResponse: &http.Response{},
		},
		nil,
	).Run(func(args mock.Arguments) {
		request := args.Get(1).(customHTTP.RequestValues)
		assert.Equal(s.T(), http.MethodPost, request.Method)
	}).Once()

	match, err := s.googleMapsRepository.ValidateAddress(mockCtx, mockRequest)

	assert.NotNil(s.T(), match)
	assert.Nil(s.T(), err)
	s.assertMockExpectations()
}

func (s *GoogleMapsRepositorySuite) Test_ValidateAddress_ReturnsNilOn404() {
	s.httpClientMock.On(
		"DoWithRetry",
		mockCtx,
		mock.Anything,
		time.Duration(5)*time.Second,
		3,
		time.Second,
		[]int{},
	).Return(
		customHTTP.CustomHTTPResponse{
			StatusCode:   http.StatusNotFound,
			BodyPayload:  []byte{},
			Headers:      http.Header{},
			BaseResponse: &http.Response{},
		},
		nil,
	).Run(func(args mock.Arguments) {
		request := args.Get(1).(customHTTP.RequestValues)
		assert.Equal(s.T(), http.MethodPost, request.Method)
	}).Once()

	match, err := s.googleMapsRepository.ValidateAddress(mockCtx, mockRequest)

	assert.Nil(s.T(), match)
	assert.Nil(s.T(), err)
	s.assertMockExpectations()
}

func (s *GoogleMapsRepositorySuite) Test_ValidateAddress_ReturnsNilWhenNoMatchesAreReturned() {
	validateAddressResponse := utils.GetJSONFileContent("validate-address-no-matches")

	s.httpClientMock.On(
		"DoWithRetry",
		mockCtx,
		mock.Anything,
		time.Duration(5)*time.Second,
		3,
		time.Second,
		[]int{},
	).Return(
		customHTTP.CustomHTTPResponse{
			StatusCode:   http.StatusOK,
			BodyPayload:  validateAddressResponse,
			Headers:      http.Header{},
			BaseResponse: &http.Response{},
		},
		nil,
	).Run(func(args mock.Arguments) {
		request := args.Get(1).(customHTTP.RequestValues)
		assert.Equal(s.T(), http.MethodPost, request.Method)
	}).Once()

	match, err := s.googleMapsRepository.ValidateAddress(mockCtx, mockRequest)

	assert.Nil(s.T(), match)
	assert.Nil(s.T(), err)
	s.assertMockExpectations()
}

func (s *GoogleMapsRepositorySuite) Test_ValidateAddress_ReturnsNilWhenNoPremiseMatchesAreReturned() {
	validateAddressResponse := utils.GetJSONFileContent("validate-address-no-premise-matches")

	s.httpClientMock.On(
		"DoWithRetry",
		mockCtx,
		mock.Anything,
		time.Duration(5)*time.Second,
		3,
		time.Second,
		[]int{},
	).Return(
		customHTTP.CustomHTTPResponse{
			StatusCode:   http.StatusOK,
			BodyPayload:  validateAddressResponse,
			Headers:      http.Header{},
			BaseResponse: &http.Response{},
		},
		nil,
	).Run(func(args mock.Arguments) {
		request := args.Get(1).(customHTTP.RequestValues)
		assert.Equal(s.T(), http.MethodPost, request.Method)
	}).Once()

	match, err := s.googleMapsRepository.ValidateAddress(mockCtx, mockRequest)

	assert.Nil(s.T(), match)
	assert.Nil(s.T(), err)
	s.assertMockExpectations()
}

func (s *GoogleMapsRepositorySuite) Test_ValidateAddress_Error() {
	s.httpClientMock.On(
		"DoWithRetry",
		mockCtx,
		mock.Anything,
		time.Duration(5)*time.Second,
		3,
		time.Second,
		[]int{},
	).Return(
		customHTTP.CustomHTTPResponse{
			StatusCode:   http.StatusBadRequest,
			BodyPayload:  []byte("some error"),
			Headers:      http.Header{},
			BaseResponse: &http.Response{},
		},
		nil,
	).Once()

	res, err := s.googleMapsRepository.ValidateAddress(mockCtx, mockRequest)

	assert.Nil(s.T(), res)
	assert.NotNil(s.T(), err)
	s.assertMockExpectations()
}
