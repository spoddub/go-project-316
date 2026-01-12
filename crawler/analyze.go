package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Options struct {
	URL         string
	Depth       int
	Retries     int
	Delay       time.Duration
	Timeout     time.Duration
	RPS         int
	UserAgent   string
	Concurrency int
	IndentJSON  bool
	HTTPClient  *http.Client
}

type Page struct {
	URL        string `json:"url"`
	Depth      int    `json:"depth"`
	HTTPStatus int    `json:"http_status"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

type Report struct {
	RootURL     string `json:"root_url"`
	Depth       int    `json:"depth"`
	GeneratedAt string `json:"generated_at"`
	Pages       []Page `json:"pages"`
}

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	if opts.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{}
	}

	reqCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	p := Page{
		URL:   opts.URL,
		Depth: 0,
	}

	attempts := opts.Retries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, opts.URL, nil)
		if err != nil {
			return nil, err
		}
		if opts.UserAgent != "" {
			req.Header.Set("User-Agent", opts.UserAgent)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		p.Status = "ok"
		p.HTTPStatus = resp.StatusCode
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		lastErr = nil
		break
	}

	if lastErr != nil {
		p.Status = "error"
		p.Error = lastErr.Error()
	}

	r := Report{
		RootURL:     opts.URL,
		Depth:       opts.Depth,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Pages:       []Page{p},
	}

	if opts.IndentJSON {
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return nil, err
		}
		return b, nil
	}

	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}
