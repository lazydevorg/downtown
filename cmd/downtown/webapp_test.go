package main

import (
	"net/http"
	"testing"
)

type MockTransport func(*http.Request) *http.Response

func (m *MockTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	panic("implement me")
}

func TestHome(t *testing.T) {
	//app := App{
	//	Config: &AppConfig{
	//		host: "localhost",
	//		addr: ":4000",
	//	},
	//	Client: Client{
	//		client: http.Client{
	//			Transport: ,
	//		},
	//		host:   "localhost",
	//	},
	//}
	//a := NewWebApp(&app)
}
