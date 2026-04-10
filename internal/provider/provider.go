package provider

import (
	"context"
	"io"
)

type EntryType string

const (
	EntryTypeFile EntryType = "file"
	EntryTypeDir  EntryType = "dir"
)

type TreeEntry struct {
	Path string
	Type EntryType
	Size int64
	SHA  string
}

type TreeOptions struct {
	Ref       string
	Path      string
	Recursive bool
}

type FileOptions struct {
	Ref string
}

type FileContent struct {
	Path    string
	Content io.ReadCloser
	Size    int64
}

type Provider interface {
	Name() string
	GetTree(ctx context.Context, owner, repo string, opts TreeOptions) ([]TreeEntry, error)
	GetFileContent(ctx context.Context, owner, repo, path string, opts FileOptions) (*FileContent, error)
	GetDefaultBranch(ctx context.Context, owner, repo string) (string, error)
}

type RepoInfo struct {
	Provider string
	Owner    string
	Repo     string
	Ref      string
	Path     string
}

type ProviderError struct {
	Provider   string
	StatusCode int
	Message    string
	Err        error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

func NewProviderError(provider string, statusCode int, message string, err error) *ProviderError {
	return &ProviderError{
		Provider:   provider,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}
