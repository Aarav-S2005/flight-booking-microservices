package http_client

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type HttpClient interface {
	GetEmailFromAuthService(userID string) (string, error)
}

type client struct {
	client *resty.Client
}

func NewClient(authServiceURL string) HttpClient {
	return &client{client: resty.New().SetBaseURL(authServiceURL)}
}

type Response struct {
	Email string `json:"email"`
}

func (c *client) GetEmailFromAuthService(userID string) (string, error) {
	var resBody Response
	resp, err := c.client.R().SetResult(&resBody).Get("/users/email/" + userID)
	if err != nil {
		return "", fmt.Errorf("get email from auth service failed: %s", err.Error())
	}
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("get email from auth service failed: %d", resp.StatusCode())
	}
	return resBody.Email, nil
}
