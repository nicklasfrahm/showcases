package gateway

const (
	DefaultGatewayPort = "8080"
	DefaultPrefork     = false
)

type Options struct {
	Port    string
	Prefork bool
}

// GetDefaultOptions returns default configuration options for the gateway.
func GetDefaultOptions() Options {
	return Options{
		Port: DefaultGatewayPort,
	}
}

// Option is a function on the options for a gateway.
type Option func(*Options) error

// Port is an Option to set the gateway port.
func Port(port string) Option {
	return func(o *Options) error {
		o.Port = port
		return nil
	}
}

// Prefork is an Option to configure preforking. Enabling it
// will create multiple processes listening on the same port.
// This is not recommended when running inside containers.
func Prefork(prefork bool) Option {
	return func(o *Options) error {
		o.Prefork = prefork
		return nil
	}
}
