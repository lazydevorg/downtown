package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	ExpectedLoginUrl = "/webapi/entry.cgi?api=SYNO.API.Auth&version=6&method=login&account=user&passwd=pass&session=DownloadStation&format=sid"
)

func TestSuccessfulDoRequest(t *testing.T) {
	c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/test-request" {
			t.Errorf("requests url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request")
		}

		w.WriteHeader(http.StatusOK)
		data := []byte("{\"data\":null,\"success\":true}")
		_, _ = w.Write(data)
	})
	defer s.Close()

	res := new(Response[string])
	req, err := c.createRequest(context.Background(), s.URL+"/test-request")
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	err = doRequest(c, "test request", req, res)
	if err != nil {
		t.Error(err)
	}
}

func TestUnsuccessfulDoRequest(t *testing.T) {
	c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/test-request" {
			t.Errorf("requests url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request")
		}

		w.WriteHeader(http.StatusOK)
		data := []byte("{\"data\":null,\"success\":false,\"error\":{\"code\":1}}")
		_, _ = w.Write(data)
	})
	defer s.Close()

	res := new(Response[string])
	req, err := c.createRequest(context.Background(), s.URL+"/test-request")
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	err = doRequest(c, "test request", req, res)
	if err == nil {
		t.Fatal("error expected when response has success field to false")
	}
	if err.Error() != "test request request error: code 1" {
		t.Errorf("error message '%s' expected 'test request request error: code 1'", err.Error())
	}
}

func TestLogin(t *testing.T) {
	c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != ExpectedLoginUrl {
			t.Errorf("login url '%s' used but expected '%s'", r.URL.RequestURI(), ExpectedLoginUrl)
		}

		w.WriteHeader(http.StatusOK)
		data := []byte("{\"data\":{\"did\":\"DIDDID\",\"is_portal_port\":false,\"sid\":\"SIDDIS\"},\"success\":true}")
		_, _ = w.Write(data)
	})
	defer s.Close()

	res, err := c.Login(context.Background(), LoginRequest{
		user: "user",
		pass: "pass",
	})
	if err != nil {
		t.Error(err)
	}
	if res.Data.SID != "SIDDIS" {
		t.Errorf("login sid is '%s' while 'SIDDIS' expected", res.Data.SID)
	}
}

func testClient(f http.HandlerFunc) (*Client, *httptest.Server) {
	ts := httptest.NewTLSServer(f)
	host := strings.TrimPrefix(ts.URL, "https://")
	return NewClient(host), ts
}

func sequantialRequests(f ...http.HandlerFunc) http.HandlerFunc {
	index := 0
	return func(w http.ResponseWriter, r *http.Request) {
		f[index](w, r)
		index++
	}
}
