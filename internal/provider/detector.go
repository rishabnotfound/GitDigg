package provider

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	shorthandPattern = regexp.MustCompile(`^([a-zA-Z0-9][-a-zA-Z0-9]*)/([a-zA-Z0-9._-]+)$`)
	githubPattern    = regexp.MustCompile(`(?:https?://)?(?:www\.)?github\.com/([^/]+)/([^/]+)(?:/(?:tree|blob)/([^/]+)(?:/(.*))?)?`)
	gitlabPattern    = regexp.MustCompile(`(?:https?://)?(?:www\.)?gitlab\.com/([^/]+)/([^/]+)(?:/-/(?:tree|blob)/([^/]+)(?:/(.*))?)?`)
	bitbucketPattern = regexp.MustCompile(`(?:https?://)?(?:www\.)?bitbucket\.org/([^/]+)/([^/]+)(?:/src/([^/]+)(?:/(.*))?)?`)
)

func ParseRepo(input string) (*RepoInfo, error) {
	input = strings.TrimSpace(input)

	if matches := shorthandPattern.FindStringSubmatch(input); matches != nil {
		return &RepoInfo{
			Provider: "github",
			Owner:    matches[1],
			Repo:     cleanRepoName(matches[2]),
		}, nil
	}

	if info := tryGitHubURL(input); info != nil {
		return info, nil
	}
	if info := tryGitLabURL(input); info != nil {
		return info, nil
	}
	if info := tryBitbucketURL(input); info != nil {
		return info, nil
	}

	if strings.Contains(input, "://") {
		return tryGenericURL(input)
	}

	return nil, fmt.Errorf("invalid repo: %s\n\nExpected: owner/repo or https://github.com/owner/repo", input)
}

func tryGitHubURL(input string) *RepoInfo {
	matches := githubPattern.FindStringSubmatch(input)
	if matches == nil {
		return nil
	}
	return &RepoInfo{
		Provider: "github",
		Owner:    matches[1],
		Repo:     cleanRepoName(matches[2]),
		Ref:      matches[3],
		Path:     matches[4],
	}
}

func tryGitLabURL(input string) *RepoInfo {
	matches := gitlabPattern.FindStringSubmatch(input)
	if matches == nil {
		return nil
	}
	return &RepoInfo{
		Provider: "gitlab",
		Owner:    matches[1],
		Repo:     cleanRepoName(matches[2]),
		Ref:      matches[3],
		Path:     matches[4],
	}
}

func tryBitbucketURL(input string) *RepoInfo {
	matches := bitbucketPattern.FindStringSubmatch(input)
	if matches == nil {
		return nil
	}
	return &RepoInfo{
		Provider: "bitbucket",
		Owner:    matches[1],
		Repo:     cleanRepoName(matches[2]),
		Ref:      matches[3],
		Path:     matches[4],
	}
}

func tryGenericURL(input string) (*RepoInfo, error) {
	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s", input)
	}

	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))

	var provider string
	switch {
	case strings.Contains(host, "github"):
		provider = "github"
	case strings.Contains(host, "gitlab"):
		provider = "gitlab"
	case strings.Contains(host, "bitbucket"):
		provider = "bitbucket"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", host)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid URL: missing owner/repo")
	}

	return &RepoInfo{
		Provider: provider,
		Owner:    parts[0],
		Repo:     cleanRepoName(parts[1]),
	}, nil
}

func cleanRepoName(repo string) string {
	return strings.TrimSuffix(repo, ".git")
}

func GetProvider(info *RepoInfo, token string) (Provider, error) {
	switch info.Provider {
	case "github":
		return NewGitHubProvider(token), nil
	case "gitlab":
		return NewGitLabProvider(token), nil
	case "bitbucket":
		return NewBitbucketProvider(token), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", info.Provider)
	}
}

func DetectAndGetProvider(repoStr, token string) (Provider, *RepoInfo, error) {
	info, err := ParseRepo(repoStr)
	if err != nil {
		return nil, nil, err
	}

	provider, err := GetProvider(info, token)
	if err != nil {
		return nil, nil, err
	}

	return provider, info, nil
}
