package controllers

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go-service-template/domain"
	"net/http"
	"testing"
)

func Test_httpStatusFromError_returns400OnBusinessErr(t *testing.T) {
	testErr := domain.BusinessErr{Msg: "err"}
	wrappedTestErr := fmt.Errorf("some err: %w", testErr)

	code := httpStatusFromError(wrappedTestErr)

	assert.Equal(t, http.StatusBadRequest, code)
}

func Test_httpStatusFromError_returns400OnAddressNotValidErr(t *testing.T) {
	testErr := domain.AddressNotValidErr{Msg: "err"}
	wrappedTestErr := fmt.Errorf("some err: %w", testErr)

	code := httpStatusFromError(wrappedTestErr)

	assert.Equal(t, http.StatusBadRequest, code)
}

func Test_httpStatusFromError_returns400OnNameAlreadyInUseErr(t *testing.T) {
	testErr := domain.NameAlreadyInUseErr{Msg: "err"}
	wrappedTestErr := fmt.Errorf("some err: %w", testErr)

	code := httpStatusFromError(wrappedTestErr)

	assert.Equal(t, http.StatusBadRequest, code)
}

func Test_httpStatusFromError_returns500OnUnhandledError(t *testing.T) {
	unknownErr := errors.New("some err")

	code := httpStatusFromError(unknownErr)

	assert.Equal(t, http.StatusInternalServerError, code)
}
