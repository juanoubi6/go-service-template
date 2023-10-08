package controllers

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go-service-template/domain"
	"go-service-template/domain/dto"
	customHTTP "go-service-template/http"
	"go-service-template/http/middleware"
	"go-service-template/monitor"
	"go-service-template/services"
	"go-service-template/utils"
	"go.opentelemetry.io/otel/codes"
	"net/http"
)

var (
	ErrNoLocationIDSend   = errors.New("no locationID sent in URL")
	ErrLocationIDMismatch = errors.New("mismatch between location ID in url and the one in the request payload")
)

type LocationController struct {
	logger          monitor.AppLogger
	locationService services.ILocationService
	validator       *validator.Validate
}

func NewLocationController(locService services.ILocationService, validator *validator.Validate) *LocationController {
	return &LocationController{
		locationService: locService,
		logger:          monitor.GetStdLogger("LocationController"),
		validator:       validator,
	}
}

// Nada godoc
// @Summary Create location mock
// @Description Receives a request and mocks a location creation
// @Produce json
// @Success 200
// @Router /v1/location-mock [post]
func (ct *LocationController) CreateLocationMockEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodPost,
		Path:    "/v1/location-mock",
		Handler: ct.createLocationMock,
	}
}

// Nada godoc
// @Summary Create location
// @Description Create a new location and a default sub location
// @Produce json
// @Param request body dto.CreateLocationRequest true "Location attributes"
// @Success 200 {object} []domain.Location
// @Router /v1/locations [post]
func (ct *LocationController) CreateLocationEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodPost,
		Path:    "/v1/locations",
		Handler: ct.createLocation,
	}
}

// Nada godoc
// @Summary Update existing location
// @Description Update an existing location
// @Produce json
// @Param locationID path string true "Location ID"
// @Param request body dto.UpdateLocationRequest true "Location attributes"
// @Success 200 {object} []domain.Location
// @Router /v1/locations/{locationID} [put]
func (ct *LocationController) UpdateLocationEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodPut,
		Path:    "/v1/locations/:locationID",
		Handler: ct.updateLocation,
	}
}

// Nada godoc
// @Summary Retrieve paginated locations
// @Description Get paginated locations
// @Produce json
// @Param name query string false "Optional location name section. Service will filter locations that include this string"
// @Param limit query int false "Pagination limit, default to 10000"
// @Param cursor query string false "Cursor value, default to empty string"
// @Param direction query string true "Indicates the cursor direction. Accepted values: 'next' or 'prev'"
// @Success 200 {object} []domain.ExampleCursorPage
// @Router /v1/locations [get]
func (ct *LocationController) PaginatedLocationsEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodGet,
		Path:    "/v1/locations",
		Handler: ct.getPaginatedLocations,
	}
}

// Nada godoc
// @Summary Get location details
// @Description Get location details
// @Produce json
// @Param locationID path string true "Location ID"
// @Success 200 {object} domain.Location
// @Router /v1/locations/{locationID} [get]
func (ct *LocationController) LocationDetailsEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodGet,
		Path:    "/v1/locations/:locationID",
		Handler: ct.getLocationDetails,
	}
}

func (ct *LocationController) createLocationMock(c echo.Context) error {
	fnName := "LocationController.createLocationMock"
	var appCtx monitor.ApplicationContext = middleware.GetAppContext(c)

	appCtx, span := appCtx.StartSpan(fnName)
	defer span.End()

	err := ct.locationService.CreateLocationMock(appCtx)
	if err != nil {
		ct.logger.ErrorCtx(appCtx, fnName, "failed to create location mock", err)
		return c.JSON(httpStatusFromError(err), buildFailResponse(err, err.Error(), appCtx.GetCorrelationID()))
	}

	return c.JSON(http.StatusOK, "ok")
}

func (ct *LocationController) createLocation(c echo.Context) error {
	fnName := "LocationController.createLocation"
	var appCtx monitor.ApplicationContext = middleware.GetAppContext(c)

	appCtx, span := appCtx.StartSpan(fnName)
	defer span.End()

	createLocationRequest, err := parseAndValidateBody[dto.CreateLocationRequest](c.Request().Body, ct.validator)
	if err != nil {
		ct.logger.ErrorCtx(appCtx, fnName, "failed to parse or validate request body", err)
		return c.JSON(http.StatusBadRequest, buildFailResponse(err, "failed to parse or validate request body", appCtx.GetCorrelationID()))
	}

	location, err := ct.locationService.CreateLocation(appCtx, createLocationRequest)
	if err != nil {
		ct.logger.ErrorCtx(appCtx, fnName, "failed to create location", err)
		return c.JSON(httpStatusFromError(err), buildFailResponse(err, err.Error(), appCtx.GetCorrelationID()))
	}

	return c.JSON(http.StatusOK, buildSuccessResponse(location))
}

func (ct *LocationController) updateLocation(c echo.Context) error {
	fnName := "LocationController.updateLocation"
	var appCtx monitor.ApplicationContext = middleware.GetAppContext(c)

	appCtx, span := appCtx.StartSpan(fnName)
	defer span.End()

	locationID := c.Param("locationID")
	if locationID == "" {
		ct.logger.ErrorCtx(appCtx, fnName, ErrNoLocationIDSend.Error(), ErrNoLocationIDSend)
		return c.JSON(http.StatusBadRequest, buildFailResponse(ErrNoLocationIDSend, ErrNoLocationIDSend.Error(), appCtx.GetCorrelationID()))
	}

	updateLocationRequest, err := parseAndValidateBody[dto.UpdateLocationRequest](c.Request().Body, ct.validator)
	if err != nil {
		ct.logger.ErrorCtx(appCtx, fnName, "failed to parse or validate request body", err)
		return c.JSON(http.StatusBadRequest, buildFailResponse(err, "failed to parse or validate request body", appCtx.GetCorrelationID()))
	}

	if locationID != updateLocationRequest.ID {
		ct.logger.ErrorCtx(appCtx, fnName, ErrLocationIDMismatch.Error(), ErrLocationIDMismatch)
		return c.JSON(http.StatusBadRequest, buildFailResponse(ErrLocationIDMismatch, ErrLocationIDMismatch.Error(), appCtx.GetCorrelationID()))
	}

	location, err := ct.locationService.UpdateLocation(appCtx, updateLocationRequest)
	if err != nil {
		ct.logger.ErrorCtx(appCtx, fnName, "failed to update location", err)
		return c.JSON(http.StatusBadRequest, buildFailResponse(err, "failed to update location", appCtx.GetCorrelationID()))
	}

	return c.JSON(http.StatusOK, buildSuccessResponse(location))
}

func (ct *LocationController) getPaginatedLocations(c echo.Context) error {
	fnName := "LocationController.getPaginatedLocations"
	var appCtx monitor.ApplicationContext = middleware.GetAppContext(c)

	appCtx, span := appCtx.StartSpan(fnName)
	defer span.End()

	filters, err := buildLocationFilters(c.Request())
	if err != nil {
		ct.logger.ErrorCtx(appCtx, fnName, "error building location filters", err)
		span.SetStatus(codes.Error, err.Error())
		return c.JSON(http.StatusBadRequest, buildFailResponse(err, err.Error(), appCtx.GetCorrelationID()))
	}

	locationPage, err := ct.locationService.GetPaginatedLocations(appCtx, filters)
	if err != nil {
		ct.logger.ErrorCtx(appCtx, fnName, "failed to get paginated locations", err)
		span.SetStatus(codes.Error, err.Error())
		return c.JSON(http.StatusBadRequest, buildFailResponse(err, "failed to get paginated locations", appCtx.GetCorrelationID()))
	}

	return c.JSON(http.StatusOK, buildSuccessResponse(locationPage))
}

func (ct *LocationController) getLocationDetails(c echo.Context) error {
	fnName := "LocationController.getLocationDetails"
	var appCtx monitor.ApplicationContext = middleware.GetAppContext(c)

	appCtx, span := appCtx.StartSpan(fnName)
	defer span.End()

	locationID := c.Param("locationID")
	if locationID == "" {
		ct.logger.ErrorCtx(appCtx, fnName, ErrNoLocationIDSend.Error(), ErrNoLocationIDSend)
		return c.JSON(http.StatusBadRequest, buildFailResponse(ErrNoLocationIDSend, ErrNoLocationIDSend.Error(), appCtx.GetCorrelationID()))
	}

	location, err := ct.locationService.GetLocationByID(appCtx, locationID)
	if err != nil {
		ct.logger.ErrorCtx(appCtx, fnName, "failed to retrieve location by ID", err)
		return c.JSON(http.StatusBadRequest, buildFailResponse(err, err.Error(), appCtx.GetCorrelationID()))
	}

	if location == nil {
		errMsg := fmt.Errorf("location with ID %v not found", locationID)
		ct.logger.ErrorCtx(appCtx, fnName, errMsg.Error(), err)
		return c.JSON(http.StatusNotFound, buildFailResponse(errMsg, errMsg.Error(), appCtx.GetCorrelationID()))
	}

	return c.JSON(http.StatusOK, buildSuccessResponse(location))
}

func buildLocationFilters(req *http.Request) (domain.LocationsFilters, error) {
	locationFilters := domain.LocationsFilters{}

	cursorPaginationFilters, err := buildCursorPaginationFilters(req)
	if err != nil {
		return locationFilters, err
	}

	locationFilters.CursorPaginationFilters = cursorPaginationFilters

	if nameVal, ok := req.URL.Query()["name"]; ok {
		locationFilters.Name = utils.ToPointer[string](nameVal[0])
	}

	return locationFilters, nil
}
