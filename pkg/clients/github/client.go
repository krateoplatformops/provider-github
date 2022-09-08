package github

import (
	"fmt"
	"net/http"
	"net/url"
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
	apiUrl       string
	apiExtraPath string
	httpClient   *http.Client
	repos        *RepoService
}

// NewClient returns a new Github Client
func NewClient(opts ClientOpts) *Client {
	res := &Client{
		apiUrl:       defaultApiURL,
		apiExtraPath: "",
		httpClient:   opts.HttpClient,
	}

	if len(opts.ApiURL) > 0 {
		u, err := url.ParseRequestURI(opts.ApiURL)
		if err != nil {
			res.apiUrl = opts.ApiURL
			res.apiExtraPath = ""
		} else {
			res.apiUrl = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			res.apiExtraPath = u.Path
		}
	}

	res.repos = newRepoService(res.httpClient, res.apiUrl, res.apiExtraPath, opts.Token)

	return res
}

func (c *Client) Repos() *RepoService {
	return c.repos
}
