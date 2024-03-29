package googlemaps

import (
	"encoding/json"
	"errors"
	"go-service-template/domain/googlemaps"
	customHTTP "go-service-template/http"
	"go-service-template/monitor"
	"net/http"
	"time"
)

const (
	premisePlaceType   = "premise"
	defaultTimeoutSecs = 5
)

var ErrGenericGoogleErr = errors.New("error from Google Maps API")

type Repository struct {
	logger     monitor.AppLogger
	httpClient customHTTP.CustomHTTPClient
}

func NewGoogleMapsRepository(httpClient customHTTP.CustomHTTPClient) *Repository {
	return &Repository{
		logger:     monitor.GetStdLogger("Repository"),
		httpClient: httpClient,
	}
}

// ValidateAddress is just a simplified example on how do we call external services
func (r *Repository) ValidateAddress(ctx monitor.ApplicationContext, request googlemaps.AddressValidationRequest) (*googlemaps.AddressValidateMatch, error) {
	fnName := "ValidateAddress"

	requestValues := customHTTP.RequestValues{
		URL:     "google maps address validation URL",
		Method:  http.MethodPost,
		Headers: nil,
		Body:    request,
	}

	res, err := r.httpClient.DoWithRetry(ctx, requestValues, defaultTimeoutSecs*time.Second, customHTTP.DefaultRetryAmount, time.Second, []int{})
	if err != nil {
		return nil, err
	}

	r.logger.InfoCtx(ctx, fnName,
		"Validate address endpoint response",
		monitor.LoggingParam{
			Name: "response_metadata",
			Value: map[string]interface{}{
				"status_code": res.StatusCode,
				"body":        string(res.BodyPayload),
			},
		},
	)

	if res.StatusCode != http.StatusOK {
		// If error is HttpNotFound, the address could not be validated
		if res.StatusCode == http.StatusNotFound {
			return nil, nil
		}

		// Else, log error and return
		r.logger.ErrorCtx(ctx, fnName, ErrGenericGoogleErr.Error(), ErrGenericGoogleErr, monitor.LoggingParam{
			Name:  "error_payload",
			Value: string(res.BodyPayload),
		})

		return nil, ErrGenericGoogleErr
	}

	var response googlemaps.AddressValidationResponse
	if err = json.Unmarshal(res.BodyPayload, &response); err != nil {
		return nil, err
	}

	// If no matches were found, return nil
	if len(response.Matches) == 0 {
		return nil, nil
	}

	// If any match was found, return the first match whose type is "premise"
	for _, match := range response.Matches {
		if match.MatchType == premisePlaceType {
			return &match, nil
		}
	}

	return nil, nil
}
