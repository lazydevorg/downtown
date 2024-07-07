package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	LOGIN_URL = "https://%s/webapi/entry.cgi?api=SYNO.API.Auth&version=6&method=login&account=%s&passwd=%s&session=DownloadStation&format=sid"
	TASKS_URL = "https://%s/webapi/DownloadStation/task.cgi?api=SYNO.DownloadStation.Task&version=1&method=list&additional=transfer"
)

type Client struct {
	client http.Client
	host   string
	user   string
	pass   string
	sid    string
}

func NewClient(host string, user string, pass string) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Client{
		client: http.Client{Transport: tr},
		host:   host,
		user:   user,
		pass:   pass,
	}
}

type Response[T any] struct {
	Success bool `json:"success"`
	Error   struct {
		Code int `json:"code"`
	} `json:"error"`
	Data T `json:"data"`
}

func doRequest[T any](c *Client, name string, u string, response *Response[T]) error {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return fmt.Errorf("creating %s request: %w", name, err)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("doing %s request: %w", name, err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("parsing %s response: %w", name, err)
	}
	if !response.Success {
		return fmt.Errorf("%s request error: code %d", name, response.Error.Code)
	}
	return nil
}

func doAuthenticatedRequest[T any](c *Client, name string, u string, response *Response[T]) error {
	if c.sid == "" {
		err := c.Login()
		if err != nil {
			return fmt.Errorf("login failed executing %s: %w", name, err)
		}
	}

	u = fmt.Sprintf("%s&_sid=%s", u, c.sid)
	err := doRequest(c, name, u, response)
	if err != nil {
		return err
	}

	if response.Error.Code == 105 {
		err := c.Login()
		if err != nil {
			return fmt.Errorf("login failed executing %s: %w", name, err)
		}
	}
	return nil
}

type LoginData struct {
	SID string `json:"sid"`
	DID string `json:"did"`
}

func (c *Client) Login() error {
	u := fmt.Sprintf(LOGIN_URL, c.host, c.user, c.pass)
	var result Response[LoginData]
	err := doRequest(c, "login", u, &result)
	if err != nil {
		return err
	}
	c.sid = result.Data.SID
	return nil
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

func (c *Client) GetTasks(response *Response[TasksData]) error {
	u := fmt.Sprintf(TASKS_URL, c.host)
	err := doAuthenticatedRequest(c, "get tasks", u, response)
	if err != nil {
		return err
	}
	return nil
}
