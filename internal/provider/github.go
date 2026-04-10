package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type GitHubProvider struct {
	client *github.Client
}

func NewGitHubProvider(token string) *GitHubProvider {
	var httpClient *http.Client

	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(context.Background(), ts)
	}

	return &GitHubProvider{client: github.NewClient(httpClient)}
}

func (g *GitHubProvider) Name() string {
	return "github"
}

func (g *GitHubProvider) GetTree(ctx context.Context, owner, repo string, opts TreeOptions) ([]TreeEntry, error) {
	ref := opts.Ref
	if ref == "" {
		var err error
		ref, err = g.GetDefaultBranch(ctx, owner, repo)
		if err != nil {
			return nil, err
		}
	}

	branch, _, err := g.client.Repositories.GetBranch(ctx, owner, repo, ref, 0)
	if err != nil {
		commit, _, err2 := g.client.Repositories.GetCommit(ctx, owner, repo, ref, nil)
		if err2 != nil {
			return nil, NewProviderError("github", 0, "couldn't get ref", err)
		}
		ref = commit.GetSHA()
	} else {
		ref = branch.GetCommit().GetSHA()
	}

	tree, _, err := g.client.Git.GetTree(ctx, owner, repo, ref, true)
	if err != nil {
		return nil, NewProviderError("github", 0, "couldn't get tree", err)
	}

	entries := make([]TreeEntry, 0, len(tree.Entries))
	for _, entry := range tree.Entries {
		entryType := EntryTypeFile
		if entry.GetType() == "tree" {
			entryType = EntryTypeDir
		}

		path := entry.GetPath()
		if opts.Path != "" && !strings.HasPrefix(path, opts.Path) && path != opts.Path {
			continue
		}

		entries = append(entries, TreeEntry{
			Path: path,
			Type: entryType,
			Size: int64(entry.GetSize()),
			SHA:  entry.GetSHA(),
		})
	}

	return entries, nil
}

func (g *GitHubProvider) GetFileContent(ctx context.Context, owner, repo, path string, opts FileOptions) (*FileContent, error) {
	ref := opts.Ref
	if ref == "" {
		var err error
		ref, err = g.GetDefaultBranch(ctx, owner, repo)
		if err != nil {
			return nil, err
		}
	}

	fileContent, _, _, err := g.client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: ref})
	if err != nil {
		return nil, NewProviderError("github", 0, fmt.Sprintf("couldn't get %s", path), err)
	}

	if fileContent == nil {
		return nil, NewProviderError("github", 404, fmt.Sprintf("%s is not a file", path), nil)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return g.getBlob(ctx, owner, repo, fileContent.GetSHA())
	}

	return &FileContent{
		Path:    path,
		Content: io.NopCloser(strings.NewReader(content)),
		Size:    int64(fileContent.GetSize()),
	}, nil
}

func (g *GitHubProvider) getBlob(ctx context.Context, owner, repo, sha string) (*FileContent, error) {
	blob, _, err := g.client.Git.GetBlob(ctx, owner, repo, sha)
	if err != nil {
		return nil, NewProviderError("github", 0, "couldn't get blob", err)
	}

	var content []byte
	if blob.GetEncoding() == "base64" {
		content, err = base64.StdEncoding.DecodeString(blob.GetContent())
		if err != nil {
			return nil, NewProviderError("github", 0, "couldn't decode blob", err)
		}
	} else {
		content = []byte(blob.GetContent())
	}

	return &FileContent{
		Content: io.NopCloser(bytes.NewReader(content)),
		Size:    int64(blob.GetSize()),
	}, nil
}

func (g *GitHubProvider) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	repository, _, err := g.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", NewProviderError("github", 0, "couldn't get repo info", err)
	}
	return repository.GetDefaultBranch(), nil
}
