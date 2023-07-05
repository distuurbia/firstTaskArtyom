// Package config represents struct Config.
package config

// Config is a structure of environment variables.
type Config struct {
	PostgresPath          string `env:"POSTGRES_PATH"`
	MongoPath             string `env:"MONGO_PATH"`
	AccessTokenSignature  string `env:"ACCESS_TOKEN_SIGNATURE"`
	RefreshTokenSignature string `env:"REFRESH_TOKEN_SIGNATURE"`
	Port                  int    `env:"PORT" envDefault:"5433"`
	RedisAddress          string `env:"REDIS_ADDRESS"`
	RedisPassword         string `env:"REDIS_PASSWORD"`
}
