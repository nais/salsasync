package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"reflect"
	"salsasync/internal/console"
	"strings"
)

type Client struct {
	url    string
	apiKey string
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Users []struct {
	User User `json:"user"`
}

type Team struct {
	Uuid      string `json:"uuid"`
	Name      string `json:"name"`
	OidcUsers []struct {
		Username string `json:"username"`
	} `json:"oidcUsers,omitempty"`
}

type Project struct {
	Name    string `json:"name"`
	Uuid    string `json:"uuid"`
	Version string `json:"version"`
}

func New(url string, apiKey string) *Client {
	return &Client{
		url:    url,
		apiKey: apiKey,
	}
}

func (c *Client) SynchronizeTeamsAndUsers(ctx context.Context, teams *console.Teams, consoleUsers *console.Users) error {
	internalTeams := []string{"Portfolio Managers", "Administrator", "Automation"}

	dtrackUsers, err := c.GetUsers(ctx)
	dtrackTeams, err := c.GetTeams(ctx)
	if err != nil {
		return fmt.Errorf("getting users or teams from dtrack: %w", err)
	}
	for _, u := range *consoleUsers {
		if !containsStructFieldValue(dtrackUsers, "Username", u.Email) {
			log.Infof("User not in dtrack  %+v, creating...", u.Email)
			err := c.CreateUser(ctx, u.Email)
			if err != nil {
				log.WithError(err).Warning("failed to create user", u.Email)
			}
		}
	}
	for _, dtu := range dtrackUsers {
		if !containsStructFieldValue(*consoleUsers, "Email", dtu.Username) {
			log.Infof("User not in console  %+v, deleting...", dtu.Username)
			if err := c.DeleteUser(ctx, dtu.Username); err != nil {
				log.WithError(err).Warning("failed to delete user ", dtu.Username)
			}
		}
	}
	for _, t := range *teams {
		if !containsStructFieldValue(dtrackTeams, "Name", t.Slug) {
			log.Infof("Team not in dtrack  %+v, creating...", t.Slug)
			err := c.CreateTeam(ctx, t.Slug)
			if err != nil {
				log.WithError(err).Warning("failed to create team ", t.Slug)
			}
		}

	}
	for _, dtt := range dtrackTeams {
		if !containsStructFieldValue(*teams, "Slug", dtt.Name) && !contains(internalTeams, dtt.Name) {
			log.Infof("Team not in console:  %s, deleting...", dtt.Name)
			if err := c.DeleteTeam(ctx, dtt.Uuid); err != nil {
				log.WithError(err).Warning("failed to delete team ", dtt.Name)
			}
		}
	}
	return nil
}

func (c *Client) CreateTeam(ctx context.Context, teamName string) error {
	body, _ := json.Marshal(map[string]string{
		"name": teamName,
	})
	_, err := c.httpRequest(ctx, http.MethodPut, "team", body)
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) CreateUser(ctx context.Context, email string) error {
	u := &User{
		Username: email,
		Email:    email,
	}
	body, err := json.Marshal(u)
	_, err = c.httpRequest(ctx, http.MethodPut, "user/oidc", body)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	resBody, err := c.httpRequest(ctx, http.MethodGet, "user/oidc", nil)
	if err != nil {
		return nil, err
	}

	var dtrackusers []User
	if err = json.Unmarshal(resBody, &dtrackusers); err != nil {
		panic(err)
	}
	return dtrackusers, err
}

func (c *Client) DeleteUser(ctx context.Context, username string) error {
	u := &User{
		Username: username,
	}
	body, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("marshalling body: %w", err)
	}

	_, err = c.httpRequest(ctx, http.MethodDelete, "user/oidc", body)
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) MapTeamWithProject(ctx context.Context, team string, project string) error {
	body, _ := json.Marshal(map[string]string{
		"team":    team,
		"project": project,
	})
	_, err := c.httpRequest(ctx, http.MethodPut, "acl/mapping", body)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetTeams(ctx context.Context) ([]Team, error) {
	resBody, err := c.httpRequest(ctx, http.MethodGet, "team", nil)
	if err != nil {
		return nil, err
	}

	var dtrackTeams []Team
	if err = json.Unmarshal(resBody, &dtrackTeams); err != nil {
		panic(err)
	}
	return dtrackTeams, nil
}

func (c *Client) DeleteTeam(ctx context.Context, uuid string) error {
	body, _ := json.Marshal(map[string]string{
		"uuid": uuid,
	})

	_, err := c.httpRequest(ctx, http.MethodDelete, "team", body)
	if err != nil {
		return err
	}
	return nil

}

func (c *Client) httpRequest(ctx context.Context, httpMethod string, path string, body []byte) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, httpMethod, c.url+path, bytes.NewReader(body))
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}
		return nil, fmt.Errorf("unexpected status code: %d, with body:\n%s\n", resp.StatusCode, string(b))
	}
	resBody, err := io.ReadAll(resp.Body)
	return resBody, err
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.ToUpper(a) == strings.ToUpper(e) {
			return true
		}
	}
	return false
}

func containsStructFieldValue(slice interface{}, fieldName string, fieldValueToCheck interface{}) bool {

	rangeOnMe := reflect.ValueOf(slice)

	for i := 0; i < rangeOnMe.Len(); i++ {
		s := rangeOnMe.Index(i)
		f := s.FieldByName(fieldName)
		if f.IsValid() {
			if f.Interface() == fieldValueToCheck {
				return true
			}
		}
	}
	return false
}
