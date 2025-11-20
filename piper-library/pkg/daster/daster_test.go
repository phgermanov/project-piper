package daster

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

// TestReadResponse handles testing of readResponse function
func TestReadResponse(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		response := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("success")),
		}

		body, err := readResponse(response, false)
		assert.NoError(t, err)
		assert.NotEmpty(t, body)
		assert.Equal(t, "success", string(body))
	})
	t.Run("error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1")
		}))
		defer ts.Close()

		resp, _ := http.Get(ts.URL)
		_, err := readResponse(resp, true)
		assert.Error(t, err)
	})
}

func TestGetRequestBody(t *testing.T) {
	t.Parallel()

	t.Run("nil request body", func(t *testing.T) {
		reqBody, err := getRequestBody(nil)
		assert.NoError(t, err)
		assert.Empty(t, reqBody)
	})

	t.Run("map in request body", func(t *testing.T) {
		body := map[string]interface{}{
			"param1": "value1",
			"param2": 2,
		}
		reqBody, err := getRequestBody(body)
		assert.NoError(t, err)
		assert.NotEmpty(t, reqBody)
		assert.True(t, strings.Contains(string(reqBody), `"param1":"value1"`))
		assert.True(t, strings.Contains(string(reqBody), `"param2":2`))
	})

	t.Run("custom struct in request body", func(t *testing.T) {
		body := struct {
			Name string
			Age  int
		}{Name: "John", Age: 30}
		reqBody, err := getRequestBody(body)
		assert.NoError(t, err)
		assert.NotEmpty(t, reqBody)
		assert.True(t, strings.Contains(string(reqBody), `"Name":"John"`))
		assert.True(t, strings.Contains(string(reqBody), `"Age":30`))
	})

	t.Run("unsupported type", func(t *testing.T) {
		body := make(chan int)
		_, err := getRequestBody(body)
		assert.Error(t, err)
	})
}

type ClientMock struct {
	attemptsLeft int
	isError      bool
	returnValue  string
}

func newClientMock(attemptsLeft int, isErrorExpected bool, returnValue string) *ClientMock {
	return &ClientMock{
		attemptsLeft: attemptsLeft,
		isError:      isErrorExpected,
		returnValue:  returnValue,
	}
}

func (c *ClientMock) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	if url == "" {
		return nil, errors.New("test error")
	}
	if c.isError {
		return &http.Response{
			StatusCode: 403,
			Body:       io.NopCloser(strings.NewReader("error")),
		}, nil
	}
	if c.attemptsLeft == 0 {
		if len(c.returnValue) == 0 {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "1")
			}))
			defer ts.Close()

			resp, _ := http.Get(ts.URL)
			return resp, nil
		}
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(c.returnValue)),
		}
		return resp, nil
	}
	c.attemptsLeft--
	return &http.Response{
		StatusCode: 100,
		Body:       io.NopCloser(strings.NewReader("retry")),
	}, nil
}

func (c *ClientMock) SetOptions(options piperhttp.ClientOptions) {}

func TestCallAPI(t *testing.T) {
	t.Parallel()
	t.Run("ok with first attempt", func(t *testing.T) {
		result, err := callAPI(newClientMock(0, false, "ok"), "url", "GET", nil, nil, false, 1, time.Second)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
	t.Run("ok with retries", func(t *testing.T) {
		result, err := callAPI(newClientMock(1, false, "ok"), "url", "GET", nil, nil, false, 2, time.Second)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
	t.Run("error from api", func(t *testing.T) {
		result, err := callAPI(newClientMock(0, true, ""), "url", "GET", nil, nil, false, 2, time.Second)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "error 403 from")
	})
	t.Run("timeout error", func(t *testing.T) {
		result, err := callAPI(newClientMock(2, false, ""), "url", "GET", nil, nil, false, 1, time.Second)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "timeout error occurred")
	})
	t.Run("fail sending request", func(t *testing.T) {
		result, err := callAPI(newClientMock(0, false, "ok"), "", "GET", nil, nil, true, 1, time.Second)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "test error")
	})
	t.Run("fail reading response", func(t *testing.T) {
		result, err := callAPI(newClientMock(0, false, ""), "url", "GET", nil, nil, true, 1, time.Second)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.ErrorContains(t, err, "unexpected EOF")
	})
}

func TestFetchOAuthToken(t *testing.T) {
	t.Parallel()
	t.Run("ok", func(t *testing.T) {
		response := OAuthTokenResponse{AccessToken: "token"}
		resp, err := json.Marshal(response)
		assert.NoError(t, err)
		client := newClientMock(0, false, string(resp))
		result, err := fetchOAuthToken(client, "url", "grandType", "source", "clientId", "secret", false, time.Second)
		assert.NoError(t, err)
		assert.Equal(t, "token", result)
	})
	t.Run("fail calling API", func(t *testing.T) {
		client := newClientMock(0, true, "")
		result, err := fetchOAuthToken(client, "url", "grandType", "source", "clientId", "secret", false, time.Second)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "error")
		assert.Empty(t, result)
	})
	t.Run("fail unmarshalling response", func(t *testing.T) {
		client := newClientMock(0, false, "token")
		result, err := fetchOAuthToken(client, "url", "grandType", "source", "clientId", "secret", false, time.Second)
		assert.Error(t, err)
		assert.Empty(t, result)
	})
}

func TestCallScanAPI(t *testing.T) {
	t.Parallel()
	t.Run("ok", func(t *testing.T) {
		resp := "ok"
		req := "request"
		client := newClientMock(0, false, resp)
		result, err := callScanAPI(client, "url", "GET", req, false, 5, time.Second)
		assert.NoError(t, err)
		assert.Equal(t, resp, string(result))
	})
	t.Run("fail calling API", func(t *testing.T) {
		req := "request"
		client := newClientMock(0, true, "")
		result, err := callScanAPI(client, "url", "GET", req, false, 5, time.Second)
		assert.Error(t, err)
		assert.Empty(t, result)
	})
	t.Run("fail getting request body", func(t *testing.T) {
		req := make(chan int)
		client := newClientMock(0, false, "")
		_, err := callScanAPI(client, "url", "GET", req, false, 5, time.Second)
		assert.Error(t, err)
	})
}
