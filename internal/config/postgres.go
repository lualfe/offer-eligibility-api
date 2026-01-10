package config

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Host string `env:"DATABASE_HOST,default=localhost"`
	Port int    `env:"DATABASE_PORT,default=5432"`
	User string `env:"DATABASE_USER"`
	Pass string `env:"DATABASE_PASSWORD"`
	DB   string `env:"DATABASE_NAME"`
}

func (p Postgres) Conn() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Pass, p.DB)
	return sql.Open("postgres", dsn)
}
