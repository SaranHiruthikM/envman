package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	apiURL string
}

func NewClient(apiURL string) *Client {
	return &Client{apiURL: apiURL}
}

func (c *Client) do(method, path string, body any, token string, out any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, c.apiURL+path, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if out != nil {
		json.NewDecoder(resp.Body).Decode(out)
	}

	return nil
}

func (c *Client) Login(email, password string) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	err := c.do("POST", "/auth/login", map[string]string{"email": email, "password": password}, "", &out)
	return out.Token, err
}

func (c *Client) Register(email, password, name string) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	err := c.do("POST", "/auth/register", map[string]string{"email": email, "password": password, "name": name}, "", &out)
	return out.Token, err
}

func (c *Client) CreateProject(token, name string) (string, string, error) {
	var out struct {
		ProjectID string `json:"project_id"`
		Name      string `json:"name"`
	}
	err := c.do("POST", "/projects", map[string]string{"name": name}, token, &out)
	return out.ProjectID, out.Name, err
}

func (c *Client) PushSecrets(token, projectID, encryptedData string) error {
	return c.do("POST", "/projects/"+projectID+"/secrets",
		map[string]string{"encryptedData": encryptedData}, token, nil)
}

func (c *Client) GetSecrets(token, projectID string) (string, error) {
	var out struct {
		Message string `json:"message"`
	}
	err := c.do("GET", "/projects/"+projectID+"/secrets", nil, token, &out)
	return out.Message, err
}

func (c *Client) InviteMember(token, projectID, email, role string) error {
	return c.do("POST", "/projects/"+projectID+"/members",
		map[string]string{"email": email, "role": role}, token, nil)
}

func (c *Client) UpdateRole(token, projectID, userID, role string) error {
	return c.do("PATCH", "/projects/"+projectID+"/members/"+userID,
		map[string]string{"role": role}, token, nil)
}

func (c *Client) RemoveMember(token, projectID, userID string) error {
	return c.do("DELETE", "/projects/"+projectID+"/members/"+userID, nil, token, nil)
}
