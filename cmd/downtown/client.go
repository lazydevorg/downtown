package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

const (
	LoginUrl = "https://%s/webapi/entry.cgi?api=SYNO.API.Auth&version=6&method=login&account=%s&passwd=%s&session=DownloadStation&format=sid"
	TasksUrl = "https://%s/webapi/DownloadStation/task.cgi?api=SYNO.DownloadStation.Task&version=1&method=list&additional=transfer"
)

type Client struct {
	client http.Client
	host   string
	logger *slog.Logger
}

func NewClient(host string, logger *slog.Logger) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Client{
		client: http.Client{Transport: tr},
		host:   host,
		logger: logger,
	}
}

type Response[T any] struct {
	Success bool `json:"success"`
	Error   struct {
		Code int `json:"code"`
	} `json:"error"`
	Data T `json:"data"`
}

func doRequest[T any](c *Client, name string, request *http.Request, response *Response[T]) error {
	c.logger.Debug("executing DS request", "name", name, "method", request.Method, "url", request.URL)
	res, err := c.client.Do(request)
	if err != nil {
		c.logger.Debug("Error executing request", "name", name, "error", err)
		return fmt.Errorf("doing %s request: %w", name, err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		c.logger.Debug("Error parsing response", "name", name, "error", err)
		return fmt.Errorf("parsing %s response: %w", name, err)
	}
	if !response.Success {
		c.logger.Debug("Request error", "name", name, "code", response.Error.Code)
		return fmt.Errorf("%s request error: code %d", name, response.Error.Code)
	}
	return nil
}

type LoginRequest struct {
	user string
	pass string
}

type LoginResponseData struct {
	SID string `json:"sid"`
	DID string `json:"did"`
}

func (c *Client) createRequest(ctx context.Context, requestUrl string, urlParams ...string) (*http.Request, error) {
	params := make([]any, len(urlParams)+1)
	params[0] = c.host
	for i, p := range urlParams {
		params[i+1] = url.QueryEscape(p)
	}
	if len(params) > 0 {
		requestUrl = fmt.Sprintf(requestUrl, params...)
	}
	return http.NewRequestWithContext(ctx, "GET", requestUrl, nil)
}

func (c *Client) createAuthenticatedRequest(ctx context.Context, requestUrl string, sid string, urlParams ...string) (*http.Request, error) {
	requestUrl = requestUrl + "&_sid=" + sid
	return c.createRequest(ctx, requestUrl, urlParams...)
}

func (c *Client) Login(ctx context.Context, data LoginRequest) (*Response[LoginResponseData], error) {
	request, err := c.createRequest(ctx, LoginUrl, data.user, data.pass)
	if err != nil {
		return nil, fmt.Errorf("creating login request: %w", err)
	}
	response := new(Response[LoginResponseData])
	err = doRequest(c, "login", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type TasksData struct {
	Offset int `json:"offset"`
	Tasks  []struct {
		Additional struct {
			Transfer struct {
				DownloadedPieces int64 `json:"downloaded_pieces"`
				SizeDownloaded   int64 `json:"size_downloaded"`
				SizeUploaded     int64 `json:"size_uploaded"`
				SpeedDownload    int64 `json:"speed_download"`
				SpeedUpload      int64 `json:"speed_upload"`
			} `json:"transfer"`
		} `json:"additional"`
		Id       string `json:"id"`
		Size     int64  `json:"size"`
		Status   string `json:"status"`
		Title    string `json:"title"`
		Type     string `json:"type"`
		Username string `json:"username"`
	} `json:"tasks"`
	Total int `json:"total"`
}

func (c *Client) GetTasks(ctx context.Context, sid string, response *Response[TasksData]) error {
	request, err := c.createAuthenticatedRequest(ctx, TasksUrl, sid)
	if err != nil {
		return fmt.Errorf("creating tasks request: %w", err)
	}
	err = doRequest(c, "tasks", request, response)
	if err != nil {
		return err
	}
	return nil
}
