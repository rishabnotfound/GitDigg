package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type GitLabProvider struct {
	client *gitlab.Client
}

func NewGitLabProvider(token string) *GitLabProvider {
	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
	}

	client, err := gitlab.NewClient(token)
	if err != nil {
		return &GitLabProvider{}
	}

	return &GitLabProvider{client: client}
}

func (g *GitLabProvider) Name() string {
	return "gitlab"
}

func (g *GitLabProvider) GetTree(ctx context.Context, owner, repo string, opts TreeOptions) ([]TreeEntry, error) {
	if g.client == nil {
		return nil, NewProviderError("gitlab", 0, "client not initialized", nil)
	}

	projectPath := fmt.Sprintf("%s/%s", owner, repo)

	ref := opts.Ref
	if ref == "" {
		var err error
		ref, err = g.GetDefaultBranch(ctx, owner, repo)
		if err != nil {
			return nil, err
		}
	}

	recursive := true
	listOpts := &gitlab.ListTreeOptions{
		Ref:         &ref,
		Recursive:   &recursive,
		Path:        &opts.Path,
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	var allEntries []TreeEntry
	for {
		trees, resp, err := g.client.Repositories.ListTree(projectPath, listOpts)
		if err != nil {
			return nil, NewProviderError("gitlab", 0, "couldn't get tree", err)
		}

		for _, tree := range trees {
			entryType := EntryTypeFile
			if tree.Type == "tree" {
				entryType = EntryTypeDir
			}
			allEntries = append(allEntries, TreeEntry{
				Path: tree.Path,
				Type: entryType,
				SHA:  tree.ID,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	return allEntries, nil
}

func (g *GitLabProvider) GetFileContent(ctx context.Context, owner, repo, path string, opts FileOptions) (*FileContent, error) {
	if g.client == nil {
		return nil, NewProviderError("gitlab", 0, "client not initialized", nil)
	}

	projectPath := fmt.Sprintf("%s/%s", owner, repo)

	ref := opts.Ref
	if ref == "" {
		var err error
		ref, err = g.GetDefaultBranch(ctx, owner, repo)
		if err != nil {
			return nil, err
		}
	}

	content, resp, err := g.client.RepositoryFiles.GetRawFile(projectPath, path, &gitlab.GetRawFileOptions{Ref: &ref})
	if err != nil {
		return nil, NewProviderError("gitlab", 0, fmt.Sprintf("couldn't get %s", path), err)
	}

	return &FileContent{
		Path:    path,
		Content: io.NopCloser(strings.NewReader(string(content))),
		Size:    resp.ContentLength,
	}, nil
}

func (g *GitLabProvider) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	if g.client == nil {
		return "", NewProviderError("gitlab", 0, "client not initialized", nil)
	}

	project, _, err := g.client.Projects.GetProject(fmt.Sprintf("%s/%s", owner, repo), nil)
	if err != nil {
		return "", NewProviderError("gitlab", 0, "couldn't get project info", err)
	}

	return project.DefaultBranch, nil
}
