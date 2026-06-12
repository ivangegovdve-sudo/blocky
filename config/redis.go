package config

import (
	"fmt"
	"log/slog"
)

// Redis configuration for the redis connection
type Redis struct {
	// Server address and port (e.g. `localhost:6379`), a unix socket path (e.g. `/var/run/redis/redis.sock`),
	// or the sentinel master name when sentinel is used. An address starting with `/` is treated as a unix socket.
	Address string `yaml:"address"`
	// Redis username (if authentication is required).
	Username string `default:"" yaml:"username"`
	// Redis password (if authentication is required).
	Password Secret `default:"" yaml:"password"`
	// Redis database index to use.
	Database int `default:"0" yaml:"database"`
	// If true, blocky will not start when the Redis connection cannot be established.
	Required bool `default:"false" yaml:"required"`
	// Maximum number of connection attempts before giving up.
	ConnectionAttempts int `default:"3" yaml:"connectionAttempts"`
	// Delay between consecutive connection attempts.
	ConnectionCooldown Duration `default:"1s" yaml:"connectionCooldown"`
	// Sentinel username (if sentinel authentication is required).
	SentinelUsername string `default:"" yaml:"sentinelUsername"`
	// Sentinel password (if sentinel authentication is required).
	SentinelPassword Secret `default:"" yaml:"sentinelPassword"`
	// List of Redis Sentinel host:port addresses; enables sentinel mode when non-empty.
	SentinelAddresses []string `yaml:"sentinelAddresses"`
}

// IsEnabled implements `config.Configurable`
func (c *Redis) IsEnabled() bool {
	return c.Address != ""
}

// LogConfig implements `config.Configurable`
func (c *Redis) LogConfig(logger *slog.Logger) {
	if len(c.SentinelAddresses) == 0 {
		logger.Info(fmt.Sprintf("address: %s", c.Address))
	}

	logger.Info(fmt.Sprintf("username: %s", c.Username))
	logger.Info(fmt.Sprintf("password: %s", secretObfuscator))
	logger.Info(fmt.Sprintf("database: %d", c.Database))
	logger.Info(fmt.Sprintf("required: %t", c.Required))
	logger.Info(fmt.Sprintf("connectionAttempts: %d", c.ConnectionAttempts))
	logger.Info(fmt.Sprintf("connectionCooldown: %s", c.ConnectionCooldown))

	if len(c.SentinelAddresses) > 0 {
		logger.Info("sentinel:")
		logger.Info(fmt.Sprintf("  master: %s", c.Address))
		logger.Info(fmt.Sprintf("  username: %s", c.SentinelUsername))
		logger.Info(fmt.Sprintf("  password: %s", secretObfuscator))
		logger.Info("  addresses:")

		for _, addr := range c.SentinelAddresses {
			logger.Info(fmt.Sprintf("    - %s", addr))
		}
	}
}
