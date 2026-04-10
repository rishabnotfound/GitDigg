package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type BitbucketProvider struct {
	client *http.Client
	token  string
}

func NewBitbucketProvider(token string) *BitbucketProvider {
	if token == "" {
		token = os.Getenv("BITBUCKET_TOKEN")
	}
	return &BitbucketProvider{client: &http.Client{}, token: token}
}

func (b *BitbucketProvider) Name() string {
	return "bitbucket"
}

type bbTreeResponse struct {
	Values []struct {
		Path   string `json:"path"`
		Type   string `json:"type"`
		Size   int64  `json:"size"`
		Commit struct {
			Hash string `json:"hash"`
		} `json:"commit"`
	} `json:"values"`
	Next string `json:"next"`
}

type bbRepoResponse struct {
	MainBranch struct {
		Name string `json:"name"`
	} `json:"mainbranch"`
}

func (b *BitbucketProvider) GetTree(ctx context.Context, owner, repo string, opts TreeOptions) ([]TreeEntry, error) {
	ref := opts.Ref
	if ref == "" {
		var err error
		ref, err = b.GetDefaultBranch(ctx, owner, repo)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/src/%s/?pagelen=100", owner, repo, ref)
	if opts.Path != "" {
		url = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/src/%s/%s?pagelen=100", owner, repo, ref, opts.Path)
	}

	var allEntries []TreeEntry
	for url != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, NewProviderError("bitbucket", 0, "request failed", err)
		}

		if b.token != "" {
			req.Header.Set("Authorization", "Bearer "+b.token)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, NewProviderError("bitbucket", 0, "couldn't fetch tree", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, NewProviderError("bitbucket", resp.StatusCode, resp.Status, nil)
		}

		var treeResp bbTreeResponse
		if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
			return nil, NewProviderError("bitbucket", 0, "couldn't parse response", err)
		}

		for _, entry := range treeResp.Values {
			entryType := EntryTypeFile
			if entry.Type == "commit_directory" {
				entryType = EntryTypeDir
			}
			allEntries = append(allEntries, TreeEntry{
				Path: entry.Path,
				Type: entryType,
				Size: entry.Size,
				SHA:  entry.Commit.Hash,
			})
		}

		url = treeResp.Next
	}

	return allEntries, nil
}

func (b *BitbucketProvider) GetFileContent(ctx context.Context, owner, repo, path string, opts FileOptions) (*FileContent, error) {
	ref := opts.Ref
	if ref == "" {
		var err error
		ref, err = b.GetDefaultBranch(ctx, owner, repo)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/src/%s/%s", owner, repo, ref, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, NewProviderError("bitbucket", 0, "request failed", err)
	}

	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, NewProviderError("bitbucket", 0, fmt.Sprintf("couldn't fetch %s", path), err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, NewProviderError("bitbucket", resp.StatusCode, fmt.Sprintf("couldn't get %s", path), nil)
	}

	return &FileContent{
		Path:    path,
		Content: resp.Body,
		Size:    resp.ContentLength,
	}, nil
}

func (b *BitbucketProvider) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", NewProviderError("bitbucket", 0, "request failed", err)
	}

	if b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return "", NewProviderError("bitbucket", 0, "couldn't fetch repo info", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", NewProviderError("bitbucket", resp.StatusCode, "couldn't get repo info", nil)
	}

	var repoResp bbRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&repoResp); err != nil {
		return "", NewProviderError("bitbucket", 0, "couldn't parse response", err)
	}

	if repoResp.MainBranch.Name == "" {
		return "main", nil
	}

	return repoResp.MainBranch.Name, nil
}
