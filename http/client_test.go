package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-service-template/config"
	customHTTP "go-service-template/http"
	"go-service-template/monitor"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var timesExecuted int
var mockCtx = monitor.CreateAppContext(context.Background(), "")

type MockBody struct {
	SomeKey string `json:"some_key"`
}

type CustomHTTPClientSuite struct {
	suite.Suite
	httpClient     *customHTTP.CustomClient
	testHTTPServer *httptest.Server
}

func (s *CustomHTTPClientSuite) SetupSuite() {
	client := customHTTP.CreateCustomHTTPClient(config.HTTPClientConfig{})
	s.httpClient = client
	timesExecuted = 0
	s.configureTestHTTPServer()
}

func (s *CustomHTTPClientSuite) configureTestHTTPServer() {
	done := make(chan bool)
	go func() {
		testMux := http.NewServeMux()
		testMux.HandleFunc("/auth-and-headers", func(w http.ResponseWriter, r *http.Request) {
			headerValue := r.Header.Get("headerVal")
			if headerValue == "" {
				w.WriteHeader(400)
				return
			}

			_, _, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(400)
				return
			}

			w.WriteHeader(200)
		})
		testMux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
			_, _, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(400)
				return
			}

			w.WriteHeader(200)
		})
		testMux.HandleFunc("/resource", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("someKey", "someVal")
			_, _ = fmt.Fprintf(w, "{\"data\":{\"id\":50}}")
		})
		testMux.HandleFunc("/resource-with-empty-body", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
		})
		testMux.HandleFunc("/slow-response-resource", func(w http.ResponseWriter, r *http.Request) {
			timesExecuted++
			time.Sleep(time.Second)
			w.WriteHeader(200)
		})
		testMux.HandleFunc("/works-on-second-attempt", func(w http.ResponseWriter, r *http.Request) {
			if timesExecuted == 0 {
				timesExecuted++
				time.Sleep(time.Second)
				w.WriteHeader(201)
			} else {
				w.WriteHeader(201)
			}
		})
		testMux.HandleFunc("/returns-400", func(w http.ResponseWriter, r *http.Request) {
			timesExecuted++
			w.WriteHeader(400)
		})
		testMux.HandleFunc("/returns-500-and-200-on-second-call", func(w http.ResponseWriter, r *http.Request) {
			if timesExecuted > 0 {
				w.WriteHeader(200)
				return
			}
			timesExecuted++
			w.WriteHeader(500)
		})
		testMux.HandleFunc("/post-retry", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(400)
				return
			}

			var payload MockBody
			err := json.NewDecoder(r.Body).Decode(&payload)
			if err != nil {
				w.WriteHeader(500)
				return
			}

			if payload.SomeKey == "" {
				w.WriteHeader(500)
				return
			}

			if timesExecuted < 2 {
				timesExecuted++
				w.WriteHeader(400)
				return
			}

			w.WriteHeader(200)
		})
		testServer := httptest.NewServer(testMux)
		s.testHTTPServer = testServer
		done <- true
	}()
	<-done
}

func (s *CustomHTTPClientSuite) SetupTest() {
	timesExecuted = 0
}

func (s *CustomHTTPClientSuite) TearDownSuite() {
	_ = s.testHTTPServer.Close
}

func TestCustomHttpClientSuite(t *testing.T) {
	suite.Run(t, new(CustomHTTPClientSuite))
}

func (s *CustomHTTPClientSuite) Test_Do_Success() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/resource",
		Method:    http.MethodGet,
		Headers:   nil,
		Body:      nil,
		BasicAuth: nil,
	}

	resp, err := s.httpClient.Do(mockCtx, requestValues)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 200, resp.StatusCode)
	assert.Equal(s.T(), "someVal", resp.Headers.Get("someKey"))
	assert.Equal(s.T(), "{\"data\":{\"id\":50}}", string(resp.BodyPayload))
	assert.NotNil(s.T(), resp.BaseResponse)
}

func (s *CustomHTTPClientSuite) Test_BasicAuthAndHeadersAreAdded() {
	header := http.Header{}
	header.Add("headerVal", "someValue")

	requestValues := customHTTP.RequestValues{
		URL:     s.testHTTPServer.URL + "/auth-and-headers",
		Method:  http.MethodGet,
		Headers: header,
		Body:    nil,
		BasicAuth: &customHTTP.BasicAuth{
			Username: "someUser",
			Password: "somePass",
		},
	}

	resp, err := s.httpClient.Do(mockCtx, requestValues)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 200, resp.StatusCode)
	assert.NotNil(s.T(), resp.BaseResponse)
}

func (s *CustomHTTPClientSuite) Test_BasicAuthWithoutHeadersAdded() {
	header := http.Header{}
	header.Add("headerVal", "someValue")

	requestValues := customHTTP.RequestValues{
		URL:     s.testHTTPServer.URL + "/auth",
		Method:  http.MethodGet,
		Headers: nil,
		Body:    nil,
		BasicAuth: &customHTTP.BasicAuth{
			Username: "someUser",
			Password: "somePass",
		},
	}

	resp, err := s.httpClient.Do(mockCtx, requestValues)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 200, resp.StatusCode)
	assert.NotNil(s.T(), resp.BaseResponse)
}

func (s *CustomHTTPClientSuite) Test_DoWithEmptyResponse_Success() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/resource-with-empty-body",
		Method:    http.MethodGet,
		Headers:   nil,
		Body:      nil,
		BasicAuth: nil,
	}

	resp, err := s.httpClient.Do(mockCtx, requestValues)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 201, resp.StatusCode)
	assert.NotNil(s.T(), resp.BaseResponse)
}

func (s *CustomHTTPClientSuite) Test_DoWithRetry_Success() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/resource",
		Method:    http.MethodGet,
		Headers:   nil,
		Body:      nil,
		BasicAuth: nil,
	}

	resp, err := s.httpClient.DoWithRetry(mockCtx, requestValues, time.Second, 2, time.Millisecond*100, []int{})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, timesExecuted)
	assert.Equal(s.T(), 200, resp.StatusCode)
	assert.Equal(s.T(), "someVal", resp.Headers.Get("someKey"))
	assert.Equal(s.T(), "{\"data\":{\"id\":50}}", string(resp.BodyPayload))
	assert.NotNil(s.T(), resp.BaseResponse)
}

func (s *CustomHTTPClientSuite) Test_DoWithRetry_ExecutesAllRetriesAndFail() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/slow-response-resource",
		Method:    http.MethodGet,
		Headers:   nil,
		Body:      nil,
		BasicAuth: nil,
	}

	startTime := time.Now()
	_, err := s.httpClient.DoWithRetry(mockCtx, requestValues, time.Millisecond*500, 3, time.Millisecond*100, []int{})
	totalTime := time.Since(startTime)

	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), 3, timesExecuted)
	assert.True(s.T(), totalTime <= time.Second*2) // 500ms timeout time * 3 + 100ms backoff * 3 = 1800ms
}

func (s *CustomHTTPClientSuite) Test_DoWithRetry_ExecutesOneRetryAndSuccess() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/works-on-second-attempt",
		Method:    http.MethodGet,
		Headers:   nil,
		Body:      nil,
		BasicAuth: nil,
	}

	startTime := time.Now()
	resp, err := s.httpClient.DoWithRetry(mockCtx, requestValues, time.Millisecond*500, 3, time.Millisecond*100, []int{})
	totalTime := time.Since(startTime)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 201, resp.StatusCode)
	assert.Equal(s.T(), 1, timesExecuted)
	assert.True(s.T(), totalTime <= time.Second*1) // 500ms timeout time first time + 100ms backoff = 600ms
}

func (s *CustomHTTPClientSuite) Test_DoWithRetry_DoesNotRetryOn4xxResponses() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/returns-400",
		Method:    http.MethodGet,
		Headers:   nil,
		Body:      nil,
		BasicAuth: nil,
	}

	resp, err := s.httpClient.DoWithRetry(mockCtx, requestValues, time.Second, 2, time.Millisecond*100, []int{})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, timesExecuted)
	assert.Equal(s.T(), 400, resp.StatusCode)
	assert.NotNil(s.T(), resp.BaseResponse)
}

func (s *CustomHTTPClientSuite) Test_DoWithRetry_DoesRetryOn5xxResponses() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/returns-500-and-200-on-second-call",
		Method:    http.MethodGet,
		Headers:   nil,
		Body:      nil,
		BasicAuth: nil,
	}

	resp, err := s.httpClient.DoWithRetry(mockCtx, requestValues, time.Millisecond*500, 3, time.Millisecond*100, []int{})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, timesExecuted)
	assert.Equal(s.T(), 200, resp.StatusCode)
	assert.NotNil(s.T(), resp.BaseResponse)
}

func (s *CustomHTTPClientSuite) Test_DoWithRetry_ExecutesRetriesOnSpecialCodes() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/returns-400",
		Method:    http.MethodGet,
		Headers:   nil,
		Body:      nil,
		BasicAuth: nil,
	}

	_, err := s.httpClient.DoWithRetry(mockCtx, requestValues, time.Second, 3, time.Millisecond*100, []int{http.StatusBadRequest})

	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), 3, timesExecuted)
}

func (s *CustomHTTPClientSuite) Test_DoWithRetry_ExecutesPostRequestWith2RetriesAndSuccess() {
	requestValues := customHTTP.RequestValues{
		URL:       s.testHTTPServer.URL + "/post-retry",
		Method:    http.MethodPost,
		Headers:   nil,
		Body:      MockBody{SomeKey: "someValue"},
		BasicAuth: nil,
	}

	resp, err := s.httpClient.DoWithRetry(mockCtx, requestValues, time.Millisecond*500, 3, time.Millisecond*100, []int{http.StatusBadRequest})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 200, resp.StatusCode)
	assert.Equal(s.T(), 2, timesExecuted)
}
