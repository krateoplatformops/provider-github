package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/krateoplatformops/provider-github/apis/repo/v1alpha1"
)

const (
	defaultApiURL = "https://api.github.com/"
)

// GithubError represents a Github API error response
// https://developer.github.com/v3/#client-errors
type GithubError struct {
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors,omitempty"`
	DocumentationURL string `json:"documentation_url"`
}

func (e GithubError) Error() string {
	return fmt.Sprintf("github: %v %+v %v", e.Message, e.Errors, e.DocumentationURL)
}

type ClientOpts struct {
	ApiURL     string
	Token      string
	HttpClient *http.Client
}

// Client is a tiny Github client
type Client struct {
	apiUrl     string
	httpClient *http.Client
	repos      *RepoService
}

// NewClient returns a new Github Client
func NewClient(opts ClientOpts) *Client {
	res := &Client{
		apiUrl:     defaultApiURL,
		httpClient: opts.HttpClient,
	}

	if len(opts.ApiURL) > 0 {
		res.apiUrl = opts.ApiURL
	}

	res.repos = newRepoService(res.httpClient, res.apiUrl, opts.Token)

	return res
}

func (c *Client) Repos() *RepoService {
	return c.repos
}

// RepoService provides methods for creating and reading repositories.
type RepoService struct {
	client *http.Client
	apiUrl string
	token  string
}

// newRepoService returns a new RepoService.
func newRepoService(httpClient *http.Client, apiUrl, token string) *RepoService {
	return &RepoService{
		client: httpClient,
		apiUrl: apiUrl,
		token:  token,
	}
}

func (s *RepoService) Create(opts *v1alpha1.RepoParams) error {
	ok, err := s.isOrg(opts.Org)
	if err != nil {
		return err
	}

	pt := "/user/repos"
	if ok {
		pt = fmt.Sprintf("orgs/%s/repos", opts.Org)
	}

	githubError := &GithubError{}

	err = requests.URL(s.apiUrl).Path(pt).
		Client(s.client).
		Method(http.MethodPost).
		Header("Authorization", fmt.Sprintf("token %s", s.token)).
		BodyJSON(map[string]interface{}{
			"name":      opts.Name,
			"private":   opts.Private,
			"auto_init": true,
		}).
		AddValidator(ErrorJSON(githubError, 201)).
		Fetch(context.Background())
	if err != nil {
		var gerr *GithubError
		if errors.As(err, &gerr) {
			return fmt.Errorf(gerr.Error())
		}
		return err
	}

	return nil
}

// Get fetches a repository.
//
// GitHub API docs: https://docs.github.com/en/free-pro-team@latest/rest/reference/repos/#get-a-repository
func (s *RepoService) Exists(opts *v1alpha1.RepoParams) (bool, error) {
	pt := fmt.Sprintf("repos/%s/%s", opts.Org, opts.Name)

	err := requests.URL(s.apiUrl).Path(pt).
		Client(s.client).
		Method(http.MethodGet).
		Header("Authorization", fmt.Sprintf("token %s", s.token)).
		CheckStatus(200).
		Fetch(context.Background())
	if err != nil {
		if requests.HasStatusErr(err, 404) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *RepoService) isOrg(owner string) (bool, error) {
	err := requests.URL(s.apiUrl).Pathf("/orgs/%s", owner).
		Client(s.client).
		Method(http.MethodGet).
		Header("Authorization", fmt.Sprintf("token %s", s.token)).
		CheckStatus(200).
		Fetch(context.Background())
	if err != nil {
		if requests.HasStatusErr(err, 404) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
