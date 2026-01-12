package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"code/crawler"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "hexlet-go-crawler",
		Usage: "analyze a website structure",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "depth", Value: 10, Usage: "crawl depth"},
			&cli.IntFlag{Name: "retries", Value: 1, Usage: "number of retries for failed requests"},
			&cli.DurationFlag{Name: "delay", Value: 0, Usage: "delay between requests (example: 200ms, 1s)"},
			&cli.DurationFlag{Name: "timeout", Value: 15 * time.Second, Usage: "per-request timeout"},
			&cli.IntFlag{Name: "rps", Value: 0, Usage: "limit requests per second (overrides delay)"},
			&cli.StringFlag{Name: "user-agent", Value: "", Usage: "custom user agent"},
			&cli.IntFlag{Name: "workers", Value: 4, Usage: "number of concurrent workers"},
		},
		Action: func(c *cli.Context) error {
			url := c.Args().First()
			if url == "" {
				_ = cli.ShowAppHelp(c)
				return nil
			}

			opts := crawler.Options{
				URL:         url,
				Depth:       c.Int("depth"),
				Retries:     c.Int("retries"),
				Delay:       c.Duration("delay"),
				Timeout:     c.Duration("timeout"),
				RPS:         c.Int("rps"),
				UserAgent:   c.String("user-agent"),
				Concurrency: c.Int("workers"),
				IndentJSON:  true,
				HTTPClient:  &http.Client{},
			}

			b, err := crawler.Analyze(context.Background(), opts)
			if len(b) > 0 {
				fmt.Fprintln(os.Stdout, string(b))
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
