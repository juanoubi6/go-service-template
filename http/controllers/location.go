package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go-service-template/domain/dto"
	customHTTP "go-service-template/http"
	"go-service-template/http/middleware"
	"go-service-template/log"
	"go-service-template/services"
	"net/http"
)

type LocationController struct {
	logger          log.StdLogger
	locationService services.ILocationService
}

func NewLocationController(locService services.ILocationService) *LocationController {
	return &LocationController{
		locationService: locService,
		logger:          log.GetStdLogger("LocationController"),
	}
}

// Nada godoc
// @Summary POSTCreateLocationEndpoint
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
// @Summary PUTUpdateLocationEndpoint
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
// @Summary GETPaginatedLocationsEndpoint
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
// @Summary GETLocationDetailsEndpoint
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

	var createLocationRequest dto.CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&createLocationRequest); err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), err)
		_ = SendFailureResponse(w, http.StatusBadRequest, err, "failed to parse location data", appCtx.GetCorrelationID())
		return
	}

	location, err := c.locationService.CreateLocation(appCtx, createLocationRequest)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), err)
		_ = SendFailureResponse(w, HTTPStatusFromError(err), err, err.Error(), appCtx.GetCorrelationID())
		return
	}

	err = SendSuccessResponse(w, location, 200)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), fmt.Errorf("error sending API response: %w", err))
	}
}

func (c *LocationController) updateLocation(w http.ResponseWriter, r *http.Request) {
	fnName := "updateLocation"
	appCtx := middleware.GetAppContext(r)

	locationID := chi.URLParam(r, "locationID")
	if locationID == "" {
		errMsg := errors.New("no locationID sent in URL")
		c.logger.Error(fnName, appCtx.GetCorrelationID(), errMsg)
		_ = SendFailureResponse(w, http.StatusBadRequest, errMsg, errMsg.Error(), appCtx.GetCorrelationID())
		return
	}

	var updateLocationRequest dto.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&updateLocationRequest); err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), err)
		_ = SendFailureResponse(w, http.StatusBadRequest, err, "failed to parse location data", appCtx.GetCorrelationID())
		return
	}

	if locationID != updateLocationRequest.ID {
		errMsg := errors.New("mismatch between location ID in url and the one in the request payload")
		c.logger.Error(fnName, appCtx.GetCorrelationID(), errMsg)
		_ = SendFailureResponse(w, http.StatusBadRequest, errMsg, errMsg.Error(), appCtx.GetCorrelationID())
		return
	}

	location, err := c.locationService.UpdateLocation(appCtx, updateLocationRequest)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), err)
		_ = SendFailureResponse(w, HTTPStatusFromError(err), err, err.Error(), appCtx.GetCorrelationID())
		return
	}

	err = SendSuccessResponse(w, location, 200)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), fmt.Errorf("error sending API response: %w", err))
	}
}

func (c *LocationController) getPaginatedLocations(w http.ResponseWriter, r *http.Request) {
	fnName := "getPaginatedLocations"
	appCtx := middleware.GetAppContext(r)

	filters, err := buildLocationFilters(r)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), fmt.Errorf("error building location filters: %w", err))
		_ = SendFailureResponse(w, http.StatusBadRequest, err, err.Error(), appCtx.GetCorrelationID())
		return
	}

	locationPage, err := c.locationService.GetPaginatedLocations(appCtx, filters)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), err)
		_ = SendFailureResponse(w, HTTPStatusFromError(err), err, err.Error(), appCtx.GetCorrelationID())
		return
	}

	err = SendSuccessResponse(w, locationPage, 200)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), fmt.Errorf("error sending API response: %w", err))
	}
}

func (c *LocationController) getLocationDetails(w http.ResponseWriter, r *http.Request) {
	fnName := "getLocationDetails"
	appCtx := middleware.GetAppContext(r)

	locationID := chi.URLParam(r, "locationID")
	if locationID == "" {
		errMsg := errors.New("no locationID sent in URL")
		c.logger.Error(fnName, appCtx.GetCorrelationID(), errMsg)
		_ = SendFailureResponse(w, http.StatusBadRequest, errMsg, errMsg.Error(), appCtx.GetCorrelationID())
		return
	}

	location, err := c.locationService.GetLocationByID(appCtx, locationID)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), err)
		_ = SendFailureResponse(w, HTTPStatusFromError(err), err, err.Error(), appCtx.GetCorrelationID())
		return
	}

	if location == nil {
		errMsg := fmt.Errorf("location with ID %v not found", locationID)
		_ = SendFailureResponse(w, http.StatusNotFound, errMsg, errMsg.Error(), appCtx.GetCorrelationID())
		return
	}

	err = SendSuccessResponse(w, location, 200)
	if err != nil {
		c.logger.Error(fnName, appCtx.GetCorrelationID(), fmt.Errorf("error sending API response: %w", err))
	}
}
