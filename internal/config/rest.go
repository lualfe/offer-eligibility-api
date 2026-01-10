package config

type API struct {
	Host string `env:"SERVER_HOST,default=localhost"`
	Port int    `env:"SERVER_PORT,default=8080"`
}
