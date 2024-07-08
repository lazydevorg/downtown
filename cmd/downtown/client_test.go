package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	EXPECTED_LOGIN_URL = "/webapi/entry.cgi?api=SYNO.API.Auth&version=6&method=login&account=user&passwd=pass&session=DownloadStation&format=sid"
)

func TestSuccessfulDoRequest(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/test-request" {
			t.Errorf("requests url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request")
		}

		w.WriteHeader(http.StatusOK)
		data := []byte("{\"data\":null,\"success\":true}")
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "https://")
	c := NewClient(host, "user", "pass")
	res := new(Response[string])
	err := doRequest(c, "test request", ts.URL+"/test-request", res)
	if err != nil {
		t.Error(err)
	}
}

func TestUnsuccessfulDoRequest(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/test-request" {
			t.Errorf("requests url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request")
		}

		w.WriteHeader(http.StatusOK)
		data := []byte("{\"data\":null,\"success\":false,\"error\":{\"code\":1}}")
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "https://")
	c := NewClient(host, "user", "pass")
	res := new(Response[string])
	err := doRequest(c, "test request", ts.URL+"/test-request", res)
	if err == nil {
		t.Fatal("error expected when response has success field to false")
	}
	if err.Error() != "test request request error: code 1" {
		t.Errorf("error message '%s' expected 'test request request error: code 1'", err.Error())
	}
}

func TestLogin(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != EXPECTED_LOGIN_URL {
			t.Errorf("login url '%s' used but expected '%s'", r.URL.RequestURI(), EXPECTED_LOGIN_URL)
		}

		w.WriteHeader(http.StatusOK)
		data := []byte("{\"data\":{\"did\":\"DIDDID\",\"is_portal_port\":false,\"sid\":\"SIDDIS\"},\"success\":true}")
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "https://")
	c := NewClient(host, "user", "pass")
	err := c.Login()
	if err != nil {
		t.Error(err)
	}
	if c.sid != "SIDDIS" {
		t.Errorf("login sid is '%s' while 'SIDDIS' expected", c.sid)
	}
}
