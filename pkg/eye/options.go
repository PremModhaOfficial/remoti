package eye

import "time"

// Options configures the Eye client.
type Options struct {
	// ServerAddress is the remoti server address. Default: "127.0.0.1:8080"
	ServerAddress string
	// Network is "tcp" or "unix". Default: "tcp"
	Network string
	// FindTimeout is the default timeout for Find operations. Default: 5s
	FindTimeout time.Duration
}

// Option is a functional option for configuring the Eye client.
type Option func(*Options)

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		ServerAddress: "127.0.0.1:8080",
		Network:       "tcp",
		FindTimeout:   5 * time.Second,
	}
}

func WithServerAddress(addr string) Option {
	return func(o *Options) { o.ServerAddress = addr }
}

func WithNetwork(network string) Option {
	return func(o *Options) { o.Network = network }
}

func WithFindTimeout(d time.Duration) Option {
	return func(o *Options) { o.FindTimeout = d }
}
