package api

import (
	"crypto/tls"
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	r *resty.Client
}

func New(baseURL string) *Client {
	c := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(10 * time.Second).
		// InsecureSkipVerify используется только для self-signed сертификатов в DEV
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true, // DEV ONLY
		})

	return &Client{r: c}
}

func (c *Client) Register(login, password string) error {
	_, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"login": login, "password": password}).
		Post("/api/v1/auth/register")
	return err
}

func (c *Client) Login(login, password string) (string, error) {
	resp, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"login": login, "password": password}).
		Post("/api/v1/auth/login")
	if err != nil {
		return "", err
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(resp.Body(), &out); err != nil {
		return "", err
	}
	return out.Token, nil
}
