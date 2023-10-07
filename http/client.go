package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-service-template/config"
	"go-service-template/monitor"
	"go-service-template/utils"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	DefaultMaxIdleConns           = 100
	DefaultMaxConnsPerHost        = 100
	DefaultMaxIdleConnsPerHost    = 50
	DefaultIdleConnTimeoutSeconds = 90
	DefaultRequestTimeoutSeconds  = 30
	DefaultRetryAmount            = 3
	DefaultDialerTimeoutSec       = 30
	DefaultDialerKeepAliveSec     = 30
	DefaultTLSHandshakeSec        = 10
)

var ErrRetryAmountExceeded = errors.New("failed to execute request, retry amount exceeded")

type CustomHTTPClient interface {
	Do(ctx monitor.ApplicationContext, requestValues RequestValues) (CustomHTTPResponse, error)
	DoWithRetry(
		ctx monitor.ApplicationContext,
		requestValues RequestValues,
		timeout time.Duration,
		retryAmount int,
		backoff time.Duration,
		specialStatusCodesToRetry []int,
	) (CustomHTTPResponse, error)
}

type RequestValues struct {
	URL       string
	Method    string
	Headers   http.Header
	Body      any
	BasicAuth *BasicAuth
}

type BasicAuth struct {
	Username string
	Password string
}

type CustomClient struct {
	baseClient *http.Client
	logger     monitor.AppLogger
}

type CustomHTTPResponse struct {
	StatusCode   int
	BodyPayload  []byte
	Headers      http.Header
	BaseResponse *http.Response
}

func CreateCustomHTTPClient(cfg config.HTTPClientConfig) *CustomClient {
	baseHTTPClient := buildClient(cfg)

	return &CustomClient{
		baseClient: baseHTTPClient,
		logger:     monitor.GetStdLogger("CustomHTTPClient"),
	}
}

func buildClient(cfg config.HTTPClientConfig) *http.Client {
	var maxIdleConns = config.GetIntValueOrDefault(cfg.MaxIdleConns, DefaultMaxIdleConns)
	var maxConnsPerHost = config.GetIntValueOrDefault(cfg.MaxConnsPerHost, DefaultMaxConnsPerHost)
	var maxIdleConnsPerHost = config.GetIntValueOrDefault(cfg.MaxIdleConnsPerHost, DefaultMaxIdleConnsPerHost)
	var idleConnTimeoutSeconds = config.GetIntValueOrDefault(cfg.IdleConnTimeoutSeconds, DefaultIdleConnTimeoutSeconds)
	var requestTimeoutSeconds = config.GetIntValueOrDefault(cfg.RequestTimeoutSeconds, DefaultRequestTimeoutSeconds)

	transport := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   DefaultDialerTimeoutSec * time.Second,
			KeepAlive: DefaultDialerKeepAliveSec * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: DefaultTLSHandshakeSec * time.Second,
		MaxIdleConns:        maxIdleConns,
		MaxConnsPerHost:     maxConnsPerHost,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(idleConnTimeoutSeconds) * time.Second,
	}

	return &http.Client{
		Transport: otelhttp.NewTransport(&transport),
		Timeout:   time.Duration(requestTimeoutSeconds) * time.Second,
	}
}

func (cli *CustomClient) Do(ctx monitor.ApplicationContext, requestValues RequestValues) (CustomHTTPResponse, error) {
	fnName := "Do"

	req, err := cli.buildHTTPRequest(ctx, requestValues)
	if err != nil {
		return CustomHTTPResponse{}, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	resp, err := cli.baseClient.Do(req) //nolint
	if err != nil {
		cli.logger.ErrorCtx(ctx, fnName, "http request failed", err)
		return CustomHTTPResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}

	return cli.handleResponse(ctx, resp, fnName)
}

func (cli *CustomClient) DoWithRetry(
	ctx monitor.ApplicationContext,
	requestValues RequestValues,
	timeout time.Duration,
	retryAmount int,
	backoff time.Duration,
	statusCodesToRetry []int,
) (CustomHTTPResponse, error) {
	fnName := "DoWithRetry"

	var resp *http.Response
	var err error
	var request *http.Request
	var attempts = retryAmount

	var timeOutCtx context.Context
	var cancelFn context.CancelFunc

	for attempts > 0 {
		// Build request
		request, err = cli.buildHTTPRequest(ctx, requestValues)
		if err != nil {
			break
		}

		// Add timeout
		timeOutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		newRequestWithTimeout := request.WithContext(timeOutCtx)

		// Execute request
		resp, err = cli.baseClient.Do(newRequestWithTimeout)
		if err == nil && resp.StatusCode < http.StatusInternalServerError && !utils.ListContains[int](statusCodesToRetry, resp.StatusCode) {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if err != nil {
			cli.logger.ErrorCtx(ctx, fnName, "http request failed", err)
		}

		attempts--
		cancelFn()
		time.Sleep(backoff)
	}

	if cancelFn != nil {
		defer cancelFn()
	}

	if err != nil {
		return CustomHTTPResponse{}, fmt.Errorf("failed to execute request, unexpected error. Error: %w", err)
	}
	if attempts == 0 {
		return CustomHTTPResponse{}, ErrRetryAmountExceeded
	}

	return cli.handleResponse(ctx, resp, fnName)
}

func (cli *CustomClient) handleResponse(
	ctx monitor.ApplicationContext,
	response *http.Response,
	functionName string,
) (CustomHTTPResponse, error) {
	defer func() {
		err := response.Body.Close()
		if err != nil {
			cli.logger.ErrorCtx(ctx, functionName, "failed to close response body", err)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		cli.logger.ErrorCtx(ctx, functionName, "failed to read response body", err)
		return CustomHTTPResponse{}, err
	}

	customResponse := CustomHTTPResponse{
		StatusCode:   response.StatusCode,
		Headers:      response.Header,
		BodyPayload:  body,
		BaseResponse: response,
	}

	cli.logger.InfoCtx(ctx, functionName, "HTTP request successful",
		monitor.LoggingParam{
			Name: "url", Value: response.Request.URL.String(),
		},
		monitor.LoggingParam{
			Name: "status_code", Value: response.StatusCode,
		},
	)

	cli.logger.DebugCtx(ctx, functionName, "HTTP request body",
		monitor.LoggingParam{
			Name: "response_body", Value: string(customResponse.BodyPayload),
		},
	)

	return customResponse, nil
}

func (cli *CustomClient) buildHTTPRequest(ctx monitor.ApplicationContext, rv RequestValues) (*http.Request, error) {
	var body io.Reader = http.NoBody

	// If body is not nil, create new body
	if rv.Body != nil {
		data, err := json.Marshal(rv.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(data)
	}

	// Build HTTP request
	request, err := http.NewRequestWithContext(ctx, rv.Method, rv.URL, body)
	if err != nil {
		return nil, err
	}

	// Append headers
	if rv.Headers != nil {
		request.Header = rv.Headers
	}

	// Set basic auth
	if rv.BasicAuth != nil {
		request.SetBasicAuth(rv.BasicAuth.Username, rv.BasicAuth.Password)
	}

	return request, nil
}
