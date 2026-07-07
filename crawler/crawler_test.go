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

func TestAnalyzeSuccess(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://example.com" {
			t.Fatalf("expected URL https://example.com, got %s", req.URL.String())
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("OK")),
			Header:     make(http.Header),
		}, nil
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
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("Not Found")),
			Header:     make(http.Header),
		}, nil
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
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("Internal Server Error")),
			Header:     make(http.Header),
		}, nil
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
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
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
