package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/carlmjohnson/requests"
	"github.com/krateoplatformops/provider-github/apis/repo/v1alpha1"
	"github.com/krateoplatformops/provider-github/pkg/helpers"
)

// RepoService provides methods for creating and reading repositories.
type RepoService struct {
	client       *http.Client
	apiUrl       string
	apiExtraPath string
	token        string
}

// newRepoService returns a new RepoService.
func newRepoService(httpClient *http.Client, apiUrl, extraPath, token string) *RepoService {
	return &RepoService{
		client:       httpClient,
		apiUrl:       apiUrl,
		apiExtraPath: extraPath,
		token:        token,
	}
}

func (s *RepoService) Create(opts *v1alpha1.RepoParams) error {
	ok, err := s.isOrg(opts.Org)
	if err != nil {
		return err
	}

	pt := path.Join(s.apiExtraPath, "/user/repos")
	if ok {
		pt = path.Join(s.apiExtraPath, fmt.Sprintf("orgs/%s/repos", opts.Org))
	}

	githubError := &GithubError{}

	err = requests.URL(s.apiUrl).Path(pt).
		Client(s.client).
		Method(http.MethodPost).
		Header("Authorization", fmt.Sprintf("token %s", s.token)).
		BodyJSON(map[string]interface{}{
			"name":      opts.Name,
			"private":   opts.Private,
			"auto_init": helpers.BoolValueOrDefault(opts.Initialize, true),
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
	pt := path.Join(s.apiExtraPath, fmt.Sprintf("repos/%s/%s", opts.Org, opts.Name))

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

// Deleting a repository requires admin access. If OAuth is used, the delete_repo scope is required.
// https://docs.github.com/en/rest/repos/repos#get-a-repository
func (s *RepoService) Delete(opts *v1alpha1.RepoParams) error {
	pt := path.Join(s.apiExtraPath, fmt.Sprintf("repos/%s/%s", opts.Org, opts.Name))

	err := requests.URL(s.apiUrl).Path(pt).
		Client(s.client).
		Method(http.MethodDelete).
		Header("Authorization", fmt.Sprintf("token %s", s.token)).
		CheckStatus(204).
		Fetch(context.Background())
	if err != nil {
		if requests.HasStatusErr(err, 404) {
			return nil
		}

		return err
	}

	return nil
}

func (s *RepoService) isOrg(owner string) (bool, error) {
	pt := path.Join(s.apiExtraPath, fmt.Sprintf("/orgs/%s", owner))
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
