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
	err := doRequest(c, "test request", s.URL+"/test-request", res)
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
	err := doRequest(c, "test request", s.URL+"/test-request", res)
	if err == nil {
		t.Fatal("error expected when response has success field to false")
	}
	if err.Error() != "test request request error: code 1" {
		t.Errorf("error message '%s' expected 'test request request error: code 1'", err.Error())
	}
}

func TestDoAuthenticatedRequest(t *testing.T) {
	t.Run("Authenticate first when SID not available", func(t *testing.T) {
		successfulRequestMade := false
		c, s := testClient(sequantialRequests(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RequestURI() != EXPECTED_LOGIN_URL {
					t.Errorf("login url '%s' used but expected '%s'", r.URL.RequestURI(), EXPECTED_LOGIN_URL)
				}

				w.WriteHeader(http.StatusOK)
				data := []byte("{\"data\":{\"did\":\"DIDDID\",\"is_portal_port\":false,\"sid\":\"SIDDIS\"},\"success\":true}")
				_, _ = w.Write(data)
			},
			func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.RequestURI(), "/test-request&_sid=SIDDIS") {
					t.Errorf("url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request&_sid=SIDDIS")
				}

				w.WriteHeader(http.StatusOK)
				data := []byte("{\"data\":\"response\",\"success\":true}")
				_, _ = w.Write(data)
				successfulRequestMade = true
			}))
		defer s.Close()

		res := new(Response[string])
		err := doAuthenticatedRequest(c, "test request", s.URL+"/test-request", res)
		if err != nil {
			t.Error(err)
		}
		if !successfulRequestMade {
			t.Error("Final successful request was not made")
		}
	})

	t.Run("Don't authenticate when SID is available", func(t *testing.T) {
		c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasSuffix(r.URL.RequestURI(), "/test-request&_sid=SIDDIS") {
				t.Errorf("url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request&_sid=SIDDIS")
			}

			w.WriteHeader(http.StatusOK)
			data := []byte("{\"data\":\"response\",\"success\":true}")
			_, _ = w.Write(data)
		})
		defer s.Close()

		res := new(Response[string])
		c.sid = "SIDDIS"
		err := doAuthenticatedRequest(c, "test request", s.URL+"/test-request", res)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Make request, authenticate and retry after the request fails when SID expired", func(t *testing.T) {
		successfulRequestMade := false
		c, s := testClient(sequantialRequests(
			func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.RequestURI(), "/test-request&_sid=EXPIRED") {
					t.Errorf("url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request&_sid=SIDDIS")
				}

				w.WriteHeader(http.StatusOK)
				data := []byte("{\"data\":null,\"success\":false,\"error\":{\"code\":105}}")
				_, _ = w.Write(data)
			},
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RequestURI() != EXPECTED_LOGIN_URL {
					t.Errorf("login url '%s' used but expected '%s'", r.URL.RequestURI(), EXPECTED_LOGIN_URL)
				}

				w.WriteHeader(http.StatusOK)
				data := []byte("{\"data\":{\"did\":\"DIDDID\",\"is_portal_port\":false,\"sid\":\"SIDDIS\"},\"success\":true}")
				_, _ = w.Write(data)
			},
			func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.RequestURI(), "/test-request&_sid=SIDDIS") {
					t.Errorf("url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request&_sid=SIDDIS")
				}

				w.WriteHeader(http.StatusOK)
				data := []byte("{\"data\":\"response\",\"success\":true}")
				_, _ = w.Write(data)
				successfulRequestMade = true
			}))
		defer s.Close()

		res := new(Response[string])
		c.sid = "EXPIRED"
		err := doAuthenticatedRequest(c, "test request", s.URL+"/test-request", res)
		if err != nil {
			t.Error(err)
		}
		if !successfulRequestMade {
			t.Error("Final successful request was not made")
		}
	})

	t.Run("Make request, authenticate and don't retry when login fails", func(t *testing.T) {
		loginRequestMade := false
		c, s := testClient(sequantialRequests(
			func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.RequestURI(), "/test-request&_sid=EXPIRED") {
					t.Errorf("url '%s' used but expected '%s'", r.URL.RequestURI(), "/test-request&_sid=SIDDIS")
				}

				w.WriteHeader(http.StatusOK)
				data := []byte("{\"data\":null,\"success\":false,\"error\":{\"code\":105}}")
				_, _ = w.Write(data)
			},
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.RequestURI() != EXPECTED_LOGIN_URL {
					t.Errorf("login url '%s' used but expected '%s'", r.URL.RequestURI(), EXPECTED_LOGIN_URL)
				}

				w.WriteHeader(http.StatusOK)
				data := []byte("{\"data\":null,\"success\":false,\"error\":{\"code\":101}}\"}")
				_, _ = w.Write(data)
				loginRequestMade = true
			}))
		defer s.Close()

		res := new(Response[string])
		c.sid = "EXPIRED"
		err := doAuthenticatedRequest(c, "test request", s.URL+"/test-request", res)
		if err == nil {
			t.Error("error expected when login fails")
		}
		if !loginRequestMade {
			t.Error("Login request was not made")
		}
	})
}

func TestLogin(t *testing.T) {
	c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != EXPECTED_LOGIN_URL {
			t.Errorf("login url '%s' used but expected '%s'", r.URL.RequestURI(), EXPECTED_LOGIN_URL)
		}

		w.WriteHeader(http.StatusOK)
		data := []byte("{\"data\":{\"did\":\"DIDDID\",\"is_portal_port\":false,\"sid\":\"SIDDIS\"},\"success\":true}")
		_, _ = w.Write(data)
	})
	defer s.Close()

	err := c.Login()
	if err != nil {
		t.Error(err)
	}
	if c.sid != "SIDDIS" {
		t.Errorf("login sid is '%s' while 'SIDDIS' expected", c.sid)
	}
}

func testClient(f http.HandlerFunc) (*Client, *httptest.Server) {
	ts := httptest.NewTLSServer(f)
	host := strings.TrimPrefix(ts.URL, "https://")
	return NewClient(host, "user", "pass"), ts
}

func sequantialRequests(f ...http.HandlerFunc) http.HandlerFunc {
	index := 0
	return func(w http.ResponseWriter, r *http.Request) {
		f[index](w, r)
		index++
	}
}
