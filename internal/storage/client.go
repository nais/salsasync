package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"

	//log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"salsasync/internal/console"
)

type Client struct {
	url    string
	apiKey string
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
				log.WithError(err).Warning("failed to delete dtrackUser ", dtu.Username)
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
				log.WithError(err).Warning("failed to delete dtrackTeam ", dtt.Name)
			}
		}
	}
	return nil
}

func (c *Client) CreateTeam(ctx context.Context, teamName string) error {
	body, _ := json.Marshal(map[string]string{
		"name": teamName,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url+"team", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d, with body:\n%s\n", resp.StatusCode, string(b))
	}
	return nil
}
func (c *Client) CreateUser(ctx context.Context, email string) error {
	u := &User{
		Username: email,
		Email:    email,
	}
	body, err := json.Marshal(u)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url+"user/oidc", bytes.NewReader(body))
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d, with body:\n%s\n", resp.StatusCode, string(b))
	}
	return nil
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Users []struct {
	User User `json:"user"`
}

func (c *Client) GetUsers(ctx context.Context) ([]User, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+"user/oidc", nil)
	req.Header.Set("X-API-Key", c.apiKey)
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

	var dtrackusers []User
	if err = json.Unmarshal(resBody, &dtrackusers); err != nil {
		panic(err)
	}
	return dtrackusers, nil
}

func (c *Client) DeleteUser(ctx context.Context, username string) error {

	u := &User{
		Username: username,
	}
	body, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("marshalling body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url+"user/oidc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d, with body:\n%s\n", resp.StatusCode, string(b))
	}
	return nil
}
func (c *Client) MapTeamWithProject(ctx context.Context, team string, project string) error {

	body, _ := json.Marshal(map[string]string{
		"team":    team,
		"project": project,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url+"acl/mapping", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d, with body:\n%s\n", resp.StatusCode, string(b))
	}
	return nil
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

func (c *Client) GetTeams(ctx context.Context) ([]Team, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+"team", nil)
	req.Header.Set("X-API-Key", c.apiKey)
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

	var dtrackTeams []Team
	if err = json.Unmarshal(resBody, &dtrackTeams); err != nil {
		panic(err)
	}
	return dtrackTeams, nil
}

func (c *Client) GetProject(ctx context.Context, name string, version string) (*Project, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+"project/lookup?name="+name+"&version="+version, nil)
	req.Header.Set("X-API-Key", c.apiKey)
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

	var dtrackProject Project
	if err = json.Unmarshal(resBody, &dtrackProject); err != nil {
		panic(err)
	}
	return &dtrackProject, nil
}

type Tag struct {
	Name string `json:"name"`
}

type Tags struct {
	Tags []Tag `json:"tags"`
}

func (c *Client) UpdateProjectTags(ctx context.Context, projectUuid string, tags []string) error {

	tagArray := make([]Tag, 0)
	for _, dtt := range tags {
		t := Tag{Name: dtt}
		tagArray = append(tagArray, t)
	}

	body, _ := json.Marshal(Tags{tagArray})

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url+"project/"+projectUuid, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d, with body:\n%s\n", resp.StatusCode, string(b))
	}
	return nil
}

func (c *Client) DeleteTeam(ctx context.Context, uuid string) error {
	body, _ := json.Marshal(map[string]string{
		"uuid": uuid,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url+"team", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d, with body:\n%s\n", resp.StatusCode, string(b))
	}
	return nil

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
