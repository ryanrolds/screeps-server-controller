package client

import "context"

type Config struct {
	host  string
	token string
}

type Interface interface {
	CommentOnPR(ctx context.Context, repo, comment string, issue int) error
}

type Client struct {
	cfg *Config
}

func New(cfg *Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

func (c *Client) CommentOnPR(ctx context.Context, repo, comment string, issue int) error {
	return nil
}
