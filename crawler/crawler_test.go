package crawler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"code/crawler"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func newResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestAnalyzeSuccess(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://example.com" {
			t.Fatalf("expected URL https://example.com, got %s", req.URL.String())
		}

		return newResponse(http.StatusOK, "OK"), nil
	})

	report, err := crawler.Analyze("https://example.com", crawler.Options{
		Client: client,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if report.RootURL != "https://example.com" {
		t.Fatalf("expected root url https://example.com, got %s", report.RootURL)
	}

	if len(report.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(report.Pages))
	}

	page := report.Pages[0]

	if page.URL != "https://example.com" {
		t.Fatalf("expected page url https://example.com, got %s", page.URL)
	}

	if page.HTTPStatus != http.StatusOK {
		t.Fatalf("expected status 200, got %d", page.HTTPStatus)
	}

	if page.Error != "" {
		t.Fatalf("expected empty error, got %s", page.Error)
	}
}

func TestAnalyzeNotFoundStatus(t *testing.T) {
	client := newTestClient(func(_ *http.Request) (*http.Response, error) {
		return newResponse(http.StatusNotFound, "Not Found"), nil
	})

	report, err := crawler.Analyze("https://example.com/missing", crawler.Options{
		Client: client,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	page := report.Pages[0]

	if page.HTTPStatus != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", page.HTTPStatus)
	}

	if page.Error == "" {
		t.Fatal("expected error message for 404 status")
	}
}

func TestAnalyzeServerErrorStatus(t *testing.T) {
	client := newTestClient(func(_ *http.Request) (*http.Response, error) {
		return newResponse(http.StatusInternalServerError, "Internal Server Error"), nil
	})

	report, err := crawler.Analyze("https://example.com/error", crawler.Options{
		Client: client,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	page := report.Pages[0]

	if page.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", page.HTTPStatus)
	}

	if page.Error == "" {
		t.Fatal("expected error message for 500 status")
	}
}

func TestAnalyzeNetworkError(t *testing.T) {
	client := newTestClient(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("network is unreachable")
	})

	report, err := crawler.Analyze("https://example.com", crawler.Options{
		Client: client,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	page := report.Pages[0]

	if page.HTTPStatus != 0 {
		t.Fatalf("expected status 0, got %d", page.HTTPStatus)
	}

	if page.Error == "" {
		t.Fatal("expected network error message")
	}
}

func TestAnalyzeTimeout(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	report, err := crawler.Analyze("https://example.com", crawler.Options{
		Client: client,
		Ctx:    ctx,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	page := report.Pages[0]

	if page.HTTPStatus != 0 {
		t.Fatalf("expected status 0, got %d", page.HTTPStatus)
	}

	if page.Error == "" {
		t.Fatal("expected timeout/context error message")
	}
}

func TestAnalyzeBrokenLinks(t *testing.T) {
	html := `
		<html>
			<head>
				<link href="/ok.css">
				<link href="/missing.css">
			</head>
			<body>
				<a href="/about.html">About</a>
			</body>
		</html>
	`

	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "http://simple.test":
			return newResponse(http.StatusOK, html), nil
		case "http://simple.test/ok.css":
			return newResponse(http.StatusOK, "OK"), nil
		case "http://simple.test/missing.css":
			return newResponse(http.StatusNotFound, "Not Found"), nil
		case "http://simple.test/about.html":
			return newResponse(http.StatusOK, "OK"), nil
		default:
			return nil, errors.New("unexpected request: " + req.URL.String())
		}
	})

	report, err := crawler.Analyze("http://simple.test", crawler.Options{
		Client: client,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(report.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(report.Pages))
	}

	page := report.Pages[0]

	if len(page.BrokenLinks) != 1 {
		t.Fatalf("expected 1 broken link, got %d", len(page.BrokenLinks))
	}

	brokenLink := page.BrokenLinks[0]

	if brokenLink.URL != "http://simple.test/missing.css" {
		t.Fatalf("expected broken link URL http://simple.test/missing.css, got %s", brokenLink.URL)
	}

	if brokenLink.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code 404, got %d", brokenLink.StatusCode)
	}

	if brokenLink.Error != "" {
		t.Fatalf("expected empty error, got %s", brokenLink.Error)
	}
}

func TestAnalyzeIgnoresUnsupportedLinks(t *testing.T) {
	html := `
		<html>
			<body>
				<a href="">Empty</a>
				<a href="#top">Anchor</a>
				<a href="mailto:test@example.com">Email</a>
				<a href="tel:+123456789">Phone</a>
				<a href="javascript:void(0)">JS</a>
			</body>
		</html>
	`

	requestsCount := 0

	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		requestsCount++

		if req.URL.String() != "http://simple.test" {
			return nil, errors.New("unexpected request: " + req.URL.String())
		}

		return newResponse(http.StatusOK, html), nil
	})

	report, err := crawler.Analyze("http://simple.test", crawler.Options{
		Client: client,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	page := report.Pages[0]

	if len(page.BrokenLinks) != 0 {
		t.Fatalf("expected 0 broken links, got %d", len(page.BrokenLinks))
	}

	if requestsCount != 1 {
		t.Fatalf("expected only 1 request to root page, got %d", requestsCount)
	}
}
