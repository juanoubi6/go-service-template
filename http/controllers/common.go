package controllers

import (
	"encoding/json"
	"fmt"
	"go-service-template/domain"
	"go-service-template/utils"
	"net/http"
	"strconv"
)

const (
	DefaultLimit                 = 10000
	DefaultCursorPaginationValue = ""
	CursorQP                     = "cursor"
	DirectionQP                  = "direction"
	LimitQP                      = "limit"
)

type APIResponse struct {
	ApiError *APIError `json:"error,omitempty"`
	Data     any       `json:"data,omitempty"`
}

type APIError struct {
	Type          string   `json:"type"`
	Title         string   `json:"title"`
	Details       []Detail `json:"details"`
	CorrelationId string   `json:"correlation_id"`
}

type Detail struct {
	Message string            `json:"message"`
	Meta    map[string]string `json:"metadata"`
}

func SendSuccessResponse(w http.ResponseWriter, payload any, statusCode int) error {
	response := APIResponse{
		Data: payload,
	}

	err := sendResponse(w, response, statusCode)
	if err != nil {
		return err
	}

	return nil
}

func SendFailureResponse(w http.ResponseWriter, statusCode int, err error, title, correlationID string) error {
	response := APIResponse{
		ApiError: &APIError{
			Title:         title,
			CorrelationId: correlationID,
			Details: []Detail{
				{Message: err.Error()},
			},
		},
	}

	err = sendResponse(w, response, statusCode)
	if err != nil {
		return err
	}

	return nil
}

func sendResponse(w http.ResponseWriter, payload any, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func HTTPStatusFromError(err error) int {
	switch err.(type) {
	case domain.NameAlreadyInUseErr, domain.AddressNotValidErr, domain.BusinessErr:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
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

func buildCursorPaginationFilters(req *http.Request) (domain.CursorPaginationFilters, error) {
	var cursorPagFilters domain.CursorPaginationFilters
	queryString := req.URL.Query()

	if cursorVal, ok := queryString[CursorQP]; ok {
		cursorPagFilters.Cursor = cursorVal[0]
	} else {
		cursorPagFilters.Cursor = DefaultCursorPaginationValue
	}

	if limit, ok := queryString[LimitQP]; ok {
		limitVal, err := strconv.Atoi(limit[0])
		if err != nil {
			return cursorPagFilters, fmt.Errorf("invalid limit value: %v", limit[0])
		}
		cursorPagFilters.Limit = limitVal
	} else {
		cursorPagFilters.Limit = DefaultLimit
	}

	if directionVal, ok := queryString[DirectionQP]; ok {
		if utils.ListContains([]string{domain.PreviousPage, domain.NextPage}, directionVal[0]) {
			cursorPagFilters.Direction = directionVal[0]
		} else {
			return cursorPagFilters, fmt.Errorf("invalid direction value: %v", directionVal[0])
		}
	} else {
		return cursorPagFilters, fmt.Errorf("'direction' query param not provided")
	}

	// Validate initial cursor
	if cursorPagFilters.Cursor == "" && cursorPagFilters.Direction != domain.NextPage {
		return cursorPagFilters, fmt.Errorf("if the cursor is empty, the only allowed direction value is '%v'", domain.NextPage)
	}

	return cursorPagFilters, nil
}
