package glue

type config struct {
	SignalRecipient string
}

type Option func(*config)

func SignalRecipient(number string) Option {
	return func(cfg *config) {
		cfg.SignalRecipient = number
	}
}
