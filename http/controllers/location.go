package controllers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go-service-template/domain"
	"go-service-template/domain/dto"
	customHTTP "go-service-template/http"
	"go-service-template/http/middleware"
	"go-service-template/log"
	"go-service-template/services"
	"go-service-template/utils"
	"net/http"
)

type LocationController struct {
	logger          log.StdLogger
	locationService services.ILocationService
	validator       *validator.Validate
}

func NewLocationController(locService services.ILocationService, validator *validator.Validate) *LocationController {
	return &LocationController{
		locationService: locService,
		logger:          log.GetStdLogger("LocationController"),
		validator:       validator,
	}
}

// Nada godoc
// @Summary Create location
// @Description Create a new location and a default sub location
// @Produce json
// @Param request body dto.CreateLocationRequest true "Location attributes"
// @Success 200 {object} []domain.Location
// @Router /v1/locations [post]
func (c *LocationController) CreateLocationEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodPost,
		Path:    "/v1/locations",
		Handler: c.createLocation,
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
func (c *LocationController) UpdateLocationEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodPut,
		Path:    "/v1/locations/{locationID}",
		Handler: c.updateLocation,
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
func (c *LocationController) PaginatedLocationsEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodGet,
		Path:    "/v1/locations",
		Handler: c.getPaginatedLocations,
	}
}

// Nada godoc
// @Summary Get location details
// @Description Get location details
// @Produce json
// @Param locationID path string true "Location ID"
// @Success 200 {object} domain.Location
// @Router /v1/locations/{locationID} [get]
func (c *LocationController) LocationDetailsEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodGet,
		Path:    "/v1/locations/{locationID}",
		Handler: c.getLocationDetails,
	}
}

func (c *LocationController) createLocation(w http.ResponseWriter, r *http.Request) {
	fnName := "createLocation"
	appCtx := middleware.GetAppContext(r)

	createLocationRequest, err := parseAndValidateBody[dto.CreateLocationRequest](r.Body, c.validator)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "failed to parse or validate request body", err)
		_ = sendFailureResponse(w, http.StatusBadRequest, err, "failed to parse or validate request body", appCtx.GetCorrelationID())
		return
	}

	location, err := c.locationService.CreateLocation(appCtx, createLocationRequest)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "failed to create location", err)
		_ = sendFailureResponse(w, httpStatusFromError(err), err, err.Error(), appCtx.GetCorrelationID())
		return
	}

	err = sendSuccessResponse(w, location, 200)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "error sending API response", err)
	}
}

func (c *LocationController) updateLocation(w http.ResponseWriter, r *http.Request) {
	fnName := "updateLocation"
	appCtx := middleware.GetAppContext(r)

	locationID := chi.URLParam(r, "locationID")
	if locationID == "" {
		errMsg := errors.New("no locationID sent in URL")
		c.logger.Error(fnName, appCtx.GetCorrelationID(), errMsg.Error(), errMsg)
		_ = sendFailureResponse(w, http.StatusBadRequest, errMsg, errMsg.Error(), appCtx.GetCorrelationID())
		return
	}

	updateLocationRequest, err := parseAndValidateBody[dto.UpdateLocationRequest](r.Body, c.validator)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "failed to parse or validate request body", err)
		_ = sendFailureResponse(w, http.StatusBadRequest, err, "failed to parse or validate request body", appCtx.GetCorrelationID())
		return
	}

	if locationID != updateLocationRequest.ID {
		errMsg := errors.New("mismatch between location ID in url and the one in the request payload")
		c.logger.Error(fnName, appCtx.GetCorrelationID(), errMsg.Error(), errMsg)
		_ = sendFailureResponse(w, http.StatusBadRequest, errMsg, errMsg.Error(), appCtx.GetCorrelationID())
		return
	}

	location, err := c.locationService.UpdateLocation(appCtx, updateLocationRequest)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "failed to update location", err)
		_ = sendFailureResponse(w, httpStatusFromError(err), err, "failed to update location", appCtx.GetCorrelationID())
		return
	}

	err = sendSuccessResponse(w, location, 200)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "error sending API response", err)
	}
}

func (c *LocationController) getPaginatedLocations(w http.ResponseWriter, r *http.Request) {
	fnName := "getPaginatedLocations"
	appCtx := middleware.GetAppContext(r)

	filters, err := buildLocationFilters(r)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "error building location filters", err)
		_ = sendFailureResponse(w, http.StatusBadRequest, err, err.Error(), appCtx.GetCorrelationID())
		return
	}

	locationPage, err := c.locationService.GetPaginatedLocations(appCtx, filters)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "failed to get paginated locations", err)
		_ = sendFailureResponse(w, httpStatusFromError(err), err, "failed to get paginated locations", appCtx.GetCorrelationID())
		return
	}

	err = sendSuccessResponse(w, locationPage, 200)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "error sending API response", err)
	}
}

func (c *LocationController) getLocationDetails(w http.ResponseWriter, r *http.Request) {
	fnName := "getLocationDetails"
	appCtx := middleware.GetAppContext(r)

	locationID := chi.URLParam(r, "locationID")
	if locationID == "" {
		errMsg := errors.New("no locationID sent in URL")
		c.logger.Error(fnName, appCtx.GetCorrelationID(), errMsg.Error(), errMsg)
		_ = sendFailureResponse(w, http.StatusBadRequest, errMsg, errMsg.Error(), appCtx.GetCorrelationID())
		return
	}

	location, err := c.locationService.GetLocationByID(appCtx, locationID)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "failed to retrieve location by ID", err)
		_ = sendFailureResponse(w, httpStatusFromError(err), err, err.Error(), appCtx.GetCorrelationID())
		return
	}

	if location == nil {
		errMsg := fmt.Errorf("location with ID %v not found", locationID)
		_ = sendFailureResponse(w, http.StatusNotFound, errMsg, errMsg.Error(), appCtx.GetCorrelationID())
		return
	}

	err = sendSuccessResponse(w, location, 200)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), "error sending API response", err)
	}
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
