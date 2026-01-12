package main

import (
	"fmt"
	"github.com/spoddub/web-crawler/internal/app/webcrawler"
	"os"
)

func main() {
	if err := webcrawler.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
