package signal

import "github.com/rs/zerolog"

type config struct {
	Number string
	Host   string
	Logger *zerolog.Logger
}

type Option func(*config)

func Host(host string) Option {
	return func(cfg *config) {
		cfg.Host = host
	}
}

func Number(number string) Option {
	return func(cfg *config) {
		cfg.Number = number
	}
}

func Logger(logger *zerolog.Logger) Option {
	return func(cfg *config) {
		cfg.Logger = logger
	}
}
