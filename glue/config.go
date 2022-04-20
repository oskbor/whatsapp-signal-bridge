package glue

import (
	"github.com/rs/zerolog"
)

type config struct {
	SignalRecipient string
	Logger          *zerolog.Logger
}

type Option func(*config)

func SignalRecipient(number string) Option {
	return func(cfg *config) {
		cfg.SignalRecipient = number
	}
}

func Logger(logger *zerolog.Logger) Option {
	return func(cfg *config) {
		cfg.Logger = logger
	}
}
