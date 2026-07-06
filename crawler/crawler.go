package crawler

import (
	"context"
	"net/http"
	"time"
)

type Options struct {
	Client *http.Client
	Ctx    context.Context
}

type Report struct {
	RootURL     string       `json:"root_url"`
	GeneratedAt string       `json:"generated_at"`
	Pages       []PageReport `json:"pages"`
}

type PageReport struct {
	URL        string `json:"url"`
	HTTPStatus int    `json:"http_status"`
	Error      string `json:"error"`
}

func Analyze(rootURL string, options Options) (*Report, error) {
	client := options.Client
	if client == nil {
		client = http.DefaultClient
	}

	ctx := options.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	report := &Report{
		RootURL:     rootURL,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Pages: []PageReport{
			{
				URL: rootURL,
			},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rootURL, nil)
	if err != nil {
		report.Pages[0].Error = err.Error()
		return report, nil
	}

	resp, err := client.Do(req)

	defer resp.Body.Close()

	report.Pages[0].HTTPStatus = resp.StatusCode

	return report, nil
}
