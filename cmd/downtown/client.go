package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	LOGIN_URL = "https://%s/webapi/entry.cgi?api=SYNO.API.Auth&version=6&method=login&account=%s&passwd=%s&session=DownloadStation&format=sid"
)

type Client struct {
	client http.Client
	host   string
	user   string
	pass   string
	sid    string
}

type Response[T any] struct {
	Success bool `json:"success"`
	Error   struct {
		Code int `json:"code"`
	} `json:"error"`
	Data T `json:"data"`
}

type LoginData struct {
	SID string `json:"sid"`
	DID string `json:"did"`
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
	return nil
}

func (c *Client) Login() error {
	u := fmt.Sprintf(LOGIN_URL, c.host, c.user, c.pass)
	var result Response[LoginData]
	err := doRequest(c, "login", u, &result)
	if err != nil {
		return err
	}
	if result.Success {
		c.sid = result.Data.SID
	}
	return nil
}
