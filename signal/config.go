package signal

type config struct {
	Number string
	Host   string
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
