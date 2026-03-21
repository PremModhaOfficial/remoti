package client

// Options configures client behavior.
type Options struct {
	Network string // "tcp" or "unix". Default: "tcp"
}

// Option is a functional option for configuring the client.
type Option func(*Options)

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Network: "tcp",
	}
}

func WithNetwork(network string) Option {
	return func(o *Options) { o.Network = network }
}
