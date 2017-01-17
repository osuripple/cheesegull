package redis

import (
	"github.com/osuripple/cheesegull"
	"gopkg.in/redis.v5"
)

// Provided is a struct containing the services implemented by this package.
type Provided interface {
	cheesegull.CommunicationService
	cheesegull.SystemService
	cheesegull.Logging
}

type impl struct {
	*redis.Client
}

// Options are the settings to start a redis connection.
type Options struct {
	// The network type, either tcp or unix.
	// Default is tcp.
	Network string
	// host:port address.
	Addr string
	// Optional password. Must match the password specified in the
	// requirepass server configuration option.
	Password string
	// Database to be selected after connecting to the server.
	DB int
}

func (o Options) toRedisOptions() *redis.Options {
	if o.Addr == "" {
		o.Addr = "localhost:6379"
	}
	return &redis.Options{
		Network:  o.Network,
		Addr:     o.Addr,
		Password: o.Password,
		DB:       o.DB,
	}
}

// New creates a new service which provides the services mentioned in the
// Provided interface.
func New(o Options) (Provided, error) {
	i := &impl{
		redis.NewClient(o.toRedisOptions()),
	}
	return i, i.Ping().Err()
}
