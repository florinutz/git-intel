package resolve

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/go-github/v62/github"
)

type RepoPair struct {
	Owner string
	Repo  string
}

type Resolver struct {
	Client *github.Client
}

func NewResolver(token string) *Resolver {
	client := github.NewClient(nil)
	if token != "" {
		client = client.WithAuthToken(token)
	}
	return &Resolver{Client: client}
}

func (r *Resolver) Resolve(rawURLs []string) ([][]RepoPair, error) {
	invalidURLsPositions := r.validate(rawURLs)
	if len(invalidURLsPositions) > 0 {
		invalidURLs := make([]string, len(invalidURLsPositions))
		for i, pos := range invalidURLsPositions {
			invalidURLs[i] = rawURLs[pos]
		}
		return nil, fmt.Errorf("invalid URLs: %v", strings.Join(invalidURLs, ", "))
	}

	var allRepos [][]RepoPair
	for _, rawURL := range rawURLs {
		repos, err := r.resolveOne(rawURL)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos)
	}
	return allRepos, nil
}

func (r *Resolver) resolveOne(rawURL string) ([]RepoPair, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s", rawURL)
	}
	parts := strings.Split(u.Path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL: %s", rawURL)
	}
	owner := parts[1]
	if len(parts) == 2 {
		// This is an organization URL, we need to list all the repos
		opt := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 10},
		}
		var allRepos []RepoPair
		for {
			repos, resp, err := r.Client.Repositories.ListByOrg(context.Background(), owner, opt)
			if err != nil {
				return nil, err
			}
			for _, repo := range repos {
				allRepos = append(allRepos, RepoPair{Owner: owner, Repo: *repo.Name})
			}
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
		return allRepos, nil
	} else {
		// This is a single repo URL
		return []RepoPair{{Owner: owner, Repo: parts[2]}}, nil
	}
}

// validate returns positions of invalid urls in the input slice
func (r *Resolver) validate(urls []string) (invalidURLs []int) {
	for i, rawURL := range urls {
		if u, err := url.ParseRequestURI(rawURL); err != nil || u.Host != "github.com" {
			matched, err := regexp.MatchString(`git@github\.com:([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)\.git`, rawURL)
			if err != nil || !matched {
				invalidURLs = append(invalidURLs, i)
			}
		}
	}
	return invalidURLs
}
