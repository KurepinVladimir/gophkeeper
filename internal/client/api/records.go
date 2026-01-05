package api

import (
	"encoding/json"
	"strconv"
)

type Record struct {
	ID            int64  `json:"id"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	Meta          string `json:"meta"`
	EncryptedData string `json:"encrypted_data"`
	Version       int64  `json:"version"`
	UpdatedAt     string `json:"updated_at"`
	Deleted       bool   `json:"deleted"`
}

func (c *Client) List(token string) ([]Record, error) {
	resp, err := c.r.R().
		SetAuthToken(token).
		Get("/api/v1/records/list")
	if err != nil {
		return nil, err
	}
	var out []Record
	return out, json.Unmarshal(resp.Body(), &out)
}

func (c *Client) Upsert(token string, rec Record) (int64, error) {
	resp, err := c.r.R().
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json").
		SetBody(rec).
		Post("/api/v1/records/upsert")
	if err != nil {
		return 0, err
	}
	var out struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

func (c *Client) Get(token string, id int64) (Record, error) {
	resp, err := c.r.R().SetAuthToken(token).Get("/api/v1/records/" + strconv.FormatInt(id, 10))
	if err != nil {
		return Record{}, err
	}
	var out Record
	return out, json.Unmarshal(resp.Body(), &out)
}

func (c *Client) Delete(token string, id int64) error {
	_, err := c.r.R().SetAuthToken(token).Delete("/api/v1/records/" + strconv.FormatInt(id, 10))
	return err
}
