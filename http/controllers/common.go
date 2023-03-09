package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go-service-template/domain"
	"go-service-template/utils"
	"io"
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

var (
	businessErr         = &domain.BusinessErr{}
	nameAlreadyInUseErr = &domain.NameAlreadyInUseErr{}
	addressNotValidErr  = &domain.AddressNotValidErr{}
)

type APIResponse struct {
	Error *APIError `json:"error,omitempty"`
	Data  any       `json:"data,omitempty"`
}

type APIError struct {
	Type          string   `json:"type,omitempty"`
	Title         string   `json:"title"`
	Details       []Detail `json:"details"`
	CorrelationID string   `json:"correlation_id"`
}

type Detail struct {
	Message string            `json:"message"`
	Meta    map[string]string `json:"metadata,omitempty"`
}

func buildSuccessResponse(payload any) APIResponse {
	return APIResponse{
		Data: payload,
	}
}

func buildFailResponse(err error, title, correlationID string) APIResponse {
	return APIResponse{
		Error: &APIError{
			Title:         title,
			CorrelationID: correlationID,
			Details:       errorDetailsFromError(err),
		},
	}
}

func httpStatusFromError(err error) int {
	switch {
	case errors.As(err, nameAlreadyInUseErr), errors.As(err, addressNotValidErr), errors.As(err, businessErr):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func errorDetailsFromError(err error) []Detail {
	var details []Detail

	switch errT := err.(type) {
	case validator.ValidationErrors:
		for _, validationErr := range errT {
			details = append(details, Detail{Message: validationErr.Error()})
		}
	default:
		details = append(details, Detail{Message: errT.Error()})
	}

	return details
}

func parseAndValidateBody[T any](body io.ReadCloser, v *validator.Validate) (T, error) {
	var bodyStruct T
	if err := json.NewDecoder(body).Decode(&bodyStruct); err != nil {
		return bodyStruct, err
	}

	if err := v.Struct(bodyStruct); err != nil {
		return bodyStruct, err
	}

	return bodyStruct, nil
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
