package crawler

import (
	"net/http"
	"time"
)

type Options struct {
	Client *http.Client
}

type Report struct {
	RootURL     string `json:"root_url"`
	GeneratedAt string `json:"generated_at"`
	Pages       []PageReport
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

	report := &Report{
		RootURL:     rootURL,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Pages: []PageReport{
			{
				URL: rootURL,
			},
		},
	}

	resp, err := client.Get(rootURL)
	if err != nil {
		report.Pages[0].Error = err.Error()
		return report, nil
	}
	defer resp.Body.Close()

	report.Pages[0].HTTPStatus = resp.StatusCode

	return report, nil
}
