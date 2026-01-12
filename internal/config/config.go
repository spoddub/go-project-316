package config

type Config struct {
	Depth     int
	Retries   int
	Delay     string
	Timeout   string
	RPS       int
	UserAgent string
	Workers   int
}
