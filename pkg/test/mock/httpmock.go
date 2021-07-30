package mock

import (
	"net/http"

	"github.com/jarcoal/httpmock"
)

func ActivateHttpMock() {
	httpmock.Activate()
}

func DeactivateHttpMock() {
	httpmock.DeactivateAndReset()
}

func NewBytesResponder(status int, body []byte) httpmock.Responder {
	return httpmock.NewBytesResponder(status, body)
}

func NewStringResponder(status int, body string) httpmock.Responder {
	return httpmock.NewStringResponder(status, body)
}

func NewStringResponse(status int, body string) *http.Response {
	return httpmock.NewStringResponse(status, body)
}

func NewJsonResponse(status int, body interface{}) (*http.Response, error) {
	return httpmock.NewJsonResponse(status, body)
}

func RegisterHttpResponder(method string, url string, responder httpmock.Responder) {
	httpmock.RegisterResponder(method, url, responder)
}
