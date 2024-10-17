package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	ExpectedLoginUrl      = "/webapi/entry.cgi?api=SYNO.API.Auth&version=6&method=login&account=user&passwd=pass&session=DownloadStation&format=sid"
	ExpectedTasksUrl      = "/webapi/DownloadStation/task.cgi?api=SYNO.DownloadStation.Task&version=1&method=list&additional=transfer&_sid=SID"
	ExpectedDeleteTaskUrl = "/webapi/DownloadStation/task.cgi?api=SYNO.DownloadStation.Task&version=1&method=delete&id=ID1&_sid=SID"
	ExpectedPauseTaskUrl  = "/webapi/DownloadStation/task.cgi?api=SYNO.DownloadStation.Task&version=1&method=pause&id=ID1&_sid=SID"
	ExpectedResumeTaskUrl = "/webapi/DownloadStation/task.cgi?api=SYNO.DownloadStation.Task&version=1&method=resume&id=ID1&_sid=SID"
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
	req, err := c.createRequest(context.Background(), "https://%s/test-request")
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
	req, err := c.createRequest(context.Background(), "https://%s/test-request")
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

func TestCreateAuthenticatedRequest(t *testing.T) {
	c := NewClient("localhost", slog.Default())
	req, err := c.createAuthenticatedRequest(context.Background(), "https://%s/test-request?", "SID")
	if err != nil {
		t.Fatalf("failed to create authenticated request: %v", err)
	}
	if req.URL.Query().Get("_sid") != "SID" {
		t.Errorf("authenticated request url '%s' used but expected '%s'", req.URL.Query().Get("_sid"), "SID")
	}
}

func TestTasks(t *testing.T) {
	c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != ExpectedTasksUrl {
			t.Errorf("tasks url '%s' used but expected '%s'", r.URL.RequestURI(), ExpectedTasksUrl)
		}

		w.WriteHeader(http.StatusOK)
		data := []byte(`{"data":{"offset":0,"tasks":[],"total":3},"success":true}`)
		_, _ = w.Write(data)
	})
	defer s.Close()

	var response Response[TasksData]
	err := c.GetTasks(context.Background(), "SID", &response)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteTask(t *testing.T) {
	c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != ExpectedDeleteTaskUrl {
			t.Errorf("delete task url '%s' used but expected '%s'", r.URL.RequestURI(), ExpectedDeleteTaskUrl)
		}

		w.WriteHeader(http.StatusOK)
		data := []byte(`{"data":[{"id":"ID1","error": 0}],"success":true}`)
		_, _ = w.Write(data)
	})
	defer s.Close()

	err := c.DeleteTask(context.Background(), "SID", "ID1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPauseTask(t *testing.T) {
	c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != ExpectedPauseTaskUrl {
			t.Errorf("delete task url '%s' used but expected '%s'", r.URL.RequestURI(), ExpectedPauseTaskUrl)
		}

		w.WriteHeader(http.StatusOK)
		data := []byte(`{"data":[{"id":"ID1","error": 0}],"success":true}`)
		_, _ = w.Write(data)
	})
	defer s.Close()

	err := c.PauseTask(context.Background(), "SID", "ID1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestResumeTask(t *testing.T) {
	c, s := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != ExpectedResumeTaskUrl {
			t.Errorf("delete task url '%s' used but expected '%s'", r.URL.RequestURI(), ExpectedResumeTaskUrl)
		}

		w.WriteHeader(http.StatusOK)
		data := []byte(`{"data":[{"id":"ID1","error": 0}],"success":true}`)
		_, _ = w.Write(data)
	})
	defer s.Close()

	err := c.ResumeTask(context.Background(), "SID", "ID1")
	if err != nil {
		t.Fatal(err)
	}
}

func testClient(f http.HandlerFunc) (*Client, *httptest.Server) {
	ts := httptest.NewTLSServer(f)
	host := strings.TrimPrefix(ts.URL, "https://")
	return NewClient(host, slog.Default()), ts
}

//func sequantialRequests(f ...http.HandlerFunc) http.HandlerFunc {
//	index := 0
//	return func(w http.ResponseWriter, r *http.Request) {
//		f[index](w, r)
//		index++
//	}
//}
