package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	nethtml "golang.org/x/net/html"
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
	URL         string       `json:"url"`
	Depth       int          `json:"depth"`
	HTTPStatus  int          `json:"http_status"`
	Error       string       `json:"error,omitempty"`
	BrokenLinks []BrokenLink `json:"broken_links"`
}

type BrokenLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

func Analyze(rootURL string, options Options) (*Report, error) {
	client := options.Client
	if client == nil {
		client = &http.Client{}
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
				URL:         rootURL,
				BrokenLinks: []BrokenLink{},
			},
		},
	}

	page := &report.Pages[0]

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rootURL, nil)
	if err != nil {
		report.Pages[0].Error = err.Error()
		return report, nil
	}

	resp, err := client.Do(req)
	if err != nil {
		report.Pages[0].Error = err.Error()
		return report, nil
	}

	if resp == nil {
		page.Error = err.Error()
		return report, nil
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	page.HTTPStatus = resp.StatusCode

	if resp.StatusCode >= 400 {
		page.Error = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		return report, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		page.Error = err.Error()
		return report, nil
	}

	rawLinks, err := extractLinks(string(body))
	if err != nil {
		page.Error = err.Error()
		return report, nil
	}

	seen := map[string]any{}

	for _, rawLink := range rawLinks {
		linkURL, ok := normalizeLink(rootURL, rawLink)
		if !ok {
			continue
		}

		if _, ok := seen[linkURL]; ok {
			continue
		}

		seen[linkURL] = struct{}{}

		brokenLink := checkLink(ctx, client, linkURL)
		if brokenLink != nil {
			page.BrokenLinks = append(page.BrokenLinks, *brokenLink)
		}
	}

	return report, nil
}

func extractLinks(html string) ([]string, error) {
	z := nethtml.NewTokenizer(strings.NewReader(html))

	var links []string

	for {
		tt := z.Next()

		switch tt {
		case nethtml.ErrorToken:
			err := z.Err()
			if errors.Is(err, io.EOF) {
				return links, nil
			}

			return nil, fmt.Errorf("error parsing html: %w", err)

		case nethtml.StartTagToken, nethtml.SelfClosingTagToken:
			token := z.Token()

			for _, attr := range token.Attr {
				key := strings.ToLower(attr.Key)

				if key == "href" || key == "src" {
					links = append(links, attr.Val)
				}
			}
		}
	}
}

func normalizeLink(baseURL, rawLink string) (string, bool) {
	rawLink = strings.TrimSpace(rawLink)
	if rawLink == "" {
		return "", false
	}

	if strings.HasPrefix(rawLink, "#") {
		return "", false
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", false
	}

	ref, err := url.Parse(rawLink)
	if err != nil {
		return "", false
	}

	resolved := base.ResolveReference(ref)

	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false
	}

	resolved.Fragment = ""

	return resolved.String(), true
}

func checkLink(ctx context.Context, client *http.Client, linkURL string) *BrokenLink {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, linkURL, nil)
	if err != nil {
		return &BrokenLink{
			URL:   linkURL,
			Error: err.Error(),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return &BrokenLink{
			URL:   linkURL,
			Error: err.Error(),
		}
	}

	if resp == nil {
		return &BrokenLink{
			URL:   linkURL,
			Error: "empty response",
		}
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode >= 400 {
		return &BrokenLink{
			URL:        linkURL,
			StatusCode: resp.StatusCode,
		}
	}

	return nil
}
