package configs

import (
	"github.com/jmoiron/sqlx"
)

var Cfg Config

type Config struct {
	Db        *sqlx.DB
	SecretKey string
}
